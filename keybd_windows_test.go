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
	"RuneToVK":           true,
	"RuneToVSC":          true,
	"KeyPressAndRelease": true,
	"TypeStr":            true,
	"TypeStrWithOpts":    true,
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
			Input:   'K',
			Output:  []byte{75, 1},
			Passing: true,
		},
		{
			Input:   '😊',
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
			Input:   'r',
			Output:  []uint16{19, 0},
			Passing: true,
		},
		{
			Input:   '😊',
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

// TestRealKeyPressAndRelease will type the letter 'r' wherever the cursor is
// currently positioned.
func TestRealKeyPressAndRelease(t *testing.T) {
	tName := "KeyPressAndRelease"
	if !enabled[tName] {
		t.Skip(tName + test.TestsDisabled)
	}

	scenes := []test.Scene{
		{
			Input:   'r',
			Output:  nil,
			Passing: true,
		},
		{
			Input:   '😊',
			Output:  test.NewError,
			Passing: false,
		},
	}

	test.Countdown(3)

	for i, s := range scenes {
		t.Run(fmt.Sprintf(tName+" #%d", i), func(t *testing.T) {
			code, _, err := keybd.RuneToVSC(s.Input.(rune), hkl)
			got := keybd.KeyPressAndRelease(code, winapi.KEYEVENTF_SCANCODE)
			want := s.Output

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

// TestRealTypeStr will type a collection of different strings wherever the
// cursor is currently positioned.
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

// TestRealTypeStrWithOpts will type a collection of different strings wherever
// the cursor is currently positioned.
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
	keybd.Global.TabSize = 8

	test.Countdown(3)

	for i, s := range scenes {
		t.Run(fmt.Sprintf(tName+" #%d", i), func(t *testing.T) {
			if err := keybd.TypeStr(s.Input.(string)); err != nil {
				t.Errorf(test.ErrWantFGotF, nil, err)
			}
		})
	}
}
