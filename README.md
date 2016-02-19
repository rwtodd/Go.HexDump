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

... and it always just gives the canonical (-C) format for now.

## Implementation Notes

Though it only gives "canoncial" formatting right now, I developed the framework
with user-defined formatting in mind.  Everything is driven from the `formatter` 
interface:

```
type formatter interface {
	bytesNeeded() int
	format(loc uint64, bytes []byte) string
}
```

So, ultimately, the plan is for all user-defined format directives follow the `formatter` interface, and plug
them into the engine I've already put in place.  You can imagine a type representing aggregates of these formats, 
which adds up the `bytesNeeded` into the total that will be needed for the entire line.  Of course there will
also be an aggregate type to combine lines, which asks for the maximum number of bytes any of the lines
require.

So, I don't know when or if I will get around to filling this functionality out, but the
hooks are in place to make it go smoothly. 
