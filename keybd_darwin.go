//go:build darwin

package keybd

// #cgo LDFLAGS: -framework Carbon
// #import "keybd_darwin.h"
import "C"

import (
	"context"
	"fmt"
	"time"
	"unsafe"
)

// Constants for virtual key codes of whitespace characters.
const (
	VK_Shift  = 0x38
	VK_Option = 0x3A
	VK_Return = 0x24
	VK_Tab    = 0x30
	VK_Space  = 0x31
	VK_None   = 0xFFFF
)

// Constants for modifier key masks.
const (
	Mod_Shift  = 0x2
	Mod_Option = 0x8
)

// Constants for modifier key flags.
const (
	Flag_Shift  = 0x20000
	Flag_Option = 0x80000
)

// StandardMods is a [Modifier] slice of the standard modifier keys.
var StandardMods = []Modifier{
	{Mask: Mod_Shift, VK: VK_Shift, Flag: Flag_Shift},
	{Mask: Mod_Option, VK: VK_Option, Flag: Flag_Option},
}

// A KeyboardLayoutInfo is a struct that contains the keyboard layout and
// keyboard type of the current machine.
type KeyboardLayoutInfo = struct {
	Layout *C.UCKeyboardLayout
	Type   C.int
}

// A Modifier is a struct that contains the mask, the virtual key code, and the
// event flag for a modifier key.
type Modifier = struct {
	Mask uint16
	VK   uint16
	Flag uint64
}

// GetKeyboardLayoutInfo retrieves the layout and type for the local machine.
// It always returns a [KeyboardLayoutInfo].
func GetKeyboardLayoutInfo() KeyboardLayoutInfo {
	r1 := C.GetKeyboardLayoutInfo()
	return KeyboardLayoutInfo{
		Layout: r1.kbLayout,
		Type:   r1.kbType,
	}
}

// RuneToVK translates r to a virtual key code and its shift state.
// It returns a pair of 0's with an error if the translation fails, otherwise it
// returns the key code, shift state, and a nil error.
func RuneToVK(r rune, kli KeyboardLayoutInfo) (code, shift uint16, err error) {
	switch r {
	case '\r':
		return VK_None, 0, nil
	case '\n':
		return VK_Return, 0, nil
	case '\t':
		return VK_Tab, 0, nil
	case ' ':
		return VK_Space, 0, nil
	}

	var info C.KeyboardLayoutInfo
	info.kbLayout = kli.Layout
	info.kbType = kli.Type

	kt := C.TranslateChar(C.UniChar(r), info)

	if kt.vk == C.kVK_None {
		return 0, 0, fmt.Errorf("%s", C.GoString(&C.LastErrorMessage[0]))
	}

	return uint16(kt.vk), uint16(kt.mods), nil
}

// KeyIsDown detects the down state of virtKey.
// It returns true if the key is currently depressed and false if it is not.
func KeyIsDown(virtKey uint16) bool { return C.KeyIsDown(C.CGKeyCode(virtKey)) != 0 }

// KeyPress sends a key-down event and is intended to be used before a call to
// [KeyRelease].
// It returns an error if the call fails.
func KeyPress(key uint16, flags uint64) error {
	if r1 := C.KeyPress(C.CGKeyCode(key), C.CGEventFlags(flags)); r1 == 0 {
		return fmt.Errorf("%s", C.GoString(&C.LastErrorMessage[0]))
	}

	return nil
}

// KeyRelease sends a key-up event and is intended to be used after a call to
// [KeyPress].
// It returns an error if the call fails.
func KeyRelease(key uint16, flags uint64) error {
	if r1 := C.KeyRelease(C.CGKeyCode(key), C.CGEventFlags(flags)); r1 == 0 {
		return fmt.Errorf("%s", C.GoString(&C.LastErrorMessage[0]))
	}

	return nil
}

// KeyTap sends a key-down event and a key-up event with a brief pause in
// between to help simulate an actual keystroke. The duration of the pause is
// defined by [Global].
// It returns an error if the call fails.
func KeyTap(key uint16, flags uint64) error {
	if r1 := C.KeyTap(C.CGKeyCode(key), C.CGEventFlags(flags), C.int(Global.KeyPressDuration)); r1 == 0 {
		return fmt.Errorf("%s", C.GoString(&C.LastErrorMessage[0]))
	}

	return nil
}

// TypeStr types str using the [Global] options. A timeout prevents the function
// call from hanging indefinitely.
// It returns an error if the call fails.
func TypeStr(str string) (err error) {
	if len(str) == 0 {
		return nil
	} else if len(str) > Global.MaxCharacters {
		return fmt.Errorf("%s", ErrMaxCharacter)
	}

	ctx, cancel := context.WithTimeout(context.Background(), Global.TypeStringTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- typeStr(str)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("%s", ErrTimeout)
	}
}

// typeStr is the base function for TypeStr that makes the underlying call to
// the C.TypeStr function.
func typeStr(str string) (err error) {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))

	if r1 := C.TypeStr(
		cStr,
		C.int(Global.ModPressDuration/time.Microsecond),
		C.int(Global.KeyPressDuration/time.Microsecond),
		C.int(Global.KeyDelay/time.Microsecond),
		C.int(boolToInt(Global.TabsToSpaces)),
		C.int(Global.TabSize),
	); r1 == 0 {
		return fmt.Errorf("%s", C.GoString(&C.LastErrorMessage[0]))
	}

	return nil
}

// boolToInt converts a bool to an int.
func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}
