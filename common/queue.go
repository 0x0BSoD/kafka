package common

import (
	"fmt"
	"sync"
)

type Node struct {
	Key   []byte
	Value []byte
}

func (n *Node) String() string {
	return fmt.Sprintf("%s:%s", string(n.Key), string(n.Value))
}

// NewStack returns a new stack.
func NewStack() *Stack {
	return &Stack{}
}

// Stack is a basic LIFO stack that resizes as needed.
type Stack struct {
	nodes []*Node
	count int
	mu    sync.Mutex
}

// Len return size of a stack
func (s *Stack) Len() int {
	return s.count
}

// Push adds a node to the stack.
func (s *Stack) Push(n *Node) {
	s.mu.Lock()
	s.nodes = append(s.nodes[:s.count], n)
	s.count++
	s.mu.Unlock()
}

// Pop removes and returns a node from the stack in last to first order.
func (s *Stack) Get() *Node {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.count == 0 {
		return nil
	}
	s.count--
	return s.nodes[s.count]
}
