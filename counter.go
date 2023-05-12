package main

import "github.com/JoshPattman/goevo"

type SaveLoadCounter struct {
	c int
}

func (c *SaveLoadCounter) Next() int {
	c.c++
	return c.c
}

func (c *SaveLoadCounter) SafeWith(gt *goevo.Genotype) {
	max := 0
	for id := range gt.Neurons {
		if id > max {
			max = id
		}
	}
	if max >= c.c {
		c.c = max + 1
	}
}
