package keymap_test

import (
	"testing"

	"github.com/kamaranl/keymap"
)

func TestSendKey(t *testing.T) {
	setup()

	t.Run("sends a valid key", func(t *testing.T) {
		key, shift := keymap.CharToVKey('K')
		keymap.SendKey(key, shift)
	})

	t.Run("sends an invalid key", func(t *testing.T) {
		key, shift := keymap.CharToVKey('†')
		if err := keymap.SendKey(key, shift); err == nil {
			t.Errorf("expected an error but got nil")
		}
	})
}

func TestTypeStr(t *testing.T) {
	setup()

	t.Run("types test string", func(t *testing.T) {
		keymap.TypeStr(testStr)
	})

	t.Run("types test code block", func(t *testing.T) {
		keymap.GlobalOptions.TabsToSpaces = true
		keymap.TypeStr(testCode)
	})

}
