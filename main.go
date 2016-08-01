// Copyright 2016 Richard Todd
// GPL v2... see license in repo

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

//go:generate go tool yacc parser.y

// Flags...
var offset = flag.Uint64("s", 0, "bytes to skip before starting")
var lenToDo = flag.Uint64("n", 1<<64-1, "max number of bytes to interpret")

// we will be building up formatters
type formatter interface {
	bytesNeeded() int
	format(loc uint64, bytes []byte) string
}

func engine(in io.Reader, f formatter) error {
	var totalRead uint64

	need := f.bytesNeeded()
	buffer := make([]byte, need)
	for {
		got, err := io.ReadFull(in, buffer)
		location := totalRead + *offset
		var line string
		switch err {
		case nil:
			line = f.format(location, buffer)
		case io.ErrUnexpectedEOF:
			line = f.format(location, buffer[:got])
		case io.EOF:
			return nil
		default:
			return err
		}
		os.Stdout.WriteString(line)
		totalRead += uint64(got)
		if totalRead > *lenToDo {
			break
		}
	}

	return nil
}

// ===================
// Serial Formatter
// ===================
type serial []formatter

func (s *serial) bytesNeeded() int {
	ans := 0
	for _, v := range *s {
		ans += v.bytesNeeded()
	}
	return ans
}

func (s *serial) format(loc uint64, bytes []byte) string {
	var lines = make([]string, 0, len(*s))
	for _, v := range *s {
		lines = append(lines, v.format(loc, bytes))
		bn := v.bytesNeeded()
		lb := len(bytes)

		loc += uint64(bn)

		if bn < lb {
			bytes = bytes[bn:]
		} else {
			if lb > 0 {
				bytes = make([]byte, 0)
			}
		}
	}
	return strings.Join(lines, "")
}

// ===================
// Parallel Formatter
// ===================
type parallel []formatter

// global master formatter
var masterFormat parallel

func (p *parallel) bytesNeeded() int {
	ans := 0
	for _, v := range *p {
		bn := v.bytesNeeded()
		if bn > ans {
			ans = bn
		}
	}
	return ans
}

func (p *parallel) format(loc uint64, bytes []byte) string {
	var lines = make([]string, 0, len(*p))
	for _, v := range *p {
		lines = append(lines, v.format(loc, bytes))
	}
	return strings.Join(lines, "")
}

// ===================
// Canonical Formatter
// ===================
func addCanonical(format *parallel, _ string) {
	addFromString(format, "@ '%08x: ' 8 '%02X ' ' ' 8 '%02X '")

	// handle the char part ....
	// TODO: addFromString(format, "'|' 16 '%_c' '|\n'")
	var canonC serial
	bar := litString("|")
	fmtC := &fmtString{repeat: 16, size: 1, str: "%c", explen: 1, conv: dotChars}
	bar2 := litString("|\n")
	canonC = append(canonC, &bar, fmtC, &bar2)
	*format = append(*format, &canonC)
}

func add2Hex(format *parallel, _ string) {
	addFromString(format, "@ '%08x: ' 4/2 '%04X  ' ' ' 4/2 '%04X  ' '\n'")
}

func add4Hex(format *parallel, _ string) {
	addFromString(format, "@ '%08x: ' 2/4 '%08X    ' ' ' 2/4 '%08X    ' '\n'")
}

func addFromString(format *parallel, arg string) {
	if parsed := parseFormat(arg); parsed != nil {
		*format = append(*format, parsed)
	}
}

// =================
// Format setter -- command line argument that sets a format
// =================
type formatSetter struct {
	setter func(p *parallel, arg string)
	text   string
}

func (fs *formatSetter) String() string {
	return fs.text
}

func (fs *formatSetter) Set(arg string) error {
	fs.setter(&masterFormat, arg)
	return nil
}

func (fs *formatSetter) IsBoolFlag() bool { return true }

func init() {
	flag.Var(&formatSetter{setter: add2Hex, text: "off"}, "x", "format rows of 8 2-byte hex values")
	flag.Var(&formatSetter{setter: add4Hex, text: "off"}, "x4", "format rows of 4 4-byte hex values")
	flag.Var(&formatSetter{setter: addCanonical, text: "on"}, "C", "Canonical mode: 16 hex bytes with characters to the side")
	flag.Var(&formatSetter{setter: addFromString, text: "off"}, "f", "format via format string")
}

// ==================
// Main
// ==================
func main() {
	flag.Parse()
	var fl *os.File = os.Stdin
	var err error

	// open the file if there is one
	if len(flag.Args()) > 0 {
		fl, err = os.Open(flag.Args()[0])

		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't open file! %s", err.Error())
			os.Exit(1)
		}
		defer fl.Close()
	}

	// seek ahead if the cmdline told us to
	if *offset > 0 {
		_, err = fl.Seek(int64(*offset), 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't seek! %s", err.Error())
			os.Exit(1)
		}
	}

	// if no formats were specified, assume canonical
	if len(masterFormat) == 0 {
		addCanonical(&masterFormat, "")
	}

	// format the output
	if err = engine(fl, &masterFormat); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}
