package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const userID = "1234567890123456789"
const b64 = "FYHpffQQIhE="

func TestConvert(t *testing.T) {
	if r, err := convert(userID); err != nil || r.Base64 != b64 {
		t.Fatalf("number->base64: %+v %v", r, err)
	}
	if r, err := convert(b64); err != nil || r.Number != userID {
		t.Fatalf("base64->number: %+v %v", r, err)
	}
	for _, bad := range []string{"AAAA", "99999999999999999999999", "not base64!"} {
		if _, err := convert(bad); err == nil {
			t.Errorf("expected error for %q", bad)
		}
	}
}

func TestExtractCode(t *testing.T) {
	ok := map[string]string{
		redirectURI + "?code=abc123&cid=x": "abc123",
		`  "abc123"  ` + "\r\n":            "abc123",
	}
	for in, want := range ok {
		if got, err := extractCode(in); err != nil || got != want {
			t.Errorf("extractCode(%q) = %q, %v", in, got, err)
		}
	}
	for _, bad := range []string{"", "https://evil.example/remoteplay/redirect?code=x", redirectURI + "?foo=1"} {
		if _, err := extractCode(bad); err == nil {
			t.Errorf("expected error for %q", bad)
		}
	}
}

func TestLoginURL(t *testing.T) {
	u := loginURL()
	if !strings.HasPrefix(u, "https://auth.api.sonyentertainmentnetwork.com/2.0/oauth/authorize?") ||
		!strings.Contains(u, clientID) || strings.Contains(u, "+") {
		t.Fatal(u)
	}
}

func TestFlowAgainstFakeSony(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LANGUAGE", "")
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, secret, ok := r.BasicAuth(); !ok || id != clientID || secret != clientSecret {
			w.WriteHeader(401)
			return
		}
		switch {
		case r.Method == "POST" && r.URL.Path == "/2.0/oauth/token":
			_ = r.ParseForm()
			if r.Form.Get("code") != "abc" {
				w.WriteHeader(400)
				return
			}
			_, _ = w.Write([]byte(`{"access_token":"AT/1"}`))
		case r.Method == "GET" && r.URL.EscapedPath() == "/2.0/oauth/token/AT%2F1":
			_, _ = w.Write([]byte(`{"user_id":"` + userID + `","online_id":"Player1"}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	old := authHost
	authHost = srv.URL
	defer func() { authHost = old }()

	var out, errOut bytes.Buffer
	in := bufio.NewReader(strings.NewReader(redirectURI + "?code=abc\n"))
	if code := run([]string{"-json", "-no-browser"}, &out, &errOut, in); code != 0 {
		t.Fatalf("exit %d: %s", code, errOut.String())
	}
	var r Result
	if err := json.Unmarshal(out.Bytes(), &r); err != nil || r.Base64 != b64 || r.OnlineID != "Player1" {
		t.Fatalf("got %q (%v)", out.String(), err)
	}

	out.Reset()
	in = bufio.NewReader(strings.NewReader("wrong\n"))
	if code := run([]string{"-no-browser"}, &out, &errOut, in); code != 1 {
		t.Fatalf("expected failure for rejected code, got %d", code)
	}
}
