package main

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Public identifiers of Sony's PS Remote Play app, as used by chiaki / chiaki-ng.
// They identify the app, not you, and grant nothing without your own Sony login.
const (
	clientID     = "ba495a24-818c-472b-b12d-ff231c1b5745"
	clientSecret = "mvaiZkRsAsI1IBkY"
	redirectURI  = "https://remoteplay.dl.playstation.net/remoteplay/redirect"
	scope        = "psn:clientapp referenceDataService:countryConfig.read " +
		"pushNotification:webSocket.desktop.connect " +
		"sessionManager:remotePlaySession.system.update"
)

// authHost is a variable only so tests can point it at a local fake server.
var authHost = "https://auth.api.sonyentertainmentnetwork.com"

// Result is what the tool prints.
type Result struct {
	Base64   string `json:"account_id_base64"`
	Number   string `json:"account_id_number"`
	OnlineID string `json:"online_id,omitempty"`
}

func loginURL() string {
	q := url.Values{}
	for k, v := range map[string]string{
		"service_entity": "urn:service-entity:psn", "response_type": "code",
		"client_id": clientID, "redirect_uri": redirectURI, "scope": scope,
		"request_locale": "en_US", "ui": "pr", "service_logo": "ps",
		"layout_type": "popup", "smcid": "remoteplay", "prompt": "always",
		"PlatformPrivacyWs1": "minimal",
	} {
		q.Set(k, v)
	}
	return authHost + "/2.0/oauth/authorize?" + strings.ReplaceAll(q.Encode(), "+", "%20")
}

// extractCode accepts the full redirect address or just the code itself.
func extractCode(pasted string) (string, error) {
	text := strings.Trim(strings.TrimSpace(pasted), `"'`)
	if text == "" {
		return "", errors.New("nothing was pasted")
	}
	lower := strings.ToLower(text)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		if !strings.HasPrefix(text, redirectURI+"?") {
			return "", fmt.Errorf("that address is not Sony's redirect page; it must start with\n  %s?code=", redirectURI)
		}
		u, err := url.Parse(text)
		if err != nil {
			return "", errors.New("could not read that address")
		}
		codes := u.Query()["code"]
		if len(codes) != 1 || codes[0] == "" {
			return "", errors.New("no 'code=' found in that address; sign in again and copy the whole address")
		}
		text = codes[0]
	}
	if strings.ContainsAny(text, " \t\r\n") || len(text) > 2048 {
		return "", errors.New("that doesn't look like a valid code")
	}
	return text, nil
}

func toBase64(id uint64) string {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], id)
	return base64.StdEncoding.EncodeToString(b[:])
}

// convert switches an Account ID between its number and Base64 forms (offline).
func convert(value string) (Result, error) {
	value = strings.TrimSpace(value)
	if n, err := strconv.ParseUint(value, 10, 64); err == nil {
		return Result{Base64: toBase64(n), Number: value}, nil
	} else if value != "" && strings.Trim(value, "0123456789") == "" {
		return Result{}, errors.New("number is too large to be a PSN Account ID")
	}
	raw, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return Result{}, errors.New("not a number and not valid Base64")
	}
	if len(raw) != 8 {
		return Result{}, errors.New("a Base64 Account ID must decode to exactly 8 bytes")
	}
	return Result{Base64: value, Number: strconv.FormatUint(binary.LittleEndian.Uint64(raw), 10)}, nil
}

var client = &http.Client{Timeout: 30 * time.Second}

func sonyRequest(method, endpoint string, form url.Values, out any) error {
	var body *strings.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	} else {
		body = strings.NewReader("")
	}
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("User-Agent", "psn-account-id/"+version)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach Sony (%v); check your internet connection", err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == 400 || resp.StatusCode == 401:
		return errors.New("Sony rejected the code; codes work once and expire within minutes: run the tool again and paste faster")
	case resp.StatusCode != 200:
		return fmt.Errorf("Sony returned HTTP %d; try again later", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// fetchAccount trades the one-time code for a short-lived token, then asks Sony whose it is.
func fetchAccount(code string) (Result, error) {
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "scope": {scope}, "redirect_uri": {redirectURI}}
	if err := sonyRequest("POST", authHost+"/2.0/oauth/token", form, &tok); err != nil {
		return Result{}, err
	}
	if tok.AccessToken == "" {
		return Result{}, errors.New("Sony did not return an access token")
	}
	var info struct {
		UserID   json.Number `json:"user_id"`
		OnlineID string      `json:"online_id"`
	}
	if err := sonyRequest("GET", authHost+"/2.0/oauth/token/"+url.PathEscape(tok.AccessToken), nil, &info); err != nil {
		return Result{}, err
	}
	r, err := convert(info.UserID.String())
	if err != nil || r.Number == "" {
		return Result{}, errors.New("Sony did not return a numeric account number")
	}
	r.OnlineID = info.OnlineID
	return r, nil
}
