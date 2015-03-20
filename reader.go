package lex

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// ReaderLexer lexes input from an io.Reader instance
type ReaderLexer struct {
	source     *bufio.Reader
	start      int
	pos        int
	peekLoc    int
	line       int
	buf        []rune
	items      chan LexItem
	entryPoint LexFn
}

// NewReaderLexer creats a ReaderLexer
func NewReaderLexer(in io.Reader, fn LexFn) *ReaderLexer {
	return &ReaderLexer{
		bufio.NewReader(in),
		0,
		-1,
		-1,
		1,
		[]rune{},
		make(chan LexItem, 1),
		fn,
	}
}

// Current returns current rune being considered
func (l *ReaderLexer) Current() (r rune) {
	if len(l.buf) == 0 {
		return l.Next()
	}

	return l.buf[l.peekLoc]
}

// Next returns the next rune
func (l *ReaderLexer) Next() (r rune) {
	/* Illustrated guide to how the cursors move:

	              pos
	              |
	              v
	buf | a | b | c |
	              ^
	              |
	              peekLoc

	-> l.Nex() -> returns d

	                  pos
	                  |
	                  v
	buf | a | b | c | d |
	                  ^
	                  |
	                  peekLoc


	-> l.Peek() -> returns e

	                  pos
	                  |
	                  v
	buf | a | b | c | d | e |
	                      ^
	                      |
	                      peekLoc

	-> l.Backup()
	-> l.Backup()

	              pos
	              |
	              v
	buf | a | b | c | d | e |
	              ^
	              |
	              peekLoc

	-> l.Next() -> returns c

	                  pos
	                  |
	                  v
	buf | a | b | c | d | e |
	                  ^
	                  |
	                  peekLoc

	-> l.Next() -> returns d

	                      pos
	                      |
	                      v
	buf | a | b | c | d | e |
	                      ^
	                      |
	                      peekLoc
	*/
	guard := Mark("Next")
	defer func() {
		Trace("return = %q", r)
		guard()
	}()

	if l.pos == -1 || len(l.buf) == l.peekLoc+1 {
		r, _, err := l.source.ReadRune()
		switch err {
		case nil:
			l.peekLoc++
			l.pos++
			l.buf = append(l.buf, r)
		case io.EOF:
			l.peekLoc++
			l.pos++
			r = -1
			if l.buf[len(l.buf)-1] != r {
				l.buf = append(l.buf, r)
			}
		}
	} else {
		l.peekLoc++
		l.pos++
	}

	loc := l.peekLoc

	if loc < 0 || len(l.buf) <= loc {
		return EOF
		//    panic(fmt.Sprintf("FATAL: loc = %d, l.buf = %q (len = %d)", loc, l.buf, len(l.buf)))
	}

	Trace("l.buf = %q, loc = %d", l.buf, loc)
	return l.buf[loc]
}

// Peek returns the next rune, but does not move the position
func (l *ReaderLexer) Peek() (r rune) {
	guard := Mark("Peek")
	defer func() {
		Trace("return = %q", r)
		guard()
	}()

	r = l.Next()
	l.Backup()
	return r
}

// Backup moves the cursor 1 position
func (l *ReaderLexer) Backup() {
	guard := Mark("Backup")
	defer guard()

	l.pos--
	l.peekLoc = l.pos // align
	Trace("Backed up l.pos = %d", l.pos)
}

// AcceptString returns true if the given string can be matched exactly.
// This is a utility function to be called from concrete Lexer types
func (l *ReaderLexer) AcceptString(word string) bool {
	guard := Mark("AcceptString")
	defer guard()
	return AcceptString(l, word, false)
}

// PeekString returns true if the given string can be matched exactly,
// but does not move the position
func (l *ReaderLexer) PeekString(word string) bool {
	guard := Mark(fmt.Sprintf("PeekString '%s'", word))
	defer guard()
	return AcceptString(l, word, true)
}

// AcceptAny takes a string which contains a set of runes that can be accepted.
// This method moves the cursor 1 rune if the rune is contained in the given
// string.
func (l *ReaderLexer) AcceptAny(valid string) bool {
	return AcceptAny(l, valid)
}

// AcceptRun takes a string, and moves the cursor forward as long as
// the input matches one of the given runes in the string
func (l *ReaderLexer) AcceptRun(valid string) bool {
	guard := Mark("AcceptRun")
	defer guard()
	return AcceptRun(l, valid)
}

// Emit creates and sends a new Item of type `t` through the output
// channel. The Item is generated using `Grab`
func (l *ReaderLexer) Emit(t ItemType) {
	Trace("Emit %s", t)
	l.items <- l.Grab(t)
}

// EmitErrorf emits an Error Item
func (l *ReaderLexer) EmitErrorf(format string, args ...interface{}) LexFn {
	l.items <- NewItem(ItemError, l.pos, l.line, fmt.Sprintf(format, args...))
	return nil
}

// BufferString returns the current buffer
func (l *ReaderLexer) BufferString() (str string) {
	guard := Mark("BufferString")
	defer func() {
		Trace("BufferString() -> l.pos = %d, l.buf = %q, return = %q\n", l.pos, l.buf, str)
		guard()
	}()

	Trace("l.buf = %q, l.pos = %d", l.buf, l.pos)

	total := 0
	for i := 0; len(l.buf) > i && i < l.pos+1; i++ {
		Trace("l.buf[%d] = %q\n", i, l.buf[i])
		if l.buf[i] == -1 {
			break
		}
		total += utf8.RuneLen(l.buf[i])
	}

	Trace("Expecting buffer to contain %d bytes", total)
	if total == 0 {
		str = ""
		return
	}

	strbuf := make([]byte, total)
	pos := 0
	for i := 0; len(l.buf) > i && i < l.pos+1; i++ {
		if l.buf[i] == -1 {
			break
		}
		Trace("Encoding rune %q into position %d", l.buf[i], pos)
		pos += utf8.EncodeRune(strbuf[pos:], l.buf[i])
	}
	Trace("%q (%d)\n", strbuf, len(strbuf))

	str = string(strbuf)
	return str
}

// Grab creates a new Item of type `t`. The value in the item is created
// from the position of the last read item to current cursor position
func (l *ReaderLexer) Grab(t ItemType) Item {
	guard := Mark("Grab")
	defer guard()
	// special case
	line := l.line

	strbuf := l.BufferString()
	if strings.ContainsRune(strbuf, '\n') {
		l.line++
	}
	strlen := len(strbuf)

	item := NewItem(t, l.start, line, strbuf)
	l.buf = l.buf[utf8.RuneCountInString(strbuf):]
	l.peekLoc = l.peekLoc - l.pos - 1
	l.pos = -1
	l.start += strlen
	Trace("Emit item %#v", item)
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
//    go lexer.Run()
//    for item := range lexer.Items() {
//      ...
//    }
//
func (l *ReaderLexer) Run() {
	LexRun(l)
}
