package gui

import (
	"net/http"
	"strconv"

	"github.com/yetone/magpie/internal/gwmw"
	"github.com/yetone/magpie/internal/maskrec"
)

// maskedRecord is a record as the page shows it: each value says whether
// the masker still knows what it stood for, so the page offers to show it
// only where it can.
type maskedRecord struct {
	maskrec.Record
	Values []maskedValue `json:"values,omitempty"`
}

type maskedValue struct {
	maskrec.Value
	Known bool `json:"known,omitempty"`
}

// teller is the served masker, where it can say what a placeholder stood
// for.
func teller() (gwmw.Teller, bool) {
	m, ok := gwmw.Served()
	if !ok {
		return nil, false
	}
	t, ok := m.(gwmw.Teller)
	return t, ok
}

func shownRecord(r maskrec.Record, t gwmw.Teller) maskedRecord {
	out := maskedRecord{Record: r}
	for _, v := range r.Values {
		mv := maskedValue{Value: v}
		if t != nil {
			_, mv.Known = t.Value(v.Placeholder)
		}
		out.Values = append(out.Values, mv)
	}
	return out
}

// maskingRoutes are Settings → Privacy's record of what masking swapped:
//   - GET /api/masking/summary?days= is maskrec.Sum;
//   - GET /api/masking/records?before=&limit=&kind=&by=&agent=&provider=
//     a page of records, newest first;
//   - GET /api/masking/record?key= one record;
//   - GET /api/masking/value?p= what a placeholder stood for, while this
//     magpie's masker remembers it (404 after a restart: it is never on
//     disk).
func maskingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/masking/summary", func(rw http.ResponseWriter, r *http.Request) {
		days, _ := strconv.Atoi(r.URL.Query().Get("days"))
		writeJSON(rw, maskrec.Sum(min(max(days, 1), 30)))
	})
	mux.HandleFunc("GET /api/masking/records", func(rw http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		before, _ := strconv.ParseInt(q.Get("before"), 10, 64)
		limit, _ := strconv.Atoi(q.Get("limit"))
		rs, more := maskrec.List(before, min(max(limit, 1), 200), maskrec.Filter{Kind: q.Get("kind"), By: q.Get("by"), Agent: q.Get("agent"), Provider: q.Get("provider")})
		t, _ := teller()
		out := []maskedRecord{}
		for _, rec := range rs {
			out = append(out, shownRecord(rec, t))
		}
		writeJSON(rw, map[string]any{"records": out, "more": more})
	})
	mux.HandleFunc("GET /api/masking/record", func(rw http.ResponseWriter, r *http.Request) {
		key, _ := strconv.ParseInt(r.URL.Query().Get("key"), 10, 64)
		rec, ok := maskrec.Get(key)
		if !ok {
			http.NotFound(rw, r)
			return
		}
		t, _ := teller()
		writeJSON(rw, shownRecord(rec, t))
	})
	mux.HandleFunc("GET /api/masking/value", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Cache-Control", "no-store")
		t, ok := teller()
		if !ok {
			http.NotFound(rw, r)
			return
		}
		v, ok := t.Value(r.URL.Query().Get("p"))
		if !ok {
			http.NotFound(rw, r)
			return
		}
		writeJSON(rw, map[string]string{"value": v})
	})
}
