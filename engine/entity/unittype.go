package entity

import (
	"github.com/bklimczak/tanks/engine/resource"
	"image/color"
)

type UnitType int

const (
	UnitTypeBasic UnitType = iota
	UnitTypeConstructor
	UnitTypeTank
	UnitTypeScout
	UnitTypeMechConstructor
)

func (t UnitType) String() string {
	switch t {
	case UnitTypeConstructor:
		return "Constructor"
	case UnitTypeTank:
		return "Tank"
	case UnitTypeScout:
		return "Scout"
	case UnitTypeMechConstructor:
		return "Mech Constructor"
	default:
		return "Unit"
	}
}

type UnitDef struct {
	Type        UnitType
	Name        string
	Description string
	Size        float64
	Speed       float64
	Color       color.Color
	Cost        map[resource.Type]float64
	BuildTime   float64
	Health      float64
	Damage      float64
	Range       float64
	FireRate    float64
	VisionRange float64
	RepairRate  float64 // Health per second when repairing
	RepairRange float64 // Range to repair units

	// Capabilities - data-driven instead of hardcoded type checks
	CanConstruct        bool           // Can build structures
	CanRepairUnits      bool           // Can repair other units
	BuildableTypes      []BuildingType // Which buildings this unit can construct (if CanConstruct)
	RotationSpeed       float64        // Body rotation speed for movement
	TurretRotationSpeed float64        // Turret rotation speed (for tanks)
}

// AllBuildableTypes is a convenience list for constructor units that can build everything
var AllBuildableTypes = []BuildingType{
	BuildingTankFactory,
	BuildingMechFactory,
	BuildingSolarPanel,
	BuildingMetalExtractor,
	BuildingMetalStorage,
	BuildingEnergyStorage,
	BuildingMetalStorageLarge,
	BuildingEnergyStorageLarge,
	BuildingLaserTower,
}

var UnitDefs = map[UnitType]*UnitDef{
	UnitTypeTank: {
		Type:        UnitTypeTank,
		Name:        "Tank",
		Description: "Heavy combat unit",
		Size:        22,
		Speed:       2.5,
		Color:       color.RGBA{80, 120, 80, 255},
		Cost: map[resource.Type]float64{
			resource.Metal:  100,
			resource.Energy: 50,
		},
		BuildTime:           5.0,
		Health:              100,
		Damage:              15,
		Range:               150,
		FireRate:            1.0,
		VisionRange:         250,
		RotationSpeed:       0.03,  // Slow body rotation (tanks are heavy)
		TurretRotationSpeed: 0.08,  // Faster turret rotation
	},
	UnitTypeScout: {
		Type:        UnitTypeScout,
		Name:        "Scout",
		Description: "Fast recon unit, light weapons",
		Size:        16,
		Speed:       5.0,
		Color:       color.RGBA{180, 180, 100, 255},
		Cost: map[resource.Type]float64{
			resource.Metal:  30,
			resource.Energy: 20,
		},
		BuildTime:     2.0,
		Health:        40,
		Damage:        5,
		Range:         100,
		FireRate:      2.0,
		VisionRange:   400,
		RotationSpeed: 0.2,
	},
	UnitTypeConstructor: {
		Type:        UnitTypeConstructor,
		Name:        "Constructor",
		Description: "Builds structures",
		Size:        24,
		Speed:       1.5,
		Color:       color.RGBA{200, 180, 50, 255},
		Cost: map[resource.Type]float64{
			resource.Metal:  150,
			resource.Energy: 75,
		},
		BuildTime:      8.0,
		Health:         60,
		Damage:         0,
		Range:          0,
		FireRate:       0,
		VisionRange:    200,
		RotationSpeed:  0.1,
		CanConstruct:   true,
		BuildableTypes: AllBuildableTypes,
	},
	UnitTypeMechConstructor: {
		Type:        UnitTypeMechConstructor,
		Name:        "Mech Constructor",
		Description: "Builds structures and repairs units",
		Size:        28,
		Speed:       2.0,
		Color:       color.RGBA{180, 140, 200, 255}, // Purple-ish
		Cost: map[resource.Type]float64{
			resource.Metal:  200,
			resource.Energy: 100,
		},
		BuildTime:      10.0,
		Health:         80,
		Damage:         0,
		Range:          0,
		FireRate:       0,
		VisionRange:    250,
		RepairRate:     20,  // 20 HP per second
		RepairRange:    100, // Must be within 100 units to repair
		RotationSpeed:  0.08,
		CanConstruct:   true,
		CanRepairUnits: true,
		BuildableTypes: AllBuildableTypes,
	},
}

