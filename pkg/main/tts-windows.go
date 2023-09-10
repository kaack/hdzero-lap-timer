//go:build windows

package main

import (
	"fmt"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"strings"
	"time"
)

func durationAsTTS(duration time.Duration) string {
	seconds := duration.Seconds()
	if seconds < 60 {
		return strings.Replace(fmt.Sprintf("%0.2f", seconds), ".", " ", 1)
	} else {
		minutes := seconds / 60
		return strings.Replace(fmt.Sprintf("%0.2f", minutes), ".", " ", 1)
	}
}

func ttsSay(text string) error {
	var err error

	if err = ole.CoInitialize(0); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	var spVoice *ole.IUnknown
	// Create object
	if spVoice, err = oleutil.CreateObject("SAPI.SpVoice"); err != nil {
		return err
	}

	// Get voice
	var voice *ole.IDispatch
	if voice, err = spVoice.QueryInterface(ole.IID_IDispatch); err != nil {
		return err
	}

	// Speak
	if _, err = oleutil.CallMethod(voice, "Speak", text); err != nil {
		return err
	}

	return err
}
