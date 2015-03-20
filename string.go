package lex

import (
	"fmt"
	"unicode/utf8"
)

// StringLexer is an implementation of Lexer interface, which lexes
// contents in a string
type StringLexer struct {
	input       string
	inputLength int
	start       int
	pos         int
	line        int
	width       int
	items       chan LexItem
	entryPoint  LexFn
}

// NewStringLexer creates a new StringLexer instance. This lexer can be
// used only once per input string. Do not try to reuse it
func NewStringLexer(input string, fn LexFn) *StringLexer {
	return &StringLexer{
		input:       input,
		inputLength: len(input),
		start:       0,
		pos:         0,
		line:        1,
		width:       0,
		items:       make(chan LexItem, 1),
		entryPoint:  fn,
	}
}

// GetEntryPoint returns the function that lexing is started with
func (l *StringLexer) GetEntryPoint() LexFn {
	return l.entryPoint
}

func (l *StringLexer) inputLen() int {
	return l.inputLength
}

// Current returns the current rune being considered
func (l *StringLexer) Current() (r rune) {
	r, _ = utf8.DecodeRuneInString(l.input[l.pos:])
	return r
}

// Next returns the next rune
func (l *StringLexer) Next() (r rune) {
	if l.pos >= l.inputLen() {
		l.width = 0
		return EOF
	}

	// if the previous char was a new line, then we're at a new line
	if l.pos >= 0 {
		if l.input[l.pos] == '\n' {
			l.line++
		}
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
	if l.width == 1 && l.pos >= 0 && l.inputLen() > l.pos {
		if l.input[l.pos] == '\n' {
			l.line--
		}
	}
}

// AcceptString returns true if the given string can be matched exactly.
// This is a utility function to be called from concrete Lexer types
func (l *StringLexer) AcceptString(word string) bool {
	return AcceptString(l, word, false)
}

// PeekString returns true if the given string can be matched exactly,
// but does not move the position
func (l *StringLexer) PeekString(word string) bool {
	return AcceptString(l, word, true)
}

// AcceptAny takes a string, and moves the cursor 1 rune if the rune is
// contained in the given string
func (l *StringLexer) AcceptAny(valid string) bool {
	return AcceptAny(l, valid)
}

// AcceptRun takes a string, and moves the cursor forward as long as
// the input matches one of the given runes in the string
func (l *StringLexer) AcceptRun(valid string) bool {
	return AcceptRun(l, valid)
}

// EmitErrorf emits an Error Item
func (l *StringLexer) EmitErrorf(format string, args ...interface{}) LexFn {
	l.items <- NewItem(ItemError, l.pos, l.line, fmt.Sprintf(format, args...))
	return nil
}

// Grab creates a new Item of type `t`. The value in the item is created
// from the position of the last read item to current cursor position
func (l *StringLexer) Grab(t ItemType) Item {
	// special case
	str := l.BufferString()
	line := l.line
	if len(str) > 0 && str[0] == '\n' {
		line--
	}
	return NewItem(t, l.start, line, str)
}

// Emit creates and sends a new Item of type `t` through the output
// channel. The Item is generated using `Grab`
func (l *StringLexer) Emit(t ItemType) {
	l.items <- l.Grab(t)
	l.start = l.pos
}

// PrevByte returns the previous byte (l.Cursor - 1)
func (l *StringLexer) PrevByte() byte {
	return l.input[l.pos-1]
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

// RemainingString returns the string starting at the current cursor
func (l *StringLexer) RemainingString() string {
	return l.input[l.pos:]
}

// Items returns the channel where lex'ed Item structs are sent to
func (l *StringLexer) Items() chan LexItem {
	return l.items
}

// NextItem returns the next Item in the processing pipeline.
// This is just a convenience function over reading l.Items()
func (l *StringLexer) NextItem() LexItem {
	return <-l.items
}

// Run starts the lexing. You should be calling this method as a goroutine:
//
//    lexer := lex.NewStringLexer(...)
//    go lexer.Run()
//    for item := range lexer.Items() {
//      ...
//    }
//
func (l *StringLexer) Run() {
	LexRun(l)
}
