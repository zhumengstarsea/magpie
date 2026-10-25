package redact

import (
	"strings"
	"testing"
)

// A password written after its name in Chinese, Japanese or Korean is
// masked, a short one and one of digits only too: 密码：786kih went to the
// vendor as written (CyberCaius on Discord). The name, the words around it
// and a sentence's closing stop stay.
func TestCJKPasswordMasked(t *testing.T) {
	cases := []struct {
		in, gone string
		kept     []string
	}{
		{"密码：786kih", "786kih", []string{"密码："}},
		{"服务器密码是 hunter2，记得改", "hunter2", []string{"服务器密码是 ", "，记得改"}},
		{"WiFi 密碼: 12345678。", "12345678", []string{"密碼: ", "。"}},
		{"口令=Abc!2024.", "Abc!2024", []string{"口令=", "."}},
		{"パスワードは pa55word です", "pa55word", []string{"パスワードは ", " です"}},
		{"비밀번호: qwer1234", "qwer1234", []string{"비밀번호: "}},
		{`{"text":"账号 admin 密码 \"s3cr3t\""}`, "s3cr3t", []string{"账号 admin 密码 "}},
	}
	for _, c := range cases {
		out, n := Mask(c.in, Options{Secrets: true})
		if n != 1 || strings.Contains(unmarked(out), c.gone) {
			t.Errorf("%q: %d masked → %q", c.in, n, out)
		}
		for _, k := range c.kept {
			if !strings.Contains(out, k) {
				t.Errorf("%q: %q gone from %q", c.in, k, out)
			}
		}
		if back := Restore(out, false); back != c.in {
			t.Errorf("round trip: %q → %q → %q", c.in, out, back)
		}
	}
	// in a request body, as the gateway masks one
	body := `{"messages":[{"role":"user","content":"密码：786kih"}]}`
	if out, n := MaskJSON([]byte(body), Options{Secrets: true}); n != 1 || strings.Contains(string(out), "786kih") || string(RestoreJSON(out)) != body {
		t.Errorf("body: %d %s", n, out)
	}
	// what isn't a password: no value, a placeholder, words, too short
	for _, in := range []string{"请输入密码：", "密码：${DB_PASS}", "密码：****", "忘记密码了怎么办", "修改密码功能", "密码：abc"} {
		if out, n := Mask(in, Options{Secrets: true}); n != 0 {
			t.Errorf("%q masked: %q", in, out)
		}
	}
}