// GetProducibleUnits returns the units that a building type can produce (data-driven)
func GetProducibleUnits(buildingType BuildingType) []*UnitDef {
	def := BuildingDefs[buildingType]
	if def == nil || !def.IsFactory || len(def.ProducesUnits) == 0 {
		return nil
	}
	result := make([]*UnitDef, 0, len(def.ProducesUnits))
	for _, unitType := range def.ProducesUnits {
		if unitDef := UnitDefs[unitType]; unitDef != nil {
			result = append(result, unitDef)
		}
	}
	return result
}

type BuildingType int

const (
	BuildingTankFactory BuildingType = iota
	BuildingSolarPanel
	BuildingMetalExtractor
	BuildingMetalStorage
	BuildingEnergyStorage
	BuildingMetalStorageLarge
	BuildingEnergyStorageLarge
	BuildingMechFactory
	BuildingLaserTower
	NumBuildingTypes
)

func (t BuildingType) String() string {
	switch t {
	case BuildingTankFactory:
		return "Tank Factory"
	case BuildingSolarPanel:
		return "Solar Panel"
	case BuildingMetalExtractor:
		return "Metal Extractor"
	case BuildingMetalStorage:
		return "Metal Storage"
	case BuildingEnergyStorage:
		return "Energy Storage"
	case BuildingMetalStorageLarge:
		return "Large Metal Storage"
	case BuildingEnergyStorageLarge:
		return "Large Energy Storage"
	case BuildingMechFactory:
		return "Mech Factory"
	case BuildingLaserTower:
		return "Laser Tower"
	default:
		return "Unknown"
	}
}

type BuildingDef struct {
	Type              BuildingType
	Name              string
	Description       string
	Size              float64
	Color             color.Color
	Cost              map[resource.Type]float64
	MetalProduction   float64
	EnergyProduction  float64
	MetalConsumption  float64
	EnergyConsumption float64
	MetalStorage      float64
	EnergyStorage     float64
	BuildTime         float64
	VisionRange       float64
	Health            float64

	// Capabilities - data-driven instead of hardcoded type checks
	IsFactory     bool       // Can produce units
	ProducesUnits []UnitType // Which units this factory can produce (if IsFactory)

	// Combat stats for defensive buildings
	CanAttack     bool    // Can attack enemy units
	Damage        float64 // Damage per shot
	AttackRange   float64 // Range to attack enemies
	FireRate      float64 // Shots per second
	EnergyPerShot float64 // Energy consumed per shot
}

