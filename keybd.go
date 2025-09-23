package keybd

import "time"

var Global struct {
	KeyDelay          time.Duration
	KeyPressDuration  time.Duration
	ModPressDuration  time.Duration
	MaxCharacters     int
	TabsToSpaces      bool
	TabSize           int
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
