// psn-account-id: get your PSN Account ID for PS5 Remote Play using only Sony's own login.
// Talks only to auth.api.sonyentertainmentnetwork.com over HTTPS. Saves nothing.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

var version = "dev" // set at build time: -ldflags "-X main.version=v1.0.0"

func openBrowser(u string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	case "darwin":
		cmd = exec.Command("open", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	_ = cmd.Start()
}

func interactiveLogin(noBrowser bool, out io.Writer, in *bufio.Reader) (Result, error) {
	u := loginURL()
	fmt.Fprintln(out, t("step1a"))
	fmt.Fprintln(out, t("step1b"))
	fmt.Fprintln(out, "\n   "+u+"\n")
	if !noBrowser {
		openBrowser(u)
	}
	fmt.Fprintln(out, t("step2a"))
	fmt.Fprintln(out, "        "+redirectURI+"?code=...")
	fmt.Fprintln(out, t("step2b"))
	fmt.Fprintln(out, t("step2c"))
	fmt.Fprint(out, "\n"+t("paste"))
	line, err := in.ReadString('\n')
	if err != nil && line == "" {
		return Result{}, fmt.Errorf("%s", t("no_input"))
	}
	code, err := extractCode(line)
	if err != nil {
		return Result{}, err
	}
	return fetchAccount(code)
}

func run(args []string, stdout, stderr io.Writer, in *bufio.Reader) int {
	lang = pickLang(args)
	fs := flag.NewFlagSet("psn-account-id", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprintln(stderr, t("usage")); fs.PrintDefaults() }
	asJSON := fs.Bool("json", false, t("flag_json"))
	noBrowser := fs.Bool("no-browser", false, t("flag_nobrowser"))
	conv := fs.String("convert", "", t("flag_convert"))
	showVersion := fs.Bool("version", false, t("flag_version"))
	_ = fs.String("lang", "", t("flag_lang"))
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *showVersion {
		fmt.Fprintln(stdout, version)
		return 0
	}
	var r Result
	var err error
	if *conv != "" {
		r, err = convert(*conv)
	} else {
		info := stdout
		if *asJSON {
			info = stderr
		}
		r, err = interactiveLogin(*noBrowser, info, in)
	}
	if err != nil {
		fmt.Fprintln(stderr, "\n"+t("error"), err)
		return 1
	}
	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(r)
		return 0
	}
	fmt.Fprintln(stdout, "\n"+t("result"))
	fmt.Fprintln(stdout, "\n   Base64:  "+r.Base64)
	fmt.Fprintln(stdout, "   "+t("number")+" "+r.Number)
	if r.OnlineID != "" {
		fmt.Fprintln(stdout, "   "+t("account")+" "+r.OnlineID+"  "+t("check_you"))
	}
	fmt.Fprintln(stdout, "\n"+t("footer"))
	return 0
}

func main() {
	in := bufio.NewReader(os.Stdin)
	code := run(os.Args[1:], os.Stdout, os.Stderr, in)
	// When double-clicked on Windows the console closes immediately; keep it open.
	if runtime.GOOS == "windows" && len(os.Args) == 1 {
		fmt.Print("\n" + t("press_enter"))
		_, _ = in.ReadString('\n')
	}
	os.Exit(code)
}
