//go:build windows

package keybd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kamaranl/winapi"
	"golang.org/x/sys/windows"
)

// Contants for virtual scan codes.
const (
	VSC_SPACE      = 0x39
	VSC_RETURN     = 0x1C
	VSC_TAB        = 0x0F
	VSC_LSHIFT     = 0x02A
	VSC_LCTRL      = 0x01D
	VSC_LMENU      = 0x038
	VSC_UNASSIGNED = 0x200
)

// Constants for modifier key masks.
const (
	MOD_LSHIFT = 1 << iota
	MOD_LCTRL
	MOD_LALT
)

// abortFlag is checked in typeStr so that the function can safely abort typing.
var abortFlag bool

// StandardMods is a [Modifier] slice of the standard modifier keys.
var StandardMods = []Modifier{
	{Mask: MOD_LSHIFT, VK: windows.VK_LSHIFT, VSC: VSC_LSHIFT},
	{Mask: MOD_LCTRL, VK: windows.VK_LCONTROL, VSC: VSC_LCTRL},
	{Mask: MOD_LALT, VK: windows.VK_LMENU, VSC: VSC_LMENU},
}

// A Modifier is a struct that contains the mask, the virtual key code, and the
// virtual scan code for a modifier key.
type Modifier struct {
	Mask byte   // high-order bitmask of the modifier key
	VK   byte   // virtual key code of the modifier key
	VSC  uint16 // virtual scan code of the modifier key
}

// RuneToVK translates r to a virtual key code and its shift state. It's
// recommended to provide hkl by using [windows.GetKeyboardLayout], however, a 0
// can be provided for hkl to skip detecting a keyboard layout.
// It returns a pair of 0's with an error if the translation fails, otherwise it
// returns the key code, shift state, and a nil error.
func RuneToVK(r rune, hkl winapi.Handle) (code byte, shift byte, err error) {
	switch r {
	case '\r':
		return winapi.VK_UNASSIGNED, 0, nil
	case '\n':
		return windows.VK_RETURN, 0, nil
	case '\t':
		return windows.VK_TAB, 0, nil
	case ' ':
		return windows.VK_SPACE, 0, nil
	}

	return winapi.VkKeyScanExW(int16(r), hkl)
}

// RuneToVSC translates r to a virtual scan code and its shift state. It's
// recommended to provide hkl by using [windows.GetKeyboardLayout], however, a 0
// can be provided for hkl to skip detecting a keyboard layout.
// It returns a pair of 0's with an error if the translation fails, otherwise it
// returns the scan code, shift state, and a nil error.
func RuneToVSC(r rune, hkl winapi.Handle) (code uint16, shift byte, err error) {
	switch r {
	case '\r':
		return VSC_UNASSIGNED, 0, nil
	case '\n':
		return VSC_RETURN, 0, nil
	case '\t':
		return VSC_TAB, 0, nil
	case ' ':
		return VSC_SPACE, 0, nil
	}

	vk, shift, err := winapi.VkKeyScanExW(int16(r), hkl)
	if err != nil {
		return 0, 0, err
	}

	vsc, err := winapi.MapVirtualKeyExW(uint32(vk), winapi.MAPVK_VK_TO_VSC_EX, hkl)
	if err != nil {
		return 0, 0, err
	}

	return uint16(vsc), shift, nil
}

// KeyIsDown detects the down state of virtKey.
// It returns true if the key is currently depressed and false if it is not.
func KeyIsDown(virtKey byte) bool {
	down, _ := winapi.GetKeyState(virtKey)
	return down
}

// KeyPress sends a key-down event and is intended to be used before a call to
// [KeyRelease].
// It returns an error if the call fails.
func KeyPress(key uint16, flags winapi.KiFlags) error {
	return winapi.SendInput(newKeyEvent(key, flags))
}

// KeyRelease sends a key-up event and is intended to be used after a call to
// [KeyPress].
// It returns an error if the call fails.
func KeyRelease(key uint16, flags winapi.KiFlags) error {
	return winapi.SendInput(newKeyEvent(key, flags|winapi.KEYEVENTF_KEYUP))
}

