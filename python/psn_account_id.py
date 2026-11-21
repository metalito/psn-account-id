#!/usr/bin/env python3
"""psn-account-id: get your PSN Account ID for PS5 Remote Play using only Sony's own login.

Talks only to auth.api.sonyentertainmentnetwork.com over HTTPS. Saves nothing.
Standard library only. Python 3.10+.  Usage:  python3 psn_account_id.py [--json] [--no-browser]
                                              python3 psn_account_id.py --convert <id>
"""
from __future__ import annotations

import argparse
import base64
import json
import sys
import urllib.error
import urllib.parse
import urllib.request
import webbrowser

VERSION = "1.0.0"
AUTH_HOST = "https://auth.api.sonyentertainmentnetwork.com"
# Public identifiers of Sony's PS Remote Play app, as used by chiaki / chiaki-ng.
# They identify the *app*, not you, and grant nothing without your own Sony login.
CLIENT_ID = "ba495a24-818c-472b-b12d-ff231c1b5745"
CLIENT_SECRET = "mvaiZkRsAsI1IBkY"
REDIRECT = "https://remoteplay.dl.playstation.net/remoteplay/redirect"
SCOPE = ("psn:clientapp referenceDataService:countryConfig.read "
         "pushNotification:webSocket.desktop.connect "
         "sessionManager:remotePlaySession.system.update")


# (English, Spanish). The language comes from --lang or the system language.
MESSAGES = {
    "step1a": ("Step 1. Sign in on Sony's page. Check the address bar shows a sony.com or",
               "Paso 1. Inicia sesión en la página de Sony. Comprueba que la barra de direcciones muestra"),
    "step1b": ("        sonyentertainmentnetwork.com address before typing your password.",
               "        sony.com o sonyentertainmentnetwork.com antes de escribir tu contraseña."),
    "step2a": ("Step 2. After signing in you land on a page starting with",
               "Paso 2. Después de iniciar sesión llegarás a una página que empieza por"),
    "step2b": ("        (it may look blank or show an error - that's normal).",
               "        (puede verse en blanco o mostrar un error: es normal)."),
    "step2c": ("        Copy the WHOLE address from the address bar and paste it below.",
               "        Copia la dirección COMPLETA de la barra de direcciones y pégala abajo."),
    "paste": ("Paste here: ", "Pega aquí: "),
    "result": ("Your PSN Account ID (paste this into Veyra or chiaki-ng):",
               "Tu Account ID de PSN (pégalo en Veyra o en chiaki-ng):"),
    "number": ("Number: ", "Número: "),
    "account": ("Account:", "Cuenta: "),
    "check_you": ("(check this is you)", "(comprueba que eres tú)"),
    "footer": ("This ID never changes - you only need it once. Nothing was saved.",
               "Este ID nunca cambia: solo lo necesitas una vez. No se ha guardado nada."),
    "error": ("Error:", "Error:"),
    "nothing_pasted": ("Nothing was pasted.", "No se ha pegado nada."),
    "not_redirect": ("That address is not Sony's redirect page. It must start with:\n  %s?code=",
                     "Esa dirección no es la página de redirección de Sony. Debe empezar por:\n  %s?code="),
    "no_code": ("No 'code=' found in that address. Sign in again and copy the whole address.",
                "No hay 'code=' en esa dirección. Vuelve a iniciar sesión y copia la dirección completa."),
    "invalid_code": ("That doesn't look like a valid code.", "Eso no parece un código válido."),
    "too_large": ("Number is too large to be a PSN Account ID.",
                  "El número es demasiado grande para ser un Account ID de PSN."),
    "not_b64": ("Not a number and not valid Base64.", "No es un número ni Base64 válido."),
    "wrong_len": ("A Base64 Account ID must decode to exactly 8 bytes.",
                  "Un Account ID en Base64 debe decodificarse en exactamente 8 bytes."),
    "rejected": ("Sony rejected the code. Codes work once and expire within minutes: "
                 "run the tool again and paste faster.",
                 "Sony ha rechazado el código. Cada código sirve una sola vez y caduca en pocos minutos: "
                 "vuelve a ejecutar la herramienta y pega más rápido."),
    "http_status": ("Sony returned HTTP %d. Try again later.", "Sony ha devuelto HTTP %d. Inténtalo más tarde."),
    "unreachable": ("Could not reach Sony (%s). Check your internet connection.",
                    "No se ha podido conectar con Sony (%s). Comprueba tu conexión a internet."),
    "no_token": ("Sony did not return an access token.", "Sony no ha devuelto un token de acceso."),
    "no_number": ("Sony did not return a numeric account number.", "Sony no ha devuelto un número de cuenta."),
    "description": ("Get your PSN Account ID using Sony's own login.",
                    "Obtén tu Account ID de PSN usando solo el inicio de sesión de Sony."),
    "flag_json": ("print the result as JSON", "muestra el resultado en JSON"),
    "flag_nobrowser": ("don't open the browser, just print the link", "no abre el navegador, solo muestra el enlace"),
    "flag_convert": ("convert an Account ID between number and Base64 (offline)",
                     "convierte un Account ID entre número y Base64 (sin conexión)"),
    "flag_lang": ("language: en or es (default: system language)", "idioma: en o es (por defecto: el del sistema)"),
}
LANG = "en"


