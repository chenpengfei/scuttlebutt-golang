package duplex

//todo.线程安全的必要性
type Buffer struct {
	items []interface{}
}

func NewBuffer() *Buffer {
	return &Buffer{items: make([]interface{}, 0)}
}

func (b *Buffer) Length() int {
	return len(b.items)
}

// Remove an item from the beginning of an array
func (b *Buffer) Shift() interface{} {
	if len(b.items) != 0 {
		elem := b.items[0]
		b.items = b.items[1:]
		return elem
	}

	return nil
}

// Add items to the beginning of an array
func (b *Buffer) Unshift(item interface{}) {
	b.items = append([]interface{}{item}, b.items...)
}

// Add items to the end of an array
func (b *Buffer) Push(item interface{}) {
	b.items = append(b.items, item)
}
