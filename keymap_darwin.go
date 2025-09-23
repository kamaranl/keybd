//go:build !(windows || linux)

package keymap

/*
#cgo CFLAGS: -x objective-c -fmodules
#cgo LDFLAGS: -framework Carbon -framework ApplicationServices -framework Cocoa
#include "keybd.h"
*/
import "C"
import (
	"context"
	"fmt"
	"time"
	"unsafe"
)

const (
	VK_None   uint = 0xFFFF
	VK_Return uint = 0x24
	VK_Tab    uint = 0x30
	VK_Space  uint = 0x31
)

type KeyboardInfo = struct {
	Layout *C.UCKeyboardLayout
	Kind   C.int
}

func GetKeyboardInfo() KeyboardInfo {
	info := C.GetKeyboardInfo()
	return KeyboardInfo{
		Layout: info.layout,
		Kind:   info.kind,
	}
}

func CharToVKAndMods(r rune, kbInfo KeyboardInfo) (err error, vk uint, mods uint) {
	switch r {
	case '\r':
		return nil, VK_None, 0
	case '\n':
		return nil, VK_Return, 0
	case '\t':
		return nil, VK_Tab, 0
	case ' ':
		return nil, VK_Space, 0
	}

	var info C.KeyboardInfo
	info.layout = kbInfo.Layout
	info.kind = kbInfo.Kind

	kMap := C.CharToVKAndMods(C.UniChar(r), info)

	if kMap.vk == C.kVK_None {
		return fmt.Errorf("No translation for char %q (0x%x)", r, r), 0, 0
	}

	return nil, uint(kMap.vk), uint(kMap.mods)
}

func KeyIsDown(vk uint) bool {
	return C.KeyIsDown(C.CGKeyCode(vk)) != 0
}

func TypeStr(str string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // global TypeStr timeout
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- typeStr(str)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("Exceeded timeout: %w", ctx.Err())
	}
}

func typeStr(str string) (err error) {
	if len(str) == 0 {
		return nil
	} else if len(str) > 5000 { // global max chars
		return fmt.Errorf("TypeStr: Exceeds max character limit")
	}

	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))

	if r1 := C.TypeStr(
		cStr,
		C.int((2*time.Millisecond)/time.Microsecond), // global modPressDur
		C.int((2*time.Millisecond)/time.Microsecond), // global keyPressDur
		C.int((2*time.Millisecond)/time.Microsecond), // global keyDelay
		C.int(6), // global tabSize
	); r1 == 0 {
		return fmt.Errorf("%v", C.LastErrorMessage)
	}

	return nil
}
