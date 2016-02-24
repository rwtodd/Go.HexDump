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
func addCanonical(format *parallel) {
	// handle the hex part.....
	var canon serial
	loc := locString("%08x: ")
	fmt8 := &fmtString{repeat: 8, size: 1, str: "%02X ", explen: 3, conv: nil}
	spacer := litString(" ")
	canon = append(canon, &loc, fmt8, &spacer, fmt8)
	*format = append(*format, &canon)

	// handle the char part ....
	var canonC serial
	bar := litString("|")
	fmtC := &fmtString{repeat: 16, size: 1, str: "%c", explen: 1, conv: dotChars}
	bar2 := litString("|\n")
	canonC = append(canonC, &bar, fmtC, &bar2)
	*format = append(*format, &canonC)
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

	// format the output
	var allFormats parallel
	addCanonical(&allFormats)

	if err = engine(fl, &allFormats); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}
