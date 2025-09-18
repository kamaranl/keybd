package keymap_test

import (
	"strings"
	"time"
)

var (
	simpleChar  = 'K'
	complexChar = '†'

	shortWord   = "cat"
	complexWord = "Pneumonoultramicroscopicsilicovolcanoconiosis"

	simpleStr  = "Hello World!\r\n\tHow are we today? :)"
	simpleStr2 = "Good morning!!!"
	invalidStr = "Praise the Lord †\n\t→Say your prayers\n\t→Count your blessings"

	multiLineStr = `
function uuidgen
{
	if which uuidgen &>/dev/null; then
		/usr/bin/uuidgen | tr [:upper:] [:lower:]
	else
		cat /proc/sys/kernel/random/uuid
	fi
}
`

	multiLineStrSp = strings.ReplaceAll(multiLineStr, string('\t'), " ")

	multiLineStr2 = `
function uuidgen
{
    if which uuidgen &>/dev/null; then
        /usr/bin/uuidgen | tr [:upper:] [:lower:]
    else
        cat /proc/sys/kernel/random/uuid
    fi
}
`
)

var testStr = "Hello World!"

var testCode = `
package keymap_test

import (
	"fmt"
	"time"
)

func countdown(seconds int) {
	for i := seconds; i > 0; i-- {
		fmt.Printf("Testing in %d...", i)
		print("\r")

		time.Sleep(1 * time.Second)
	}
	println()
}

func setup() {
	countdown(5)
}
`

func countdown(seconds int) {
	for i := seconds; i > 0; i-- {
		time.Sleep(1 * time.Second)
	}
}

func setup() {
	countdown(5)

}
