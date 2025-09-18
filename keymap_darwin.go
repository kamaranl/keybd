//go:build !(windows || linux)

package keymap

const KeyNotMapped = -1

const (
	Raw KeySet = iota
	Shift
	// TODO?: Alt
)

var GlobalOptions struct {
	TabsToSpaces bool // insert spaces instead of tabs
	TabSize      int  // number of spaces - only works when TabsToSpaces == true
}

var Whitespace = KeyMap{
	"\n": 0x24, // Newline
	"\t": 0x34, // Tab
	" ":  0x31, // Space
}

var ANSI = KeySetMap{
	Raw: KeyMap{
		"a": 0x00,
		"s": 0x01,
		"d": 0x02,
		"f": 0x03,
		"h": 0x04,
		"g": 0x05,
		"z": 0x06,
		"x": 0x07,
		"c": 0x08,
		"v": 0x09,
		"b": 0x0B,
		"q": 0x0C,
		"w": 0x0D,
		"e": 0x0E,
		"r": 0x0F,
		"y": 0x10,
		"t": 0x11,
		"1": 0x12,
		"2": 0x13,
		"3": 0x14,
		"4": 0x15,
		"6": 0x16,
		"5": 0x17,
		"=": 0x18,
		"9": 0x19,
		"7": 0x1A,
		"-": 0x1B,
		"8": 0x1C,
		"0": 0x1D,
		"]": 0x1E,
		"o": 0x1F,
		"u": 0x20,
		"[": 0x21,
		"i": 0x22,
		"p": 0x23,
		"l": 0x25,
		"j": 0x26,
		"'": 0x27,
		"k": 0x28,
		";": 0x29,
		`\`: 0x2A,
		",": 0x2B,
		"/": 0x2C,
		"n": 0x2D,
		"m": 0x2E,
		".": 0x2F,
		"`": 0x32,
	},
}

type KeySetMap = map[KeySet]KeyMap

type KeySet = int

type KeyMap = map[string]uint8

func CharToVKey(r rune) (int, bool) {
	ch := string(r)

	if key, ok := ANSI[Raw][ch]; ok {
		return int(key), false
	} else if key, ok := ANSI[Shift][ch]; ok {
		return int(key), true
	} else if key, ok := Whitespace[ch]; ok {
		return int(key), false
	}

	return KeyNotMapped, false
}

func setGlobalOptions() {
	GlobalOptions.TabsToSpaces = false
	GlobalOptions.TabSize = 4
}

func setShiftMappings() {
	ANSI[Shift] = KeyMap{
		"A": ANSI[Raw]["a"],
		"S": ANSI[Raw]["s"],
		"D": ANSI[Raw]["d"],
		"F": ANSI[Raw]["f"],
		"H": ANSI[Raw]["h"],
		"G": ANSI[Raw]["g"],
		"Z": ANSI[Raw]["z"],
		"X": ANSI[Raw]["x"],
		"C": ANSI[Raw]["c"],
		"V": ANSI[Raw]["v"],
		"B": ANSI[Raw]["b"],
		"Q": ANSI[Raw]["q"],
		"W": ANSI[Raw]["w"],
		"E": ANSI[Raw]["e"],
		"R": ANSI[Raw]["r"],
		"Y": ANSI[Raw]["y"],
		"T": ANSI[Raw]["t"],
		"!": ANSI[Raw]["1"],
		"@": ANSI[Raw]["2"],
		"#": ANSI[Raw]["3"],
		"$": ANSI[Raw]["4"],
		"^": ANSI[Raw]["6"],
		"%": ANSI[Raw]["5"],
		"+": ANSI[Raw]["="],
		"(": ANSI[Raw]["9"],
		"&": ANSI[Raw]["7"],
		"_": ANSI[Raw]["-"],
		"*": ANSI[Raw]["8"],
		")": ANSI[Raw]["0"],
		"}": ANSI[Raw]["]"],
		"O": ANSI[Raw]["o"],
		"U": ANSI[Raw]["u"],
		"{": ANSI[Raw]["["],
		"I": ANSI[Raw]["i"],
		"P": ANSI[Raw]["p"],
		"L": ANSI[Raw]["l"],
		"J": ANSI[Raw]["j"],
		`"`: ANSI[Raw]["'"],
		"K": ANSI[Raw]["k"],
		":": ANSI[Raw][";"],
		"|": ANSI[Raw][`\`],
		"<": ANSI[Raw][","],
		"?": ANSI[Raw]["/"],
		"N": ANSI[Raw]["n"],
		"M": ANSI[Raw]["m"],
		">": ANSI[Raw]["."],
		"~": ANSI[Raw]["`"],
	}
}

func init() {
	setGlobalOptions()
	setShiftMappings()
}
