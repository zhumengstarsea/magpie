package redact

import (
	"strings"
	"testing"
)

// An image a Mac saved at twice its size is named like an address
// (…@2x.jpg), and was masked as one, so Pi could no longer read the file
// (#3029, the reporter's path). A file's name stays; an address beside it
// is still masked.
func TestFileNameIsNoEmail(t *testing.T) {
	o := Options{Secrets: true, Personal: true}
	for _, s := range []string{
		"/Users/vdxz/Pictures/261020_14-16-11@2x.jpg",
		"read assets/icon@3x.png and logo@2x.PNG",
		"open ~/Desktop/scan@1x.pdf",
	} {
		if out, n := Mask(s, o); n != 0 || out != s {
			t.Errorf("%q masked: %q", s, out)
		}
	}
	s := `{"role":"user","content":"look at /Users/vdxz/Pictures/261020_14-16-11@2x.jpg and mail rina.k@fastmail.com"}`
	out, n := MaskJSON([]byte(s), o)
	if n != 1 || !strings.Contains(string(out), "261020_14-16-11@2x.jpg") || strings.Contains(string(out), "rina.k@fastmail.com") {
		t.Fatalf("%d masked: %s", n, out)
	}

	// nor does the local model's EMAIL take one
	f := &fakeModel{values: map[string]string{"261020_14-16-11@2x.jpg": "private_email"}}
	if out, n := Mask("/Users/vdxz/Pictures/261020_14-16-11@2x.jpg", Options{Personal: true, Model: f.find}); n != 0 {
		t.Fatalf("model's: %d masked: %q", n, out)
	}
}
