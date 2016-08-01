# hexdump
An implementation of hexdump in Go

Like my implementation of sed, I'm really just writing this for fun, and to make my life 
easier on random non-cygwin Windows machines.  It also gives me a chance to really learn
the nooks and crannies of these tools.  For instance, investigating this implementation
I learned a LOT about the types of fancy formats the linux hexdump utility lets you
produce.

## Status

I now consider the program "finished"... it does all I need from it.  A couple features 
of the GNU hexdump are missing, and may get implemented one day just for
completeness.  Here are the command-line switches supported:

  * __-s__: skip ahead in the file
  * __-n__: only read so much of the file
  * __-C__: set the canned "canonical" hexdump format, with 16 hex bytes followed by those
same bytes as characters.
  * __-x__: 8 16-bit words in hex (little-endian)
  * __-x4__: 4 32-bit dwords in hex (little-endian)
  * __-e__: define a custom format (see below)


## Custom Format Strings

The format strings supported by hexdump-go are the same as the ones you'd give to the 
GNU tool, with the following exceptions:

  * The `%_p` special format is supported, but I haven't implemented `%_c` or `%_u` yet, since I don't use those.
  * Instead of using special format `%_a` for the location, use a `@` (at-sign) before the format string. 
  * I didn't implement the '_A' "only print once at the end" version of location, since I never use that option
personally.
  * Since I use the tool on windows, where one uses double-quotes to surround an argument, I support either single
or double quotes inside the format strings.  So on UNIX you'd say `-e ' 16/2 "%04X "'` and on Windows you'd say
`-e  " 16/2 '%04X '"`, and it works both ways.

_Example custom format_: two lines for each set of 16 bytes... the top line in octal, and the bottom line in
characters (special characters replaced by '.'):

    > hexdump-go -e "@ '%08X: ' 16 '%03o ' '\n'" -e "'\t  ' 16 '%-4_p' '\n'" .gitignore 
    00000000: 043 040 103 157 155 160 151 154 145 144 040 117 142 152 145 143
              #       C   o   m   p   i   l   e   d       O   b   j   e   c
    00000010: 164 040 146 151 154 145 163 054 040 123 164 141 164 151 143 040
              t       f   i   l   e   s   ,       S   t   a   t   i   c
    00000020: 141 156 144 040 104 171 156 141 155 151 143 040 154 151 142 163
              a   n   d       D   y   n   a   m   i   c       l   i   b   s
    00000030: 040 050 123 150 141 162 145 144 040 117 142 152 145 143 164 163
                  (   S   h   a   r   e   d       O   b   j   e   c   t   s
    00000040: 051 012 052 056 157 012 052 056 141 012 052 056 163 157 012 012
              )   .   *   .   o   .   *   .   a   .   *   .   s   o   .   .

(note the example shows how to use the `@` indicator to mark the format string for the 
current location.  I think it's better than the %_a marker, personally, though there
is room to disagree.)

## Implementation Notes

I developed the framework with user-defined formatting in mind.  Everything is driven from 
the `formatter` interface:

```
type formatter interface {
	bytesNeeded() int
	format(loc uint64, bytes []byte) string
}
```

Composite formats come in "parallel" and "series" varieties.  When you define a format with `-e`, you are making a serial
list of formats -- for example: a location formatter, then a 16-byte double-word formatter.  Each of those components 
takes some bytes and gives what's left to its neighbors in the series.  All of those `-e` series formats are combined
into a single parallel format, so that they each get to read all of the bytes.  
Series and parallel slices of formats can be arbitrarily nested.

A parser (written with `go tool yacc`) turns string formats into instances of `formatter`.  All of the canned formats (`-x`, `-C`, etc.) are stored as format strings in `main.go` exactly as they could have been given via the `-e` flag on the command
line, and they get run through the same parser that `-e` formats use.

