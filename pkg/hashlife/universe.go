package hashlife

import (
	"hash/maphash"
	"math"
)

type LeafParentGrid [2 * LeafSize][2 * LeafSize]uint8

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
	for i := 1; i <= min(LeafLevel, generations); i++ {
		nextGrid := LeafParentGrid{}
		for y := i; y < 2*LeafSize-i; y++ {
			for x := i; x < 2*LeafSize-i; x++ {

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

type key struct {
	nodeHash    uint64
	generations uint64
}

type Universe struct {
	Root       *Node
	Cache      map[key]*Node // TODO: LRU
	Hasher     maphash.Hash
	Generation uint64
	// TODO:history
}

func NewUniverse(level uint64) *Universe {
	return &Universe{
		Root:   NewNode(level),
		Cache:  make(map[key]*Node),
		Hasher: maphash.Hash{},
	}
}

func (u *Universe) StepNode(n *Node, generations uint64) *Node {
	if n.Population < 3 {
		return NewNode(n.Level - 1)
	}

	u.Hasher.Reset()
	nodeHash := n.Hash(u.Hasher)
	key := key{nodeHash, uint64(math.Min(float64(generations), float64(n.Level-2)))}
	if cached, ok := u.Cache[key]; ok {
		return cached
	}

	next := NewNode(n.Level - 1)
	if n.Level == LeafLevel+1 {
		grid := LeafParentGrid{}
		for y := -LeafSize; y < LeafSize; y++ {
			for x := -LeafSize; x < LeafSize; x++ {
				grid[y+LeafSize][x+LeafSize] = n.Get(x, y)
			}
		}

		grid.Step(1 << generations)

		for y := -LeafHalfSize; y < LeafHalfSize; y++ {
			for x := -LeafHalfSize; x < LeafHalfSize; x++ {
				next.Set(x, y, grid[y+LeafSize][x+LeafSize])
			}
		}
	} else {
		s := [3][3]*Node{}
		for y := -1; y <= 1; y++ {
			for x := -1; x <= 1; x++ {
				s[y+1][x+1] = u.StepNode(n.GetPseudoChild(x, y), generations)
			}
		}

		children := [4]*Node{
			NewNodeWithChildren(s[0][0], s[0][1], s[1][0], s[1][1]),
			NewNodeWithChildren(s[0][1], s[0][2], s[1][1], s[1][2]),
			NewNodeWithChildren(s[1][0], s[1][1], s[2][0], s[2][1]),
			NewNodeWithChildren(s[1][1], s[1][2], s[2][1], s[2][2]),
		}
		for i, c := range children {
			if generations+2 >= n.Level {
				children[i] = u.StepNode(c, generations)
			} else {
				children[i] = c.GetPseudoChild(0, 0)
			}
		}

		next.SetChildren(children)
	}

	if generations >= n.Level-2 { // FIX: breaks without this check for some reason
		u.Cache[key] = next
	}
	return next
}

func (u *Universe) Step(generations int) {
	generations = min(generations, int(u.Root.Level-2))
	u.Root.Grow(0, 0)
	u.Root = u.StepNode(u.Root, uint64(generations))
	u.Generation += 1 << generations
}

func (u *Universe) Get(x, y int) uint8 {
	return u.Root.Get(x, y)
}
func (u *Universe) Set(x, y int, value uint8) {
	u.Root.Set(x, y, value)
	u.Generation = 0
}
