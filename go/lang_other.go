//go:build !windows

package main

// systemLang: on macOS and Linux the LANG / LC_* variables (checked in pickLang) decide.
func systemLang() string { return "en" }
