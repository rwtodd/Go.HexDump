package main

import "strconv"

func dotChars(b byte) byte {
	ans := b
	if !strconv.IsGraphic(rune(b)) {
		ans = '.'
	}
	return ans
}
