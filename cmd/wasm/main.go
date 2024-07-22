package main

import (
	"fmt"
)

func main() {
	n1 := NewNode(5)
	n1.set(0, 0, 23)
	fmt.Println(n1.get(0, 0))
}
