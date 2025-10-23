//go:build windows

package keybd

import (
	"errors"
	"fmt"
	"time"

	"github.com/kamaranl/keybd/winapi"
)

const (
	ScanSpace      = 0x39
	ScanReturn     = 0x1C
	ScanTab        = 0x0F
	ScanLShift     = 0x02A
	ScanLCtrl      = 0x01D
	ScanLAlt       = 0x038
	ScanUnassigned = 0x200
)

const (
	ModLShift ModifierMask = 1 << iota // 0x01
	ModLCtrl                           // 0x02
	ModLAlt                            // 0x04
)

var StandardMods = []ModifierSet{
	{Mask: ModLShift, Scan: ScanLShift, VK: winapi.VK_LSHIFT},
	{Mask: ModLCtrl, Scan: ScanLCtrl, VK: winapi.VK_LCONTROL},
	{Mask: ModLAlt, Scan: ScanLAlt, VK: winapi.VK_LMENU},
}

type KeyCode = uint16

type ModifierMask = uint8

type ModifierSet struct {
	Mask ModifierMask
	Scan KeyCode
	VK   KeyCode
}

func CharToVKAndMods(r rune, hkl uintptr) (err error, vk KeyCode, mods ModifierMask) {
	switch r {
	case '\r':
		return nil, winapi.VK_UNASSIGNED, 0
	case '\n':
		return nil, winapi.VK_RETURN, 0
	case '\t':
		return nil, winapi.VK_TAB, 0
	}

	err, vkShort := winapi.VkKeyScanEx(r, hkl)
	if err != nil {
		return err, 0, 0
	}

	vkMods := uint16(vkShort & 0xFFFF)
	vk = KeyCode(vkMods & 0xFF)
	mods = ModifierMask((vkMods >> 8) & 0xFF)

	return nil, vk, mods
}

func CharToScanAndMods(r rune, hkl uintptr) (err error, sc KeyCode, mods ModifierMask) {
	switch r {
	case '\r':
		return nil, ScanUnassigned, 0
	case '\n':
		return nil, ScanReturn, 0
	case '\t':
		return nil, ScanTab, 0
	}

	err, vk, mods := CharToVKAndMods(r, hkl)
	if err != nil {
		return err, 0, 0
	}

	err, code := VKtoSC(vk, hkl)
	if err != nil {
		return err, 0, 0
	}

	return nil, KeyCode(code), mods
}

func KeyIsDown(vk KeyCode) bool {
	return int16(winapi.GetKeyState(uint8(vk))) < 0
}

func KeyPress(key KeyCode, flags winapi.KeyEventFlags) error {
	return winapi.SendInput([]winapi.Input{keyDown(key, flags)})
}

func KeyPressAndRelease(key KeyCode, flags winapi.KeyEventFlags, noDelay ...bool) (err error) {
	var errs []error

	if err = KeyPress(key, flags); err != nil {
		errs = append(errs, err)
	}

	if len(noDelay) == 0 || (len(noDelay) > 0 && !noDelay[0]) {
		time.Sleep(Global.KeyPressDuration)
	}

	if err = KeyRelease(key, flags); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func KeyRelease(key KeyCode, flags winapi.KeyEventFlags) error {
	return winapi.SendInput([]winapi.Input{keyUp(key, flags)})
}

func SCtoVK(scan KeyCode, hkl uintptr) (err error, vk KeyCode) {
	err, code := winapi.MapVirtualKeyEx(uint(scan), winapi.MAPVK_VSC_TO_VK_EX, hkl)
	if err != nil {
		return err, 0
	}
	return nil, KeyCode(code)
}

func TypeStr(str string) (err error) {
	if len(str) == 0 {
		return nil
	} else if len(str) > Global.MaxCharacters {
		return fmt.Errorf("Exceeds max character limit")
	}

	var (
		errs   []error
		fwThId uintptr
	)

	hfw := winapi.GetForegroundWindow()
	if hfw != 0 {
		fwThId = winapi.GetWindowThreadProcessId(hfw)
	}

	thId := winapi.GetCurrentThreadId()
	attached := false
	if fwThId != 0 && thId != fwThId {
		if err = winapi.AttachThreadInput(thId, fwThId, true); err != nil {
			errs = append(errs, err)
		} else {
			attached = true
			defer func() {
				if err = winapi.AttachThreadInput(thId, fwThId, false); err != nil {
					errs = append(errs, err)
				}
			}()
		}
	}

	if err = winapi.BringWindowToTop(hfw); err != nil {
		errs = append(errs, err)
	}

	if err = winapi.SetForegroundWindow(hfw); err != nil {
		errs = append(errs, err)
	}

	_ = winapi.SetFocus(hfw)

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

	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(Global.TypeStringTimeout):
			fmt.Println("Exceeded timeout... forcing cleanup")
			if blocked {
				fmt.Println("Unblocking input...")
				_ = winapi.BlockInput(false)
			}
			if attached {
				fmt.Println("Detaching thread...")
				_ = winapi.AttachThreadInput(thId, fwThId, false)
			}

			fmt.Println("Done")
		case <-done:
		}
	}()

	hkl := winapi.GetKeyboardLayout(fwThId)
	if err = typeStr(str, hkl); err != nil {
		errs = append(errs, err)
	}
	close(done)

	return errors.Join(errs...)
}

