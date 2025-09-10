package keymap

/*
#cgo CFLAGS: -x objective-c -fmodules -framework ApplicationServices
#include "key_ev.h"
*/
import "C"
import (
	"fmt"
)

func SendKey(key int, shift ...bool) error {
	if key == KeyNotMapped {
		return fmt.Errorf("Key %q is not mapped", key)
	}

	if len(shift) < 1 {
		shift = append(shift, false)
	}

	C.SendKey(C.int(key), C._Bool(shift[0]))

	return nil
}

func TypeStr(str string) {
	for _, r := range str {
		key, shift := CharToVKey(r)

		if key == KeyNotMapped {
			continue
		}

		if GlobalOptions.TabsToSpaces && key == int(Whitespace["\t"]) {
			for i := 0; i < GlobalOptions.TabSize; i++ {
				C.SendKey(C.int(Whitespace[" "]), C._Bool(shift))
			}

			continue
		}

		C.SendKey(C.int(key), C._Bool(shift))
	}
}
