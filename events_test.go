package keymap_test

import (
	"fmt"
	"time"
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
		fmt.Printf("Testing in %d...", i)
		print("\r")

		time.Sleep(1 * time.Second)
	}
	println()
}

func setup() {
	countdown(5)
}
