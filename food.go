package main

import (
	"math"

	"github.com/faiface/pixel"
)

type Food struct {
	Pos      pixel.Vec
	Rot      float64
	IsVeggie bool
	Energy   float64
}

func NewFood(energy float64, veggie bool) *Food {
	return &Food{
		Pos:      pixel.V(0, 0),
		Rot:      0,
		IsVeggie: veggie,
		Energy:   energy,
	}
}

func (f *Food) HMPos() pixel.Vec {
	return f.Pos
}
func (f *Food) Eq(o HashMappable) bool {
	if o == nil {
		return false
	}
	return f.Pos == o.(*Food).Pos
}

func (f *Food) Radius() float64 {
	return math.Pow(f.Energy, 1/3.0)
}
