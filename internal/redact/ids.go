package redact

import (
	"slices"
	"strings"
)

// An id a backend made (a UUID, 32 hex digits or more: a row's, a
// request's, a session's, a commit's) goes as written. The vendor's model
// reads it, matches it and writes it back into a tool call, and a backend
// looks it up by every one of its characters: one masked, or a piece of one,
// is a lookup that finds nothing or the wrong row. A piece of one is never
// a value of its own: the phone rule found 13800138000 at the end of
// …-a13800138000, the bank card rule 4111-1111-1111-9007 across a UUID's
// groups, and the id went half a placeholder. A whole one is masked only
// as a credential, in a field named for one (password=, api_key:,
// "refresh_token":), never in one named for an id (secret_id, token_id),
// a cursor (next_page_token), a request (Idempotency-Key, csrf_token), a
// digest (secret_hash) or a bare "key"; never as a serial number (sn:),
// and never as what the model takes it for (a name, an account), but for
// hex digits without hyphens it calls a secret (an API key in prose). What
// the user chose (their words, their rules) stays theirs.
//
// An element token cua-driver made for computer use is one too: s and the
// snapshot's counter in 8 hex digits, a colon and the element's index
// (s0000002a:14). The model clicks by it, and a masked one is a click on
// nothing (#3782). Its field is named for an element (element_token), not
// a credential. Without the index, s and 8 hex digits is one only as such
// a field's value: alone in a text it may be a student's number.

// noIDs is found without what is in an id of s, but for what the user
// chose and the credentials named for one.
func noIDs(s string, found []span) []span {
	return slices.DeleteFunc(found, func(f span) bool {
		if f.by == ByWord || f.by == ByYours {
			return false
		}
		l, r, ok := idAround(s, f.start, f.end)
		if !ok {
			return false
		}
		if f.start > l || f.end < r {
			return true // a piece of an id
		}
		switch f.by {
		case ByKnown:
			// masked before as what it is in full: a credential. Not an
			// element token: one the model labelled in prose a turn before
			// ("click s00000001") would be a click on nothing in every
			// element_token after it
			return elementToken(s, l, r)
		case ByModel:
			// a key the model found in prose: 32 hex digits it calls a
			// secret ("the api key is 9f8e…"); never a UUID or an element
			// token
			return f.kind != "SECRET" || uuidRe.MatchString(s[l:r]) || elementToken(s, l, r)
		case ByRule:
			switch f.kind {
			case "PASSWORD", "API_KEY", "TOKEN", "PRIVATE_KEY":
				return false
			case "SECRET":
				return !credentialField(s, l)
			}
		}
		return true
	})
}

// idAround is the id that s[a:b] is in, or is: the run of hex digits and
// hyphens around it, when that run is a UUID or 32 hex digits or more, or
// s[a:b] when it is an element token. ok is false where s[a:b] reaches
// past such a run.
func idAround(s string, a, b int) (l, r int, ok bool) {
	if elementToken(s, a, b) {
		return a, b, true
	}
	l, r = a, b
	for l > 0 && hexOrHyphen(s[l-1]) {
		l--
	}
	for r < len(s) && hexOrHyphen(s[r]) {
		r++
	}
	for l < r && s[l] == '-' {
		l++
	}
	for r > l && s[r-1] == '-' {
		r--
	}
	if a < l || b > r || strings.IndexFunc(s[l:r], func(c rune) bool { return c > 0x7f || !hexOrHyphen(byte(c)) }) >= 0 ||
		!idShaped(s[l:r]) {
		return 0, 0, false
	}
	return l, r, true
}

// elementToken says s[a:b] is an element token, or its snapshot's part
// (s0000002a, where the secret rule's value stops, at the colon): s and 8
// lowercase hex digits, then a colon and the index, with no letter, digit
// or _ on either side. Without an index, only the value of an element_token
// field is one ("element_token": "s00000001").
func elementToken(s string, a, b int) bool {
	if b-a < 9 || s[a] != 's' || a > 0 && wordByte(s[a-1]) {
		return false
	}
	for _, c := range []byte(s[a+1 : a+9]) {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	end := a + 9
	if end < len(s) && s[end] == ':' {
		i := end + 1
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i == end+1 {
			return false // a colon and no index
		}
		end = i
	}
	if b != a+9 && b != end || end < len(s) && wordByte(s[end]) {
		return false
	}
	return end > a+9 || elementField(fieldWords(s, a))
}

// elementField says a field's name words are cua-driver's element_token
// (elementToken, element-token): only that name, not any other with
// "element" in it (element_api_key, element_password), which is a
// credential's as before.
func elementField(words []string) bool {
	return slices.Equal(words, []string{"element", "token"})
}

func hexOrHyphen(c byte) bool {
	return c == '-' || c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

// idShaped says v, hex digits and hyphens, is a UUID or 32 hex digits or
// more without a hyphen.
func idShaped(v string) bool {
	return uuidRe.MatchString(v) || len(v) >= 32 && !strings.Contains(v, "-")
}

// credentialField says the field whose value starts at s[at] is named for
// a credential: refresh_token, DB_PASSWORD, x-api-key; not one named for
// an id, a cursor, a request or cua-driver's element_token, and not a
// bare "key".
func credentialField(s string, at int) bool {
	words := fieldWords(s, at)
	if len(words) < 2 && (len(words) == 0 || words[0] == "key") || elementField(words) {
		return false
	}
	switch words[len(words)-1] {
	case "id", "ids", "uuid", "guid":
		return false
	}
	for _, w := range words {
		switch w {
		case "page", "next", "cursor", "continuation", "csrf", "xsrf", "trace", "request", "idempotency", "sync", "nonce",
			"hash", "digest", "checksum", "sha", "fingerprint", "etag":
			return false
		}
	}
	return true
}

// fieldWords are the words of the name of the field whose value starts at
// s[at] (nameWords), none where no field's name is before it.
func fieldWords(s string, at int) []string {
	i := at
	for i > 0 && strings.IndexByte(" \t\"'\\", s[i-1]) >= 0 {
		i--
	}
	if i == 0 || s[i-1] != ':' && s[i-1] != '=' {
		return nil
	}
	i--
	for i > 0 && strings.IndexByte(" \t\"'\\", s[i-1]) >= 0 {
		i--
	}
	j := i
	for j > 0 && strings.IndexByte(alnum+"_.-", s[j-1]) >= 0 {
		j--
	}
	return nameWords(s[j:i])
}

// nameWords are a field's name's words in lower case: next_page_token,
// nextPageToken and X-Next-Page-Token are next, page, token.
func nameWords(name string) []string {
	var out []string
	w := []byte{}
	flush := func() {
		if len(w) > 0 {
			out = append(out, string(w))
			w = w[:0]
		}
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c == '_' || c == '-' || c == '.':
			flush()
		case c >= 'A' && c <= 'Z':
			if i > 0 && name[i-1] >= 'a' && name[i-1] <= 'z' {
				flush()
			}
			w = append(w, c+'a'-'A')
		default:
			w = append(w, c)
		}
	}
	flush()
	return out
}
