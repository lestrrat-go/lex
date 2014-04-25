package lex

// Consumer is a base implementation for things that consume the Lexer interface
type Consumer interface {
	Peek() LexItem
	Consume() LexItem
	Backup()
	Backup2(LexItem)
}

// ItemConsume is a simple Consumer impementation.
type ItemConsume struct {
	lexer     Lexer
	items     [3]LexItem
	peekCount int
}

// NewItemConsume creates a new ItemConsume instance
func NewItemConsume(l Lexer) *ItemConsume {
	return &ItemConsume{
		l,
		[3]LexItem{},
		0,
	}
}

// Peek returns the next item, but does not consume it
func (c *ItemConsume) Peek() LexItem {
	if c.peekCount > 0 {
		return c.items[c.peekCount-1]
	}
	c.peekCount = 1
	c.items[0] = c.lexer.NextItem()
	return c.items[0]
}

// Consume returns the next item, and consumes it.
func (c *ItemConsume) Consume() LexItem {
	if c.peekCount > 0 {
		c.peekCount--
	} else {
		c.items[0] = c.lexer.NextItem()
	}

	return c.items[c.peekCount]
}

// Backup moves 1 item back
func (c *ItemConsume) Backup() {
	c.peekCount++
}

// Backup2 pushes `t1` into the buffer, and moves 2 items back
func (c *ItemConsume) Backup2(t1 LexItem) {
	c.items[1] = t1
	c.peekCount = 2
}
