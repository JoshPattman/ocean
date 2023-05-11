package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/aquilax/go-perlin"
	"github.com/faiface/pixel"
	"golang.org/x/image/colornames"
)

type Environment struct {
	TexelsWall [][]bool
	Food       *HashMap[*Food]
	Creatures  *HashMap[*Creature]
	Radius     int
	Plants     *HashMap[*Plant]
}

func NewEnvironment(radius int) *Environment {
	env := &Environment{
		TexelsWall: nil,
		Radius:     radius,
		Food:       NewHashMap[*Food](10),
		Creatures:  NewHashMap[*Creature](10),
		Plants:     NewHashMap[*Plant](10),
	}
	env.regenerateTerrain()
	env.regrowPlants()
	return env
}

func (env *Environment) regenerateTerrain() {
	radius := env.Radius
	radiusFloat := float64(radius)
	tw := make([][]bool, radius*2)
	perlinGen := perlin.NewPerlin(1.8, 2, 3, time.Now().UnixNano())
	center := pixel.V(radiusFloat, radiusFloat)
	for i := range tw {
		tw[i] = make([]bool, radius*2)
		for j := range tw[i] {
			xc, yc := float64(i), float64(j)
			p := perlinGen.Noise2D(xc/25, yc/25)
			d := center.Sub(pixel.V(float64(i), float64(j))).Len()
			if d < 0.25*radiusFloat {
				tw[i][j] = false
			} else if d < 0.5*radiusFloat {
				tw[i][j] = p > 0.3
			} else if d < 0.75*radiusFloat {
				tw[i][j] = p > 0.1
			} else if d < radiusFloat-2 {
				tw[i][j] = p > 0.0
			} else if d < radiusFloat {
				tw[i][j] = true
			} else {
				tw[i][j] = true
			}
		}
	}
	env.TexelsWall = tw
}

func (env *Environment) regrowPlants() {
	env.Plants = NewHashMap[*Plant](10)
	perlinGen := perlin.NewPerlin(1.8, 2, 3, time.Now().UnixNano())
	radiusFloat := float64(env.Radius)
	for x := -radiusFloat; x < radiusFloat; x += 1 {
		for y := -radiusFloat; y < radiusFloat; y += 1 {
			p := pixel.V(x, y)
			if !env.sampleWallAt(p, true) {
				// Only if this is a free space with some distance to the side
				if perlinGen.Noise2D(p.X/100, p.Y/100) > rand.Float64()/SPPlantDensity {
					env.Plants.Add(&Plant{
						Pos:     p,
						Radius:  3 + rand.Float64()*2,
						Rot:     rand.Float64() * 2 * math.Pi,
						Shading: uint8(rand.Intn(100)),
					})
				}
			}
		}
	}
	env.Plants.Refresh()
}

func (env *Environment) GetTerrainSprite() *pixel.Sprite {
	img := image.NewRGBA(image.Rect(0, 0, len(env.TexelsWall), len(env.TexelsWall[0])))
	for i := range env.TexelsWall {
		for j := range env.TexelsWall[i] {
			jImg := len(env.TexelsWall[i]) - j - 1
			if env.TexelsWall[i][j] {
				img.Set(i, jImg, colornames.Black)
			} else {
				d := pixel.V(float64(i), float64(j)).Sub(pixel.V(float64(env.Radius), float64(env.Radius))).Len() / float64(env.Radius)
				if d >= 1 {
					img.Set(i, jImg, color.RGBA{0, 0, 0, 0})
				} else {
					img.Set(i, jImg, lerpColor(colornames.Skyblue, color.RGBA{28, 40, 90, 255}, d))
				}
			}
		}
	}

	pic := pixel.PictureDataFromImage(img)
	return pixel.NewSprite(pic, pic.Bounds())
}

func (e *Environment) ScatterFood(density float64) {
	numFood := int(density * float64(e.Radius*e.Radius) * math.Pi)
	for i := 0; i < numFood; i++ {
		position := pixel.V(0, math.Sqrt(rand.Float64())*float64(e.Radius)).Rotated(rand.Float64() * 2 * math.Pi)
		if !e.sampleWallAt(position, true) {
			energy := rand.Float64()*2 + 1
			if rand.Float64() < 0.1 {
				energy *= 10
			}
			f := NewFood(energy, true)
			f.Pos = position
			f.Rot = rand.Float64() * 2 * math.Pi
			e.Food.Add(f)
		}
	}
}

func (e *Environment) sampleWallAt(pos pixel.Vec, smooth bool) bool {
	// The position we get has a center of 0,0 (center of the screen)
	// The position from the texels wall has a center of radius,radius (center of the texels wall)
	px, py := e.worldPosToMapPos(pos)
	if !(px >= 0 && px < len(e.TexelsWall) && py >= 0 && py < len(e.TexelsWall[0])) {
		return true
	}
	if smooth {
		return e.TexelsWall[px][py] || e.TexelsWall[px+1][py] || e.TexelsWall[px][py+1] || e.TexelsWall[px-1][py] || e.TexelsWall[px][py-1] ||
			e.TexelsWall[px+1][py+1] || e.TexelsWall[px-1][py-1] || e.TexelsWall[px+1][py-1] || e.TexelsWall[px-1][py+1]
	} else {
		return e.TexelsWall[px][py]
	}

}

// Converts a world pos to an integer position on the texels wall
func (e *Environment) worldPosToMapPos(p pixel.Vec) (int, int) {
	return int(p.X-0.5) + e.Radius, int(p.Y-0.5) + e.Radius
}

func (e *Environment) worldPosToClosestTexel(p pixel.Vec) pixel.Vec {
	return pixel.V(float64(int(p.X-0.5))+0.5, float64(int(p.Y-0.5))+0.5)
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	ra, ga, ba, aa := a.R, a.G, a.B, a.A
	rb, gb, bb, ab := b.R, b.G, b.B, b.A
	cr := uint8(float64(ra)*(1-t) + float64(rb)*t)
	cg := uint8(float64(ga)*(1-t) + float64(gb)*t)
	cb := uint8(float64(ba)*(1-t) + float64(bb)*t)
	ca := uint8(float64(aa)*(1-t) + float64(ab)*t)
	return color.RGBA{cr, cg, cb, ca}
}

func randomColor() color.RGBA {
	return color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255}
}
