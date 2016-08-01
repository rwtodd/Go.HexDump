
%{ 

package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

var escapes *strings.Replacer

func init() {
	escapes = strings.NewReplacer(`\n`, "\n", `\t`, "\t", `\r`, "\r", `\\`, "\\")
}

%}

%union{
  val   int
  str   string
  frag  formatter
  ser   serial
}

%type <val> number
%type <frag> start fragment
%type <ser> format

%token <val> DIGIT SLASH LOC
%token <str> STRING

%%

start : format 
	{
		yylex.(*formatLex).result = &($1)  // a hack, but it works
	}
	;

format : fragment 
	{
		$$ = serial{$1};
	}
	| format fragment
	{
		$$ = append($1, $2); 
	}
	;

fragment : number SLASH number STRING
	{
		$$ = newFmtString($1, $3, escapes.Replace($4))
	}
	| number STRING
	{
		$$ = newFmtString($1, 1, escapes.Replace($2)) 
	}
	| LOC STRING
	{
		tmp := locString(escapes.Replace($2))
		$$ = &tmp
	}
	| STRING
	{
		tmp := litString(escapes.Replace($1))
		$$ = &tmp
	}
	;

number : DIGIT
	{
		$$ = $1;
	}
	| number DIGIT
	{
		$$ = $1 * 10 + $2;
	}
	;

%%  /* start of program */

type formatLex struct {
	s 	string
	pos 	int
	result 	formatter
}

func (l *formatLex) skipWS() {
	orig := l.s
	l.s = strings.TrimLeftFunc(orig, unicode.IsSpace)	
	l.pos +=  ( len(orig) - len(l.s) )  // FIXME this won't be right for multibyte spaces
}

func (l *formatLex) readStr(terminator rune) string {
 	idx := strings.IndexRune(l.s, terminator)
	if idx < 0 {
		return "" // FIXME how do I declare a lex error?
	}
	ans := l.s[:idx]

	// we know both possible terminators are 1-byte, thus the +1's
	l.pos += utf8.RuneCountInString(ans) + 1
	l.s = l.s[idx+1:] 
	return ans
}

func (l *formatLex) Lex(lval *yySymType) int {
	l.skipWS()
	if len(l.s) == 0 {
		return 0
	}
	ch, wid := utf8.DecodeRuneInString(l.s) 
	l.s = l.s[wid:]

	switch {
	case unicode.IsDigit(ch):
		lval.val = int(ch) - '0' 
		return DIGIT	
	case ch == '/':
		lval.val = int(ch)
		return SLASH
	case ch == '@':
		lval.val = int(ch)
		return LOC 
	case ch == '"', ch == '\'':
		lval.str = l.readStr(ch) 	
		return STRING
	}	

	return int(ch)  // this shouldn't really happen...
}


func (l *formatLex) Error(s string) {
	fmt.Printf("syntax error: %s (last pos was %d)\n", s, l.pos)
}

// helper func
func parseFormat(s string) formatter {
	lxr := &formatLex{s: s}
	if yyParse(lxr) > 0 {
		return nil
	}
	return lxr.result
}
