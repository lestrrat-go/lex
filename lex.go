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
)

// LexFn defines the lexing function.
type LexFn func(Lexer, interface {}) LexFn

// EOF is used to signal that we have reached EOF
const EOF = -1

// Lexer defines the interface for Lexers
type Lexer interface {
  Run(Lexer)
  GetEntryPoint() LexFn
  Next() rune
  Peek() rune
  Backup()
  Accept(string) bool
  AcceptRun(string) bool
  EmitErrorf(string, ...interface {}) LexFn
  Emit(ItemType)
  Items() chan LexItem
  NextItem() LexItem
}

// LexRun starts lexing using Lexer l, and a context Lexer ctx. "Context" in 
// this case can be thought as the concret lexer, and l is the parent class.
// This is a utility function to be called from concrete Lexer types
func LexRun(l, ctx Lexer) {
  for fn := ctx.GetEntryPoint(); fn != nil; {
    fn = fn(l, ctx)
  }
  close(l.Items())
}

// Accept takes a string, and moves the cursor 1 rune if the rune is
// contained in the given string. 
// This is a utility function to be called from concrete Lexer types
func Accept(l Lexer, valid string) bool {
  if strings.IndexRune(valid, l.Next()) >= 0 {
    return true
  }
  l.Backup()
  return false
}

// AcceptRun takes a string, and moves the cursor forward as long as 
// the input matches one of the given runes in the string
// This is a utility function to be called from concrete Lexer types
func AcceptRun(l Lexer, valid string) bool {
  count := 0
  for strings.IndexRune(valid, l.Next()) >= 0 {
    count++
  }
  l.Backup()
  return count > 0
}