// KeyTap sends a key-down event and a key-up event with a brief pause in
// between to help simulate an actual keystroke. The duration of the pause is
// defined by [KeyPressDuration].
// It returns an error if the call fails.
func KeyTap(key uint16, flags winapi.KiFlags) error {
	var errs []error

	if err := KeyPress(key, flags); err != nil {
		errs = append(errs, err)
	}

	time.Sleep(KeyPressDuration)

	if err := KeyRelease(key, flags); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// TypeStr types str using [TypeString] options and ensures accuracy by
// attaching the current thread to the thread of the foreground window and
// temporary blocking input while attached. A timeout prevents the function call
// from hanging indefinitely while an abort channel allows aborting the
// operation.
// It returns an error if the call fails.
func TypeStr(str string) (err error) {
	if len(str) == 0 {
		return nil
	} else if len(str) > TypeString.MaxCharacters {
		return fmt.Errorf("%s", ErrMaxCharacter)
	}

	hwnd := windows.GetForegroundWindow()

	var (
		pid         uint32
		tidAttachTo uint32
	)

	if hwnd != 0 {
		tidAttachTo, _ = windows.GetWindowThreadProcessId(hwnd, &pid)
	}

	tidAttach := windows.GetCurrentThreadId()
	attached := false
	if tidAttachTo != 0 && tidAttach != tidAttachTo {
		if err = winapi.AttachThreadInput(tidAttach, tidAttachTo, true); err == nil {
			attached = true
			defer func() { _ = winapi.AttachThreadInput(tidAttach, tidAttachTo, false) }()
		}
	}

	_ = winapi.BringWindowToTop(hwnd)
	_ = winapi.SetForegroundWindow(hwnd)
	_, _ = winapi.SetFocus(hwnd)

	blocked := false
	if err = winapi.BlockInput(true); err == nil {
		blocked = true
		defer func() { _ = winapi.BlockInput(false) }()
	}

	TypeString.mu.Lock()
	TypeString.abort = make(chan struct{})
	abort := TypeString.abort
	TypeString.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), TypeString.Timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		hkl := windows.GetKeyboardLayout(tidAttachTo)
		done <- typeStr(str, hkl)
	}()

	cleanup := func() {
		if blocked {
			_ = winapi.BlockInput(false)
		}
		if attached {
			_ = winapi.AttachThreadInput(tidAttach, tidAttachTo, false)
		}
	}

	select {
	case typeStrErr := <-done:
		if typeStrErr != nil {
			return fmt.Errorf("%s: %v", ErrUncaught, typeStrErr)
		}
		return nil
	case <-ctx.Done():
		abortFlag = true
		cleanup()
		return fmt.Errorf("%s", ErrTimeout)
	case <-abort:
		abortFlag = true
		cleanup()
		return fmt.Errorf("%s", ErrAborted)
	}
}

// newKeyEvent creates an input that can be processed by [winapi.SendInput].
func newKeyEvent(key uint16, flags winapi.KiFlags) []winapi.INPUT_Ki {
	ki := winapi.KEYBDINPUT{Vk: 0, Scan: 0, Flags: flags}

	if flags&(winapi.KEYEVENTF_SCANCODE|winapi.KEYEVENTF_UNICODE) != 0 {
		ki.Scan = key
	} else {
		ki.Vk = key
	}

	return []winapi.INPUT_Ki{winapi.NewKeybdInput(ki)}
}

// setMods sets the modifier key state for the current and next iteration of a
// modifier key.
func setMods(flags winapi.KiFlags, mods byte, modsNext byte) bool {
	var modsSetCount uint
	for _, m := range StandardMods {
		if mods&m.Mask != 0 {
			if flags&winapi.KEYEVENTF_KEYUP != 0 {
				if modsNext&m.Mask == 0 {
					_ = KeyRelease(m.VSC, flags)
					modsSetCount++
				}
			} else {
				if !KeyIsDown(m.VK) {
					_ = KeyPress(m.VSC, flags)
					modsSetCount++
				}
			}
		}
	}

	return modsSetCount > 0
}

// typeStr is the base function for TypeStr that primarily handles the rune
// translation and the actual key presses.
func typeStr(str string, hkl winapi.Handle) (err error) {
	runes := []rune(str)
	iLast := len(runes) - 1

	var (
		errCount int
		vscNext  uint16
		modsNext byte
	)

	vsc, mods, err := RuneToVSC(runes[0], hkl)
	if err != nil {
		errCount++
	}

	for i, r := range runes {
		if abortFlag {
			abortFlag = false
			return fmt.Errorf("%s", ErrAborted)
		}

		modFlags := winapi.KEYEVENTF_SCANCODE
		if modsSet := setMods(modFlags, mods, 0); modsSet {
			time.Sleep(TypeString.ModPressDuration)
		}

		numTaps := 1
		if r == '\t' && TypeString.TabsToSpaces {
			vsc = VSC_SPACE
			numTaps = TypeString.TabSize
		}

		for range numTaps {
			if err = KeyTap(vsc, modFlags); err != nil {
				errCount++
			}
		}

		if i < iLast {
			vscNext, modsNext, err = RuneToVSC(runes[i+1], hkl)
			if err != nil {
				errCount++
			}
		} else if i == iLast {
			modsNext = 0
		}

		modFlags |= winapi.KEYEVENTF_KEYUP
		_ = setMods(modFlags, mods, modsNext)

		if i < iLast {
			vsc = vscNext
			mods = modsNext
			time.Sleep(TypeString.KeyDelay)
		}
	}

	if errCount == 0 {
		return nil
	}

	return err
}
