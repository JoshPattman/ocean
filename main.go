package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/JoshPattman/goevo"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var (
	debugCreatureSensors int = 0 // 0 = off, 1 = food, 2 = creatures, 3 = walls
)

var (
	gtCounter *SaveLoadCounter
)

func main() {
	ensureDataDir()
	err := reloadSimParams()
	if err != nil {
		fmt.Println(err)
		data, _ := json.MarshalIndent(GlobalSP, "", "  ")
		if err := os.WriteFile(getParamsPath(), data, 0644); err != nil {
			panic(err)
		}
	}
	pixelgl.Run(run)
}

func run() {
	startTime := time.Now()

	// Setup goevo
	gtCounter = &SaveLoadCounter{}
	gtOrig := goevo.NewGenotype(gtCounter, NewCreature(CreatureDNA{}).NumInputs(), 3, goevo.ActivationLinear, goevo.ActivationTanh)

	// Setup Environment
	env := NewEnvironment(GlobalSP.MapParams.MapRadius)
	//env.ScatterFood(0.01)
	for i := 0; i < GlobalSP.MapParams.InitialCreaturesNumber; i++ {
		gt := goevo.NewGenotypeCopy(gtOrig)
		goevo.AddRandomSynapse(gtCounter, gt, 1, false, 5)
		goevo.AddRandomSynapse(gtCounter, gt, 1, false, 5)
		goevo.AddRandomSynapse(gtCounter, gt, 1, false, 5)
		c := NewCreature(CreatureDNA{
			Size:     1 + (rand.Float64()-0.5)*2,
			Speed:    1 + (rand.Float64()-0.5)*2,
			Diet:     rand.Float64(),
			Genotype: gt,
			Color:    RandomHSV(),
			Vision:   1,
		})
		c.Pos = pixel.V(math.Sqrt(rand.Float64())*float64(GlobalSP.MapParams.MapRadius)*0.25, 0).Rotated(rand.Float64() * 2 * math.Pi)
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
	imd := imdraw.New(nil)

	// Create UI elements
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	numCreaturesText := text.New(pixel.ZV, atlas)
	timerText := text.New(pixel.ZV, atlas)

	// Create creature stats elements
	creatureStats := text.New(pixel.ZV, atlas)
	var activeCreature *Creature
	vis := goevo.NewGenotypeVisualiser()
	vis.ImgSizeX = 400
	vis.ImgSizeY = 800
	vis.NeuronSize = 5
	var currentCreatureBrainSprite *pixel.Sprite
	instructionsText := text.New(pixel.ZV, atlas)
	isActiveGrabbed := false

	// Define player control variables
	offset := pixel.V(500, 400)
	scale := 4.0

	// Main Loop
	for !win.Closed() {
		// Default instructions
		instructionsText.Clear()
		fmt.Fprintf(instructionsText, "(I)mport Creature, Sca(t)ter Food, (L)oad Sim Params")
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
		if win.Pressed(pixelgl.KeyL) {
			err := reloadSimParams()
			if err != nil {
				fmt.Println(err)
			}
		}

		newCreatures := make([]*Creature, 0)
		for _, c := range env.Creatures.Objects {
			me := c.DNA.MaxEnergy()
			if c.Energy >= me*0.8 && rand.Float64() < (1/60.0)/5 {
				c1 := c.Child()
				c1.Pos = c.Pos
				c1.Energy = me * 0.79
				c.Energy = me * 0.79
				newCreatures = append(newCreatures, c1)
			}
		}
		for _, c1 := range newCreatures {
			env.Creatures.Add(c1)
		}

		if win.JustPressed(pixelgl.KeyT) {
			env.ScatterFood(0.01)
		}

		// Update Sim
		// Grow new food on plants
		for _, p := range env.Plants.Objects {
			if rand.Float64() < (1/60.0)/GlobalSP.PlantParams.FoodGrowthDelay {
				// Check if there is already a food under us
				if len(env.Food.Query(p.Pos, 0.1)) == 0 {
					energy := math.Pow(p.Fertility, 3) * GlobalSP.PlantParams.GrownFoodEnergy
					f := NewFood(energy, true)
					f.Pos = p.Pos
					f.Rot = rand.Float64() * 2 * math.Pi
					env.Food.Add(f)
				}
			}
		}
		// Decay Food
		for _, f := range env.Food.Objects {
			f.Energy -= GlobalSP.EnvironmentalParams.FoodDecayRate * (1 / 60.0)
			if f.Energy <= 0 {
				env.Food.Remove(f)
			}
		}

		// Update hash maps
		env.Creatures.Refresh()
		env.Food.Refresh()
		// We dont need to update the plants map as they never move
		// Update creatures
		for _, c := range env.Creatures.Objects {
			c.updateTimer += 1 / 60.0
			if c.updateTimer >= GlobalSP.EnvironmentalParams.BrainUpdateDelay {
				c.updateTimer -= GlobalSP.EnvironmentalParams.BrainUpdateDelay
				c.Update(1/60.0, env, true)
			} else {
				c.Update(1/60.0, env, false)
			}
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
			creatureSprite.DrawColorMask(creatureBatch, pixel.IM.Scaled(pixel.ZV, c.Radius/creatureSprite.Frame().W()).Rotated(pixel.ZV, c.Rot).Moved(c.Pos).Moved(offset).Scaled(win.Bounds().Center(), scale), c.DNA.Color.ToColor())
		}
		creatureBatch.Draw(win)
		// Draw plants
		for _, p := range env.Plants.Objects {
			colorGreen := lerpColor(colornames.Lightgreen, colornames.Darkgreen, p.Shading)
			colorBrown := lerpColor(colornames.Yellow, colornames.Brown, p.Shading)
			colorMask := lerpColor(colorGreen, colorBrown, 1-p.Fertility)
			plantSprite.DrawColorMask(plantBatch, pixel.IM.Rotated(pixel.ZV, p.Rot).Scaled(pixel.ZV, p.Radius/plantSprite.Frame().W()).Moved(p.Pos).Moved(offset).Scaled(win.Bounds().Center(), scale), colorMask)
		}
		plantBatch.Draw(win)

		// UI
		// Clear Stats
		timerText.Clear()
		numCreaturesText.Clear()
		// Update Stats
		fmt.Fprintf(timerText, "Sim Time: %s", time.Since(startTime).String())
		fmt.Fprintf(numCreaturesText, "Num Creatures: %d\nNum Food: %d", len(env.Creatures.Objects), len(env.Food.Objects))
		// Draw Stats
		timerText.Draw(win, pixel.IM.Moved(pixel.V(10, win.Bounds().H()-20)))
		numCreaturesText.Draw(win, pixel.IM.Moved(pixel.V(10, win.Bounds().H()-40)))

		// Creature UI
		// Find the creature under the mouse
		mousePos := win.MousePosition().Sub(win.Bounds().Center()).Scaled(1 / scale).Add(win.Bounds().Center()).Sub(offset)
		pressedNumKey := getJustPressedNumKey(win)
		if win.JustPressed(pixelgl.MouseButtonLeft) {
			creatureUnderMouse := env.Creatures.Query(mousePos, 1)
			isActiveGrabbed = false
			if len(creatureUnderMouse) > 0 {
				activeCreature = creatureUnderMouse[0]
				nnimg := vis.DrawImage(activeCreature.DNA.Genotype)
				nnPic := pixel.PictureDataFromImage(nnimg)
				currentCreatureBrainSprite = pixel.NewSprite(nnPic, nnPic.Bounds())
			} else {
				activeCreature = nil
				currentCreatureBrainSprite = nil
			}
		}
		if activeCreature != nil {
			instructionsText.Clear()
			fmt.Fprintf(instructionsText, "Sca(t)ter Food, (K)ill, (C)lone, (F)eed, (G)rab, (R)andomize Color, Exp(o)rt Creature, (I)mport Creature, (L)oad Sim Params")
			// Update actions
			if win.JustPressed(pixelgl.KeyK) {
				activeCreature.Die(env)
			}
			if win.JustPressed(pixelgl.KeyC) {
				newDNA := activeCreature.DNA.Copied()
				newCreature := NewCreature(newDNA)
				newCreature.Pos = activeCreature.Pos
				env.Creatures.Add(newCreature)
			}
			if win.JustPressed(pixelgl.KeyF) {
				activeCreature.Energy = activeCreature.DNA.MaxEnergy()
			}
			if win.JustPressed(pixelgl.KeyG) || win.JustPressed(pixelgl.MouseButtonRight) {
				isActiveGrabbed = !isActiveGrabbed
			}
			if isActiveGrabbed {
				activeCreature.Pos = mousePos
				activeCreature.Vel = pixel.ZV
			}
			if win.JustPressed(pixelgl.KeyR) {
				activeCreature.DNA.Color = RandomHSV()
			}
			if win.JustPressed(pixelgl.KeyF1) {
				debugCreatureSensors = 0
			}
			if win.JustPressed(pixelgl.KeyF2) {
				debugCreatureSensors = 1
			}
			if win.JustPressed(pixelgl.KeyF3) {
				debugCreatureSensors = 2
			}
			if win.JustPressed(pixelgl.KeyF4) {
				debugCreatureSensors = 3
			}
			if win.Pressed(pixelgl.KeyO) {
				instructionsText.Clear()
				fmt.Fprintf(instructionsText, "Press A Number Key To Save The Creature's DNA To That Slot")
			}
			if win.Pressed(pixelgl.KeyO) && pressedNumKey != -1 {
				serialisedDNA, err := json.MarshalIndent(activeCreature.DNA, "", "  ")
				if err != nil {
					fmt.Println(err)
				} else {
					ensureDataDir()
					err = os.WriteFile(getSaveSlotPath(pressedNumKey), serialisedDNA, 0644)
					if err != nil {
						fmt.Println(err)
					}
				}
			}

			// Draw stats
			creatureStats.Clear()
			creatureStats.Color = colornames.White
			fmt.Fprintf(creatureStats, "Creature Stats:\n"+
				"Energy ------------ %.2f/%.2f\n"+
				"Energy (Adjusted) - %.2f/%.2f\n"+
				"Size -------------- %.2f\n"+
				"Speed ------------- %.2f\n"+
				"Sight Range ------- %.2f\n"+
				"Diet -------------- %.2f\n"+
				"Plant Efficiency -- %.2f\n"+
				"Meat Efficiency --- %.2f\n"+
				"Predator Met Mult - %.2f\n"+
				"Metabolism -------- %.2f\n",

				activeCreature.Energy, activeCreature.DNA.MaxEnergy(),
				activeCreature.Energy-activeCreature.DNA.DeathEnergy(), activeCreature.DNA.MaxEnergy()-activeCreature.DNA.DeathEnergy(),
				activeCreature.DNA.Size,
				activeCreature.DNA.Speed,
				activeCreature.DNA.Vision,
				activeCreature.DNA.Diet,
				activeCreature.DNA.PlantConversionEfficiency(),
				activeCreature.DNA.MeatConversionEfficiency(),
				activeCreature.DNA.PredatoryMetabolismMultiplier(),
				activeCreature.DNA.Metabolism())

			statsLoc := pixel.V(win.Bounds().W()-250, win.Bounds().H()-20)
			// Background box
			imd.Clear()
			imd.Color = color.RGBA{0, 0, 0, 150}
			imd.Push(statsLoc.Add(pixel.V(0, 10)))
			imd.Push(statsLoc.Add(pixel.V(0, -130)))
			imd.Push(statsLoc.Add(pixel.V(250, -130)))
			imd.Push(statsLoc.Add(pixel.V(250, 10)))
			imd.Polygon(0)
			// Creature circle
			imd.Color = colornames.White
			imd.Push(activeCreature.Pos.Add(offset).Sub(win.Bounds().Center()).Scaled(scale).Add(win.Bounds().Center()))
			imd.Circle(activeCreature.DNA.VisionRange()*scale, 2)
			imd.Draw(win)
			// Debug Sensors
			if debugCreatureSensors != 0 {
				imd.Clear()
				var sensorValues []float64
				var sensorOffColor color.RGBA
				switch debugCreatureSensors {
				case 1:
					sensorValues = activeCreature.debugFoodSensorValues
					sensorOffColor = colornames.Green
				case 2:
					sensorValues = activeCreature.debugAnimalSensorValues
					sensorOffColor = colornames.Blue
				case 3:
					sensorValues = activeCreature.debugWallSensorValues
					sensorOffColor = colornames.Black
				}
				for i := range activeCreature.sensorAngles {
					imd.Color = lerpColor(sensorOffColor, colornames.Red, sensorValues[i])
					imd.Push(activeCreature.Pos.Add(offset).Sub(win.Bounds().Center()).Scaled(scale).Add(win.Bounds().Center()))
					imd.Push(activeCreature.Pos.Add(pixel.V(0, 10).Rotated(activeCreature.sensorAngles[i] + activeCreature.Rot)).Add(offset).Sub(win.Bounds().Center()).Scaled(scale).Add(win.Bounds().Center()))
					imd.Line(2)
				}
				imd.Draw(win)
			}
			// Stats
			creatureStats.Draw(win, pixel.IM.Moved(statsLoc))

			// Draw neural network
			imd.Clear()
			imd.Color = color.RGBA{0, 0, 0, 150}
			imd.Push(pixel.V(win.Bounds().W()-200, 400))
			imd.Push(pixel.V(win.Bounds().W()-200, 0))
			imd.Push(pixel.V(win.Bounds().W(), 0))
			imd.Push(pixel.V(win.Bounds().W(), 400))
			imd.Polygon(0)
			imd.Draw(win)
			currentCreatureBrainSprite.Draw(win, pixel.IM.Scaled(pixel.ZV, 0.5).Moved(pixel.V(win.Bounds().W()-100, 200)))

		}

		// Check for import
		if win.Pressed(pixelgl.KeyI) {
			instructionsText.Clear()
			fmt.Fprintf(instructionsText, "Press A Number Key To Load A Creature's DNA From That Slot")
		}
		if win.Pressed(pixelgl.KeyI) && pressedNumKey != -1 {
			serialisedDNA, err := os.ReadFile(getSaveSlotPath(pressedNumKey))
			if err != nil {
				fmt.Println("No creature DNA file found")
			} else {
				var dna CreatureDNA
				dna.Genotype = goevo.NewGenotypeEmpty()
				err = json.Unmarshal(serialisedDNA, &dna)
				if err != nil {
					fmt.Println(err)
				} else {
					activeCreature = NewCreature(dna)
					env.Creatures.Add(activeCreature)
					isActiveGrabbed = true
					nnimg := vis.DrawImage(activeCreature.DNA.Genotype)
					nnPic := pixel.PictureDataFromImage(nnimg)
					currentCreatureBrainSprite = pixel.NewSprite(nnPic, nnPic.Bounds())
				}
				fmt.Println("Counter was", gtCounter.c)
				gtCounter.SafeWith(dna.Genotype)
				fmt.Println("Counter is", gtCounter.c)
			}
		}

		// Draw instructions
		instructionsText.Draw(win, pixel.IM.Moved(pixel.V(math.Round(win.Bounds().W()/2-instructionsText.Bounds().Center().X), 5)))

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

func getJustPressedNumKey(win *pixelgl.Window) int {
	if win.JustPressed(pixelgl.Key1) {
		return 1
	} else if win.JustPressed(pixelgl.Key2) {
		return 2
	} else if win.JustPressed(pixelgl.Key3) {
		return 3
	} else if win.JustPressed(pixelgl.Key4) {
		return 4
	} else if win.JustPressed(pixelgl.Key5) {
		return 5
	} else if win.JustPressed(pixelgl.Key6) {
		return 6
	} else if win.JustPressed(pixelgl.Key7) {
		return 7
	} else if win.JustPressed(pixelgl.Key8) {
		return 8
	} else if win.JustPressed(pixelgl.Key9) {
		return 9
	} else if win.JustPressed(pixelgl.Key0) {
		return 0
	} else {
		return -1
	}
}

func ensureDataDir() {
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		os.Mkdir("data", 0755)
	}
}

func getSaveSlotPath(slot int) string {
	return "data/creature_dna_" + strconv.Itoa(slot) + ".json"
}

func getParamsPath() string {
	return "data/simulation_params.json"
}

func reloadSimParams() error {
	data, err := os.ReadFile(getParamsPath())
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &GlobalSP)
	if err != nil {
		return err
	}
	return nil
}
