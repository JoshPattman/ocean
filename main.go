package main

import (
	"fmt"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

// Global Sim Params
var (
	SPEnergyDecrease     float64 = 0.05
	SPDrag               float64 = 8
	SPPropulsionForce    float64 = 20
	SPFoodDrainRate      float64 = 5
	SPIdleEnergyDecrease float64 = 0.01
	SPPlantDensity       float64 = 0.1
	SPPlantDrag          float64 = 3
	SPDeathEnergy        float64 = 1
)

func main() {
	pixelgl.Run(run)
}

func run() {
	startTime := time.Now()

	// Setup Environment
	env := NewEnvironment(500)
	env.ScatterFood(0.01)
	for i := 0; i < 200; i++ {
		c := NewCreature(CreatureDNA{
			Size:  1.5 + (rand.Float64()-0.5)*2,
			Speed: 1.5 + (rand.Float64()-0.5)*2,
			Diet:  rand.Float64(),
		})
		c.Pos = pixel.V(rand.Float64()*30-15, rand.Float64()*30-15)
		env.Creatures.Add(c)
	}

	// Setup Window
	cfg := pixelgl.WindowConfig{
		Title:     "Evo Sim",
		Bounds:    pixel.R(0, 0, 1024, 768),
		VSync:     true,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Load Sprites
	terrainSprite := env.GetTerrainSprite()
	veggieFoodSprite, meatFoodSprite, foodPic := getFoodSprites()
	creatureSprite, creaturePic := getCreatureSprite()
	plantSprite, plantPic := getPlantSprite()

	// Create Batch Renderers
	foodBatch := pixel.NewBatch(&pixel.TrianglesData{}, foodPic)
	creatureBatch := pixel.NewBatch(&pixel.TrianglesData{}, creaturePic)
	plantBatch := pixel.NewBatch(&pixel.TrianglesData{}, plantPic)

	// Create UI elements
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	numCreaturesText := text.New(pixel.ZV, atlas)
	timerText := text.New(pixel.ZV, atlas)

	// Define player control variables
	offset := pixel.V(500, 400)
	scale := 4.0

	// Main Loop
	for !win.Closed() {
		// Update user controls
		if win.Pressed(pixelgl.KeyA) {
			offset.X += 10 / scale
		}
		if win.Pressed(pixelgl.KeyD) {
			offset.X -= 10 / scale
		}
		if win.Pressed(pixelgl.KeyW) {
			offset.Y -= 10 / scale
		}
		if win.Pressed(pixelgl.KeyS) {
			offset.Y += 10 / scale
		}
		if win.Pressed(pixelgl.KeyQ) {
			scale /= 1.01
		}
		if win.Pressed(pixelgl.KeyE) {
			scale *= 1.01
		}

		if win.JustPressed(pixelgl.KeyM) || len(env.Creatures.Objects) < 50 {
			newCreatures := make([]*Creature, 0)
			for _, c := range env.Creatures.Objects {
				dna := c.DNA
				dna.Diet += (rand.Float64()*2 - 1) * 0.2
				dna.Size += (rand.Float64()*2 - 1) * 0.2
				dna.Speed += (rand.Float64()*2 - 1) * 0.2
				c1 := NewCreature(dna)
				c1.Pos = c.Pos
				newCreatures = append(newCreatures, c1)
			}
			for _, c1 := range newCreatures {
				env.Creatures.Add(c1)
			}
		}

		if win.JustPressed(pixelgl.KeyN) {
			env.ScatterFood(0.01)
		}

		// Update Sim
		// Update hash maps
		env.Creatures.Refresh()
		env.Food.Refresh()
		// We dont need to update the plants map as they never move
		// Update creatures
		for _, c := range env.Creatures.Objects {
			c.Update(1/60.0, env)
		}

		// Render
		// Clear window
		win.Clear(colornames.Black)
		foodBatch.Clear()
		creatureBatch.Clear()
		plantBatch.Clear()
		// Draw terrain
		terrainSprite.Draw(win, pixel.IM.Moved(offset).Scaled(win.Bounds().Center(), scale))
		// Draw food
		for _, f := range env.Food.Objects {
			var s *pixel.Sprite
			if f.IsVeggie {
				s = veggieFoodSprite
			} else {
				s = meatFoodSprite
			}
			s.Draw(foodBatch, pixel.IM.Rotated(pixel.ZV, f.Rot).Scaled(pixel.ZV, f.Radius()/s.Frame().W()).Moved(f.Pos).Moved(offset).Scaled(win.Bounds().Center(), scale))
		}
		foodBatch.Draw(win)
		// Draw creatures
		for _, c := range env.Creatures.Objects {
			creatureSprite.Draw(creatureBatch, pixel.IM.Scaled(pixel.ZV, c.Radius/creatureSprite.Frame().W()).Rotated(pixel.ZV, c.Rot).Moved(c.Pos).Moved(offset).Scaled(win.Bounds().Center(), scale))
		}
		creatureBatch.Draw(win)
		// Draw plants
		for _, p := range env.Plants.Objects {
			l := 255 - p.Shading
			plantSprite.DrawColorMask(plantBatch, pixel.IM.Rotated(pixel.ZV, p.Rot).Scaled(pixel.ZV, p.Radius/plantSprite.Frame().W()).Moved(p.Pos).Moved(offset).Scaled(win.Bounds().Center(), scale), color.RGBA{l, l, l, 255})
		}
		plantBatch.Draw(win)

		// UI
		// Clear UI
		timerText.Clear()
		numCreaturesText.Clear()
		// Update UI
		fmt.Fprintf(timerText, "Sim Time: %s", time.Since(startTime).String())
		fmt.Fprintf(numCreaturesText, "Num Creatures: %d", len(env.Creatures.Objects))
		// Draw UI
		timerText.Draw(win, pixel.IM.Moved(pixel.V(10, win.Bounds().H()-20)))
		numCreaturesText.Draw(win, pixel.IM.Moved(pixel.V(10, win.Bounds().H()-40)))

		// Update window
		win.Update()
	}
}

func getFoodSprites() (*pixel.Sprite, *pixel.Sprite, pixel.Picture) {
	f, err := os.Open("sprites/food.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}
	pic := pixel.PictureDataFromImage(img)
	return pixel.NewSprite(pic, pixel.Rect{Min: pixel.V(0, 0), Max: pixel.V(8, 8)}),
		pixel.NewSprite(pic, pixel.Rect{Min: pixel.V(8, 0), Max: pixel.V(16, 8)}),
		pic
}

func getCreatureSprite() (*pixel.Sprite, pixel.Picture) {
	f, err := os.Open("sprites/creature.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}
	pic := pixel.PictureDataFromImage(img)
	return pixel.NewSprite(pic, pic.Bounds()), pic
}

func getPlantSprite() (*pixel.Sprite, pixel.Picture) {
	f, err := os.Open("sprites/plant.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}
	pic := pixel.PictureDataFromImage(img)
	return pixel.NewSprite(pic, pic.Bounds()), pic
}
