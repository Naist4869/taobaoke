//+build : trie

package dao

import "taobaoke/internal/model"

type Trie struct {
	root *Node
	size int
}

func (t *Trie) String() string {
	panic("implement me")
}

func (t *Trie) Size() int {
	return t.size
}

func NewTrie() *Trie {
	return &Trie{
		root: NewNode(nil),
		size: 0,
	}
}

type Node struct {
	handlers []HandlerFunc
	Next     map[model.OrderStatus]*Node
}

func NewNode(handlers []HandlerFunc) *Node {
	return &Node{handlers: handlers, Next: make(map[model.OrderStatus]*Node)}
}

// 非递归写法
func (t *Trie) Insert(handler []HandlerFunc, statuses ...model.OrderStatus) {
	cur := t.root
	for _, status := range statuses {
		if _, exist := cur.Next[status]; !exist {
			cur.Next[status] = NewNode(nil)
		}
		cur = cur.Next[status]
	}
	if len(cur.handlers) == 0 {
		cur.handlers = handler
		t.size++
	}
}

// 递归写法
func (t *Trie) Insert2(handler []HandlerFunc, statuses ...model.OrderStatus) {
	end := t.root.insert2(handler, statuses...)
	if len(end.handlers) == 0 {
		end.handlers = handler
		t.size++
	}
}

func (n *Node) insert2(handler []HandlerFunc, statuses ...model.OrderStatus) *Node {
	if len(statuses) == 0 {
		return n
	}
	status := statuses[0]
	if _, exist := n.Next[status]; !exist {
		n.Next[status] = NewNode(nil)
	}
	return n.Next[status].insert2(handler, statuses[1:]...)
}

// 查询对应的状态链是否有方法存在
func (t *Trie) Search(statuses ...model.OrderStatus) bool {
	cur := t.root
	for _, status := range statuses {
		if _, exist := cur.Next[status]; !exist {
			return false
		}
		cur = cur.Next[status]
	}
	if len(cur.handlers) > 0 {
		return true
	}
	return false
}

// 递归
func (t *Trie) Search2(statuses ...model.OrderStatus) bool {
	return t.root.search2(statuses...)
}
func (n *Node) search2(statuses ...model.OrderStatus) bool {
	if len(statuses) == 0 {
		return len(n.handlers) > 0
	}
	status := statuses[0]
	if _, exist := n.Next[status]; exist {
		return n.Next[status].search2(statuses[1:]...)
	}
	return false
}
func (t *Trie) StartsWith(statuses ...model.OrderStatus) bool {
	cur := t.root
	for _, status := range statuses {
		if _, exist := cur.Next[status]; !exist {
			return false
		}
		cur = cur.Next[status]
	}
	return true
}
func (t *Trie) HandlerChain(prefix ...model.OrderStatus) []HandlerFunc {
	cur := t.root
	for _, status := range prefix {
		if _, exist := cur.Next[status]; !exist {
			return nil
		}
		cur = cur.Next[status]
	}
	return cur.handlers
}

func (n *Node) fillHandlerChain() []HandlerFunc {
	res := n.handlers
	for _, node := range n.Next {
		res = append(res, node.fillHandlerChain()...)
	}
	return res
}
