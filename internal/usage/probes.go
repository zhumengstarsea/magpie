package usage

import (
	"strings"
	"time"

	"github.com/yetone/magpie/internal/probelog"
	"github.com/yetone/magpie/internal/provider"
)

// SourceProbe is Record.Source for a request magpie sent a provider on its
// own, outside any agent's turn (probelog): the editor's Test, Detect
// APIs, a model list's refresh. Its Kind says which (probelog.KindTest,
// KindDetect, KindList). Such a row is listed on the request log, told
// apart there, and counted nowhere: no total, cost, chart, group, budget,
// window hold or observer reads it. The log's index keeps it apart from
// the calls (logSnapshot.probes) so a reader of those never meets it.
const SourceProbe = "probe"

// IsProbe is a request magpie sent on its own (SourceProbe).
func (r Record) IsProbe() bool { return r.Source == SourceProbe }

// IsProbe is a row of such a request. A packed row keeps the row's Source,
// not its record's, so the request log asks the row.
func (r Row) IsProbe() bool { return r.Source == SourceProbe || r.Record.IsProbe() }

// probeRow is a probe's record and where its line ends in the log, as a
// gateway row's Order is.
type probeRow struct {
	Record
	off int64
}

func init() {
	probelog.Sink = func(p probelog.Row) {
		Append(Record{Time: p.Time, Agent: "magpie", Provider: p.Provider, ProviderAccount: p.Account,
			Host: p.Host, Model: p.Model, Requested: p.Model, Status: p.Status, Millis: p.Millis,
			Error: clipProbeError(p.Error), Endpoint: p.Endpoint, Kind: p.Kind, Source: SourceProbe})
	}
}

// clipProbeError keeps an error's first 500 bytes, as a vendor's answer can
// be a whole page.
func clipProbeError(s string) string {
	const most = 500
	if len(s) <= most {
		return s
	}
	cut := most
	for cut > 0 && s[cut]&0xC0 == 0x80 { // not inside a character
		cut--
	}
	return s[:cut] + "…"
}

// probeChunk is the probes made since since as one chunk of the request
// log's rows, each in the order its line has in the log as a gateway
// row's is; nil for none.
// A provider renamed since is listed by the id it has now (Renames): the
// Requests page's provider list named a renamed one twice (cybercaius).
// One asked while a provider was being added, under its host, is listed
// by the id of the provider on that host (Renames.ProbeOf), and once no
// provider is on it, a deleted one's, by the id the log's calls to that
// host were made under (hosts, logSnapshot.hosts): cybercaius on Discord,
// a deleted provider's tests were listed under its host beside its id.
func probeChunk(probes []probeRow, since time.Time, renames provider.Renames, hosts map[string]string) *rowChunk {
	var c *rowChunk
	for _, p := range probes {
		if !since.IsZero() && p.Time.Before(since) {
			continue
		}
		if c == nil {
			c = &rowChunk{}
		}
		r := p.Record
		r.Provider = probeProvider(r, renames, hosts)
		c.add(Row{Record: r, Source: SourceProbe}, "", p.off, false)
	}
	if c == nil {
		return nil
	}
	c.dict, c.shapeIDs = nil, nil
	return c
}

// probeProvider is the provider a probe row is listed under: its id as it
// is now, and for one recorded under its host before the provider had an
// id, the provider saved on that host, or, with none on it now, the one
// the log's calls to that host went to. A host two providers were on, or
// none, stays the host.
func probeProvider(r Record, renames provider.Renames, hosts map[string]string) string {
	id := renames.ProbeOf(r.Provider, r.Time)
	if id != r.Provider || !strings.EqualFold(r.Provider, r.Host) || renames.OnHost(r.Host) {
		return id
	}
	if to := hosts[strings.ToLower(r.Host)]; to != "" {
		return renames.Of(to, r.Time)
	}
	return strings.ToLower(id) // one host, however it was typed
}
