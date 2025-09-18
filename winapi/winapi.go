//go:build !(darwin || linux)

package winapi

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	VK_RETURN     = 0x0D
	VK_TAB        = 0x09
	VK_SPACE      = 0x20
	VK_LSHIFT     = 0xA0
	VK_LCONTROL   = 0xA2
	VK_LMENU      = 0xA4 // menu == alt
	VK_UNASSIGNED = 0x9F
)

const (
	_ = iota
	INPUT_KEYBOARD
	_
)

const (
	MAPVK_VK_TO_VSC MapType = iota
	MAPVK_VSC_TO_VK
	MAPVK_VK_TO_CHAR
	MAPVK_VSC_TO_VK_EX
	MAPVK_VK_TO_VSC_EX
)

const (
	KEYEVENTF_EXTENDEDKEY KeyEventFlags = 1 << iota
	KEYEVENTF_KEYUP
	KEYEVENTF_UNICODE
	KEYEVENTF_SCANCODE
)

// dlls
var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
)

// procs
var (
	procAttachThreadInput = user32.NewProc("AttachThreadInput")

	procBlockInput = user32.NewProc("BlockInput")

	procBringWindowToTop = user32.NewProc("BringWindowToTop")

	procGetCurrentThreadId = kernel32.NewProc("GetCurrentThreadId")

	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")

	procGetKeyboardLayout = user32.NewProc("GetKeyboardLayout")

	procGetKeyState = user32.NewProc("GetKeyState")

	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")

	procMapVirtualKeyEx = user32.NewProc("MapVirtualKeyExW")

	procSendInput = user32.NewProc("SendInput")

	procSetFocus = user32.NewProc("SetFocus")

	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")

	procVkKeyScanEx = user32.NewProc("VkKeyScanExW")
)

type Input struct {
	Type InputType
	Ki   KeybdInput

	_ [8]byte
}

type KeybdInput struct {
	Vk        uint16
	Scan      uint16
	Flags     KeyEventFlags
	Time      uint32
	ExtraInfo uintptr
}
type InputType = uint32

type MapType = uintptr

type KeyEventFlags = uint32

func AttachThreadInput(attachThId uintptr, attachToThId uintptr, attach bool) (err error) {
	var pAttach uintptr
	if attach {
		pAttach = 1
	}

	if r1, _, err := procAttachThreadInput.Call(attachThId, attachToThId, pAttach); r1 == 0 {
		return fmt.Errorf("Failed to AttachThreadInput: %v", err)
	}
	return nil
}

func BlockInput(block bool) (err error) {
	var pBlock uintptr
	if block {
		pBlock = 1
	}

	if r1, _, err := procBlockInput.Call(pBlock); r1 == 0 {
		return fmt.Errorf("Failed to BlockInput: %v", err)
	}
	return nil
}

func BringWindowToTop(hwnd uintptr) (err error) {
	if r1, _, err := procBringWindowToTop.Call(hwnd); r1 == 0 {
		return fmt.Errorf("Failed to BringWindowToTop: %v", err)
	}
	return nil
}

func GetCurrentThreadId() (thId uintptr) {
	thId, _, _ = procGetCurrentThreadId.Call()
	return thId
}

func GetForegroundWindow() (hwnd uintptr) {
	hwnd, _, _ = procGetForegroundWindow.Call()
	return hwnd
}

func GetKeyboardLayout(thId uintptr) (hkl uintptr) {
	hkl, _, _ = procGetKeyboardLayout.Call(thId)
	return hkl
}

func GetKeyState(vk uint8) (state uintptr) {
	state, _, _ = procGetKeyState.Call(uintptr(vk))
	return state
}

func GetWindowThreadProcessId(hwnd uintptr) (thId uintptr) {
	thId, _, _ = procGetWindowThreadProcessId.Call(hwnd)
	return thId
}

func MapVirtualKeyEx(key uint, mapType MapType, hkl uintptr) (err error, code uintptr) {
	if code, _, _ = procMapVirtualKeyEx.Call(uintptr(key), mapType, hkl); code == 0 {
		return fmt.Errorf("Failed to MapVirtualKey: No translation available"), 0
	} else {
		return nil, code
	}
}

func SendInput(inputs []Input) (err error) {
	if len(inputs) == 0 {
		return nil
	}

	if r1, _, err := procSendInput.Call(uintptr(len(inputs)), uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(inputs[0])); r1 == 0 {
		if err != syscall.Errno(0) {
			return fmt.Errorf("Failed to SendInput: %v", err)
		}
		return syscall.EINVAL
	}
	return nil
}

func SetFocus(hwnd uintptr) uintptr {
	r1, _, _ := procSetFocus.Call(hwnd)
	return r1
}

func SetForegroundWindow(hwnd uintptr) (err error) {
	if r1, _, err := procSetForegroundWindow.Call(hwnd); r1 == 0 {
		return fmt.Errorf("Failed to SetForegroundWindow: %v", err)
	}
	return nil
}

func VkKeyScanEx(r rune, hkl uintptr) (err error, vkShort uintptr) {
	if vkShort, _, _ = procVkKeyScanEx.Call(uintptr(r), uintptr(hkl)); vkShort == 0xFFFF {
		return fmt.Errorf("Failed to VkKeyScan: No translation for char %q (0x%x)", r, r), 0
	} else {
		return nil, vkShort
	}
}
