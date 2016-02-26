# hexdump
An implementation of hexdump in Go

Like my implementation of sed, I'm really just writing this for fun, and to make my life 
easier on random non-cygwin Windows machines.  It also gives me a chance to really learn
the nooks and crannies of these tools.  For instance, investigating this implementation
I learned a LOT about the types of fancy formats the linux hexdump utility lets you
produce.

## Status

I just threw together the skeleton of the program so far.  It accepts:

  * __-s__: skip ahead in the file
  * __-n__: only read so much of the file

... and for formats you have any combination of:

  * __-C__: "Canonical" format: 16 hex bytes followed by the same bytes as characters.
  * __-x__: 8 16-bit words in hex
  * __-x4__: 4 32-bit dwords in hex

... the main thing it's missing is user-defined (`-e`) formats.  That will come eventually.

## Implementation Notes

I developed the framework with user-defined formatting in mind.  Everything is driven from 
the `formatter` interface:

```
type formatter interface {
	bytesNeeded() int
	format(loc uint64, bytes []byte) string
}
```

There are "parallel" formatters that give their bytes to all sub-formats.  The master format is one of these... each format
the user selects gets to format all the bytes.  Then there are "series" formats, which spread their bytes across their sub-formats.
This way, you can format 8 bytes, then a spacer, then 8 more bytes (it would be a 3-part series of bytes, literal string, and bytes).
Series and parallel slices of formats can be arbitrarily nested.

The canned formats offered by the command-line switches are just _pre-parsed_ formats.  The only thing missing is for me to write 
a lexer and parser for `-e` arguments and turn them into instances of `formatter`.  

So, I don't know when or if I will get around to filling this functionality out, but the
hooks are in place to make it go smoothly. 
