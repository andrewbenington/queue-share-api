package util

type Node[T any] struct {
	Data T
	Prev *Node[T]
	Next *Node[T]
}

func (n *Node[T]) InsertAfter(data T) *Node[T] {
	newNode := Node[T]{
		Data: data,
		Next: n.Next,
		Prev: n,
	}

	if n.Next != nil {
		n.Next.Prev = &newNode
	}

	n.Next = &newNode

	return &newNode
}

func (n *Node[T]) InsertBefore(data T) *Node[T] {
	newNode := Node[T]{
		Data: data,
		Next: n,
		Prev: n.Prev,
	}

	if n.Prev != nil {
		n.Prev.Next = &newNode
	}

	n.Prev = &newNode

	return &newNode
}

func (this *Node[T]) SwapWithNext() {
	if this.Next == nil {
		return
	}

	afterThis := this.Next
	beforeThis := this.Prev
	twoAfterThis := afterThis.Next

	afterThis.Prev = beforeThis
	afterThis.Next = this
	this.Prev = afterThis
	this.Next = twoAfterThis

	if twoAfterThis != nil {
		twoAfterThis.Prev = this
	}

	if beforeThis != nil {
		beforeThis.Next = afterThis
	}
}

type DoublyLinkedList[T any] struct {
	first *Node[T]
	last  *Node[T]
	size  int
}

func (l *DoublyLinkedList[T]) Size() int {
	return l.size
}

func (l *DoublyLinkedList[T]) Last() *T {
	if l.last == nil {
		return nil
	}
	return &l.last.Data
}

func (l *DoublyLinkedList[T]) PeekFirst() *T {
	if l.first == nil {
		return nil
	}
	return &l.first.Data
}

func (l *DoublyLinkedList[T]) PeekLast() *T {
	if l.last == nil {
		return nil
	}
	return &l.last.Data
}

func (l *DoublyLinkedList[T]) PopFirst() *T {
	if l.first == nil {
		return nil
	}

	first := l.first.Data
	if l.first.Next != nil {
		l.first.Next.Prev = nil
		l.first = l.first.Next
	} else {
		l.first = nil
		l.last = nil
	}

	l.size--
	return &first
}

func (l *DoublyLinkedList[T]) PushStart(data T) {
	l.size++
	if l.first != nil {
		l.first = l.first.InsertBefore(data)
	} else {
		newNode := Node[T]{Data: data}
		l.first = &newNode
		l.last = &newNode
	}
}

func (l *DoublyLinkedList[T]) PushEnd(data T) {
	l.size++
	if l.last != nil {
		l.last = l.last.InsertAfter(data)
	} else {
		newNode := Node[T]{Data: data}
		l.first = &newNode
		l.last = &newNode
	}
}

func (l *DoublyLinkedList[T]) ToSlice() []T {
	slice := make([]T, 0, l.size)
	if l.first != nil {
		current := l.first
		for current != nil {
			slice = append(slice, current.Data)
			current = current.Next
		}
	}
	return slice
}

func DoublyLinkedListFromSlice[T any](slice []T) *DoublyLinkedList[T] {
	if len(slice) == 0 {
		return &DoublyLinkedList[T]{}
	}

	firstNode := Node[T]{Data: slice[0]}
	list := DoublyLinkedList[T]{
		first: &firstNode,
		last:  &firstNode,
		size:  1,
	}

	for _, node := range slice[1:] {
		list.PushEnd(node)
	}

	return &list
}
