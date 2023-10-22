package main

import (
	"math"
	"math/rand"

	E "github.com/JoshPattman/goevo"
	"github.com/faiface/pixel"
)

var creatureSubstrate *E.LayeredSubstrate

func makeCreatureSubstrate() *E.LayeredSubstrate {
	numInputs := NewCreature(CreatureDNA{}).NumInputs()
	numHidden := 10
	numOutputs := 3

	inputs := make([]E.Pos, numInputs)
	hidden := make([]E.Pos, numHidden)
	outputs := make([]E.Pos, numOutputs)

	for i := 0; i < numInputs; i++ {
		inputs[i] = E.P(float64(i+1) / float64(numInputs)) // This will always leave position 0 for the bias
	}

	for i := 0; i < numHidden; i++ {
		hidden[i] = E.P(float64(i+1) / float64(numHidden))
	}

	for i := 0; i < numOutputs; i++ {
		outputs[i] = E.P(float64(i+1) / float64(numOutputs))
	}

	return E.NewLayeredSubstrate(
		[][]E.Pos{inputs, hidden, outputs},
		[]E.Activation{E.AcLin, E.AcReLU, E.AcTanh},
		E.P(0),
	)
}

type Creature struct {
	Pos                     pixel.Vec
	Vel                     pixel.Vec
	Radius                  float64
	Rot                     float64
	RotVel                  float64
	Energy                  float64
	DNA                     CreatureDNA
	debugFoodSensorValues   []float64
	debugAnimalSensorValues []float64
	debugWallSensorValues   []float64
	sensorAngles            []float64
	phenotype               E.Forwarder
	updateTimer             float64
	nnOutput                []float64
}

