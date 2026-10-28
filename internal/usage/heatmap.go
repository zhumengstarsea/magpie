package usage

// The heatmap (#1369): the last 53 weeks a day each, as a calendar of
// contributions lays them out, Monday on top. It is the request page's own
// index read over that span — QueryPage's cached chunks, priced once — and
// not another pass over the logs, so a page asking for it again reads only
// what changed. It counts what the Requests tab's chart counts: the rows its
// filters keep, but for the day picked, and not the calls the gateway turned
// away itself.

import (
	"slices"
	"strings"
	"time"

	"github.com/yetone/magpie/internal/sessions"
)

// HeatmapWeeks is how many weeks the heatmap shows, this one the last.
const HeatmapWeeks = 53

// heatmapPeriod is the span the heatmap is read over: QueryPage's period for
// it, from the Monday HeatmapWeeks-1 weeks before this week's. Not one the
// page's period switch offers.
const heatmapPeriod Period = "heatmap"

// heatmapSince is the Monday the heatmap starts on, as of now.
func heatmapSince(now time.Time) time.Time {
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return day.AddDate(0, 0, -(int(day.Weekday())+6)%7-7*(HeatmapWeeks-1))
}

// Day is one day's calls, by the local calendar day they fell on.
type Day struct {
	Date   string  `json:"date"` // YYYY-MM-DD
	Calls  int     `json:"calls"`
	Tokens int     `json:"tokens"` // in, out and through the cache, as the chart counts them
	Cost   float64 `json:"cost"`   // of the priced calls, USD
	// Unpriced: calls with tokens but no known price, not in Cost
	Unpriced int `json:"unpriced,omitempty"`
}

// Heatmap is the days from From to To (today), both YYYY-MM-DD in the
// reader's time zone (Filter.Zone): only those with calls are listed, oldest first.
type Heatmap struct {
	From string `json:"from"`
	To   string `json:"to"`
	Days []Day  `json:"days"`
}

// HeatmapOf is the last HeatmapWeeks weeks of the calls the filter keeps, a
// day each. The filter's Day is not one of its dimensions here: the heatmap
// is every day's.
func HeatmapOf(f Filter) Heatmap {
	f.Day = ""
	if h := QueryPage(heatmapPeriod, f, 0, 1).Heat; h != nil {
		return *h
	}
	return Heatmap{Days: []Day{}}
}

// heatmapOf sums the rows visit gives it by their day in now's zone (the
// reader's), from since to now's day; a row outside those days is left out.
func heatmapOf(since, now time.Time, visit func(add func(Row))) *Heatmap {
	days := heatDays{}
	visit(func(r Row) { days.add(&r, since, now.Location()) })
	return heatmapFrom(since, now, days.list())
}

// heatDays adds rows up a day each, by the day's name.
type heatDays struct {
	at map[string]*Day
	// a day's name is made once for each day, not for each of its calls
	names map[[2]int]string
}

// add adds r to its day in loc, unless the heatmap doesn't count it.
func (h *heatDays) add(r *Row, since time.Time, loc *time.Location) {
	if r.IsRejected() || r.IsProbe() || r.Time.Before(since) {
		return
	}
	if h.at == nil {
		h.at, h.names = map[string]*Day{}, map[[2]int]string{}
	}
	t := r.Time.In(loc)
	k := [2]int{t.Year(), t.YearDay()}
	d, ok := h.names[k]
	if !ok {
		d = t.Format(time.DateOnly)
		h.names[k] = d
	}
	x := h.at[d]
	if x == nil {
		x = &Day{Date: d}
		h.at[d] = x
	}
	if !r.HandedOn {
		x.Calls++ // a request, not each try at it
	}
	x.Tokens += r.Input + r.Output + r.CacheRead + r.CacheWrite
	switch {
	case r.Priced:
		x.Cost += r.Cost
	case r.Input+r.Output > 0:
		x.Unpriced++
	}
}

// addDay adds a day's sums to its day.
func (h *heatDays) addDay(d Day) {
	if h.at == nil {
		h.at = map[string]*Day{}
	}
	x := h.at[d.Date]
	if x == nil {
		x = &Day{Date: d.Date}
		h.at[d.Date] = x
	}
	x.Calls += d.Calls
	x.Tokens += d.Tokens
	x.Cost += d.Cost
	x.Unpriced += d.Unpriced
}

// list is the days, oldest first.
func (h *heatDays) list() []Day {
	out := make([]Day, 0, len(h.at))
	for _, x := range h.at {
		out = append(out, *x)
	}
	slices.SortFunc(out, func(a, b Day) int { return strings.Compare(a.Date, b.Date) })
	return out
}

// heatmapFrom is days from since to now's day, a day after it left out.
func heatmapFrom(since, now time.Time, days []Day) *Heatmap {
	out := &Heatmap{From: since.Format(time.DateOnly), To: now.Format(time.DateOnly), Days: []Day{}}
	for _, d := range days {
		if d.Date <= out.To {
			out.Days = append(out.Days, d)
		}
	}
	return out
}

// heatOf is heatmapOf of ps's rows f keeps, added up a chunk at a time:
// cache's sums of a chunk paired as before are taken as they are. A call
// through magpie changes the gateway's log, and the Requests tab's refresh
// read every row of the year again for it, and paired them: ~4 GB a
// minute and seconds of CPU a refresh while an agent worked (#2278).
func (ps *pairedRows) heatOf(since, now time.Time, f Filter, cache *partCache, identities map[string]sessions.CodexIdentity) *Heatmap {
	sigs := ps.sigs(identities)
	var rows chunkRows
	defer rows.release()
	var r Row
	all := heatDays{}
	for _, c := range ps.all {
		sig := sigs[c]
		var part *chunkPart
		if cache != nil && c.serial != 0 {
			if old := cache.old[c.serial]; old != nil && old.sig == sig {
				part = old
			}
		}
		if part == nil {
			if heatReading != nil {
				heatReading(c)
			}
			one := heatDays{}
			ps.visitChunk(&rows, c, &r, func(_ timedRef, r *Row) {
				if f.keepsRow(r) {
					one.add(r, since, now.Location())
				}
			})
			part = &chunkPart{sig: sig, days: one.list()}
		}
		if cache != nil && c.serial != 0 {
			cache.next[c.serial] = part
		}
		for _, d := range part.days {
			all.addDay(d)
		}
	}
	return heatmapFrom(since, now, all.list())
}

// heatReading, in tests, is told each chunk heatOf adds up again.
var heatReading func(*rowChunk)
