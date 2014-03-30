package lex

import (
  "bufio"
  "io"
  "fmt"
  "unicode/utf8"
)

// ReaderLexer lexes input from an io.Reader instance
type ReaderLexer struct {
  source      *bufio.Reader
  start       int
  pos         int
  line        int
  peekCount   int
  buf         []rune
  items       chan LexItem
  entryPoint  LexFn
}

// NewReaderLexer creats a ReaderLexer
func NewReaderLexer(in io.Reader, fn LexFn) *ReaderLexer {
  return &ReaderLexer {
    bufio.NewReader(in),
    0,
    0,
    1,
    -1,
    []rune {},
    make(chan LexItem, 1),
    fn,
  }
}

// Next returns the next rune
func (l *ReaderLexer) Next() (r rune) {
  if l.peekCount == len(l.buf) - 1 {
    r, _, err := l.source.ReadRune()
    if err == nil {
      l.peekCount++
      l.buf = append(l.buf, r)
    } else if err == io.EOF {
      l.peekCount++
      l.buf = append(l.buf, -1)
    } else {
      panic(err.Error())
    }
  } else {
    l.peekCount = 0
  }
  l.pos++

  if l.buf[l.peekCount] == '\n' {
    l.line++
  }
  return l.buf[l.peekCount]
}

// Peek returns the next rune, but does not move the position
func (l *ReaderLexer) Peek() (r rune) {
  r = l.Next()
  l.Backup()
  return r
}

// Backup moves the cursor 1 position 
func (l *ReaderLexer) Backup() {
  if l.peekCount >= 0 && len(l.buf) > l.peekCount && l.buf[l.peekCount] == '\n' {
    l.line--
  }
  l.peekCount--
  l.pos--
}

// Accept takes a string, and moves the cursor 1 rune if the rune is
// contained in the given string
func (l *ReaderLexer) Accept(valid string) bool {
  return Accept(l, valid)
}

// AcceptRun takes a string, and moves the cursor forward as long as 
// the input matches one of the given runes in the string
func (l *ReaderLexer) AcceptRun(valid string) bool {
  return AcceptRun(l, valid)
}

// Emit creates and sends a new Item of type `t` through the output
// channel. The Item is generated using `Grab`
func (l *ReaderLexer) Emit(t ItemType) {
  l.items <- l.Grab(t)
}

// EmitErrorf emits an Error Item
func (l *ReaderLexer) EmitErrorf(format string, args ...interface {}) LexFn {
  l.items <- NewItem(ItemError, l.pos, l.line, fmt.Sprintf(format, args...))
  return nil
}

// Grab creates a new Item of type `t`. The value in the item is created
// from the position of the last read item to current cursor position
func (l *ReaderLexer) Grab(t ItemType) Item {
  // special case
  line := l.line
  if len(l.buf) > 0 && l.buf[0] == '\n' {
    line--
  }

  total := 0
  for i := 0; i < l.pos; i++ {
    total += utf8.RuneLen(l.buf[i])
  }
  strbuf := make([]byte, total)
  pos := 0
  for i := 0; i < l.pos; i++ {
    pos += utf8.EncodeRune(strbuf[pos:], l.buf[i])
  }

  item := NewItem( t, l.start, line, string(strbuf) )
  l.buf = l.buf[l.pos:]
  if len(l.buf) > 0 {
    l.peekCount = l.peekCount - l.pos
  } else {
    l.peekCount = -1
  }
  l.pos = 0
  l.start += len(strbuf)
  return item
}

// GetEntryPoint returns the function that lexing is started with
func (l *ReaderLexer) GetEntryPoint() LexFn {
  return l.entryPoint
}

// Items returns the channel where lex'ed Item structs are sent to
func (l *ReaderLexer) Items() chan LexItem {
  return l.items
}

// NextItem returns the next Item in the processing pipeline.
// This is just a convenience function over reading l.Items()
func (l *ReaderLexer) NextItem() LexItem {
  return <-l.items
}

// Run starts the lexing. You should be calling this method as a goroutine:
//
//    lexer := lex.NewStringLexer(...)
//    go lexer.Run(lexer)
//    for item := range lexer.Items() {
//      ...
//    }
//
func (l *ReaderLexer) Run(ctx Lexer) {
  LexRun(l, ctx)
}



