# psn-account-id

**Get your PlayStation Network Account ID safely, using only Sony's own login page, so you can stream your PS5 to your PC.**

No third-party website, no account to create, nothing saved, and the code is open for anyone to read.

---

## In plain words

### What is this?
A small free tool. You sign in on **Sony's own website**, and the tool shows you your **PSN Account ID**, a number that identifies your PlayStation account.

### Why would I need my Account ID?
PC apps that stream your PS5 to your computer (PS5 "Remote Play"), such as **[Veyra](https://github.com/Likely7/Veyra-NRVideo)** (used to play GTA 6 with NVIDIA AI upscaling) and **[chiaki-ng](https://github.com/chiaki-ng/chiaki-ng)**, ask for it once when you pair the app with your console. The PS5 only accepts a pairing when the Account ID matches a user on that console **and** you type the 8-digit code the PS5 shows on screen.

The Account ID **is not a password**. On its own it gives nobody access to your account or your console. You only need it **once**: it never changes, even if you change your password or your PSN name.

### Why not just use the website Veyra links to?
Veyra's **"Get Account ID"** button opens **psntools.com**, a third-party website ([source](https://github.com/Likely7/Veyra-NRVideo/blob/fe8df5328a62849eb2d79053564b9574ab68aecd/src/ui/bridge/RemotePlay.cpp#L515-L522)). Veyra's own code sends nothing to that site; it only opens the page. Here is what that site is, checked on 2026-11-21:

- **It asks for your NPSSO login cookie to sign up.** Its own sign-up says *"Unable to register without the NPSSO cookie"* and tells you to *"Paste your NPSSO cookie here"*. The **NPSSO cookie works like your password**: whoever has it can act as you on PlayStation Network. The site's own documentation says it gives access to your *"transaction history, real name, … date of birth"* and more, and its privacy policy says it **stores** an encrypted copy.
- **Nobody knows who runs it.** The domain owner is hidden in the public registry (WHOIS "DATA REDACTED"), and no company is named.
- **It also sells PlayStation accounts.** The site's code contains account-sale listings ("rare" and "fresh" PSN accounts) and links to seller feedback on game-account marketplaces.

We are **not** saying that site steals accounts. We are saying **you should never hand your login cookie to an anonymous website** just to read one number that Sony will show you directly. This tool does exactly that: Sony shows you the number.

> **If you already pasted your NPSSO cookie into any website:** go to your PlayStation account settings, choose **Sign out of all devices**, then change your password. That makes the old cookie useless. Also turn on **2-step verification**.

---

## How to use it (Windows, nothing to install)

1. Go to the **[Releases page](../../releases/latest)** and download **`psn-account-id-windows-amd64.exe`**. Use `windows-arm64` only on ARM laptops such as Snapdragon.
2. Double-click it. Windows may show **"Windows protected your PC"**, because the app is new and not signed with a paid certificate. Click **More info → Run anyway**. (Unsure? Use the PowerShell version below, where you can read every line first, or check the download as described in "Is it safe?".)
3. Your browser opens **Sony's sign-in page**. Check the address bar shows **`my.account.sony.com`** or **`sonyentertainmentnetwork.com`**, then sign in as usual.
4. After signing in you'll land on a page starting with `https://remoteplay.dl.playstation.net/remoteplay/redirect?code=…`. **It may look blank or show an error; that's normal.** Click the address bar, select everything and copy it.
5. Paste it into the tool's window (right-click pastes in the Windows console) and press Enter. **Do it within a minute or two**, because the code expires quickly.
6. The tool shows your Account ID. Copy the **Base64** line, for example `FYHpffQQIhE=`.

### Then, in Veyra
Open the PS5 / Remote Play panel. **Don't press "Get Account ID"**; paste your ID into the Account ID field instead. Then, on the PS5, go to **Settings → System → Remote Play → Link Device** and type the 8-digit code it shows into Veyra. Veyra remembers the console afterwards.

### Or in chiaki-ng
chiaki-ng has its own built-in Sony login and can fetch the ID itself. If you prefer, paste the ID from this tool into its registration screen.

---

## Three versions, same behaviour

| Version | Best for | How to run |
|---|---|---|
| **Windows / macOS / Linux app** (`go/`) | Most people | Download from [Releases](../../releases/latest) and double-click (macOS/Linux: run it in a terminal) |
| **PowerShell script** (`powershell/Get-PsnAccountId.ps1`) | Windows users who want to **read the code before running it**; uses the PowerShell built into Windows 10/11 | See below |
| **Python script** (`python/psn_account_id.py`) | People who already have Python 3.9+ | `python psn_account_id.py` |

### Running the PowerShell version
Windows blocks downloaded scripts by default. Open the file in Notepad and read it first, then open **PowerShell** in the folder where you saved it and run:

```powershell
powershell -ExecutionPolicy Bypass -File .\Get-PsnAccountId.ps1
```

`-ExecutionPolicy Bypass` applies only to this one run and does not change your system settings. **General safety tip:** never paste commands from the internet into PowerShell unless you've read what they do. Scams use exactly that trick.

### Options (all versions)
| Option | What it does |
|---|---|
| `--json` / `-Json` | Print the result as JSON, for scripts |
| `--no-browser` / `-NoBrowser` | Print Sony's link instead of opening the browser |
| `--convert <id>` / `-Convert <id>` | Convert an Account ID between its number and Base64 forms, offline |

---

## Is it safe? (details)

- **It talks only to Sony.** The only server contacted is `auth.api.sonyentertainmentnetwork.com`, over HTTPS. Your password is typed into Sony's page in your browser, never into this tool.
- **It saves nothing.** It writes no files and keeps no token. The short-lived access token is used for two requests, then thrown away and left to expire.
- **It asks for limited permissions:** basic app access, country settings, push notifications and Remote Play sessions. That's the same set Sony's own Remote Play app requests. It has no access to purchases, your wallet or account settings.
- **It uses no third-party libraries.** Go, Python and PowerShell are each standard library only, and each is about 150 lines.
- **Releases are built in public.** Every release file is built by GitHub Actions from the code in this repo ([`release.yml`](.github/workflows/release.yml)), comes with `SHA256SUMS.txt`, and carries a **signed build-provenance attestation**. You can verify a download with:
  ```
  gh attestation verify psn-account-id-windows-amd64.exe --repo <owner>/psn-account-id
  ```

### What are the "client ID" and "client secret" in the code?
They identify the **app** to Sony, not you. They belong to Sony's official **PS Remote Play** app, were published years ago by the open-source [Chiaki](https://github.com/chiaki-ng/chiaki-ng) project, and are used by chiaki-ng and Veyra too. Desktop apps can't keep such values secret, so they open nothing without your own Sony login. Sony could stop accepting them at any time; if that happens, this tool (and chiaki-ng's login) would need an update.

## How it works (technical)
1. Opens Sony's OAuth page: `GET /2.0/oauth/authorize?response_type=code&client_id=…&redirect_uri=https://remoteplay.dl.playstation.net/remoteplay/redirect&scope=…`
2. You sign in. Sony redirects to `…/remoteplay/redirect?code=<one-time code>`, and you paste that address.
3. `POST /2.0/oauth/token` with `grant_type=authorization_code`, authenticated with HTTP Basic using the client ID and secret, returns an access token.
4. `GET /2.0/oauth/token/<access token>` returns `user_id` and `online_id`.
5. Account ID = `user_id` as an unsigned 64-bit little-endian integer, Base64-encoded. That's the same conversion chiaki-ng uses.

## Build and test it yourself
```bash
cd go && go test ./... && go build .               # Go 1.24+
cd python && python -m unittest -v                  # Python 3.9+
pwsh -File powershell/Get-PsnAccountId.ps1 -Convert 1234567890123456789
```
CI ([`ci.yml`](.github/workflows/ci.yml)) runs the Go and Python tests and tests the PowerShell script on real **Windows PowerShell 5.1**. Pushing a tag such as `v1.0.0` builds and publishes a release.

## Disclaimer
Not affiliated with or endorsed by Sony Interactive Entertainment. "PlayStation", "PS5" and "PSN" are Sony trademarks. Like chiaki-ng, this tool signs in as Sony's Remote Play app, which may not be permitted by Sony's terms of service; use it at your own discretion. It was written as a safer alternative to third-party lookup sites. Licensed under the [MIT License](LICENSE).
