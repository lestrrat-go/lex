package lex

import (
  "testing"
)

const (
  ItemNumber = ItemDefaultMax + iota
  ItemOperator
  ItemWhitespace
)

func lexStart(l Lexer, ctx interface {}) LexFn {
  b := l.Peek()
  switch {
  case b == EOF:
    l.Emit(ItemEOF)
  case b >= 0x31 && b <= 0x39:
    return lexNumber
  case b == '+':
    return lexOperator
  case b == ' ' || b == 0x0a || b == 0x13 || b == 0x09:
    return lexWhitespace
  default:
    l.EmitErrorf("Unexpected char: %q", b)
  }
  return nil
}

func lexWhitespace(l Lexer, ctx interface {}) LexFn {
  if l.AcceptRun(" \t\r\n") {
    l.Emit(ItemWhitespace)
    return lexStart
  }
 return l.EmitErrorf("Expected whitespace")
}

func lexOperator(l Lexer, ctx interface {}) LexFn {
  if l.Accept("+") {
    l.Emit(ItemOperator)
    return lexStart
  }

  return l.EmitErrorf("Expected operator")
}

func lexNumber(l Lexer, ctx interface {}) LexFn {
  if l.AcceptRun("0123456789") {
    l.Emit(ItemNumber)
    return lexStart
  }
  return l.EmitErrorf("Expected number")
}

func TestLex(t *testing.T) {
  l := NewStringLexer("1 + 2")
  go l.Run(l, lexStart)

  expectedItems := []LexItem {
    NewLexItem( ItemNumber, 0, "1" ),
    NewLexItem( ItemWhitespace, 1, " "),
    NewLexItem( ItemOperator, 2, "+" ),
    NewLexItem( ItemWhitespace, 3, " "),
    NewLexItem( ItemNumber, 4, "2"),
    NewLexItem( ItemEOF, 5, "" ),
  }

  i := 0
  for item := range l.Items() {
    expected := expectedItems[i]
    if expected.Type() != item.Type() {
      t.Errorf("Type did not match: Expected %d, got %d", expected.Type(), item.Type())
    }

    if expected.Pos() != item.Pos() {
      t.Errorf("Pos did not match: Expected %d, got %d", expected.Pos(), item.Pos())
    }
    i++
  }

  if i != len(expectedItems) {
    t.Errorf("Expected %d items, only got %d", len(expectedItems), i)
  }
}
