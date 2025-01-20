//go:build windows

package main

import "golang.org/x/sys/windows"

func setConsoleColors() error {
    console := windows.Stdout
    var consoleMode uint32
    windows.GetConsoleMode(console, &consoleMode)
    consoleMode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
    return windows.SetConsoleMode(console, consoleMode)
}