package main

type SimulationParameters struct {
	MapParams               SimulationParametersMapGen           `json:"map_generation"`          // Environment generation
	PlantParams             SimulationParametersPlant            `json:"plant_growth"`            // Plant growth
	CreatureBaseMultipliers SimulationParametersCreatureBases    `json:"creature_base_values"`    // Multipliers (any number > 0)
	CreatureBalances        SimulationParametersCreatureBalances `json:"creature_balance_values"` // Balances (between 0 and 1)
	MutationParameters      MutationParameters                   `json:"mutation_parameters"`     // Mutation parameters
}

type SimulationParametersMapGen struct {
	PlantDensity           float64 `json:"plant_density"`            // The number of plants per unit area
	PlantCoverage          float64 `json:"plant_coverage"`           // The percentage of the map covered in plants
	MapRadius              int     `json:"map_radius"`               // The radius of the map
	CaveSize               float64 `json:"cave_size"`                // The size of the cave
	InitialCreaturesNumber int     `json:"initial_creatures_number"` // The number of creatures to start with
}

type SimulationParametersPlant struct {
	FoodGrowthDelay float64 `json:"food_growth_delay"` // The number of seconds between plant growth ticks
	GrownFoodEnergy float64 `json:"grown_food_energy"` // The amount of energy in a food from a plant
}

type SimulationParametersCreatureBases struct {
	MaxEnergy   float64 `json:"max_energy"`    // The energy a creature would have if its max energy multiplier was 1
	PushForce   float64 `json:"push_force"`    // The force a creature would have if its push force multiplier was 1
	Metabolism  float64 `json:"metabolism"`    // The energy a creature would lose if its metabolism multiplier was 1
	Vision      float64 `json:"vision"`        // The range a creature would have if its vision multiplier was 1
	PlantDrag   float64 `json:"plant_drag"`    // The drag a creature would have if its drag multiplier was 1
	FoodEatRate float64 `json:"food_eat_rate"` // The rate at which a creature would eat food if its food eat rate multiplier was 1
	Drag        float64 `json:"drag"`          // The drag a creature would have if its drag multiplier was 1
	AngularDrag float64 `json:"angular_drag"`  // The angular drag a creature would have if its angular drag multiplier was 1
	RotateForce float64 `json:"rotate_force"`  // The force a creature would have if its rotate force multiplier was 1
}

type SimulationParametersCreatureBalances struct {
	ConversionEfficiencySlopePlant float64 `json:"conversion_efficiency_slope_plant"` // The slope of the conversion efficiency curve for plants
	ConversionEfficiencySlopeMeat  float64 `json:"conversion_efficiency_slope_meat"`  // The slope of the conversion efficiency curve for meat
	DeathEnergyThreshold           float64 `json:"death_energy_threshold"`            // The percent energy a creature must have to survive
	PredatorEfficiencySlope        float64 `json:"predator_efficiency_slope"`         // The slope of the predator efficiency curve. At 0, effect is linear. At 1, only a perfect predator gets a boost
	PredatorMetabolismPercentage   float64 `json:"predator_percentage"`               // The maximum percentage drop in metabolism of predators
}

type MutationParameters struct {
	TraitMutationRate          float64 `json:"trait_mutation_rate"`          // The chance that a trait will mutate
	TraitMutationSize          float64 `json:"trait_mutation_size"`          // The size of a trait mutation
	SynapseMutationProbability float64 `json:"synapse_mutation_probability"` // The chance that a synapse will mutate
	SynapseMutationSize        float64 `json:"synapse_mutation_size"`        // The size of a synapse mutation
	SynapseGrowthProbability   float64 `json:"synapse_growth_probability"`   // The chance that a synapse will grow
	SynapseGrowthSize          float64 `json:"synapse_growth_size"`          // The size of a synapse growth
	NeuronGrowProbability      float64 `json:"neuron_grow_probability"`      // The chance that a neuron will grow
	SynapsePruneProbability    float64 `json:"synapse_prune_probability"`    // The chance that a synapse will be pruned
}

var GlobalSP = SimulationParameters{
	MapParams: SimulationParametersMapGen{
		MapRadius:              400,
		PlantDensity:           0.3,
		PlantCoverage:          0.8,
		CaveSize:               1,
		InitialCreaturesNumber: 300,
	},

	PlantParams: SimulationParametersPlant{
		FoodGrowthDelay: 30,
		GrownFoodEnergy: 5,
	},

	CreatureBaseMultipliers: SimulationParametersCreatureBases{
		MaxEnergy:   1,
		PushForce:   20,
		Metabolism:  0.015,
		Vision:      10,
		PlantDrag:   3,
		FoodEatRate: 5,
		Drag:        8,
		AngularDrag: 7,
		RotateForce: 10,
	},

	CreatureBalances: SimulationParametersCreatureBalances{
		ConversionEfficiencySlopePlant: 0.5,
		ConversionEfficiencySlopeMeat:  0.5,
		DeathEnergyThreshold:           0.2,
		PredatorEfficiencySlope:        0.7,
		PredatorMetabolismPercentage:   0.5,
	},
	MutationParameters: MutationParameters{
		TraitMutationRate:          0.2,
		TraitMutationSize:          0.1,
		SynapseMutationProbability: 0.2,
		SynapseMutationSize:        0.1,
		SynapseGrowthProbability:   0.15,
		SynapseGrowthSize:          0.5,
		NeuronGrowProbability:      0.05,
		SynapsePruneProbability:    0.1,
	},
}
