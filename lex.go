/*

Package lex contains a lexer based on text/template from the main golang
distribution. I'ma big fan of how that parser works, have found that it
suits my brain better than other forms of tokenization.

I found myself cutting and pasting this code a lot, so I decided to cut it
out as a generic piece of library so I don't have to keep doing it over and
over.

*/
package lex

import (
	"strings"
	"unicode/utf8"
)

// LexFn defines the lexing function. It takes the lexer (i.e. StringLexer
// or ReaderLexer) as its argument. If you have no state, you can just
// use a regular functions. Otherwise, use an object and a method bound
// to that object:
// 
//     type Foo strcut { ... }
//     func (f *Foo) lexFoo(l lex.Lexer) lex.LexFn {
//       ...
//     }
//     
//     src := "...."
//     f := &Foo{}
//     l := lex.NewStringReader(src, f.lexFoo)
//     l.Run()
//     
type LexFn func(Lexer) LexFn

// EOF is used to signal that we have reached EOF
const EOF = -1

// Lexer defines the interface for Lexers
type Lexer interface {
	Run()
	GetEntryPoint() LexFn
	Current() rune
	Next() rune
	Peek() rune
	Backup()
	PeekString(string) bool
	AcceptString(string) bool
	AcceptRun(string) bool
	AcceptRunFunc(func(r rune) bool) bool
	AcceptRunExcept(string) bool
	EmitErrorf(string, ...interface{}) LexFn
	Emit(ItemType)
	Items() chan LexItem
	BufferString() string
	NextItem() LexItem
}

// LexRun starts lexing using Lexer l, and a context Lexer ctx. "Context" in
// this case can be thought as the concret lexer, and l is the parent class.
// This is a utility function to be called from concrete Lexer types
func LexRun(l Lexer) {
	for fn := l.GetEntryPoint(); fn != nil; {
		fn = fn(l)
	}
	close(l.Items())
}

// This method moves the cursor 1 rune if the rune is contained in the given
// string. This is a utility function to be called from concrete Lexer types
func AcceptAny(l Lexer, valid string) bool {
	if strings.IndexRune(valid, l.Next()) >= 0 {
		return true
	}
	l.Backup()
	return false
}

func AcceptRunFunc(l Lexer, fn func(rune) bool) bool {
	guard := Mark("lex.AcceptRunFunc")
	defer guard()

	count := 0
	for {
		n := l.Next()
		Trace("%d: n -> %q\n", count, n)
		if !fn(n) {
			break
		}

		count++
	}
	l.Backup()
	Trace("%d matches\n", count)
	return count > 0
}

// AcceptRun takes a string, and moves the cursor forward as long as
// the input matches one of the given runes in the string
// This is a utility function to be called from concrete Lexer types
func AcceptRun(l Lexer, valid string) bool {
	guard := Mark("lex.AcceptRun %q", valid)
	defer guard()
	return AcceptRunFunc(l, func(r rune) bool {
		return strings.IndexRune(valid, r) >= 0
	})
}

// AcceptRunExcept takes a string, and moves the cursor forward as 
// long as the input DOES NOT match one of the given runes in the string
// This is a utility function to be called from concrete Lexer types
func AcceptRunExcept(l Lexer, valid string) bool {
	guard := Mark("lex.AcceptRunExcept %q", valid)
	defer guard()
	return AcceptRunFunc(l, func(r rune) bool {
		return strings.IndexRune(valid, r) < 0
	})
}

// AcceptString returns true if the given string can be matched exactly.
// This is a utility function to be called from concrete Lexer types
func AcceptString(l Lexer, word string, rewind bool) (ok bool) {
	i := 0
	defer func() {
		if rewind {
			Trace("Rewinding AccepString(%q) (%d runes)\n", word, i)
			for j := i; j > 0; j-- {
				l.Backup()
			}
		}
		Trace("AcceptString returning %s\n", ok)
	}()

	for pos := 0; pos < len(word); {
		r, width := utf8.DecodeRuneInString(word[pos:])
		pos += width
		var n rune
		if pos == 0 {
			n = l.Current()
		} else {
			n = l.Next()
		}
		i++
		Trace("r (%q) == n (%q) %s ? \n", r, n, r == n)
		if r != n {
			rewind = true
			ok = false
			return
		}
	}
	ok = true
	return
}
