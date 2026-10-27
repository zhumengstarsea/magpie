package masking

import (
	"context"
	"slices"
	"testing"

	"github.com/yetone/magpie/internal/pii"
	"github.com/yetone/magpie/internal/redact"
	"github.com/yetone/magpie/internal/settings"
)

// Masking asks the model the user picked; with none picked, AstrLink
// Guard (Privacy Filter isn't on disk in a test).
func TestMaskingUsesThePickedModel(t *testing.T) {
	for name, want := range map[string]string{"": pii.AstrLinkGuard, pii.PrivacyFilter: pii.PrivacyFilter, pii.AstrLinkGuard: pii.AstrLinkGuard} {
		if got := picked(settings.Settings{RedactModel: true, RedactModelName: name}).ID(); got != want {
			t.Errorf("picked %q: %s, want %s", name, got, want)
		}
	}
}

// scored is a model that finds three names, as sure as their scores.
type scored struct{ pii.Model }

func (scored) ID() string        { return pii.AstrLinkGuard }
func (scored) MinScore() float64 { return 0.8 }
func (scored) Find(context.Context, []string) ([][]redact.Found, error) {
	return [][]redact.Found{{{Start: 0, End: 1, Label: "private_person", Score: 0.6},
		{Start: 2, End: 3, Label: "private_person", Score: 0.85}, {Start: 4, End: 5, Label: "private_person", Score: 0.95}}}, nil
}

// What the model is less sure of than the threshold the user set for it
// goes as written; with none set, or one out of range, its own default
// (@Yayoi_no_yume on X).
func TestModelThreshold(t *testing.T) {
	for _, c := range []struct {
		set  map[string]float64
		want []int
	}{
		{nil, []int{2, 4}},
		{map[string]float64{pii.AstrLinkGuard: 0.9}, []int{4}},
		{map[string]float64{pii.AstrLinkGuard: 0.5}, []int{0, 2, 4}},
		{map[string]float64{pii.PrivacyFilter: 0.95}, []int{2, 4}},
		{map[string]float64{pii.AstrLinkGuard: 0.3}, []int{2, 4}},
		{map[string]float64{pii.AstrLinkGuard: 1}, []int{2, 4}},
	} {
		var m pii.Model = scored{}
		var got []int
		for _, f := range model(context.Background(), m, pii.Threshold(c.set, m))([]string{"a b c"})[0] {
			got = append(got, f.Start)
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("threshold set %v: kept %v, want %v", c.set, got, c.want)
		}
	}
}
