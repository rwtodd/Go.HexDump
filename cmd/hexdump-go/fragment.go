package main

import (
	"fmt"
	"regexp"
	"strings"
)

// fragments of formats...

// L I T E R A L  S T R I N G ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
type litString string

func (ls *litString) bytesNeeded() int {
	return 0
}

func (ls *litString) format(loc uint64, bytes []byte) string {
	return string(*ls)
}

// L O C A T I O N   S T R I N G ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
type locString string

func (ls *locString) bytesNeeded() int {
	return 0
}

func (ls *locString) format(loc uint64, bytes []byte) string {
	return fmt.Sprintf(string(*ls), loc)
}

// F O R M A T  S T R I N G ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
type fmtString struct {
	repeat int             // number of times to repeat
	size   int             // number of bytes to feed it
	str    string          // format string
	explen int             // expected length
	conv   func(byte) byte // conversion function
}

var p_format *regexp.Regexp // used by newFmtString to check for %_p format

func init() {
	p_format = regexp.MustCompile(`%( ?[+-]?[\d.]*)_p`)
}

func newFmtString(rpt int, sz int, s string) *fmtString {
	ans := &fmtString{repeat: rpt, size: sz, str: s}

	// now check for the _p special format...
	if p_format.MatchString(s) {
		ans.conv = dotChars
		ans.str = p_format.ReplaceAllString(s, "%${1}c")
	}

	// TODO, if I feel like it one day:
	//  support _u and _c special formats as well

	return ans
}

func (fs *fmtString) bytesNeeded() int {
	return fs.repeat * fs.size
}

func format1b(fs *fmtString, bytes []byte) string {
	var ans = make([]string, 0, fs.repeat)

	var mx = len(bytes)
	if mx > fs.repeat {
		mx = fs.repeat
	}
	for idx := 0; idx < mx; idx++ {
		b := bytes[idx]
		if fs.conv != nil {
			b = fs.conv(b)
		}
		ans = append(ans, fmt.Sprintf(fs.str, b))
	}

	if fs.explen == 0 && mx > 0 {
		fs.explen = len(ans[0])
	}

	if mx < fs.repeat {
		ans = append(ans, strings.Repeat(" ", (fs.repeat-mx)*fs.explen))
	}
	return strings.Join(ans, "")
}

func format2b(fs *fmtString, bytes []byte) string {
	// first convert to 2-byte ints ...
	var words = make([]uint16, 0, fs.repeat)

	var blen = len(bytes)
	for idx := 0; idx < blen; {
		var word = uint16(bytes[idx])
		idx++
		if idx < blen {
			word |= uint16(bytes[idx]) << 8
		}
		idx++

		words = append(words, word)
	}

	// now, run across the ints like we did bytes in format1b
	var ans = make([]string, 0, fs.repeat)

	var mx = len(words)
	if mx > fs.repeat {
		mx = fs.repeat
	}
	for idx := 0; idx < mx; idx++ {
		// FIXME RWT do I need to support non-byte conversions?  I think not...
		ans = append(ans, fmt.Sprintf(fs.str, words[idx]))
	}

	if fs.explen == 0 && mx > 0 {
		fs.explen = len(ans[0])
	}

	if mx < fs.repeat {
		ans = append(ans, strings.Repeat(" ", (fs.repeat-mx)*fs.explen))
	}
	return strings.Join(ans, "")
}

func format4b(fs *fmtString, bytes []byte) string {
	// first convert to 2-byte ints ...
	var dwords = make([]uint32, 0, fs.repeat)

	var blen = len(bytes)
	for idx := 0; idx < blen; {
		var dword = uint32(bytes[idx])
		idx++
		if idx < blen {
			dword |= uint32(bytes[idx]) << 8
		}
		idx++
		if idx < blen {
			dword |= uint32(bytes[idx]) << 16
		}
		idx++
		if idx < blen {
			dword |= uint32(bytes[idx]) << 24
		}
		idx++

		dwords = append(dwords, dword)
	}

	// now, run across the ints like we did bytes in format1b
	var ans = make([]string, 0, fs.repeat)

	var mx = len(dwords)
	if mx > fs.repeat {
		mx = fs.repeat
	}
	for idx := 0; idx < mx; idx++ {
		// FIXME RWT do I need to support non-byte conversions?  I think not...
		ans = append(ans, fmt.Sprintf(fs.str, dwords[idx]))
	}

	if fs.explen == 0 && mx > 0 {
		fs.explen = len(ans[0])
	}

	if mx < fs.repeat {
		ans = append(ans, strings.Repeat(" ", (fs.repeat-mx)*fs.explen))
	}
	return strings.Join(ans, "")
}

func (fs *fmtString) format(loc uint64, bytes []byte) string {
	var ans string

	switch fs.size {
	case 1:
		ans = format1b(fs, bytes)
	case 2:
		ans = format2b(fs, bytes)
	case 4:
		ans = format4b(fs, bytes)
	default:
		ans = "ERROR!!!!!" // RWT FIXME do something better with that
	}

	return ans
}
