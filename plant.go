package main

import "github.com/faiface/pixel"

type Plant struct {
	Pos     pixel.Vec
	Radius  float64
	Rot     float64
	Shading uint8
}

func (p *Plant) HMPos() pixel.Vec {
	return p.Pos
}
func (p *Plant) Eq(o HashMappable) bool {
	if o == nil {
		return false
	}
	return p.Pos == o.(*Plant).Pos
}
