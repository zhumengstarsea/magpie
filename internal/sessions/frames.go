package sessions

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"

	"github.com/klauspost/compress/zstd"
)

// Reading on through a file of zstd frames, as dsh writes its sessions: a
// batch of whole lines a frame, appended. A frame can't be read from its
// middle, but it can be found without decoding it (RFC 8878: its header,
// then blocks that each say how long they are), so a read goes on from the
// frame it left off in, passing over what that frame gave it before. A
// last frame not yet written whole waits for the next read.

var errFrame = errors.New("not a zstd frame")

// readFrame reads the frame at r into buf: a zstd frame whole, nil for a
// skippable one. io.ErrUnexpectedEOF when the file ends inside it, io.EOF
// when it ends before it.
func readFrame(r *bufio.Reader, buf []byte) ([]byte, int64, error) {
	buf = buf[:0]
	var n int64
	take := func(k int) ([]byte, error) {
		start := len(buf)
		buf = append(buf, make([]byte, k)...)
		if _, err := io.ReadFull(r, buf[start:]); err != nil {
			if errors.Is(err, io.EOF) {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}
		n += int64(k)
		return buf[start:], nil
	}
	if _, err := r.Peek(1); err != nil {
		return nil, 0, io.EOF
	}
	magic, err := take(4)
	if err != nil {
		return nil, 0, err
	}
	m := binary.LittleEndian.Uint32(magic)
	if m&0xFFFFFFF0 == 0x184D2A50 {
		size, err := take(4)
		if err != nil {
			return nil, 0, err
		}
		k := int64(binary.LittleEndian.Uint32(size))
		if d, err := r.Discard(int(k)); err != nil || int64(d) < k {
			return nil, 0, io.ErrUnexpectedEOF
		}
		return nil, n + k, nil
	}
	if m != 0xFD2FB528 {
		return nil, 0, errFrame
	}
	fhd, err := take(1)
	if err != nil {
		return nil, 0, err
	}
	d := fhd[0]
	single := d>>5&1 == 1
	extra := 0
	if !single {
		extra++ // window descriptor
	}
	extra += []int{0, 1, 2, 4}[d&3] // dictionary id
	switch d >> 6 {
	case 0:
		if single {
			extra++
		}
	case 1:
		extra += 2
	case 2:
		extra += 4
	case 3:
		extra += 8
	}
	if _, err := take(extra); err != nil {
		return nil, 0, err
	}
	for {
		h, err := take(3)
		if err != nil {
			return nil, 0, err
		}
		v := uint32(h[0]) | uint32(h[1])<<8 | uint32(h[2])<<16
		size := int(v >> 3)
		switch v >> 1 & 3 {
		case 1:
			size = 1 // RLE: one byte, repeated
		case 3:
			return nil, 0, errFrame
		}
		if _, err := take(size); err != nil {
			return nil, 0, err
		}
		if v&1 == 1 {
			break
		}
	}
	if d>>2&1 == 1 {
		if _, err := take(4); err != nil {
			return nil, 0, err
		}
	}
	return buf, n, nil
}

// scanFrames hands fn each whole line of a file of zstd frames, from the
// frame at off, skip bytes into what that frame holds. It returns where to
// go on from (the frame the first line not yet whole starts in, and how far
// into it) and where the last whole frame read ends.
func scanFrames(path string, off, skip int64, fn func([]byte) bool) (next, nextSkip, end int64, err error) {
	if scanHook != nil {
		scanHook(path, off)
	}
	next, nextSkip, end = off, skip, off
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	if _, err = f.Seek(off, io.SeekStart); err != nil {
		return
	}
	dec, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(1), zstd.WithDecoderLowmem(true))
	if err != nil {
		return
	}
	defer dec.Close()
	r := bufio.NewReaderSize(f, scanBuffer(f, off))
	var frame, out, pending []byte
	at := off
	first := true
	for {
		var n int64
		frame, n, err = readFrame(r, frame)
		if err != nil {
			// the end, a frame still being written, or not one at all:
			// what was read stands, the rest waits for the next read
			err = nil
			return
		}
		start := at
		at += n
		end = at
		if frame == nil {
			if len(pending) == 0 {
				next, nextSkip = at, 0
			}
			continue
		}
		if out, err = dec.DecodeAll(frame, out[:0]); err != nil {
			err = nil
			end = start
			return
		}
		from := int64(0)
		if first {
			first = false
			if skip > int64(len(out)) {
				err = errFrame
				return
			}
			from = skip
		}
		data := out[from:]
		for {
			i := bytes.IndexByte(data, '\n')
			if i < 0 {
				break
			}
			line := data[:i]
			if len(pending) > 0 {
				line = append(pending, line...)
				pending = pending[:0]
			}
			if !fn(line) {
				return
			}
			from += int64(i + 1)
			data = data[i+1:]
		}
		if len(data) == 0 {
			if len(pending) == 0 {
				next, nextSkip = at, 0
			}
			continue
		}
		if len(pending) == 0 {
			next, nextSkip = start, from
		}
		pending = append(pending, data...)
	}
}
