package main

import (
	"fmt"
	"os"
	"strings"
)

// lang is "en" or "es". Set once in run() from --lang or the system language.
var lang = "en"

var messages = map[string][2]string{ // {English, Spanish}
	"step1a":         {"Step 1. Sign in on Sony's page. Check the address bar shows a sony.com or", "Paso 1. Inicia sesión en la página de Sony. Comprueba que la barra de direcciones muestra"},
	"step1b":         {"        sonyentertainmentnetwork.com address before typing your password.", "        sony.com o sonyentertainmentnetwork.com antes de escribir tu contraseña."},
	"step2a":         {"Step 2. After signing in you land on a page starting with", "Paso 2. Después de iniciar sesión llegarás a una página que empieza por"},
	"step2b":         {"        (it may look blank or show an error - that's normal).", "        (puede verse en blanco o mostrar un error: es normal)."},
	"step2c":         {"        Copy the WHOLE address from the address bar and paste it below.", "        Copia la dirección COMPLETA de la barra de direcciones y pégala abajo."},
	"paste":          {"Paste here (right-click pastes in the Windows console): ", "Pega aquí (en la consola de Windows se pega con clic derecho): "},
	"result":         {"Your PSN Account ID (paste this into Veyra or chiaki-ng):", "Tu Account ID de PSN (pégalo en Veyra o en chiaki-ng):"},
	"number":         {"Number: ", "Número: "},
	"account":        {"Account:", "Cuenta: "},
	"check_you":      {"(check this is you)", "(comprueba que eres tú)"},
	"footer":         {"This ID never changes - you only need it once. Nothing was saved.", "Este ID nunca cambia: solo lo necesitas una vez. No se ha guardado nada."},
	"press_enter":    {"Press Enter to close...", "Pulsa Intro para cerrar..."},
	"error":          {"Error:", "Error:"},
	"no_input":       {"no input", "no se ha recibido nada"},
	"nothing_pasted": {"nothing was pasted", "no se ha pegado nada"},
	"not_redirect":   {"that address is not Sony's redirect page; it must start with\n  %s?code=", "esa dirección no es la página de redirección de Sony; debe empezar por\n  %s?code="},
	"bad_address":    {"could not read that address", "no se ha podido leer esa dirección"},
	"no_code":        {"no 'code=' found in that address; sign in again and copy the whole address", "no hay 'code=' en esa dirección; vuelve a iniciar sesión y copia la dirección completa"},
	"invalid_code":   {"that doesn't look like a valid code", "eso no parece un código válido"},
	"too_large":      {"number is too large to be a PSN Account ID", "el número es demasiado grande para ser un Account ID de PSN"},
	"not_b64":        {"not a number and not valid Base64", "no es un número ni Base64 válido"},
	"wrong_len":      {"a Base64 Account ID must decode to exactly 8 bytes", "un Account ID en Base64 debe decodificarse en exactamente 8 bytes"},
	"unreachable":    {"could not reach Sony (%v); check your internet connection", "no se ha podido conectar con Sony (%v); comprueba tu conexión a internet"},
	"rejected":       {"Sony rejected the code; codes work once and expire within minutes: run the tool again and paste faster", "Sony ha rechazado el código; cada código sirve una sola vez y caduca en pocos minutos: vuelve a ejecutar la herramienta y pega más rápido"},
	"http_status":    {"Sony returned HTTP %d; try again later", "Sony ha devuelto HTTP %d; inténtalo más tarde"},
	"no_token":       {"Sony did not return an access token", "Sony no ha devuelto un token de acceso"},
	"no_number":      {"Sony did not return a numeric account number", "Sony no ha devuelto un número de cuenta"},
	"flag_json":      {"print the result as JSON", "muestra el resultado en JSON"},
	"flag_nobrowser": {"don't open the browser, just print the link", "no abre el navegador, solo muestra el enlace"},
	"flag_convert":   {"convert an Account ID between number and Base64 (offline)", "convierte un Account ID entre número y Base64 (sin conexión)"},
	"flag_lang":      {"language: en or es (default: system language)", "idioma: en o es (por defecto: el del sistema)"},
	"flag_version":   {"print the version", "muestra la versión"},
	"usage":          {"Usage of psn-account-id:", "Uso de psn-account-id:"},
}

// t returns the message for key in the current language, formatted with args.
func t(key string, args ...any) string {
	m, ok := messages[key]
	if !ok {
		return key
	}
	s := m[0]
	if lang == "es" {
		s = m[1]
	}
	if len(args) > 0 {
		return fmt.Sprintf(s, args...)
	}
	return s
}

// sonyLocale makes Sony's sign-in page match the tool's language.
func sonyLocale() string {
	if lang == "es" {
		return "es_ES"
	}
	return "en_US"
}

// pickLang resolves --lang (if any) or falls back to the system language.
func pickLang(args []string) string {
	for i, a := range args {
		v := ""
		switch {
		case strings.HasPrefix(a, "-lang=") || strings.HasPrefix(a, "--lang="):
			v = a[strings.Index(a, "=")+1:]
		case (a == "-lang" || a == "--lang") && i+1 < len(args):
			v = args[i+1]
		}
		if v != "" {
			return normalizeLang(v)
		}
	}
	for _, env := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(env); v != "" && v != "C" && v != "POSIX" {
			return normalizeLang(v)
		}
	}
	return normalizeLang(systemLang())
}

func normalizeLang(v string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(v)), "es") {
		return "es"
	}
	return "en"
}
