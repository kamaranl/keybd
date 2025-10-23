//go:build darwin

package keybd_test

import (
	"fmt"
	"testing"

	"github.com/kamaranl/keybd"
)

func Test_GetKeyboardInfo(t *testing.T) {
	kbInfo := keybd.GetKeyboardInfo()

	fmt.Printf("layout=%v\ntype=%v", kbInfo.Layout, kbInfo.Kind)
}

func Test_CharToVK(t *testing.T) {
	// alt := "ß"
	// shiftAlt := "Í"

	for _, r := range " \t\r\n" {
		kbInfo := keybd.GetKeyboardInfo()
		err, vk, mods := keybd.CharToVKAndMods(r, kbInfo)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("%q  (%d): 0x%x (%d) | 0x%x\n", r, r, vk, vk, mods)
		}
	}
}

func Test_KeyIsDown(t *testing.T) {
	r1 := keybd.KeyIsDown(56)
	print(r1)
}

func Test_TypeStr(t *testing.T) {
	if err := keybd.TypeStr(multiLineStr); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func init() {
	setup()
}
