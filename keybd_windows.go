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

// Constants for errors.
const (
	E_LOOKUP      = "failed to retrieve scan code"
	E_SET_MODS    = "failed to set modifier(s)"
	E_KEY_DOWN_UP = "unexpected error during key press and release"
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
	return winapi.SendInput(keyEvent(key, flags))
}

// KeyRelease sends a key-up event and is intended to be used after a call to
// [KeyPress].
// It returns an error if the call fails.
func KeyRelease(key uint16, flags winapi.KiFlags) error {
	return winapi.SendInput(keyEvent(key, flags|winapi.KEYEVENTF_KEYUP))
}

// KeyTap sends a key-down event and a key-up event with a brief pause in
// between to help simulate an actual keystroke. The duration of the pause is
// defined by [Global].
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

// TypeStr types str using the [Global] options and ensures accuracy by
// attaching the current thread to the thread of the foreground window and
// temporary blocking input while attached. A timeout prevents the function call
// from hanging indefinitely, thus allowing the block lift if the timeout is
// exceeded.
// It returns an error if the call fails.
func TypeStr(str string) (err error) {
	if len(str) == 0 {
		return nil
	} else if len(str) > TypeString.MaxCharacters {
		return fmt.Errorf("%s", ErrMaxCharacter)
	}

	hwnd := windows.GetForegroundWindow()

	var (
		errs        []error
		pid         uint32
		tidAttachTo uint32
	)

	if hwnd != 0 {
		tidAttachTo, err = windows.GetWindowThreadProcessId(hwnd, &pid)
		if err != nil {
			errs = append(errs, err)
		}
	}

	tidAttach := windows.GetCurrentThreadId()
	attached := false
	if tidAttachTo != 0 && tidAttach != tidAttachTo {
		if err = winapi.AttachThreadInput(tidAttach, tidAttachTo, true); err != nil {
			errs = append(errs, err)
		} else {
			attached = true
			defer func() {
				if err = winapi.AttachThreadInput(tidAttach, tidAttachTo, false); err != nil {
					errs = append(errs, err)
				}
			}()
		}
	}
	if err = winapi.BringWindowToTop(hwnd); err != nil {
		errs = append(errs, err)
	}
	if err = winapi.SetForegroundWindow(hwnd); err != nil {
		errs = append(errs, err)
	}
	if _, err = winapi.SetFocus(hwnd); err != nil {
		errs = append(errs, err)
	}

	blocked := false
	if err = winapi.BlockInput(true); err != nil {
		errs = append(errs, err)
	} else {
		blocked = true
		defer func() {
			if err = winapi.BlockInput(false); err != nil {
				errs = append(errs, err)
			}
		}()
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
	case err := <-done:
		errs = append(errs, err)
		return errors.Join(errs...)
	case <-ctx.Done():
		cleanup()
		abortFlag = true
		return fmt.Errorf("%s", ErrTimeout)
	case <-abort:
		cleanup()
		abortFlag = true
		return fmt.Errorf("%s", ErrAborted)
	}
}

// keyEvent creates an input that can be processed by [winapi.SendInput].
func keyEvent(key uint16, flags winapi.KiFlags) []winapi.INPUT_Ki {
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
func setMods(flags winapi.KiFlags, mods byte, modsNext byte) (modsSet bool, errCount uint) {
	var modsSetCount uint

	for _, m := range StandardMods {
		if mods&m.Mask != 0 {
			if flags&winapi.KEYEVENTF_KEYUP != 0 {
				if modsNext&m.Mask == 0 {
					if err := KeyRelease(m.VSC, flags); err != nil {
						errCount++
					}
				}
			} else {
				if !KeyIsDown(m.VK) {
					if err := KeyPress(m.VSC, flags); err != nil {
						errCount++
					} else {
						modsSetCount++
					}
				}
			}
		}
	}

	return modsSetCount > 0, errCount
}

// typeStr is the base function for TypeStr that primarily handles the rune
// translation and the actual key presses.
func typeStr(str string, hkl winapi.Handle) (err error) {
	errMap := make(map[string]uint)
	runes := []rune(str)
	iLast := len(runes) - 1

	var (
		vscNext  uint16
		modsNext byte
	)

	vsc, mods, err := RuneToVSC(runes[0], hkl)
	if err != nil {
		errMap[E_LOOKUP]++
	}

	for i, r := range runes {
		if abortFlag {
			abortFlag = false
			return fmt.Errorf("%s", ErrAborted)
		}

		modFlags := winapi.KEYEVENTF_SCANCODE
		if modsSet, errCount := setMods(modFlags, mods, 0); errCount != 0 {
			errMap[E_SET_MODS] += errCount
		} else if modsSet {
			time.Sleep(TypeString.ModPressDuration)
		}

		numTaps := 1
		if r == '\t' && TypeString.TabsToSpaces {
			vsc = VSC_SPACE
			numTaps = TypeString.TabSize
		}

		for range numTaps {
			if err = KeyTap(vsc, modFlags); err != nil {
				errMap[E_KEY_DOWN_UP]++
			}
		}

		modFlags |= winapi.KEYEVENTF_KEYUP

		if i < iLast {
			vscNext, modsNext, err = RuneToVSC(runes[i+1], hkl)
			if err != nil {
				errMap[E_LOOKUP]++
			}
		} else if i == iLast {
			modsNext = 0
		}
		if _, errCount := setMods(modFlags, mods, modsNext); errCount != 0 {
			errMap[E_SET_MODS] += errCount
		}
		if i < iLast {
			vsc = vscNext
			mods = modsNext
			time.Sleep(TypeString.KeyDelay)
		}
	}

	var errs []error
	if len(errMap) > 0 {
		for msg, msgCount := range errMap {
			errs = append(errs, fmt.Errorf("(x %d) %s", msgCount, msg))
		}
	}

	return errors.Join(errs...)
}
