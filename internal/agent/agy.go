package agent

// Antigravity's CLI, agy, keeps its settings in
// ~/.gemini/antigravity-cli/settings.json. With "modelProvider": "gemini" it
// asks a Gemini API endpoint rather than Google's own backend, and lists the
// models of customModelsConfig.customModels, each {apiProvider, modelName},
// the one picked being "model". magpie's gateway speaks the Gemini API, so
// magpie adds one custom model per catalog model, keyed
// "magpie/<provider>/<model>", and sets modelProvider and model; the user's
// own are stashed and put back when magpie steps out. The endpoint and key
// agy takes only from its environment (GOOGLE_GEMINI_BASE_URL and
// GEMINI_API_KEY; no .env file), so magpie can't set them: the agent's row
// gives the command that starts agy with them (see Agent.Launch). That
// names the model with --model too: agy 1.2 asks for a custom model picked
// in settings.json or its /model picker with no model name ("model is
// empty"); only one given on the command line resolves.
//
// So that typing agy just works (bestony on Discord: the command was
// copied at each start), magpie puts a function named agy in the profile
// of each shell the user has, between "# >>> magpie agy >>>" and
// "# <<< magpie agy <<<" (update.ShellFunction), while agy is on one of
// magpie's models. It runs `magpie agy run`, which starts the real agy
// with the two variables and --model when settings.json's model is still
// magpie's and as it is otherwise (cli/agy_cli.go); PowerShell's asks
// `magpie agy env` for them, since the app it would run is a GUI program
// PowerShell doesn't wait for. A terminal open before takes the command
// from the row.

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/jsonc"

	"github.com/yetone/magpie/internal/edit"
	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/update"
)

const agyModels = "customModelsConfig.customModels"

// AgyLaunch is the command that starts agy on magpie's gateway and one of
// its custom models, in the shell of this system. Its key names agy: its
// requests go out as Google's Gemini SDK's (User-Agent google-genai-sdk/…),
// with nothing of agy's own, so the gateway knows them by the key
// (x-goog-api-key) alone. A command copied before, with the plain token,
// still reaches magpie.
func AgyLaunch(model string) string {
	key := agentKey("agy")
	if runtime.GOOS == "windows" {
		return `$env:GEMINI_API_KEY="` + key + `"; $env:GOOGLE_GEMINI_BASE_URL="` + gateway.URL() + `"; agy --model '` + strings.ReplaceAll(model, "'", "''") + `'`
	}
	return "GEMINI_API_KEY=" + key + " GOOGLE_GEMINI_BASE_URL=" + gateway.URL() + " agy --model '" + strings.ReplaceAll(model, "'", `'\''`) + "'"
}

// AgyStart is what starts agy on magpie, the variables and the model,
// while settings.json's model is one of magpie's; nil and "" otherwise,
// when agy starts as it is.
func AgyStart() (env []string, model string) {
	home, _ := homes()
	m := jsonGet(filepath.Join(home, ".gemini", "antigravity-cli", "settings.json"), "model")()
	if !usesMagpie(m) {
		return nil, ""
	}
	return []string{"GEMINI_API_KEY=" + agentKey("agy"), "GOOGLE_GEMINI_BASE_URL=" + gateway.URL()}, m
}

// agyShell gives the user's shells the function agy, or takes it out of
// them; only for agy on this machine. What fails leaves the setting as it
// is: the row's command still starts it.
func agyShell(home string, on bool) {
	if here, _ := homes(); here != home {
		return
	}
	if on {
		update.ShellFunction("agy", agyFunction)
	} else {
		update.DropShellFunction("agy")
	}
}

// agyTyped is "agy" while a shell's profile has magpie's function for it.
func agyTyped() string {
	if len(update.ShellFunctionIn("agy")) > 0 {
		return "agy"
	}
	return ""
}

