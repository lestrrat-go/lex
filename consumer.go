package lex

type Consumer interface {
  Peek() LexItem
  Consume() LexItem
  Backup()
  Backup2(LexItem)
}

type ItemConsume struct {
  lexer Lexer
  items [3]LexItem
  peekCount int
}

func NewItemConsume(l Lexer) *ItemConsume {
  return &ItemConsume {
    l,
    [3]LexItem {},
    0,
  }
}

func (c *ItemConsume) Peek() LexItem {
  if c.peekCount > 0 {
    return c.items[c.peekCount - 1]
  }
  c.peekCount = 1
  c.items[0] = c.lexer.NextItem()
  return c.items[0]
}

func (c *ItemConsume) Consume() LexItem {
  if c.peekCount > 0 {
    c.peekCount--
  } else {
    c.items[0] = c.lexer.NextItem()
  }

  return c.items[c.peekCount]
}

func (c *ItemConsume) Backup() {
  c.peekCount++

}

func (c *ItemConsume) Backup2(t1 LexItem) {
  c.items[1] = t1
  c.peekCount = 2
}
