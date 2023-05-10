package main

import (
	"math"

	"github.com/faiface/pixel"
)

type Creature struct {
	Pos    pixel.Vec
	Vel    pixel.Vec
	Radius float64
	Rot    float64
	Energy float64
	DNA    CreatureDNA
}

// No state, but carries info about how to make a creature
type CreatureDNA struct {
	// Multipliers
	Size  float64
	Speed float64

	// Balances
	Diet float64 // 0 = veggie, 1 = meat
}

func (c CreatureDNA) MaxEnergy() float64 { return c.Size * c.Size * c.Size }
func (c CreatureDNA) EnergyDecreaseRate() float64 {
	return c.Size*c.Size*c.Speed*c.Speed + SPIdleEnergyDecrease
}
func (c CreatureDNA) FoodDrainRate() float64 { return c.Size * c.Size }
func (c CreatureDNA) PlantDrag() float64     { return c.Size }

func (c CreatureDNA) Validated() CreatureDNA {
	newDNA := c
	newDNA.Diet = math.Min(math.Max(c.Diet, 0), 1)
	newDNA.Size = math.Max(c.Size, 0.1)
	newDNA.Speed = math.Max(c.Speed, 0.1)
	return newDNA
}

func NewCreature(dna CreatureDNA) *Creature {
	dna = dna.Validated()
	return &Creature{
		Pos:    pixel.V(0, 0),
		Vel:    pixel.V(0, 0),
		Radius: 1 * dna.Size,
		Rot:    0,
		Energy: dna.MaxEnergy(),
		DNA:    dna,
	}
}

func (c *Creature) HMPos() pixel.Vec {
	return c.Pos
}

func (c *Creature) Eq(o HashMappable) bool {
	if o == nil {
		return false
	}
	return c.Pos == o.(*Creature).Pos
}

func (c *Creature) Update(deltaTime float64, e *Environment) {
	// Update knowlege
	sight := 10.0
	neighbors := e.Creatures.Query(c.Pos, sight)
	nearbyFood := e.Food.Query(c.Pos, sight)
	// We just leave this as 10 because visibility does not make a difference to drag due to plants
	nearbyPlants := e.Plants.Query(c.Pos, math.Max(10, sight))

	// Update non physical attributes
	c.Energy -= SPEnergyDecrease * deltaTime * c.DNA.EnergyDecreaseRate()
	if c.Energy <= 0 {
		e.Creatures.Remove(c)
		f := NewFood(c.DNA.MaxEnergy(), false)
		f.Pos = c.Pos
		f.Rot = c.Rot
		e.Food.Add(f)
		return
	}

	// Eat food if we are touching
	for _, f := range nearbyFood {
		offset := c.Pos.Sub(f.Pos)
		if offset.Len() < (c.Radius+f.Radius())/2 {
			// Take energy from food
			takenEnergy := math.Min(f.Energy, SPFoodDrainRate*c.DNA.FoodDrainRate()*deltaTime)
			f.Energy -= takenEnergy
			if f.Energy <= 0 {
				e.Food.Remove(f)
			}
			// Use that energy
			if f.IsVeggie {
				c.Energy += (1 - c.DNA.Diet) * takenEnergy
			} else {
				c.Energy += c.DNA.Diet * takenEnergy
			}
			// Push the food away
			lenDiff := offset.Len() - (c.Radius+f.Radius())/2
			f.Pos = f.Pos.Add(offset.Unit().Scaled(5 * deltaTime * lenDiff))
		}
	}
	if c.Energy > c.DNA.MaxEnergy() {
		c.Energy = c.DNA.MaxEnergy()
	}

	// Setup the physics
	resultantForce := pixel.ZV
	drag := SPDrag

	// Wall collisions
	{
		circleCenters := make([]pixel.Vec, 0)
		// Find the walls around us
		for xo := -2; xo <= 2; xo++ {
			for yo := -2; yo <= 2; yo++ {
				p := pixel.V(float64(xo), float64(yo)).Add(c.Pos)
				p = e.worldPosToClosestTexel(p)
				if e.sampleWallAt(p, false) {
					circleCenters = append(circleCenters, p)
				}
			}
		}
		bounceForce := pixel.V(0, 0)
		isTouching := false
		for _, p := range circleCenters {
			if c.Pos.Sub(p).Len() < (c.Radius+1)/2 {
				bounceForce = bounceForce.Add(c.Pos.Sub(p).Unit().Scaled(100))
				isTouching = true
			}
		}
		resultantForce = resultantForce.Add(bounceForce)
		if isTouching {
			// Friction
			drag += 0.2
		}
	}

	// Bounce off neighbors
	{
		neighborBounceForce := pixel.ZV
		for _, n := range neighbors {
			if n != c && c.Pos.Sub(n.Pos).Len() < (c.Radius+n.Radius)/2 {
				neighborBounceForce = neighborBounceForce.Add(c.Pos.Sub(n.Pos).Unit().Scaled(10))
			}
		}
		resultantForce = resultantForce.Add(neighborBounceForce)
	}

	// Check if we are touching a plant, and add drag if we are
	for _, p := range nearbyPlants {
		offset := c.Pos.Sub(p.Pos)
		if offset.Len() < (c.Radius+p.Radius)/2 {
			drag += SPPlantDrag * c.DNA.PlantDrag()
			break
		}
	}

	// Apply chosen motion
	forwardsPush := c.DNA.Speed * SPPropulsionForce
	resultantForce = resultantForce.Add(pixel.V(0, 1).Rotated(c.Rot).Scaled(forwardsPush))

	// Add the force and apply drag
	c.Vel = c.Vel.Add(resultantForce.Scaled(deltaTime)).Scaled(1 - drag*deltaTime)
	// Update pos and rot
	c.Pos = c.Pos.Add(c.Vel.Scaled(deltaTime))
	c.Rot = c.Vel.Angle() - math.Pi/2
}
