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
func NewNodeWithChildren(nw, ne, sw, se *Node) *Node {
	level := nw.level
	if ne.level != level || sw.level != level || se.level != level {
		return nil
	}

	n := NewNode(level + 1)
	n.SetChildren([4]*Node{nw, ne, sw, se})
	return n
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
	grown.SetPseudoChild(-x, -y, n)
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
func (n *Node) SetChildren(children [4]*Node) {
	if n.level <= leafLevel {
		return
	}
	for _, c := range children {
		if c.level != n.level-1 {
			return
		}
	}
	if n.hash != 0 {
		*n = *n.DeepCopy()
	}

	n.children = children
	n.population = 0
	for _, c := range n.children {
		n.population += c.population
	}
}

func (n *Node) Get(x, y int) uint8 {
	if n.level == leafLevel {
		x += leafHalfSize
		y += leafHalfSize
		return n.value[y*2*leafHalfSize+x]
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
	gcs := [4][4]*Node{} // grandchildren
	for i, c := range n.children {
		c.Subdivide()
		for j, gc := range c.children {
			gcs[i][j] = gc
		}
	}
	for i := 0; i < 4; i += 2 {
		for j := 0; j < 2; j++ {
			gcs[i][j+2], gcs[i+1][j] = gcs[i+1][j], gcs[i][j+2] // swap
		}
	}

	return [4]*Node{gcs[y+1][x+1], gcs[y+1][x+2], gcs[y+2][x+1], gcs[y+2][x+2]}
}
func (n *Node) GetPseudoChild(x, y int) *Node {
	if n.level < leafLevel+1 {
		return nil
	} else if n.level == leafLevel+1 { // edge case
		pseudoNode := NewNode(leafLevel)
		for yy := -leafHalfSize; yy < leafHalfSize; yy++ {
			for xx := -leafHalfSize; xx < leafHalfSize; xx++ {
				pseudoNode.Set(xx, yy, n.Get(x+xx, y+yy))
			}
		}
		return pseudoNode
	}

	pseudoNode := NewNode(n.level - 1)
	if n.children[0] == nil {
		return pseudoNode
	}

	pseudoNode.SetChildren(n.GetPseudoQuads(x, y))
	return pseudoNode
}
func (n *Node) SetPseudoChild(x, y int, node *Node) {
	if n.level < leafLevel+2 || node.level != n.level-1 {
		return
	}

	if n.hash != 0 {
		*n = *n.DeepCopy()
	}

	for i, q := range n.GetPseudoQuads(x, y) {
		n.population -= q.population
		*q = *node.children[i]
		n.population += q.population
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

	if n.level == leafLevel {
		h.Write(n.value[:])
	} else if n.population == 0 {

		levelBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(levelBytes, n.level)
		h.Write(levelBytes)
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
