/*

Package lex contains a lexer based on text/template from the main golang
distribution. I'ma big fan of how that parser works, have found that it
suits my brain better than other forms of tokenization.

I found myself cutting and pasting this code a lot, so I decided to cut it 
out as a generic piece of library so I don't have to keep doing it over and
over.

*/
package lex

// LexFn defines the lexing function.
type LexFn func(Lexer, interface {}) LexFn

const EOF = -1

// Lexer defines the interface for Lexers
type Lexer interface {
  Run(interface {})
  Next() rune
  Peek() rune
  Backup()
  Accept(string) bool
  AcceptRun(string) bool
  Emit(LexItemType)
  Items() <-chan LexItem
  NextItem() LexItem
  SetLexFn(string, LexFn)
  GetLexFn(string) (LexFn, error)
  MustGetLexFn(string) LexFn
}
