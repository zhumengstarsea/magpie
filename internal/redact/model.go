package redact

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// A local model (internal/pii, Settings → Privacy's "Find with a local
// model") finds what no rule can: a name or an address in free text, a
// password with no password= before it. Its spans are masked beside the
// rules', under the same switches: a name, an address, a phone number, an
// email or an account number is personal data, a secret is a secret.

// Found is a span the model found in a text: byte offsets, and its label.
type Found struct {
	Start, End int
	Label      string
}

// Model finds spans in each of texts, in order. It never fails: one that
// can't find (the model didn't start) panics, so that the request is held
// back rather than sent unmasked (gwmw.Safely).
type Model func(texts []string) [][]Found

// modelKinds are the placeholders' kinds for the model's labels. Its
// private_date and private_url are left out: a date in a request is
// nearly always a log's or a version's, and a URL the agent's own.
var modelKinds = map[string]string{
	"private_person":  "PERSON",
	"private_address": "ADDRESS",
	"private_phone":   "PHONE",
	"private_email":   "EMAIL",
	"account_number":  "ACCOUNT",
	"secret":          "SECRET",
}

// modelCat is the category (Options.covers) of each kind the model finds.
func modelCat(kind string) string {
	if kind == "SECRET" {
		return ""
	}
	return catPersonal
}

// modelSpans are o.found's spans for s that o covers, trimmed of the
// spaces the model takes in with a word (and a Chinese name of what it
// ran on into, zhTrim), and with what it gets wrong often left out
// (modelTakes); and the Chinese names it missed (zhNameSpans).
func modelSpans(s string, o Options) []span {
	if o.Model == nil || !o.covers(catPersonal) && len(o.found[s]) == 0 {
		return nil
	}
	var out []span
	for _, f := range o.found[s] {
		kind := modelKinds[f.Label]
		if kind == "" || !o.covers(modelCat(kind)) || f.Start < 0 || f.End > len(s) || f.Start >= f.End {
			continue
		}
		a, b := f.Start, f.End
		for a < b && isSpace(s[a]) {
			a++
		}
		for b > a && isSpace(s[b-1]) {
			b--
		}
		if kind == "PERSON" {
			a, b = zhTrim(s, a, b)
		}
		if a < b && utf8.ValidString(s[a:b]) && modelTakes(kind, s[a:b]) {
			out = append(out, span{a, b, kind, ByModel})
		}
	}
	if o.covers(catPersonal) {
		var person []span
		for _, f := range out {
			if f.kind == "PERSON" {
				person = append(person, f)
			}
		}
		out = append(out, zhNameSpans(s, person)...)
	}
	return out
}

func isSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

// modelTakes says v is masked as the model's kind: a secret as long as the
// rules ask of one, never an id a vendor or a tool made (req_…, toolu_…,
// call_…), which the agent and the vendor match against another; a name
// of two letters at least.
func modelTakes(kind, v string) bool {
	switch kind {
	case "SECRET":
		return len(v) >= 8 && !madeID(v)
	case "PERSON":
		return utf8.RuneCountInString(v) >= 2 && strings.IndexFunc(v, unicode.IsLetter) >= 0
	case "ACCOUNT":
		return !madeID(v)
	}
	return true
}

// madeID says v is an id of the shape vendors and tools give theirs: a
// short lower-case prefix, an underscore, then letters and digits
// (req_011CS…, toolu_01A…, call_x9…, msg_…).
func madeID(v string) bool {
	p, rest, ok := strings.Cut(v, "_")
	if !ok || len(p) < 2 || len(p) > 6 || len(rest) < 8 {
		return false
	}
	for _, c := range p {
		if c < 'a' || c > 'z' {
			return false
		}
	}
	for _, c := range rest {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// withModel is o with what o.Model finds in texts, the texts that may be
// masked: each once, the short ones left out.
func withModel(o Options, texts []string) Options {
	if o.Model == nil || !o.Secrets && !o.Personal {
		return o
	}
	var ask []string
	seen := map[string]bool{}
	for _, t := range texts {
		if len(t) >= 3 && !seen[t] {
			seen[t] = true
			ask = append(ask, t)
		}
	}
	if len(ask) == 0 {
		return o
	}
	got := o.Model(ask)
	o.found = make(map[string][]Found, len(ask))
	for i, t := range ask {
		if i < len(got) && len(got[i]) > 0 {
			o.found[t] = got[i]
		}
	}
	return o
}
