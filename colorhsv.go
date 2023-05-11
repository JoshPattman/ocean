package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/PerformLine/go-stockutil/colorutil"
)

type ColorHSV struct {
	H, S, V float64
}

func RandomHSV() ColorHSV {
	return ColorHSV{
		H: rand.Float64() * 360,
		S: rand.Float64()*0.5 + 0.5,
		V: rand.Float64()*0.5 + 0.5,
	}
}

func (c ColorHSV) Randomised(diff float64) ColorHSV {
	h := math.Mod(float64(c.H)+(rand.Float64()-0.5)*diff*2, 360)
	s := math.Max(math.Min(float64(c.S)+(rand.Float64()-0.5)*diff*2, 1), 0.5)
	v := math.Max(math.Min(float64(c.V)+(rand.Float64()-0.5)*diff*2, 1), 0.5)
	return ColorHSV{
		H: h,
		S: s,
		V: v,
	}
}

func (c ColorHSV) ToColor() color.Color {
	r, g, b := colorutil.HsvToRgb(c.H, c.S, c.V)
	return color.RGBA{r, g, b, 255}
}
