package lex

import (
  "fmt"
)

// LexItemType describes the type of a LexItem
type LexItemType int
var TypeNames = make(map[LexItemType]string)
const (
  // ItemEOF is emiteed upon EOF
  ItemEOF LexItemType = iota
  // ItemError is emitted upon Error
  ItemError
  // ItemDefaultMax is used as marker for your own ItemType. 
  // Start your types from this + 1
  ItemDefaultMax
)

func init () {
  TypeNames[ItemEOF] = "EOF"
  TypeNames[ItemError] = "Error"
  TypeNames[ItemDefaultMax] = "Special (DefaultMax)"
}

func (t LexItemType) String() string {
  name, ok := TypeNames[t]
  if ! ok {
    return fmt.Sprintf("Unknown Item (%d)", t)
  }
  return name
}

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

// String returns the string representation of the LexItem
func (l LexItem) String() string {
  return fmt.Sprintf("%s (%s)", l.typ, l.val)
}
