// Copyright 2016 Richard Todd
// GPL v2... see license in repo

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
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
// Canonical Formatter
// ===================
type canonical struct{}

func (c canonical) bytesNeeded() int {
	return 16
}

func (c canonical) format(loc uint64, bytes []byte) string {
	hexbuf := make([]string, 0, 16)
	chbuf := make([]byte, 0, 16)

	for idx, b := range bytes {
		hexbuf = append(hexbuf, fmt.Sprintf("%02X ", b))
		if idx == 7 {
			hexbuf[7] += " " // space in the middle
		}
		if !strconv.IsGraphic(rune(b)) {
			b = '.'
		}
		chbuf = append(chbuf, b)
	}
	hexpart := strings.Join(hexbuf, "")
	chpart := string(chbuf)

	return fmt.Sprintf("%08X:  %-49s |%-16s|\n", loc, hexpart, chpart)
}

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
	var masterFormat canonical
	if err = engine(fl, masterFormat); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}
