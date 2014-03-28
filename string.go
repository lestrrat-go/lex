package lex

import (
  "fmt"
  "strings"
  "unicode/utf8"
)

// StringLexer is an implementation of Lexer interface, which lexes
// contents in a string
type StringLexer struct {
  input string
  inputLength int
  start int
  pos int
  width int
  items chan LexItem
  entryPoint LexFn
}

// NewStringLexer creates a new StringLexer instance. This lexer can be
// used only once per input string. Do not try to reuse it
func NewStringLexer(input string, fn LexFn) *StringLexer {
  return &StringLexer {
    input,
    len(input),
    0,
    0,
    0,
    make(chan LexItem, 1),
    fn,
  }
}

func (l *StringLexer) GetEntryPoint() LexFn {
  return l.entryPoint
}

func (l *StringLexer) inputLen() int {
  return l.inputLength
}

// Next returns the next rune
func (l *StringLexer) Next() (r rune) {
  if l.pos >= l.inputLen() {
    l.width = 0
    return EOF
  }

  r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
  l.pos += l.width
  return r
}

// Peek returns the next rune, but does not move the position
func (l *StringLexer) Peek() (r rune) {
  r = l.Next()
  l.Backup()
  return r
}

// Backup moves the cursor position (as many bytes as the last read rune)
func (l *StringLexer) Backup() {
  l.pos -= l.width
}

// Accept takes a string, and moves the cursor 1 rune if the rune is
// contained in the given string
func (l *StringLexer) Accept(valid string) bool {
  if strings.IndexRune(valid, l.Next()) >= 0 {
    return true
  }
  l.Backup()
  return false
}

// AcceptRun takes a string, and moves the cursor forward as long as 
// the input matches one of the given runes in the string
func (l *StringLexer) AcceptRun(valid string) bool {
  count := 0
  for strings.IndexRune(valid, l.Next()) >= 0 {
    count++
  }
  l.Backup()
  return count > 0
}

// EmitErrorf emits an Error Item
func (l *StringLexer) EmitErrorf(format string, args ...interface {}) LexFn {
  l.items <- NewLexItem(ItemError, l.pos, fmt.Sprintf(format, args...))
  return nil
}

// Grab creates a new LexItem of type `t`. The value in the item is created
// from the position of the last read item to current cursor position
func (l *StringLexer) Grab(t LexItemType) LexItem {
  return LexItem { t, l.start, l.BufferString() }
}

// Emit creates and sends a new LexItem of type `t` through the output
// channel. The LexItem is generated using `Grab`
func (l *StringLexer) Emit(t LexItemType) {
  l.items <-l.Grab(t)
  l.start = l.pos
}

// PrevByte returns the previous byte (l.Cursor - 1)
func (l *StringLexer) PrevByte() byte {
  return l.input[l.pos - 1]
}

// Cursor returns the current cursor position
func (l *StringLexer) Cursor() int {
  return l.pos
}

// LastCursor returns the end position of the last Grab
func (l *StringLexer) LastCursor() int {
  return l.start
}

// AdvanceCursor advances the cursor position by `n`
func (l *StringLexer) AdvanceCursor(n int) {
  l.pos += n
}

// BufferString reutrns the string beween LastCursor and Cursor
func (l *StringLexer) BufferString() string {
  return l.input[l.start:l.pos]
}

// ReaminingString returns the string starting at the current cursor
func (l *StringLexer) RemainingString() string {
  return l.input[l.pos:]
}

// Items returns the channel where lex'ed LexItem structs are sent to
func (l *StringLexer) Items() <-chan LexItem {
  return l.items
}

// NextItem returns the next LexItem in the processing pipeline.
// This is just a convenience function over reading l.Items()
func (l *StringLexer) NextItem() LexItem {
  return <-l.items
}

// Run starts the lexing. You should be calling this method as a goroutine:
//
//    lexer := lex.NewStringLexer(...)
//    go lexer.Run(nil)
//    for item := range lexer.Items() {
//      ...
//    }
//
func (l *StringLexer) Run(ctx interface {}) {
  for fn := l.GetEntryPoint(); fn != nil; {
    fn = fn(l, ctx)
  }
  close(l.items)
}

