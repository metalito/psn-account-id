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
	fmt.Fprintln(out, "Step 1. Sign in on Sony's page. Check the address bar shows a sony.com or")
	fmt.Fprintln(out, "        sonyentertainmentnetwork.com address before typing your password.")
	fmt.Fprintln(out, "\n   "+u+"\n")
	if !noBrowser {
		openBrowser(u)
	}
	fmt.Fprintln(out, "Step 2. After signing in you land on a page starting with")
	fmt.Fprintln(out, "        "+redirectURI+"?code=...")
	fmt.Fprintln(out, "        (it may look blank or show an error - that's normal).")
	fmt.Fprintln(out, "        Copy the WHOLE address from the address bar and paste it below.")
	fmt.Fprint(out, "\nPaste here (right-click pastes in the Windows console): ")
	line, err := in.ReadString('\n')
	if err != nil && line == "" {
		return Result{}, fmt.Errorf("no input")
	}
	code, err := extractCode(line)
	if err != nil {
		return Result{}, err
	}
	return fetchAccount(code)
}

func run(args []string, stdout, stderr io.Writer, in *bufio.Reader) int {
	fs := flag.NewFlagSet("psn-account-id", flag.ContinueOnError)
	fs.SetOutput(stderr)
	asJSON := fs.Bool("json", false, "print the result as JSON")
	noBrowser := fs.Bool("no-browser", false, "don't open the browser, just print the link")
	conv := fs.String("convert", "", "convert an Account ID between number and Base64 (offline)")
	showVersion := fs.Bool("version", false, "print the version")
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
		fmt.Fprintln(stderr, "\nError:", err)
		return 1
	}
	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(r)
		return 0
	}
	fmt.Fprintln(stdout, "\nYour PSN Account ID (paste this into Veyra or chiaki-ng):")
	fmt.Fprintln(stdout, "\n   Base64:  "+r.Base64)
	fmt.Fprintln(stdout, "   Number:  "+r.Number)
	if r.OnlineID != "" {
		fmt.Fprintln(stdout, "   Account: "+r.OnlineID+"  (check this is you)")
	}
	fmt.Fprintln(stdout, "\nThis ID never changes - you only need it once. Nothing was saved.")
	return 0
}

func main() {
	in := bufio.NewReader(os.Stdin)
	code := run(os.Args[1:], os.Stdout, os.Stderr, in)
	// When double-clicked on Windows the console closes immediately; keep it open.
	if runtime.GOOS == "windows" && len(os.Args) == 1 {
		fmt.Print("\nPress Enter to close...")
		_, _ = in.ReadString('\n')
	}
	os.Exit(code)
}
