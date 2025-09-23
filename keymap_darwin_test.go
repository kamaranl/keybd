//go:build !(windows || linux)

package keymap_test

import (
	"fmt"
	"testing"

	"github.com/kamaranl/keymap"
)

func Test_GetKeyboardInfo(t *testing.T) {
	kbInfo := keymap.GetKeyboardInfo()

	fmt.Printf("layout=%v\ntype=%v", kbInfo.Layout, kbInfo.Kind)
}

func Test_CharToVK(t *testing.T) {
	// alt := "ß"
	// shiftAlt := "Í"

	for _, r := range " \t\r\n" {
		kbInfo := keymap.GetKeyboardInfo()
		err, vk, mods := keymap.CharToVKAndMods(r, kbInfo)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("%q  (%d): 0x%x (%d) | 0x%x\n", r, r, vk, vk, mods)
		}
	}
}

func Test_KeyIsDown(t *testing.T) {
	r1 := keymap.KeyIsDown(56)
	print(r1)
}

func Test_TypeStr(t *testing.T) {
	if err := keymap.TypeStr(multiLineStr); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func init() {
	setup()
}
