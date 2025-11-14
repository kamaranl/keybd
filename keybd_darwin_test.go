//go:build darwin

package keybd_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/kamaranl/go_private/tools/test"
	"github.com/kamaranl/keybd"
)

var enabled = map[string]bool{
	"GetKeyboardLayoutInfo": true,
	"RuneToVK":              true,
	"KeyIsDown":             true,
	"KeyPress|KeyRelease":   true,
	"KeyTap":                true,
	"TypeStr":               true,
	"TypeStrWithOpts":       true,
}

var kli = keybd.GetKeyboardLayoutInfo()

func TestGetKeyboardLayoutInfo(t *testing.T) {
	tName := "GetKeyboardLayoutInfo"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	var zero keybd.KeyboardLayoutInfo

	if r1 := keybd.GetKeyboardLayoutInfo(); r1 == zero {
		t.Fatalf(test.ErrUnexpectedF, "no keyboard layout")
	}
}

func TestRuneToVK(t *testing.T) {
	tName := "RuneToVK"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	scenes := []test.Scene{
		{
			Input:   testRunes["upper"],
			Output:  []uint16{40, 2},
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
			code, shift, err := keybd.RuneToVK(s.Input.(rune), kli)
			got := []uint16{code, shift}
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

	code, _, _ := keybd.RuneToVK(testRunes["lower"], kli)

	test.Countdown(3)

	if err := keybd.KeyPress(code, 0); err != nil {
		t.Errorf(test.ErrUnexpectedF, err)
	}
	if err := keybd.KeyRelease(code, 0); err != nil {
		t.Errorf(test.ErrUnexpectedF, err)
	}
}

func TestRealKeyTap(t *testing.T) {
	tName := "KeyTap"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	code, _, _ := keybd.RuneToVK(testRunes["lower"], kli)

	test.Countdown(3)

	if err := keybd.KeyTap(code, 0); err != nil {
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

	test.Countdown(3)

	for i, s := range scenes {
		first := i == 0
		t.Run(fmt.Sprintf(tName+" #%d", i), func(t *testing.T) {
			code, _, _ := keybd.RuneToVK(s.Input.(rune), kli)

			if first {
				if err := keybd.KeyPress(code, 0); err != nil {
					t.Fatalf(test.ErrUnexpectedF, err)
				}

				time.Sleep(50 * time.Millisecond)
			}

			if got, want := keybd.KeyIsDown(code), s.Output.(bool); got != want {
				t.Errorf(test.ErrWantFGotF, got, want)
			}

			if first {
				if err := keybd.KeyRelease(code, 0); err != nil {
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
			Input: testStrings["shortWord"] + "\r\n",
		},
		{
			Input: testStrings["complexWord"] + "\r\n",
		},
		{
			Input: testStrings["shortSentence"] + "\r\n",
		},
		{
			Input: testStrings["multiLineStringWithTabs"] + "\r\n",
		},
		{
			Input: testStrings["multiLineStringWithSpaces"] + "\r\n",
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

func TestRealTypeStrWithOpts(t *testing.T) {
	tName := "TypeStrWithOpts"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	scenes := []test.Scene{
		{
			Input: testStrings["shortWord"] + "\r\n",
		},
		{
			Input: testStrings["complexWord"] + "\r\n",
		},
		{
			Input: testStrings["shortSentence"] + "\r\n",
		},
		{
			Input: testStrings["multiLineStringWithTabs"] + "\r\n",
		},
		{
			Input: testStrings["multiLineStringWithSpaces"] + "\r\n",
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
