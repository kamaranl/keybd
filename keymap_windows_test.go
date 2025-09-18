//go:build !(darwin || linux)

package keymap_test

import (
	"fmt"
	"testing"

	"github.com/kamaranl/keymap"
)

func Test_CharToScan(t *testing.T) {
	for _, r := range simpleStr {
		if err, scan, modLR := keymap.CharToScanAndMods(r, 0); err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("%q  (%d): 0x%x (%d) | 0x%x\n", r, r, scan, scan, modLR)
		}
	}
}

func Test_CharToVK(t *testing.T) {
	for _, r := range simpleStr {
		if err, vk, modLR := keymap.CharToVKAndMods(r, 0); err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("%q  (%d): 0x%x | 0x%x\n", r, r, vk, modLR)
		}
	}
}

func TestRandom(t *testing.T) {
	keymap.Global.TabsToSpaces = true
	keymap.Global.TabSize = 5

	if err := keymap.TypeStr(multiLineStr); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func init() {
	setup()
}
