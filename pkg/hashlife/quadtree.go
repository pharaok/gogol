package hashlife

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

func hasherWriteUint64(h *maphash.Hash, x uint64) {
	hashBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(hashBytes, x)
	h.Write(hashBytes)
}

const LeafLevel = 1
const LeafHalfSize = 1 << (LeafLevel - 1)
const LeafSize int = LeafHalfSize << 1

type Node struct {
	Value      [LeafHalfSize * LeafHalfSize * 4]uint8
	Children   [4]*Node // nw ne sw se
	hash       uint64
	Population uint64
	Level      uint64
}

func NewNode(level uint64) *Node {
	return &Node{
		Level: level,
	}
}
func NewNodeWithChildren(nw, ne, sw, se *Node) *Node {
	level := nw.Level
	if ne.Level != level || sw.Level != level || se.Level != level {
		return nil
	}

	n := NewNode(level + 1)
	n.SetChildren([4]*Node{nw, ne, sw, se})
	return n
}

func (n *Node) Subdivide() {
	if n.Children[0] != nil {
		return
	}
	for i := range n.Children {
		n.Children[i] = NewNode(n.Level - 1)
	}
}
func (n *Node) Grow(x, y int) {
	n.Subdivide()
	grown := NewNode(n.Level + 1)
	grown.SetPseudoChild(-x, -y, n)
	*n = *grown
}

func (n *Node) Child(x, y int) *Node {
	switch {
	case x < 0 && y < 0:
		return n.Children[0]
	case x >= 0 && y < 0:
		return n.Children[1]
	case x < 0 && y >= 0:
		return n.Children[2]
	case x >= 0 && y >= 0:
		return n.Children[3]
	default:
		return nil
	}
}
func (n *Node) ToChildCoords(x, y int) (int, int) {
	quarterSize := 1 << (n.Level - 2)
	halfSize := quarterSize << 1
	x = (x+halfSize)%halfSize - quarterSize
	y = (y+halfSize)%halfSize - quarterSize
	return x, y
}
func (n *Node) SetChildren(children [4]*Node) {
	if n.Level <= LeafLevel {
		return
	}
	for _, c := range children {
		if c.Level != n.Level-1 {
			return
		}
	}
	if n.hash != 0 {
		*n = *n.DeepCopy()
	}

	n.Children = children
	n.Population = 0
	for _, c := range n.Children {
		n.Population += c.Population
	}
}

func (n *Node) Get(x, y int) uint8 {
	if n.Level == LeafLevel {
		x += LeafHalfSize
		y += LeafHalfSize
		return n.Value[y*2*LeafHalfSize+x]
	}

	if n.Children[0] == nil {
		return 0
	}

	return n.Child(x, y).Get(n.ToChildCoords(x, y))
}
func (n *Node) Set(x, y int, value uint8) int {
	if n.hash != 0 {
		*n = *n.DeepCopy()
	}

	if n.Level == LeafLevel {
		x += LeafHalfSize
		y += LeafHalfSize
		i := x + y*2*LeafHalfSize

		d := int(sign(int(value)) - sign(int(n.Value[i])))
		n.Population = uint64(int(n.Population) + d)
		n.Value[i] = value
		return d
	}

	n.Subdivide()
	cx, cy := n.ToChildCoords(x, y)
	d := n.Child(x, y).Set(cx, cy, value)
	n.Population = uint64(int(n.Population) + d)
	return d
}

func (n *Node) GetPseudoQuads(x, y int) [4]*Node { // nw ne sw se
	if n.Level < LeafLevel+2 {
		return [4]*Node{}
	}

	n.Subdivide()
	gcs := [4][4]*Node{} // grandchildren
	for i, c := range n.Children {
		c.Subdivide()
		for j, gc := range c.Children {
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
	if n.Level < LeafLevel+1 {
		return nil
	} else if n.Level == LeafLevel+1 { // edge case
		pseudoNode := NewNode(LeafLevel)
		for yy := -LeafHalfSize; yy < LeafHalfSize; yy++ {
			for xx := -LeafHalfSize; xx < LeafHalfSize; xx++ {
				pseudoNode.Set(xx, yy, n.Get(x+xx, y+yy))
			}
		}
		return pseudoNode
	}

	pseudoNode := NewNode(n.Level - 1)
	if n.Children[0] == nil {
		return pseudoNode
	}

	pseudoNode.SetChildren(n.GetPseudoQuads(x, y))
	return pseudoNode
}
func (n *Node) SetPseudoChild(x, y int, node *Node) {
	if n.Level < LeafLevel+2 || node.Level != n.Level-1 {
		return
	}

	if n.hash != 0 {
		*n = *n.DeepCopy()
	}

	for i, q := range n.GetPseudoQuads(x, y) {
		n.Population -= q.Population
		*q = *node.Children[i]
		n.Population += q.Population
	}
}

func (n *Node) DeepCopy() *Node {
	newNode := NewNode(n.Level)
	newNode.Population = n.Population
	if n.Children[0] == nil {
		copy(newNode.Value[:], n.Value[:])
	} else {
		for i, c := range n.Children {
			newNode.Children[i] = c.DeepCopy()
		}
	}
	return newNode
}

func (n *Node) Hash(h maphash.Hash) uint64 {
	if n.hash != 0 {
		return n.hash
	}

	if n.Level == LeafLevel {
		h.Write(n.Value[:])
	} else if n.Population == 0 {
		hasherWriteUint64(&h, n.Level)
	} else {
		for _, c := range n.Children {
			ch := maphash.Hash{}
			ch.SetSeed(h.Seed())
			hasherWriteUint64(&h, c.Hash(ch))
		}
	}

	n.hash = h.Sum64()
	return n.hash
}
