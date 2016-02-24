package main

import (
	"fmt"
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

func (fs *fmtString) bytesNeeded() int {
	return fs.repeat * fs.size
}

func (fs *fmtString) format(loc uint64, bytes []byte) string {
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

	if mx < fs.repeat {
		ans = append(ans, strings.Repeat(" ", (fs.repeat-mx)*fs.explen))
	}
	return strings.Join(ans, "")
}
