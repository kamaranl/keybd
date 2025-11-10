// Package keybd implements functions that allow keyboard synthesization on both
// MacOS and Windows desktops.
package keybd

import "time"

// Global is a struct that contains settings for changing how some of the
// functions in [keybd] operate. These settings are initialized upon import.
var Global struct {
	// KeyDelay is how long to wait after releasing a key and before proceeding
	// with pressing the next key.
	//
	// Default: 2ms
	KeyDelay time.Duration

	// KeyPressDuration is how long to wait after pressing a key and before
	// releasing the same key.
	//
	// Default: 2ms
	KeyPressDuration time.Duration

	// ModPressDuration is how long to wait after pressing a modifier key and
	// before releasing the same modifier key.
	//
	// Default: 2ms
	ModPressDuration time.Duration

	// MaxCharacters is the maximum amount of characters in a string that can be
	// sent to [TypeStr].
	//
	// Default: 5000
	MaxCharacters int

	// TabsToSpaces is a switch to enable the conversion of tabs to spaces as
	// they are typed with [TypeStr].
	//
	// Default: false
	TabsToSpaces bool

	// TabSize is the number of spaces to use in place of tabs when TabsToSpaces
	// is true.
	//
	// Default: 4
	TabSize int

	// TypeStringTimeout is how long [TypeStr] can run before finally aborting.
	//
	// Default: 30s
	TypeStringTimeout time.Duration
}

func init() {
	Global.KeyDelay = 2 * time.Millisecond
	Global.KeyPressDuration = 2 * time.Millisecond
	Global.ModPressDuration = 2 * time.Millisecond
	Global.MaxCharacters = 5000
	Global.TabsToSpaces = false
	Global.TabSize = 4
	Global.TypeStringTimeout = 30 * time.Second
}
