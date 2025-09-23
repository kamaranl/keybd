//go:build !(darwin || linux)

package keybd_test

import (
	"fmt"
	"testing"

	"github.com/kamaranl/keybd"
)

func Test_CharToScan(t *testing.T) {
	for _, r := range simpleStr {
		if err, scan, modLR := keybd.CharToScanAndMods(r, 0); err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("%q  (%d): 0x%x (%d) | 0x%x\n", r, r, scan, scan, modLR)
		}
	}
}

func Test_CharToVK(t *testing.T) {
	for _, r := range simpleStr {
		if err, vk, modLR := keybd.CharToVKAndMods(r, 0); err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("%q  (%d): 0x%x | 0x%x\n", r, r, vk, modLR)
		}
	}
}

func TestRandom(t *testing.T) {
	keybd.Global.TabsToSpaces = true
	keybd.Global.TabSize = 5

	if err := keybd.TypeStr(multiLineStr); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func init() {
	setup()
}
