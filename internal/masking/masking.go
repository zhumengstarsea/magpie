// Package masking is Settings → Privacy's masking as the gateway's masker
// (gwmw.Masker): secrets, personal data and the user's own words go to a
// vendor as placeholders (internal/redact), and come back as they were.
// It reads the settings on each call. plugins/redact serves it.
package masking

import (
	"net/http"

	"github.com/yetone/magpie/internal/pii"
	"github.com/yetone/magpie/internal/redact"
	"github.com/yetone/magpie/internal/settings"
)

// Masker masks as the settings say, read on each call.
type Masker struct{}

// Wanted says settings st ask for anything to be masked.
func Wanted(st settings.Settings) bool {
	return st.Redact || st.RedactPersonal || len(st.RedactWords) > 0
}

func options() redact.Options {
	st := settings.Load()
	o := redact.Options{Secrets: st.Redact, Personal: st.RedactPersonal, Kinds: st.RedactKinds, Words: st.RedactWords, Rules: st.RedactRules}
	if st.RedactModel {
		o.Model = model
	}
	return o
}

// model is the local model (internal/pii) as redact's Model. One that
// can't answer panics, and the request is held back (gwmw.Safely) rather
// than sent with only the rules' masking: the user asked for the model's.
func model(texts []string) [][]redact.Found {
	got, err := pii.Find(texts)
	if err != nil {
		panic(err)
	}
	return got
}

func (Masker) JSON(body []byte) ([]byte, int) {
	o := options()
	if !o.Secrets && !o.Personal && len(o.Words) == 0 {
		return body, 0
	}
	return redact.MaskJSON(body, o)
}

func (Masker) Text(s string) (string, int) { return redact.Mask(s, options()) }

func (Masker) Restore(s string, escaped bool) string { return redact.Restore(s, escaped) }

func (Masker) Writer(w http.ResponseWriter) (http.ResponseWriter, func()) {
	rw := redact.NewWriter(w)
	return rw, rw.Finish
}