func NewCreature(dna CreatureDNA) *Creature {
	dna = dna.Validated()
	sa := make([]float64, 0)
	visionAngle := math.Pi
	numSensors := 5.0
	anglePerSensor := visionAngle / (numSensors - 1)
	for a := -visionAngle / 2; a <= visionAngle/2; a += anglePerSensor {
		sa = append(sa, a)
	}

	var pheno E.Forwarder
	if dna.Genotype != nil {
		pheno = E.NewPhenotype(dna.Genotype)
		pheno = creatureSubstrate.NewPhenotype(pheno)
	}
	return &Creature{
		Pos:          pixel.V(0, 0),
		Vel:          pixel.V(0, 0),
		Radius:       1 * dna.Size,
		Rot:          rand.Float64() * math.Pi * 2,
		Energy:       dna.MaxEnergy(),
		DNA:          dna,
		sensorAngles: sa,
		phenotype:    pheno,
		updateTimer:  rand.Float64() * GlobalSP.EnvironmentalParams.BrainUpdateDelay,
		nnOutput:     make([]float64, 3),
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

func (c *Creature) NumInputs() int {
	return len(c.sensorAngles)*3 + 2 + 1
}

func (c *Creature) Die(e *Environment) {
	e.Creatures.Remove(c)
	f := NewFood(c.Energy, false)
	f.Pos = c.Pos
	f.Rot = c.Rot
	e.Food.Add(f)
}

func (c *Creature) Fwd() pixel.Vec {
	return pixel.V(0, 1).Rotated(c.Rot)
}

func (c *Creature) Update(deltaTime float64, e *Environment, updateBrain bool) {
	// Update knowlege
	sight := c.DNA.VisionRange()
	neighbors := e.Creatures.Query(c.Pos, sight)
	nearbyFood := e.Food.Query(c.Pos, sight)
	// We just leave this as 10 because visibility does not make a difference to drag due to plants
	nearbyPlants := e.Plants.Query(c.Pos, 10)
	currentDepth := c.Pos.Len() / float64(e.Radius)
	currentDepthAlignment := c.Fwd().Dot(c.Pos.Unit())

	// Update non physical attributes
	c.Energy -= deltaTime * c.DNA.Metabolism()
	if c.Energy <= c.DNA.DeathEnergy() {
		c.Die(e)
		return
	}

	// Eat food if we are touching within an angle
	for _, f := range nearbyFood {
		offset := c.Pos.Sub(f.Pos)
		if offset.Len() < (c.Radius+f.Radius())/2 {
			if -offset.Unit().Dot(c.Fwd()) > 0.9 { // On mouth
				// Take energy from food
				takenEnergy := math.Min(f.Energy, c.DNA.FoodEatRate()*deltaTime)
				f.Energy -= takenEnergy
				if f.Energy <= 0 {
					e.Food.Remove(f)
				}
				// Use that energy
				if f.IsVeggie {
					c.Energy += c.DNA.PlantConversionEfficiency() * takenEnergy
				} else {
					c.Energy += c.DNA.MeatConversionEfficiency() * takenEnergy
				}
			}
			// Push the food away
			lenDiff := offset.Len() - (c.Radius+f.Radius())/2
			f.Pos = f.Pos.Add(offset.Unit().Scaled(15 * deltaTime * lenDiff))
		}
	}
	if c.Energy > c.DNA.MaxEnergy() {
		c.Energy = c.DNA.MaxEnergy()
	}

	// Setup the physics
	resultantForce := pixel.ZV
	resultantTorque := 0.0
	drag := GlobalSP.CreatureBaseMultipliers.Drag

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
	neighborsOnMouth := make([]*Creature, 0)
	{
		neighborBounceForce := pixel.ZV
		for _, n := range neighbors {
			diff := c.Pos.Sub(n.Pos)
			if n != c && diff.Len() < (c.Radius+n.Radius)/2 {
				overlap := diff.Len() - (c.Radius+n.Radius)/2
				neighborBounceForce = neighborBounceForce.Add(diff.Unit().Scaled(-overlap * 50))
				if -diff.Unit().Dot(c.Fwd()) > 0.9 { // On mouth
					neighborsOnMouth = append(neighborsOnMouth, n)
				}
			}
		}
		resultantForce = resultantForce.Add(neighborBounceForce)
	}

	// Check if we are touching a plant, and add drag if we are
	for _, p := range nearbyPlants {
		offset := c.Pos.Sub(p.Pos)
		if offset.Len() < (c.Radius+p.Radius)/2 {
			drag += c.DNA.PlantDrag()
			break
		}
	}

	// Detect food
	sensorFoodValues := make([]float64, 0)
	sensorAnimalValues := make([]float64, 0)
	sensorWallValues := make([]float64, 0)
	sensorAngles := make([]float64, 0)
	sensorWidth := c.sensorAngles[1] - c.sensorAngles[0]
	if updateBrain {
		for _, sensorAngle := range c.sensorAngles {
			// Find the sensor dir
			sensorDir := pixel.V(0, 1).Rotated(c.Rot + sensorAngle)
			// Set up the unsensed values
			sensorFoodValue := 0.0
			sensorAnimalValue := 0.0
			sensorWallValue := 0.0
			sensorWallDist := math.Inf(1)

			// Check Wall sensors
			sectionSamples := math.Round(sight * 2)
			sectionSampleLength := sight / sectionSamples
			for i := 0.0; i <= sectionSamples; i++ {
				dist := i * sectionSampleLength
				samplePos := c.Pos.Add(sensorDir.Scaled(dist))
				if e.sampleWallAt(samplePos, false) {
					sensorWallValue = 1 - dist/sight
					sensorWallDist = dist
					break
				}
			}

			// Check Food sensors
			for _, f := range nearbyFood {
				dirToFood := f.Pos.Sub(c.Pos)
				distToFood := dirToFood.Len()
				dotSensorDir := dirToFood.Dot(sensorDir)
				if distToFood < sensorWallDist-0.5 && dotSensorDir > 0 {
					allowedDistFromLine := math.Sin(sensorWidth) / 2 * distToFood
					distToLine := math.Abs(dirToFood.Sub(sensorDir.Scaled(dotSensorDir)).Len())
					if distToLine <= allowedDistFromLine {
						newValue := 1 - distToFood/sight
						if f.IsVeggie {
							newValue = c.DNA.PlantConversionEfficiency() * newValue
						} else {
							newValue = c.DNA.MeatConversionEfficiency() * newValue
						}
						newValue *= f.Energy / c.DNA.MaxEnergy() // Multiply by what percent that food could fill us up
						if newValue > sensorFoodValue {
							sensorFoodValue = newValue
						}
					}
				}
			}
			// Check Animal sensors
			for _, f := range neighbors {
				dirToAnimal := f.Pos.Sub(c.Pos)
				distToAnimal := dirToAnimal.Len()
				dotSensorDir := dirToAnimal.Dot(sensorDir)
				if distToAnimal < sensorWallDist-0.5 && dotSensorDir > 0 {
					allowedDistFromLine := math.Sin(sensorWidth) / 2 * distToAnimal
					distToLine := math.Abs(dirToAnimal.Sub(sensorDir.Scaled(dotSensorDir)).Len())
					if distToLine <= allowedDistFromLine {
						newValue := 1 - distToAnimal/sight
						if newValue > sensorAnimalValue {
							sensorAnimalValue = newValue
						}
					}
				}
			}

			// Add the values to the list
			sensorAngles = append(sensorAngles, sensorAngle)
			sensorFoodValues = append(sensorFoodValues, sensorFoodValue)
			sensorAnimalValues = append(sensorAnimalValues, sensorAnimalValue)
			sensorWallValues = append(sensorWallValues, sensorWallValue)
		}
		c.debugFoodSensorValues = sensorFoodValues
		c.sensorAngles = sensorAngles
		c.debugAnimalSensorValues = sensorAnimalValues
		c.debugWallSensorValues = sensorWallValues

		// Calculate neural net
		nnInput := make([]float64, 0)
		nnInput = append(nnInput, sensorFoodValues...)
		nnInput = append(nnInput, sensorAnimalValues...)
		nnInput = append(nnInput, sensorWallValues...)
		nnInput = append(nnInput, currentDepth, currentDepthAlignment)
		nnInput = append(nnInput, 1)
		c.nnOutput = c.phenotype.Forward(nnInput)
	}

	// Parse the output
	turn := c.nnOutput[0] * math.Pi / 2
	power := c.nnOutput[1]/2 + 0.5
	isAttack := c.nnOutput[2] > 0

	// Apply chosen motion
	forwardsPush := c.DNA.PushForce() * power
	resultantForce = resultantForce.Add(c.Fwd().Scaled(forwardsPush))
	resultantTorque += turn * GlobalSP.CreatureBaseMultipliers.RotateForce

	// Attack enemies if we want to
	if isAttack {
		for _, n := range neighborsOnMouth {
			n.Die(e)
		}
	}

	// Add the force and apply drag
	c.Vel = c.Vel.Add(resultantForce.Scaled(deltaTime)).Scaled(1 - drag*deltaTime)
	c.RotVel = (c.RotVel + resultantTorque*deltaTime) * (1 - GlobalSP.CreatureBaseMultipliers.AngularDrag*deltaTime)
	// Update pos and rot
	c.Pos = c.Pos.Add(c.Vel.Scaled(deltaTime))
	c.Rot += c.RotVel * deltaTime //c.Vel.Angle() - math.Pi/2
}

func (c *Creature) Child() *Creature {
	// Copy DNA
	dna := c.DNA.Copied()
	// Mutate traits
	if rand.Float64() < GlobalSP.MutationParameters.TraitMutationRate {
		dna.Diet += (rand.Float64()*2 - 1) * GlobalSP.MutationParameters.TraitMutationSize
	}
	if rand.Float64() < GlobalSP.MutationParameters.TraitMutationRate {
		dna.Size += (rand.Float64()*2 - 1) * GlobalSP.MutationParameters.TraitMutationSize
	}
	if rand.Float64() < GlobalSP.MutationParameters.TraitMutationRate {
		dna.Speed += (rand.Float64()*2 - 1) * GlobalSP.MutationParameters.TraitMutationSize
	}
	if rand.Float64() < GlobalSP.MutationParameters.TraitMutationRate {
		dna.Vision += (rand.Float64()*2 - 1) * GlobalSP.MutationParameters.TraitMutationSize
	}
	if rand.Float64() < GlobalSP.MutationParameters.TraitMutationRate {
		dna.Color = c.DNA.Color.Randomised(GlobalSP.MutationParameters.TraitMutationSize)
	}
	// Mutate brain
	maxReps := 4.0
	for i := 0; i < int(maxReps); i++ {
		if rand.Float64() < GlobalSP.MutationParameters.SynapseMutationProbability/maxReps {
			E.MutateRandomSynapse(dna.Genotype, GlobalSP.MutationParameters.SynapseMutationSize)
		}
	}
	for i := 0; i < int(maxReps); i++ {
		if rand.Float64() < GlobalSP.MutationParameters.SynapseGrowthProbability/maxReps {
			E.AddRandomSynapse(gtCounter, dna.Genotype, GlobalSP.MutationParameters.SynapseGrowthSize, false, 5)
		}
	}
	for i := 0; i < int(maxReps); i++ {
		if rand.Float64() < GlobalSP.MutationParameters.NeuronGrowProbability/maxReps {
			E.AddRandomNeuron(gtCounter, dna.Genotype, E.ChooseActivationFrom([]E.Activation{E.AcCos, E.AcSin, E.AcReLU, E.AcReLUM, E.AcTanh, E.AcSig, E.AcStep}))
		}
	}
	for i := 0; i < int(maxReps); i++ {
		if rand.Float64() < GlobalSP.MutationParameters.SynapsePruneProbability/maxReps {
			E.PruneRandomSynapse(dna.Genotype)
		}
	}
	// Create creture
	c1 := NewCreature(dna)
	return c1
}
