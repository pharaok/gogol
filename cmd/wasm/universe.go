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

func (g *LeafParentGrid) Step() {
	for i := 1; i <= leafLevel; i++ {
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
	root   *Node
	cache  map[uint64]*Node
	hasher maphash.Hash
	// TODO:history
}

func NewUniverse() *Universe {
	return &Universe{
		root:   NewNode(leafLevel + 2),
		cache:  make(map[uint64]*Node),
		hasher: maphash.Hash{},
	}
}

func (u *Universe) step(n *Node) *Node {
	if n.population == 0 {
		return NewNode(n.level - 1)
	}

	u.hasher.Reset()
	hash := n.Hash(u.hasher)
	if cached, ok := u.cache[hash]; ok {
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

		grid.Step()

		for y := -leafHalfSize; y < leafHalfSize; y++ {
			for x := -leafHalfSize; x < leafHalfSize; x++ {
				next.Set(x, y, grid[y+leafSize][x+leafSize])
			}
		}
	} else {
		// TODO:
	}

	u.cache[hash] = next
	return next
}
