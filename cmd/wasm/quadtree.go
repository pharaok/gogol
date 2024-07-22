package main

import (
	"encoding/binary"
	"hash/maphash"
)

const leafLevel uint64 = 1
const leafHalfSize uint64 = 1 << (leafLevel - 1)

type Node struct {
	value      [leafHalfSize * leafHalfSize * 4]uint8
	children   [4]*Node // nw ne sw se
	_hash      uint64
	population uint64
	level      uint64
}

func NewNode(level uint64) *Node {
	return &Node{
		level: level,
	}
}

func (n *Node) subdivide() {
	for i := range n.children {
		n.children[i] = NewNode(n.level - 1)
	}
}

func (n *Node) child(x, y int32) *Node {
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
func (n *Node) toChildCoords(x, y int32) (int32, int32) {
	quarterSize := int32(1 << (n.level - 2))
	halfSize := quarterSize << 1
	x = (x+halfSize)%halfSize - quarterSize
	y = (y+halfSize)%halfSize - quarterSize
	return x, y
}

func (n *Node) get(x, y int32) uint8 {
	if n.level == leafLevel {
		x += int32(leafHalfSize)
		y += int32(leafHalfSize)
		return n.value[x+y*2*int32(leafHalfSize)]
	}

	if n.children[0] == nil {
		return 0
	}

	return n.child(x, y).get(n.toChildCoords(x, y))
}
func (n *Node) set(x, y int32, value uint8) {
	if n._hash != 0 {
		*n = *n.deepCopy()
	}

	if n.level == leafLevel {
		x += int32(leafHalfSize)
		y += int32(leafHalfSize)
		n.value[x+y*2*int32(leafHalfSize)] = value
		return
	}

	if n.children[0] == nil {
		n.subdivide()
	}

	cx, cy := n.toChildCoords(x, y)
	n.child(x, y).set(cx, cy, value)
}

func (n *Node) deepCopy() *Node {
	newNode := NewNode(n.level)
	if n.children[0] == nil {
		copy(newNode.value[:], n.value[:])
	} else {
		for i, c := range n.children {
			newNode.children[i] = c.deepCopy()
		}
	}
	return newNode
}

func (n *Node) hash(h maphash.Hash) uint64 {
	if n._hash != 0 {
		return n._hash
	}

	if n.children[0] == nil {
		if n.level == 1 {
			h.Write(n.value[:])
		} else {
			levelBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(levelBytes, n.level)
			h.Write(levelBytes)
		}
	} else {
		for _, c := range n.children {
			hashBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(hashBytes, c.hash(h))
			h.Write(hashBytes)
		}
	}

	n._hash = h.Sum64()
	return n._hash
}
