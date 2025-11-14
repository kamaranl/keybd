//go:build windows

package keybd_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kamaranl/go_private/tools/test"
	"github.com/kamaranl/keybd"
	"github.com/kamaranl/winapi"
	"golang.org/x/sys/windows"
)

var enabled = map[string]bool{
	"RuneToVK":            true,
	"RuneToVSC":           true,
	"KeyIsDown":           true,
	"KeyPress|KeyRelease": true,
	"KeyTap":              true,
	"TypeStr":             true,
	"TypeStrWithOpts":     true,
}

var (
	tid = windows.GetCurrentThreadId()
	hkl = windows.GetKeyboardLayout(tid)
)

func TestRuneToVK(t *testing.T) {
	tName := "RuneToVK"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	scenes := []test.Scene{
		{
			Input:   testRunes["upper"],
			Output:  []byte{75, 1},
			Passing: true,
		},
		{
			Input:   testRunes["emoji"],
			Output:  []byte{},
			Passing: false,
		},
	}

	for i, s := range scenes {
		t.Run(fmt.Sprintf(tName+" #%d", i), func(t *testing.T) {
			code, shift, err := keybd.RuneToVK(s.Input.(rune), hkl)

			got := []byte{code, shift}
			want := s.Output.([]byte)

			if s.Passing {
				if err != nil {
					t.Fatalf(test.ErrUnexpectedF, err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf(test.ErrWantFGotF, want, got)
				}
			} else {
				if err == nil {
					t.Errorf(test.ErrWantFGotF, "error", "none")
				}
			}
		})
	}
}

func TestRuneToVSC(t *testing.T) {
	tName := "RuneToVSC"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	scenes := []test.Scene{
		{
			Input:   testRunes["lower"],
			Output:  []uint16{37, 0},
			Passing: true,
		},
		{
			Input:   testRunes["emoji"],
			Output:  []uint16{},
			Passing: false,
		},
	}

	for i, s := range scenes {
		t.Run(fmt.Sprintf(tName+" #%d", i), func(t *testing.T) {
			code, shift, err := keybd.RuneToVSC(s.Input.(rune), hkl)

			got := []uint16{code, uint16(shift)}
			want := s.Output.([]uint16)

			if s.Passing {
				if err != nil {
					t.Fatalf(test.ErrUnexpectedF, err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf(test.ErrWantFGotF, want, got)
				}
			} else {
				if err == nil {
					t.Errorf(test.ErrWantFGotF, "error", "none")
				}
			}
		})
	}
}

func TestRealKeyPress_KeyRelease(t *testing.T) {
	tName := "KeyPress|KeyRelease"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	code, _, _ := keybd.RuneToVSC(testRunes["lower"], hkl)
	flags := winapi.KEYEVENTF_SCANCODE

	test.Countdown(3)

	if err := keybd.KeyPress(code, flags); err != nil {
		t.Errorf(test.ErrUnexpectedF, err)
	}
	if err := keybd.KeyRelease(code, flags); err != nil {
		t.Errorf(test.ErrUnexpectedF, err)
	}
}

func TestRealKeyTap(t *testing.T) {
	tName := "KeyTap"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	code, _, _ := keybd.RuneToVSC(testRunes["lower"], hkl)

	test.Countdown(3)

	if err := keybd.KeyTap(code, winapi.KEYEVENTF_SCANCODE); err != nil {
		t.Errorf(test.ErrUnexpectedF, err)
	}
}

func TestRealKeyIsDown(t *testing.T) {
	tName := "KeyIsDown"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	scenes := []test.Scene{
		{
			Input:  testRunes["lower"],
			Output: true,
		},
		{
			Input:  testRunes["lower"],
			Output: false,
		},
	}

	var flags winapi.KiFlags

	test.Countdown(3)

	for i, s := range scenes {
		first := i == 0
		t.Run(fmt.Sprintf(tName+" #%d", i), func(t *testing.T) {
			code, _, _ := keybd.RuneToVK(s.Input.(rune), hkl)

			if first {
				if err := keybd.KeyPress(uint16(code), flags); err != nil {
					t.Fatalf(test.ErrUnexpectedF, err)
				}
			}

			if got, want := keybd.KeyIsDown(code), s.Output.(bool); got != want {
				t.Errorf(test.ErrWantFGotF, got, want)
			}

			if first {
				if err := keybd.KeyRelease(uint16(code), flags); err != nil {
					t.Fatalf(test.ErrUnexpectedF, err)
				}
			}
		})
	}
}

func TestRealTypeStr(t *testing.T) {
	tName := "TypeStr"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	scenes := []test.Scene{
		{
			Input:  testStrings["shortWord"] + "\r\n",
			Output: nil,
		},
		{
			Input:  testStrings["complexWord"] + "\r\n",
			Output: nil,
		},
		{
			Input:  testStrings["shortSentence"] + "\r\n",
			Output: nil,
		},
		{
			Input:  testStrings["multiLineStringWithTabs"] + "\r\n",
			Output: nil,
		},
		{
			Input:  testStrings["multiLineStringWithSpaces"] + "\r\n",
			Output: nil,
		},
	}

	test.Countdown(3)

	for i, s := range scenes {
		t.Run(fmt.Sprintf(tName+" #%d", i), func(t *testing.T) {
			if err := keybd.TypeStr(s.Input.(string)); err != nil {
				t.Errorf(test.ErrWantFGotF, nil, err)
			}
		})
	}
}

func TestTypeStrWithOpts(t *testing.T) {
	tName := "TypeStrWithOpts"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	scenes := []test.Scene{
		{
			Input:  testStrings["shortWord"] + "\r\n",
			Output: nil,
		},
		{
			Input:  testStrings["complexWord"] + "\r\n",
			Output: nil,
		},
		{
			Input:  testStrings["shortSentence"] + "\r\n",
			Output: nil,
		},
		{
			Input:  testStrings["multiLineStringWithTabs"] + "\r\n",
			Output: nil,
		},
		{
			Input:  testStrings["multiLineStringWithSpaces"] + "\r\n",
			Output: nil,
		},
	}

	keybd.Global.TabsToSpaces = true
	keybd.Global.TabSize = 11

	test.Countdown(3)

	for i, s := range scenes {
		t.Run(fmt.Sprintf(tName+" #%d", i), func(t *testing.T) {
			if err := keybd.TypeStr(s.Input.(string)); err != nil {
				t.Errorf(test.ErrWantFGotF, nil, err)
			}
		})
	}
}