func VKtoSC(vk KeyCode, hkl uintptr) (err error, sc KeyCode) {
	err, code := winapi.MapVirtualKeyEx(uint(vk), winapi.MAPVK_VK_TO_VSC_EX, hkl)
	if err != nil {
		return err, 0
	}
	return nil, KeyCode(code)
}

func keyDown(key uint16, flags winapi.KeyEventFlags) winapi.Input {
	return keyEvent(key, flags)
}

func keyEvent(key uint16, flags winapi.KeyEventFlags) winapi.Input {
	input := winapi.Input{
		Type: winapi.INPUT_KEYBOARD,
		Ki:   winapi.KeybdInput{Vk: 0, Scan: 0, Flags: flags},
	}

	if flags&(winapi.KEYEVENTF_SCANCODE|winapi.KEYEVENTF_UNICODE) != 0 {
		input.Ki.Scan = key
	} else {
		input.Ki.Vk = key
	}

	return input
}

func keyUp(key uint16, flags winapi.KeyEventFlags) winapi.Input {
	return keyEvent(key, flags|winapi.KEYEVENTF_KEYUP)
}

func setMods(flags winapi.KeyEventFlags, mods ModifierMask, modsNext ModifierMask) (errCount uint, modsWereSet bool) {
	var modsSetCount uint

	for _, m := range StandardMods {
		if mods&m.Mask != 0 {
			if flags&winapi.KEYEVENTF_KEYUP != 0 {
				if modsNext&m.Mask == 0 {
					if err := KeyRelease(m.Scan, flags); err != nil {
						errCount++
					}
				}
			} else {
				if !KeyIsDown(m.VK) {
					if err := KeyPress(m.Scan, flags); err != nil {
						errCount++
					} else {
						modsSetCount++
					}
				}
			}
		}
	}

	return errCount, modsSetCount > 0
}

func typeStr(str string, hkl uintptr) (err error) {
	const (
		E_LOOKUP      = "Failed to retrieve scan code"
		E_SET_MODS    = "Failed to set modifier(s)"
		E_KEY_DOWN_UP = "Unexpected while attempting key press/release"
	)

	var (
		sc       KeyCode
		mods     ModifierMask
		scNext   KeyCode
		modsNext ModifierMask
	)

	errMap := make(map[string]uint)
	runes := []rune(str)
	iLast := len(runes) - 1

	for i, r := range runes {
		if i == 0 {
			err, sc, mods = CharToScanAndMods(r, hkl)
			if err != nil {
				errMap[E_LOOKUP]++
			}
		}

		if i < iLast {
			err, scNext, modsNext = CharToScanAndMods(runes[i+1], hkl)
			if err != nil {
				errMap[E_LOOKUP]++
			}
		} else if i == iLast {
			modsNext = 0
		}

		flags := winapi.KEYEVENTF_SCANCODE

		errCount, modsWereSet := setMods(flags, mods, 0)
		if errCount != 0 {
			errMap[E_SET_MODS] += errCount
		}
		if modsWereSet {
			time.Sleep(Global.ModPressDuration)
		}

		j := 1

		if r == '\t' && Global.TabsToSpaces {
			sc = ScanSpace
			j = Global.TabSize
		}

		for range j {
			if err = KeyPressAndRelease(sc, flags); err != nil {
				errMap[E_KEY_DOWN_UP]++
			}
		}

		flags |= winapi.KEYEVENTF_KEYUP

		errCount, _ = setMods(flags, mods, modsNext)
		if errCount != 0 {
			errMap[E_SET_MODS] += errCount
		}

		if i < iLast {
			sc = scNext
			mods = modsNext
			time.Sleep(Global.KeyDelay)
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
