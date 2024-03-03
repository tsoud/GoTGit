package main

import (
	"cmp"
	"fmt"
	"slices"
)

type someone struct {
	Name     string
	LastName string
	Age      int
}

func main() {
	henry := &someone{
		Name:     "Henry",
		LastName: "Anderson",
		Age:      47,
	}

	john := &someone{
		Name:     "John",
		LastName: "Johnson",
		Age:      25,
	}

	frank := &someone{
		Name:     "Frank",
		LastName: "Zephyr",
		Age:      60,
	}

	s := []someone{*john, *frank, *henry}

	slices.SortFunc(s, func(A, B someone) int {
		return cmp.Compare(A.LastName, B.LastName)
	})

	fmt.Println(s)
}
