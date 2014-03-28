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
type LexFn func(Lexer) LexFn
// LexItemType describes the type of a LexItem
type LexItemType int

const eof = -1
const (
  // ItemEOF is emiteed upon EOF
  ItemEOF LexItemType = iota
  // ItemError is emitted upon Error
  ItemError
  // ItemDefaultMax is used as marker for your own ItemType. 
  // Start your types from this + 1
  ItemDefaultMax
)

// LexItem is the struct that gets generated upon finding *something*
type LexItem struct {
  typ LexItemType
  pos int
  val string
}

// NewLexItem creates a new LexItem
func NewLexItem(t LexItemType, pos int, v string) LexItem {
  return LexItem { t, pos, v }
}

// Type returns the associated LexItemType
func (l LexItem) Type() LexItemType {
  return l.typ
}

// Pos returns the associated position
func (l LexItem) Pos() int {
  return l.pos
}

// Value returns the associated text value
func (l LexItem) Value() string {
  return l.val
}

// Lexer defines the interface for Lexers
type Lexer interface {
  Run()
  Next() rune
  Peek() rune
  Backup()
  AcceptRun(string) bool
  Emit(LexItemType)
  Items() <-chan LexItem
  NextItem() LexItem
  SetLexFn(string, LexFn)
  GetLexFn(string) (LexFn, error)
  MustGetLexFn(string) LexFn
}
