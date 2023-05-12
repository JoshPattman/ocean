package main

import (
	"math"

	"github.com/JoshPattman/goevo"
)

// No state, but carries info about how to make a creature
type CreatureDNA struct {
	// Multipliers
	Size   float64 `json:"size"`
	Speed  float64 `json:"speed"`
	Vision float64 `json:"vision"`

	// Balances
	Diet float64 `json:"diet"` // 0 = veggie, 1 = meat

	// Brain
	Genotype *goevo.Genotype `json:"brain"`

	// Cosmetic
	Color ColorHSV `json:"color"`
}

func (c CreatureDNA) MeatConversionEfficiency() float64 {
	return math.Pow(c.Diet, 1/(1-GlobalSP.ConversionEfficiencySlopeMeat))
}
func (c CreatureDNA) PlantConversionEfficiency() float64 {
	return math.Pow(1-c.Diet, 1/(1-GlobalSP.ConversionEfficiencySlopePlant))
}
func (c CreatureDNA) PredatoryMetabolismMultiplier() float64 {
	return 1 - (GlobalSP.PredatorMetabolismPercentage * math.Pow(c.Diet, 1/(1-GlobalSP.PredatorEfficiencySlope)))
}
func (c CreatureDNA) MaxEnergy() float64 {
	return GlobalSP.MaxEnergy * (c.Size * c.Size)
}
func (c CreatureDNA) Metabolism() float64 {
	return GlobalSP.Metabolism * (c.Size*c.Size + c.Vision + c.Speed) * c.PredatoryMetabolismMultiplier()
}
func (c CreatureDNA) FoodEatRate() float64 {
	return GlobalSP.FoodEatRate * c.Size
}
func (c CreatureDNA) PlantDrag() float64 {
	return GlobalSP.PlantDrag * c.Size
}
func (c CreatureDNA) DeathEnergy() float64 {
	return c.MaxEnergy() * GlobalSP.DeathEnergyThreshold
}
func (c CreatureDNA) VisionRange() float64 {
	return GlobalSP.Vision * c.Vision
}
func (c CreatureDNA) PushForce() float64 {
	return GlobalSP.PushForce * c.Speed
}

func (c CreatureDNA) Validated() CreatureDNA {
	newDNA := c
	newDNA.Diet = math.Min(math.Max(c.Diet, 0), 1)
	newDNA.Size = math.Max(c.Size, 0.1)
	newDNA.Speed = math.Max(c.Speed, 0.1)
	return newDNA
}

func (c CreatureDNA) Copied() CreatureDNA {
	newDNA := c
	newDNA.Genotype = goevo.NewGenotypeCopy(c.Genotype)
	return newDNA
}