var BuildingDefs = map[BuildingType]*BuildingDef{
	BuildingTankFactory: {
		Type:        BuildingTankFactory,
		Name:        "Tank Factory",
		Description: "Produces combat tanks",
		Size:        60,
		Color:       color.RGBA{100, 100, 100, 255},
		Cost: map[resource.Type]float64{
			resource.Metal:  200,
			resource.Energy: 100,
		},
		EnergyConsumption: 5,
		BuildTime:         10,
		VisionRange:       150,
		Health:            500,
		IsFactory:         true,
		ProducesUnits:     []UnitType{UnitTypeTank, UnitTypeScout},
	},
	BuildingSolarPanel: {
		Type:        BuildingSolarPanel,
		Name:        "Solar Panel",
		Description: "Generates energy",
		Size:        40,
		Color:       color.RGBA{50, 100, 200, 255},
		Cost: map[resource.Type]float64{
			resource.Metal: 50,
		},
		EnergyProduction: 10,
		BuildTime:        5,
		VisionRange:      100,
		Health:           150,
	},
	BuildingMetalExtractor: {
		Type:        BuildingMetalExtractor,
		Name:        "Metal Extractor",
		Description: "Extracts metal from the ground",
		Size:        35,
		Color:       color.RGBA{150, 150, 170, 255},
		Cost: map[resource.Type]float64{
			resource.Metal:  50,
			resource.Energy: 25,
		},
		MetalProduction:   3,
		EnergyConsumption: 2,
		BuildTime:         5,
		VisionRange:       100,
		Health:            200,
	},
	BuildingMetalStorage: {
		Type:        BuildingMetalStorage,
		Name:        "Metal Storage",
		Description: "Stores 500 metal",
		Size:        30,
		Color:       color.RGBA{120, 120, 140, 255},
		Cost: map[resource.Type]float64{
			resource.Metal: 75,
		},
		MetalStorage: 500,
		BuildTime:    4,
		VisionRange:  80,
		Health:       300,
	},
	BuildingEnergyStorage: {
		Type:        BuildingEnergyStorage,
		Name:        "Energy Storage",
		Description: "Stores 500 energy",
		Size:        30,
		Color:       color.RGBA{80, 120, 180, 255},
		Cost: map[resource.Type]float64{
			resource.Metal: 75,
		},
		EnergyStorage: 500,
		BuildTime:     4,
		VisionRange:   80,
		Health:        300,
	},
	BuildingMetalStorageLarge: {
		Type:        BuildingMetalStorageLarge,
		Name:        "Large Metal Storage",
		Description: "Stores 2000 metal",
		Size:        50,
		Color:       color.RGBA{100, 100, 120, 255},
		Cost: map[resource.Type]float64{
			resource.Metal:  200,
			resource.Energy: 50,
		},
		MetalStorage: 2000,
		BuildTime:    8,
		VisionRange:  100,
		Health:       600,
	},
	BuildingEnergyStorageLarge: {
		Type:        BuildingEnergyStorageLarge,
		Name:        "Large Energy Storage",
		Description: "Stores 3000 energy",
		Size:        50,
		Color:       color.RGBA{60, 100, 160, 255},
		Cost: map[resource.Type]float64{
			resource.Metal:  200,
			resource.Energy: 50,
		},
		EnergyStorage: 3000,
		BuildTime:     8,
		VisionRange:   100,
		Health:        600,
	},
	BuildingMechFactory: {
		Type:        BuildingMechFactory,
		Name:        "Mech Factory",
		Description: "Produces mech constructors",
		Size:        70,
		Color:       color.RGBA{120, 90, 140, 255}, // Purple-ish
		Cost: map[resource.Type]float64{
			resource.Metal:  300,
			resource.Energy: 150,
		},
		EnergyConsumption: 8,
		BuildTime:         15,
		VisionRange:       150,
		Health:            600,
		IsFactory:         true,
		ProducesUnits:     []UnitType{UnitTypeMechConstructor},
	},
	BuildingLaserTower: {
		Type:        BuildingLaserTower,
		Name:        "Laser Tower",
		Description: "Defensive laser turret, uses energy to fire",
		Size:        30,
		Color:       color.RGBA{200, 50, 50, 255}, // Red
		Cost: map[resource.Type]float64{
			resource.Metal:  150,
			resource.Energy: 100,
		},
		BuildTime:     6,
		VisionRange:   300,
		Health:        250,
		CanAttack:     true,
		Damage:        25,
		AttackRange:   250,
		FireRate:      2.0,
		EnergyPerShot: 5,
	},
}

// GetBuildableDefs returns the buildings that a unit type can construct (data-driven)
func GetBuildableDefs(unitType UnitType) []*BuildingDef {
	def := UnitDefs[unitType]
	if def == nil || !def.CanConstruct || len(def.BuildableTypes) == 0 {
		return nil
	}
	result := make([]*BuildingDef, 0, len(def.BuildableTypes))
	for _, buildingType := range def.BuildableTypes {
		if buildingDef := BuildingDefs[buildingType]; buildingDef != nil {
			result = append(result, buildingDef)
		}
	}
	return result
}