def t(key: str, *args) -> str:
    text = MESSAGES[key][1 if LANG == "es" else 0]
    return text % args if args else text


def pick_lang(argv) -> str:
    """--lang wins; otherwise LANGUAGE/LC_*/LANG; otherwise the OS locale (e.g. 'Spanish_Spain' on Windows)."""
    argv = list(argv)
    for i, a in enumerate(argv):
        if a.startswith("--lang="):
            return "es" if a[7:].lower().startswith("es") else "en"
        if a == "--lang" and i + 1 < len(argv):
            return "es" if argv[i + 1].lower().startswith("es") else "en"
    import locale
    import os
    for env in ("LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"):
        value = os.environ.get(env, "")
        if value and value not in ("C", "POSIX"):
            return "es" if value.lower().startswith("es") else "en"
    try:
        loc = (locale.getlocale()[0] or "").lower()
    except ValueError:
        loc = ""
    return "es" if loc.startswith(("es", "spanish")) else "en"


class PsnError(Exception):
    """A user-facing error."""


def login_url() -> str:
    query = urllib.parse.urlencode({
        "service_entity": "urn:service-entity:psn", "response_type": "code",
        "client_id": CLIENT_ID, "redirect_uri": REDIRECT, "scope": SCOPE,
        "request_locale": "es_ES" if LANG == "es" else "en_US", "ui": "pr", "service_logo": "ps",
        "layout_type": "popup", "smcid": "remoteplay", "prompt": "always",
        "PlatformPrivacyWs1": "minimal",
    }, quote_via=urllib.parse.quote)
    return f"{AUTH_HOST}/2.0/oauth/authorize?{query}"


def extract_code(pasted: str) -> str:
    """Accept the full redirect address or just the code itself."""
    text = pasted.strip().strip('"').strip("'")
    if not text:
        raise PsnError(t("nothing_pasted"))
    if text.lower().startswith(("http://", "https://")):
        if not text.startswith(REDIRECT + "?"):
            raise PsnError(t("not_redirect", REDIRECT))
        codes = urllib.parse.parse_qs(urllib.parse.urlsplit(text).query).get("code", [])
        if len(codes) != 1 or not codes[0]:
            raise PsnError(t("no_code"))
        text = codes[0]
    if any(c.isspace() for c in text) or len(text) > 2048:
        raise PsnError(t("invalid_code"))
    return text


def to_base64(user_id: int) -> str:
    return base64.b64encode(user_id.to_bytes(8, "little")).decode()


