package main

type SimulationParameters struct {
	// Environment generation
	PlantDensity  float64 `json:"plant_density"`  // The number of plants per unit area
	PlantCoverage float64 `json:"plant_coverage"` // The percentage of the map covered in plants

	// Plant growth
	FoodGrowthDelay float64 `json:"food_growth_delay"` // The number of seconds between plant growth ticks
	GrownFoodEnergy float64 `json:"grown_food_energy"` // The amount of energy in a food from a plant

	// Multipliers (any number > 0)
	MaxEnergy   float64 `json:"max_energy"`    // The energy a creature would have if its max energy multiplier was 1
	PushForce   float64 `json:"push_force"`    // The force a creature would have if its push force multiplier was 1
	Metabolism  float64 `json:"metabolism"`    // The energy a creature would lose if its metabolism multiplier was 1
	Vision      float64 `json:"vision"`        // The range a creature would have if its vision multiplier was 1
	PlantDrag   float64 `json:"plant_drag"`    // The drag a creature would have if its drag multiplier was 1
	FoodEatRate float64 `json:"food_eat_rate"` // The rate at which a creature would eat food if its food eat rate multiplier was 1
	Drag        float64 `json:"drag"`          // The drag a creature would have if its drag multiplier was 1
	AngularDrag float64 `json:"angular_drag"`  // The angular drag a creature would have if its angular drag multiplier was 1
	RotateForce float64 `json:"rotate_force"`  // The force a creature would have if its rotate force multiplier was 1

	// Balances (between 0 and 1)
	ConversionEfficiencySlopePlant float64 `json:"conversion_efficiency_slope_plant"` // The slope of the conversion efficiency curve for plants
	ConversionEfficiencySlopeMeat  float64 `json:"conversion_efficiency_slope_meat"`  // The slope of the conversion efficiency curve for meat
	DeathEnergyThreshold           float64 `json:"death_energy_threshold"`            // The percent energy a creature must have to survive
	PredatorEfficiencySlope        float64 `json:"predator_efficiency_slope"`         // The slope of the predator efficiency curve. At 0, effect is linear. At 1, only a perfect predator gets a boost
	PredatorMetabolismPercentage   float64 `json:"predator_percentage"`               // The maximum percentage drop in metabolism of predators
}

var GlobalSP = SimulationParameters{
	PlantDensity:  0.3,
	PlantCoverage: 0.8,

	FoodGrowthDelay: 30,
	GrownFoodEnergy: 5,

	MaxEnergy:   1,
	PushForce:   20,
	Metabolism:  0.02,
	Vision:      10,
	PlantDrag:   3,
	FoodEatRate: 5,
	Drag:        8,
	AngularDrag: 7,
	RotateForce: 10,

	ConversionEfficiencySlopePlant: 0.5,
	ConversionEfficiencySlopeMeat:  0.5,
	DeathEnergyThreshold:           0.2,
	PredatorEfficiencySlope:        0.7,
	PredatorMetabolismPercentage:   0.5,
}
