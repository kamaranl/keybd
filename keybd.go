// Package keybd implements functions that allow keyboard synthesization on both
// MacOS and Windows desktops.
package keybd

import (
	"sync"
	"time"
)

// Constants for common cross-platform errors.
const (
	ErrAborted      = "operation aborted"
	ErrMaxCharacter = "character limit exceeded"
	ErrTimeout      = "timeout exceeded"
)

var (
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

	// TypeString is a struct that contains specific settings for [TypeStr].
	// These settings are initialized upon import.
	TypeString struct {
		// Timeout is how long [TypeStr] can run before finally aborting.
		//
		// Default: 30s
		Timeout time.Duration

		abort chan struct{}
		mu    sync.Mutex
	}
)

func init() {
	KeyDelay = 2 * time.Millisecond
	KeyPressDuration = 2 * time.Millisecond
	ModPressDuration = 2 * time.Millisecond
	MaxCharacters = 5000
	TabsToSpaces = false
	TabSize = 4
	TypeString.Timeout = 30 * time.Second
}
