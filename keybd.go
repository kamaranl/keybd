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

// KeyPressDuration is how long to wait after pressing a key before releasing
// that same key.
//
// Default: 2 ms
var KeyPressDuration time.Duration

// TypeString is a struct that contains specific settings for [TypeStr].
var TypeString struct {
	// KeyDelay is how long to wait after releasing a key and before proceeding
	// with pressing the next key.
	//
	// Default: 2 ms
	KeyDelay time.Duration

	// ModPressDuration is how long to wait after pressing a modifier key and
	// before releasing the same modifier key.
	//
	// Default: 2 ms
	ModPressDuration time.Duration

	// MaxCharacters is the maximum amount of characters in a string that can be
	// processed.
	//
	// Default: 5000
	MaxCharacters int

	// TabsToSpaces is a switch to enable the conversion of tabs to spaces as
	// they are typed.
	//
	// Default: false
	TabsToSpaces bool

	// TabSize is the number of spaces to use in place of tabs when TabsToSpaces
	// is true.
	//
	// Default: 4
	TabSize int

	// Timeout is how long [TypeStr] can run before aborting.
	//
	// Default: 30 s
	Timeout time.Duration

	abort chan struct{}
	mu    sync.Mutex
}

// AbortTypeStr safely aborts any previous calls to [TypeStr].
func AbortTypeStr() {
	TypeString.mu.Lock()
	defer TypeString.mu.Unlock()

	if TypeString.abort == nil {
		return
	}

	select {
	case <-TypeString.abort:
		return
	default:
		close(TypeString.abort)
	}
}

func init() {
	KeyPressDuration = 2 * time.Millisecond
	TypeString.KeyDelay = 2 * time.Millisecond
	TypeString.ModPressDuration = 2 * time.Millisecond
	TypeString.MaxCharacters = 5000
	TypeString.TabsToSpaces = false
	TypeString.TabSize = 4
	TypeString.Timeout = 30 * time.Second
}
