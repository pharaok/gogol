package main

import (
	"encoding/binary"
	"fmt"
	"hash/maphash"
)

type Node struct {
	value      [4]uint8
	child      [4]*Node
	_hash      uint64
	population uint64
	level      uint64
}

func (n *Node) hash(h maphash.Hash) uint64 {
	if n._hash != 0 {
		return n._hash
	}

	if n.child[0] == nil {
		if n.level == 0 {
			h.Write(n.value[:])
		} else {
			levelBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(levelBytes, n.level)
			h.Write(levelBytes)
		}
	} else {
		for _, c := range n.child {
			hashBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(hashBytes, c.hash(h))
			h.Write(hashBytes)
		}
	}

	n._hash = h.Sum64()
	return n._hash
}

func main() {
	h1 := maphash.Hash{}
	h2 := maphash.Hash{}
	h2.SetSeed(h1.Seed())

	n1 := Node{value: [4]uint8{1, 2, 3, 4}}
	n2 := Node{value: [4]uint8{1, 2, 3, 4}}
	fmt.Println("node1 hash:", n1.hash(h1))
	fmt.Println("node2 hash:", n2.hash(h2))
}
