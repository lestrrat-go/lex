package lex

import (
	"bytes"
	"testing"
)

const (
	ItemNumber = ItemDefaultMax + iota
	ItemOperator
	ItemWhitespace
)

func ExampleLexer() {
	c := &testLexCtx{}
	l := NewStringLexer("1 + 1", c.lexStart)
	go l.Run()

	for item := range l.Items() {
		// Do your processing here
		_ = item
	}
}

type testLexCtx struct {}
func (tlc *testLexCtx) lexStart(l Lexer) LexFn {
	guard := Mark("lexStart")
	defer guard()

	b := l.Peek()
	switch {
	case b == EOF:
		l.Emit(ItemEOF)
	case b >= 0x31 && b <= 0x39:
		return tlc.lexNumber
	case b == '+':
		return tlc.lexOperator
	case b == ' ' || b == 0x13 || b == 0x09 || b == 0x0a:
		return tlc.lexWhitespace
	default:
		l.EmitErrorf("Unexpected char: %q", b)
	}
	return nil
}

func (tlc *testLexCtx) lexWhitespace(l Lexer) LexFn {
	guard := Mark("lexWhitespace")
	defer guard()
	if l.AcceptString("\n") {
		l.Emit(ItemWhitespace)
		return tlc.lexStart
	}
	if l.AcceptRun(" \t\r") {
		l.Emit(ItemWhitespace)
		return tlc.lexStart
	}
	return l.EmitErrorf("Expected whitespace")
}

func (tlc *testLexCtx) lexOperator(l Lexer) LexFn {
	guard := Mark("lexOperator")
	defer guard()
	if l.AcceptString("+") {
		l.Emit(ItemOperator)
		return tlc.lexStart
	}

	return l.EmitErrorf("Expected operator")
}

func (tlc *testLexCtx) lexNumber(l Lexer) LexFn {
	guard := Mark("lexNumber")
	defer guard()
	if l.AcceptRun("0123456789") {
		l.Emit(ItemNumber)
		return tlc.lexStart
	}
	return l.EmitErrorf("Expected number")
}

func TestStringLexer(t *testing.T) {
	tlc := &testLexCtx{}
	l := NewStringLexer("1 +\n 2", tlc.lexStart)
	go l.Run()

	verify(t, l)
}

func TestLexer_AcceptString(t *testing.T) {
	var l Lexer
	var c LexItem

	t.Logf("-----> String")
	l = NewStringLexer("HELLO user", nil)
	if l.AcceptString("HELLONEARTH") {
		t.Errorf("Accepted HELLONEARTH?!")
	}
	if !l.AcceptString("HELLO") {
		t.Errorf("Failed to accept HELLO")
	}
	l.Emit(ItemOperator)
	c = <-l.Items()
	t.Logf("%#v", c)

	t.Logf("-----> Reader")
	l = NewReaderLexer(bytes.NewBufferString("HELLO user"), nil)
	if l.AcceptString("HELLONEARTH") {
		t.Errorf("Accepted HELLONEARTH?!")
	}
	if !l.AcceptString("HELLO") {
		t.Errorf("Failed to accept HELLO")
	}
	l.Emit(ItemOperator)
	c = <-l.Items()
	t.Logf("%#v", c)
}

func TestReaderLexer(t *testing.T) {
	tlc := &testLexCtx{}
	l := NewReaderLexer(bytes.NewBufferString("1 +\n 2"), tlc.lexStart)
	go l.Run()

	verify(t, l)
}

func verify(t *testing.T, l Lexer) {
	expectedItems := []Item{
		NewItem(ItemNumber, 0, 1, "1"),
		NewItem(ItemWhitespace, 1, 1, " "),
		NewItem(ItemOperator, 2, 1, "+"),
		NewItem(ItemWhitespace, 3, 1, "\n"),
		NewItem(ItemWhitespace, 4, 2, " "),
		NewItem(ItemNumber, 5, 2, "2"),
		NewItem(ItemEOF, 6, 2, ""),
	}

	i := 0
	for item := range l.Items() {
		t.Logf("----")
		if i >= len(expectedItems) {
			t.Fatalf("expected %d items, received more than that (%#v)", len(expectedItems), item)
		}
		t.Logf("got %#v", item)
		expected := expectedItems[i]
		if expected.Type() != item.Type() {
			t.Errorf("Type did not match: Expected %d, got %d", expected.Type(), item.Type())
		}

		if expected.Pos() != item.Pos() {
			t.Errorf("Pos did not match: Expected %d, got %d", expected.Pos(), item.Pos())
		}

		if expected.Line() != item.Line() {
			t.Errorf("Line did not match: Expected %d, got %d", expected.Line(), item.Line())
		}

		if expected.Value() != item.Value() {
			t.Errorf("Value did not match: Expected '%s', got '%s'", expected.Value(), item.Value())
		}
		i++
	}

	if i != len(expectedItems) {
		t.Errorf("Expected %d items, only got %d", len(expectedItems), i)
	}
}
