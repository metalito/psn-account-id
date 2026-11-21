#!/usr/bin/env python3
"""psn-account-id: get your PSN Account ID for PS5 Remote Play using only Sony's own login.

Talks only to auth.api.sonyentertainmentnetwork.com over HTTPS. Saves nothing.
Standard library only. Python 3.9+.  Usage:  python3 psn_account_id.py [--json] [--no-browser]
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


class PsnError(Exception):
    """A user-facing error."""


def login_url() -> str:
    query = urllib.parse.urlencode({
        "service_entity": "urn:service-entity:psn", "response_type": "code",
        "client_id": CLIENT_ID, "redirect_uri": REDIRECT, "scope": SCOPE,
        "request_locale": "en_US", "ui": "pr", "service_logo": "ps",
        "layout_type": "popup", "smcid": "remoteplay", "prompt": "always",
        "PlatformPrivacyWs1": "minimal",
    }, quote_via=urllib.parse.quote)
    return f"{AUTH_HOST}/2.0/oauth/authorize?{query}"


def extract_code(pasted: str) -> str:
    """Accept the full redirect address or just the code itself."""
    text = pasted.strip().strip('"').strip("'")
    if not text:
        raise PsnError("Nothing was pasted.")
    if text.lower().startswith(("http://", "https://")):
        if not text.startswith(REDIRECT + "?"):
            raise PsnError(f"That address is not Sony's redirect page. It must start with:\n  {REDIRECT}?code=")
        codes = urllib.parse.parse_qs(urllib.parse.urlsplit(text).query).get("code", [])
        if len(codes) != 1 or not codes[0]:
            raise PsnError("No 'code=' found in that address. Sign in again and copy the whole address.")
        text = codes[0]
    if any(c.isspace() for c in text) or len(text) > 2048:
        raise PsnError("That doesn't look like a valid code.")
    return text


def to_base64(user_id: int) -> str:
    return base64.b64encode(user_id.to_bytes(8, "little")).decode()


def convert(value: str) -> dict:
    """Convert an Account ID between its number and Base64 forms (offline)."""
    value = value.strip()
    if value.isdigit():
        number = int(value)
        if number >= 2 ** 64:
            raise PsnError("Number is too large to be a PSN Account ID.")
        return {"account_id_base64": to_base64(number), "account_id_number": str(number)}
    try:
        raw = base64.b64decode(value, validate=True)
    except ValueError:
        raise PsnError("Not a number and not valid Base64.") from None
    if len(raw) != 8:
        raise PsnError("A Base64 Account ID must decode to exactly 8 bytes.")
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
            raise PsnError("Sony rejected the code. Codes work once and expire within minutes: "
                           "run the tool again and paste faster.") from None
        raise PsnError(f"Sony returned HTTP {e.code}. Try again later.") from None
    except (urllib.error.URLError, TimeoutError) as e:
        raise PsnError(f"Could not reach Sony ({e}). Check your internet connection.") from None


def fetch_account(code: str) -> dict:
    token = sony_request(f"{AUTH_HOST}/2.0/oauth/token", {
        "grant_type": "authorization_code", "code": code,
        "scope": SCOPE, "redirect_uri": REDIRECT})
    access = token.get("access_token")
    if not access:
        raise PsnError("Sony did not return an access token.")
    info = sony_request(f"{AUTH_HOST}/2.0/oauth/token/{urllib.parse.quote(access, safe='')}")
    user_id = str(info.get("user_id", ""))
    if not user_id.isdigit():
        raise PsnError("Sony did not return a numeric account number.")
    result = convert(user_id)
    result["online_id"] = info.get("online_id", "")
    return result


def interactive_login(no_browser: bool, out) -> dict:
    url = login_url()
    say = lambda text="": print(text, file=out)
    say("Step 1. Sign in on Sony's page. Check the address bar shows a sony.com or")
    say("        sonyentertainmentnetwork.com address before typing your password.\n")
    say("   " + url + "\n")
    if not no_browser:
        webbrowser.open(url)
    say("Step 2. After signing in you land on a page starting with")
    say(f"        {REDIRECT}?code=...")
    say("        (it may look blank or show an error - that's normal).")
    say("        Copy the WHOLE address from the address bar and paste it below.\n")
    print("Paste here: ", end="", file=out, flush=True)
    return fetch_account(extract_code(input()))


def main(argv=None) -> int:
    ap = argparse.ArgumentParser(description="Get your PSN Account ID using Sony's own login.")
    ap.add_argument("--json", action="store_true", help="print the result as JSON")
    ap.add_argument("--no-browser", action="store_true", help="don't open the browser, just print the link")
    ap.add_argument("--convert", metavar="ID", help="convert an Account ID between number and Base64 (offline)")
    ap.add_argument("--version", action="version", version=VERSION)
    args = ap.parse_args(argv)
    try:
        if args.convert:
            result = convert(args.convert)
        else:
            result = interactive_login(args.no_browser, sys.stderr if args.json else sys.stdout)
    except PsnError as e:
        print(f"\nError: {e}", file=sys.stderr)
        return 1
    except (KeyboardInterrupt, EOFError):
        return 130
    if args.json:
        print(json.dumps(result, indent=2))
        return 0
    print("\nYour PSN Account ID (paste this into Veyra or chiaki-ng):\n")
    print(f"   Base64:  {result['account_id_base64']}")
    print(f"   Number:  {result['account_id_number']}")
    if result.get("online_id"):
        print(f"   Account: {result['online_id']}  (check this is you)")
    print("\nThis ID never changes - you only need it once. Nothing was saved.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
