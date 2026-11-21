//go:build windows

package main

import "syscall"

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

// Use UTF-8 in the console so accented Spanish text (Número, aquí) shows correctly
// in the classic Windows console, which otherwise uses a legacy code page.
func init() {
	if p := kernel32.NewProc("SetConsoleOutputCP"); p.Find() == nil {
		_, _, _ = p.Call(65001)
	}
}

// systemLang reads the Windows display language (primary language ID 0x0a = Spanish).
func systemLang() string {
	proc := kernel32.NewProc("GetUserDefaultUILanguage")
	if proc.Find() != nil {
		return "en"
	}
	id, _, _ := proc.Call()
	if id&0x3ff == 0x0a {
		return "es"
	}
	return "en"
}