def convert(value: str) -> dict:
    """Convert an Account ID between its number and Base64 forms (offline)."""
    value = value.strip()
    if value.isdigit():
        number = int(value)
        if number >= 2 ** 64:
            raise PsnError(t("too_large"))
        return {"account_id_base64": to_base64(number), "account_id_number": str(number)}
    try:
        raw = base64.b64decode(value, validate=True)
    except ValueError:
        raise PsnError(t("not_b64")) from None
    if len(raw) != 8:
        raise PsnError(t("wrong_len"))
    return {"account_id_base64": value, "account_id_number": str(int.from_bytes(raw, "little"))}


def sony_request(url: str, form: dict | None = None) -> dict:
    basic = base64.b64encode(f"{CLIENT_ID}:{CLIENT_SECRET}".encode()).decode()
    data = urllib.parse.urlencode(form).encode() if form is not None else None
    req = urllib.request.Request(url, data=data, headers={
        "Authorization": f"Basic {basic}", "User-Agent": f"psn-account-id/{VERSION}"})
    if data is not None:
        req.add_header("Content-Type", "application/x-www-form-urlencoded")
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return json.load(resp)
    except urllib.error.HTTPError as e:
        if e.code in (400, 401):
            raise PsnError(t("rejected")) from None
        raise PsnError(t("http_status", e.code)) from None
    except (urllib.error.URLError, TimeoutError) as e:
        raise PsnError(t("unreachable", e)) from None


def fetch_account(code: str) -> dict:
    token = sony_request(f"{AUTH_HOST}/2.0/oauth/token", {
        "grant_type": "authorization_code", "code": code,
        "scope": SCOPE, "redirect_uri": REDIRECT})
    access = token.get("access_token")
    if not access:
        raise PsnError(t("no_token"))
    info = sony_request(f"{AUTH_HOST}/2.0/oauth/token/{urllib.parse.quote(access, safe='')}")
    user_id = str(info.get("user_id", ""))
    if not user_id.isdigit():
        raise PsnError(t("no_number"))
    result = convert(user_id)
    result["online_id"] = info.get("online_id", "")
    return result


def interactive_login(no_browser: bool, out) -> dict:
    url = login_url()
    say = lambda text="": print(text, file=out)
    say(t("step1a"))
    say(t("step1b") + "\n")
    say("   " + url + "\n")
    if not no_browser:
        webbrowser.open(url)
    say(t("step2a"))
    say(f"        {REDIRECT}?code=...")
    say(t("step2b"))
    say(t("step2c") + "\n")
    print(t("paste"), end="", file=out, flush=True)
    return fetch_account(extract_code(input()))


def main(argv=None) -> int:
    global LANG
    LANG = pick_lang(sys.argv[1:] if argv is None else argv)
    ap = argparse.ArgumentParser(description=t("description"))
    ap.add_argument("--json", action="store_true", help=t("flag_json"))
    ap.add_argument("--no-browser", action="store_true", help=t("flag_nobrowser"))
    ap.add_argument("--convert", metavar="ID", help=t("flag_convert"))
    ap.add_argument("--lang", choices=["en", "es"], help=t("flag_lang"))
    ap.add_argument("--version", action="version", version=VERSION)
    args = ap.parse_args(argv)
    try:
        if args.convert:
            result = convert(args.convert)
        else:
            result = interactive_login(args.no_browser, sys.stderr if args.json else sys.stdout)
    except PsnError as e:
        print(f"\n{t('error')} {e}", file=sys.stderr)
        return 1
    except (KeyboardInterrupt, EOFError):
        return 130
    if args.json:
        print(json.dumps(result, indent=2))
        return 0
    print("\n" + t("result") + "\n")
    print(f"   Base64:  {result['account_id_base64']}")
    print(f"   {t('number')} {result['account_id_number']}")
    if result.get("online_id"):
        print(f"   {t('account')} {result['online_id']}  {t('check_you')}")
    print("\n" + t("footer"))
    return 0


if __name__ == "__main__":
    sys.exit(main())
