package lex

import (
  "testing"
)

const (
  ItemNumber = ItemDefaultMax + iota
  ItemOperator
  ItemWhitespace
)

func TestLex(t *testing.T) {
  l := NewStringLexer("1 + 2")
  l.SetLexFn("__START__", func(l Lexer, ctx interface {}) LexFn {
    b := l.Peek()
    switch {
    case b >= 0x31 && b <= 39:
      return l.MustGetLexFn("Number")
    case b == '+':
      return l.MustGetLexFn("Operator")
    case b == ' ' || b == 0x0a || b == 0x13 || b == 0x09:
      return l.MustGetLexFn("Whitespace")
    }
    return nil
  })
  l.SetLexFn("Number", func(l Lexer, ctx interface {}) LexFn {
    l.AcceptRun("0123456789")
    l.Emit(ItemNumber)
    return l.MustGetLexFn("__START__")
  })

  go l.Run(l)

  expectedItems := []LexItem {
    NewLexItem( ItemNumber, 0, "1" ),
    NewLexItem( ItemWhitespace, 1, " "),
    NewLexItem( ItemOperator, 2, "+" ),
    NewLexItem( ItemWhitespace, 3, " "),
    NewLexItem( ItemNumber, 4, "2"),
  }

  i := 0
  for item := range l.Items() {
    expected := expectedItems[i]
    if expected.Type() != item.Type() {
      t.Errorf("Type did not match: Expected %d, got %d", expected.Type(), item.Type())
    }
  }
}