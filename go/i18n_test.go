package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestEveryMessageTranslated(t *testing.T) {
	for k, m := range messages {
		if m[0] == "" || m[1] == "" {
			t.Errorf("%s: missing translation", k)
		}
		if strings.Count(m[0], "%") != strings.Count(m[1], "%") {
			t.Errorf("%s: placeholders differ between English and Spanish", k)
		}
	}
}

func TestPickLang(t *testing.T) {
	t.Setenv("LANGUAGE", "")
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "es_ES.UTF-8")
	if pickLang(nil) != "es" {
		t.Error("LANG=es_ES should pick Spanish")
	}
	if pickLang([]string{"-lang", "en"}) != "en" || pickLang([]string{"--lang=es"}) != "es" {
		t.Error("--lang should override the system language")
	}
	t.Setenv("LANG", "en_GB.UTF-8")
	if pickLang(nil) != "en" {
		t.Error("LANG=en_GB should pick English")
	}
}

func TestSpanishOutput(t *testing.T) {
	defer func() { lang = "en" }()
	var out, errOut bytes.Buffer
	in := bufio.NewReader(strings.NewReader(""))
	if run([]string{"-lang", "es", "-convert", userID}, &out, &errOut, in) != 0 {
		t.Fatal(errOut.String())
	}
	for _, want := range []string{"Tu Account ID de PSN", "Número:", b64, "No se ha guardado nada"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in %q", want, out.String())
		}
	}
	if !strings.Contains(loginURL(), "request_locale=es_ES") {
		t.Error("Sony's page should be requested in Spanish")
	}
	errOut.Reset()
	if run([]string{"-lang", "es", "-convert", "AAAA"}, &out, &errOut, in) != 1 || !strings.Contains(errOut.String(), "exactamente 8 bytes") {
		t.Errorf("Spanish error expected, got %q", errOut.String())
	}
}