// agyFunction is the function agy in a shell of kind, which runs prog
// (this magpie) to start it; the agy on PATH when prog is gone.
func agyFunction(kind, prog string) string {
	switch kind {
	case "fish":
		q := "'" + strings.NewReplacer(`\`, `\\`, "'", `\'`).Replace(prog) + "'"
		return "function agy\n    if test -x " + q + "\n        " + q + " agy run $argv\n    else\n        command agy $argv\n    end\nend"
	case "powershell":
		q := "'" + strings.ReplaceAll(prog, "'", "''") + "'"
		return `function agy {
    $real = Get-Command agy -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
    if (-not $real) { Write-Error 'agy is not installed'; return }
    $start = @{}
    if (Test-Path -LiteralPath ` + q + `) {
        foreach ($l in @(& ` + q + ` agy env)) { $k, $v = "$l" -split '=', 2; if ($v) { $start[$k] = $v } }
    }
    if (-not $start.model) { & $real @args; return }
    $was = $env:GEMINI_API_KEY, $env:GOOGLE_GEMINI_BASE_URL
    $env:GEMINI_API_KEY = $start.GEMINI_API_KEY; $env:GOOGLE_GEMINI_BASE_URL = $start.GOOGLE_GEMINI_BASE_URL
    try {
        if (@($args | Where-Object { "$_" -match '^--?model(=|$)' }).Count) { & $real @args } else { & $real --model $start.model @args }
    } finally { $env:GEMINI_API_KEY = $was[0]; $env:GOOGLE_GEMINI_BASE_URL = $was[1] }
}`
	}
	q := "'" + strings.ReplaceAll(prog, "'", `'\''`) + "'"
	// "function": a name the user aliased isn't expanded where it is defined
	return "function agy {\n  if [ -x " + q + " ]; then " + q + ` agy run "$@"; else command agy "$@"; fi` + "\n}"
}

// agyCustom reads customModels as it is, raw JSON by key.
func agyCustom(path string) map[string]json.RawMessage {
	out := map[string]json.RawMessage{}
	raw, err := edit.Read(path)
	if err != nil || len(raw) == 0 {
		return out
	}
	gjson.GetBytes(jsonc.ToJSONInPlace(raw), agyModels).ForEach(func(k, v gjson.Result) bool {
		out[k.String()] = json.RawMessage(v.Raw)
		return true
	})
	return out
}

func agy(home string) *Agent {
	dir := filepath.Join(home, ".gemini", "antigravity-cli")
	path := filepath.Join(dir, "settings.json")
	get := jsonGet(path, "model")
	provider := jsonGet(path, "modelProvider")
	stashModel, stashProvider := "agy:"+path+":model", "agy:"+path+":modelProvider"
	// setModels writes customModels with the user's own and, when with,
	// magpie's; none left takes the key out
	setModels := func(with bool) error {
		ms := map[string]any{}
		for k, v := range agyCustom(path) {
			if !strings.HasPrefix(k, magpieID+"/") {
				ms[k] = v
			}
		}
		if with {
			for _, m := range magpieModels("agy") {
				ms[magpieID+"/"+m.ID] = map[string]string{"apiProvider": "API_PROVIDER_GOOGLE_GEMINI", "modelName": m.ID}
			}
		}
		if len(ms) > 0 {
			return edit.SetJSON(path, edit.KV{Path: agyModels, Value: ms})
		}
		if err := edit.DelJSON(path, agyModels); err != nil {
			return err
		}
		if v, ok := edit.GetJSON(path, "customModelsConfig"); ok && strings.TrimSpace(v) == "{}" {
			return edit.DelJSON(path, "customModelsConfig")
		}
		return nil
	}
	// put sets key to v, or takes it out for ""
	put := func(key, v string) error {
		if v == "" {
			return edit.DelJSON(path, key)
		}
		return edit.SetJSON(path, edit.KV{Path: key, Value: v})
	}
	return &Agent{
		ID: "agy", Name: "Antigravity CLI", Icon: "antigravity-color", Aliases: []string{"antigravity-cli"}, Spelled: prefixed, NoNames: true,
		Bin: "agy", Dir: dir, Path: path,
		Launch: func() string {
			if v := get(); usesMagpie(v) {
				return AgyLaunch(v)
			}
			return ""
		},
		Typed: func() string {
			if !usesMagpie(get()) {
				return ""
			}
			return agyTyped()
		},
		Sync: func() error {
			on := usesMagpie(get())
			agyShell(home, on)
			if !on {
				return nil
			}
			return setModels(true)
		},
		Notice: func() string {
			v := get()
			if !usesMagpie(v) {
				return ""
			}
			if in := update.ShellFunctionIn("agy"); len(in) > 0 {
				for i := range in {
					in[i] = update.TildeHome(in[i])
				}
				return noticeAgyTyped.say("profiles", strings.Join(in, ", "), "command", AgyLaunch(v))
			}
			return noticeAgy.say("command", AgyLaunch(v))
		},
		Check: func() string {
			v := get()
			if !usesMagpie(v) {
				return ""
			}
			if _, ok := agyCustom(path)[v]; !ok {
				return "Antigravity CLI's custom model " + v + " (settings.json) is gone, so it no longer reaches magpie"
			}
			if p := provider(); p != "gemini" {
				return "Antigravity CLI's modelProvider (settings.json) is " + orDefault(p) + ", not gemini, so it asks Google directly"
			}
			return ""
		},
		Fields: []Field{{
			Key: "model", Label: "model",
			Get: get,
			Set: func(v string) error {
				cur := get()
				if ref, ok := strings.CutPrefix(v, magpieID+"/"); ok && isMagpie(ref) {
					if !usesMagpie(cur) {
						stash(map[string]string{stashModel: cur, stashProvider: provider()})
					}
					if err := setModels(true); err != nil {
						return err
					}
					if err := edit.SetJSON(path, edit.KV{Path: "modelProvider", Value: "gemini"}, edit.KV{Path: "model", Value: v}); err != nil {
						return err
					}
					agyShell(home, true)
					return nil
				}
				if usesMagpie(cur) {
					// back to what the user had
					if v == "" {
						v = unstash(stashModel)
					} else {
						forget(stashModel)
					}
					if err := put("modelProvider", unstash(stashProvider)); err != nil {
						return err
					}
				}
				if err := put("model", v); err != nil {
					return err
				}
				if usesMagpie(cur) {
					agyShell(home, false)
				}
				return setModels(false)
			},
			Options: func(cur map[string]string) []Option {
				custom := agyCustom(path)
				// its own say they go to their provider, not through magpie,
				// as every agent's own do (EZN7L2C3, #834)
				var own []Option
				for k := range custom {
					if !strings.HasPrefix(k, magpieID+"/") {
						to := "Google"
						if p := gjson.GetBytes(custom[k], "apiProvider").String(); strings.Contains(p, "ANTHROPIC") {
							to = "Anthropic"
						} else if strings.Contains(p, "OPENAI") {
							to = "OpenAI"
						}
						own = append(own, Option{Value: k, Icon: modelIcon("", gjson.GetBytes(custom[k], "modelName").String()), Direct: to})
					}
				}
				sort.Slice(own, func(i, j int) bool { return own[i].Value < own[j].Value })
				if c := cur["model"]; c != "" && !usesMagpie(c) && custom[c] == nil {
					own = append([]Option{{Value: c, Icon: modelIcon("", c), Direct: "Google"}}, own...)
				}
				return append(group("Antigravity CLI", own), viaMagpie("agy", magpieID+"/")...)
			},
		}},
	}
}

// what agy says after a change (notice.go)
var (
	noticeAgy      = newNotice("agy takes magpie's gateway and model only from how it is started — start it with {command}")
	noticeAgyTyped = newNotice("Type agy in a new terminal: magpie added it to {profiles}, to start agy on magpie's gateway and model (in a terminal open now: {command}).")
)
