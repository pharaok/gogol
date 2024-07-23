package main

import (
	"encoding/binary"
	"hash/maphash"
)

func sign(x int) int {
	if x < 0 {
		return -1
	} else if x > 0 {
		return 1
	}
	return 0
}

const leafLevel = 1
const leafHalfSize = 1 << (leafLevel - 1)

type Node struct {
	value      [leafHalfSize * leafHalfSize * 4]uint8
	children   [4]*Node // nw ne sw se
	hash       uint64
	population uint64
	level      uint64
}

func NewNode(level uint64) *Node {
	return &Node{
		level: level,
	}
}
func NewNodeWithChildren(children [4]*Node) *Node {
	level := children[0].level
	for _, c := range children {
		if c.level != level {
			return nil
		}
	}

	return &Node{
		level:    level + 1,
		children: children,
	}
}

func (n *Node) Subdivide() {
	if n.children[0] != nil {
		return
	}
	for i := range n.children {
		n.children[i] = NewNode(n.level - 1)
	}
}
func (n *Node) Grow(x, y int) {
	grown := NewNode(n.level + 1)
	grown.setPseudoChild(-x, -y, n)
	*n = *grown
}

func (n *Node) Child(x, y int) *Node {
	switch {
	case x < 0 && y < 0:
		return n.children[0]
	case x >= 0 && y < 0:
		return n.children[1]
	case x < 0 && y >= 0:
		return n.children[2]
	case x >= 0 && y >= 0:
		return n.children[3]
	default:
		return nil
	}
}
func (n *Node) ToChildCoords(x, y int) (int, int) {
	quarterSize := 1 << (n.level - 2)
	halfSize := quarterSize << 1
	x = (x+halfSize)%halfSize - quarterSize
	y = (y+halfSize)%halfSize - quarterSize
	return x, y
}

func (n *Node) Get(x, y int) uint8 {
	if n.level == leafLevel {
		x += leafHalfSize
		y += leafHalfSize
		return n.value[x+y*2*leafHalfSize]
	}

	if n.children[0] == nil {
		return 0
	}

	return n.Child(x, y).Get(n.ToChildCoords(x, y))
}
func (n *Node) Set(x, y int, value uint8) int {
	if n.hash != 0 {
		*n = *n.DeepCopy()
	}

	if n.level == leafLevel {
		x += leafHalfSize
		y += leafHalfSize
		i := x + y*2*leafHalfSize

		d := int(sign(int(value)) - sign(int(n.value[i])))
		n.population = uint64(int(n.population) + d)
		n.value[i] = value
		return d
	}

	n.Subdivide()
	cx, cy := n.ToChildCoords(x, y)
	d := n.Child(x, y).Set(cx, cy, value)
	n.population = uint64(int(n.population) + d)
	return d
}

func (n *Node) GetPseudoQuads(x, y int) [4]*Node { // nw ne sw se
	if n.level < leafLevel+2 {
		return [4]*Node{}
	}

	n.Subdivide()
	gcs := make([]*Node, 16) // grandchildren
	for i, c := range n.children {
		c.Subdivide()
		for j, gc := range c.children {
			gcs[i*4+j] = gc
		}
	}

	// index map
	// 0 1 4 5
	// 2 3 6 7
	// 8 9 C D
	// A B E F

	switch {
	case x == -1 && y == -1:
		return [4]*Node{gcs[0], gcs[1], gcs[2], gcs[3]}
	case x == 0 && y == -1:
		return [4]*Node{gcs[1], gcs[4], gcs[3], gcs[6]}
	case x == 1 && y == -1:
		return [4]*Node{gcs[4], gcs[5], gcs[6], gcs[7]}
	case x == -1 && y == 0:
		return [4]*Node{gcs[2], gcs[3], gcs[8], gcs[9]}
	case x == 0 && y == 0:
		return [4]*Node{gcs[3], gcs[6], gcs[9], gcs[12]}
	case x == 1 && y == 0:
		return [4]*Node{gcs[6], gcs[7], gcs[12], gcs[13]}
	case x == -1 && y == 1:
		return [4]*Node{gcs[8], gcs[9], gcs[10], gcs[11]}
	case x == 0 && y == 1:
		return [4]*Node{gcs[9], gcs[12], gcs[11], gcs[14]}
	case x == 1 && y == 1:
		return [4]*Node{gcs[12], gcs[13], gcs[14], gcs[15]}
	}
	return [4]*Node{}
}
func (n *Node) GetPseudoChild(x, y int) *Node {
	if n.level < leafLevel+2 {
		return nil
	}

	pseudoNode := NewNode(n.level - 1)
	if n.children[0] == nil {
		return pseudoNode
	}

	pseudoNode.children = n.GetPseudoQuads(x, y)
	return pseudoNode
}
func (n *Node) setPseudoChild(x, y int, node *Node) {
	if n.level < leafLevel+2 || node.level != n.level-1 {
		return
	}

	if n.hash != 0 {
		*n = *n.DeepCopy()
	}

	for i, q := range n.GetPseudoQuads(x, y) {
		*q = *node.children[i]
	}
}

func (n *Node) DeepCopy() *Node {
	newNode := NewNode(n.level)
	newNode.population = n.population
	if n.children[0] == nil {
		copy(newNode.value[:], n.value[:])
	} else {
		for i, c := range n.children {
			newNode.children[i] = c.DeepCopy()
		}
	}
	return newNode
}

func (n *Node) Hash(h maphash.Hash) uint64 {
	if n.hash != 0 {
		return n.hash
	}

	if n.children[0] == nil {
		if n.level == leafLevel {
			h.Write(n.value[:])
		} else {
			levelBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(levelBytes, n.level)
			h.Write(levelBytes)
		}
	} else {
		for _, c := range n.children {
			hashBytes := make([]byte, 8)
			ch := maphash.Hash{}
			ch.SetSeed(h.Seed())
			binary.LittleEndian.PutUint64(hashBytes, c.Hash(ch))
			h.Write(hashBytes)
		}
	}

	n.hash = h.Sum64()
	return n.hash
}
