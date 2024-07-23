package main

import (
	"hash/maphash"
)

const leafSize int = leafHalfSize << 1

type LeafParentGrid [2 * leafSize][2 * leafSize]uint8

func (g *LeafParentGrid) NeighborCount(x, y int) int {
	pop := -int(g[y][x])
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			pop += int(g[y+dy][x+dx])
		}
	}
	return pop
}

func (g *LeafParentGrid) Step(generations int) {
	for i := 1; i <= min(leafLevel, generations); i++ {
		nextGrid := LeafParentGrid{}
		for y := i; y < 2*leafSize-i; y++ {
			for x := i; x < 2*leafSize-i; x++ {

				switch g.NeighborCount(x, y) {
				case 2:
					nextGrid[y][x] = g[y][x]
				case 3:
					nextGrid[y][x] = 1
				default:
					nextGrid[y][x] = 0
				}

			}
		}
		*g = nextGrid
	}
}

type Universe struct {
	root       *Node
	cache      map[uint64]*Node // TODO: LRU
	hasher     maphash.Hash
	generation uint64
	// TODO:history
}

func NewUniverse(level uint64) *Universe {
	return &Universe{
		root:   NewNode(level),
		cache:  make(map[uint64]*Node),
		hasher: maphash.Hash{},
	}
}

func (u *Universe) stepNode(n *Node, generations int) *Node {
	if n.population < 3 {
		return NewNode(n.level - 1)
	}

	u.hasher.Reset()
	hash := n.Hash(u.hasher)
	if cached, ok := u.cache[hash]; ok && uint64(generations)+2 >= n.level {
		return cached
	}

	next := NewNode(n.level - 1)
	if n.level == leafLevel+1 {
		grid := LeafParentGrid{}
		for y := -leafSize; y < leafSize; y++ {
			for x := -leafSize; x < leafSize; x++ {
				grid[y+leafSize][x+leafSize] = n.Get(x, y)
			}
		}

		grid.Step(1 << generations)

		for y := -leafHalfSize; y < leafHalfSize; y++ {
			for x := -leafHalfSize; x < leafHalfSize; x++ {
				next.Set(x, y, grid[y+leafSize][x+leafSize])
			}
		}
	} else {
		s := [3][3]*Node{}
		for y := -1; y <= 1; y++ {
			for x := -1; x <= 1; x++ {
				s[y+1][x+1] = u.stepNode(n.GetPseudoChild(x, y), generations)
			}
		}

		children := [4]*Node{
			NewNodeWithChildren(s[0][0], s[0][1], s[1][0], s[1][1]),
			NewNodeWithChildren(s[0][1], s[0][2], s[1][1], s[1][2]),
			NewNodeWithChildren(s[1][0], s[1][1], s[2][0], s[2][1]),
			NewNodeWithChildren(s[1][1], s[1][2], s[2][1], s[2][2]),
		}
		if uint64(generations)+2 >= n.level {
			for i, c := range children {
				children[i] = u.stepNode(c, generations)
			}
		} else {
			for i, c := range children {
				children[i] = c.GetPseudoChild(0, 0)
			}
		}

		next.SetChildren(children)
	}

	if uint64(generations)+2 >= n.level {
		u.cache[hash] = next
	}
	return next
}

func (u *Universe) Step(generations int) {
	generations = min(generations, int(u.root.level-2))
	u.root = u.stepNode(u.root, generations)
	u.root.Grow(0, 0)
	u.generation += 1 << generations
}

func (u *Universe) Get(x, y int) uint8 {
	return u.root.Get(x, y)
}
func (u *Universe) Set(x, y int, value uint8) {
	u.root.Set(x, y, value)
	u.generation = 0
}
