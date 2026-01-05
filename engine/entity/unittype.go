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
)

func (t UnitType) String() string {
	switch t {
	case UnitTypeConstructor:
		return "Constructor"
	case UnitTypeTank:
		return "Tank"
	case UnitTypeScout:
		return "Scout"
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
	Color       color.Color // Base color (modified by faction)
	Cost        map[resource.Type]float64
	BuildTime   float64
	Health      float64
	Damage      float64 // Damage per shot
	Range       float64 // Attack range
	FireRate    float64 // Shots per second
}

var UnitDefs = map[UnitType]*UnitDef{
	UnitTypeTank: {
		Type:        UnitTypeTank,
		Name:        "Tank",
		Description: "Heavy combat unit",
		Size:        22,
		Speed:       2.5,
		Color:       color.RGBA{80, 120, 80, 255}, // Base color (tinted by faction)
		Cost: map[resource.Type]float64{
			resource.Metal:  100,
			resource.Energy: 50,
		},
		BuildTime: 5.0,
		Health:    100,
		Damage:    15,
		Range:     150,
		FireRate:  1.0,
	},
	UnitTypeScout: {
		Type:        UnitTypeScout,
		Name:        "Scout",
		Description: "Fast recon unit, light weapons",
		Size:        16,
		Speed:       5.0,
		Color:       color.RGBA{180, 180, 100, 255}, // Base color (tinted by faction)
		Cost: map[resource.Type]float64{
			resource.Metal:  30,
			resource.Energy: 20,
		},
		BuildTime: 2.0,
		Health:    40,
		Damage:    5,
		Range:     100,
		FireRate:  2.0,
	},
	UnitTypeConstructor: {
		Type:        UnitTypeConstructor,
		Name:        "Constructor",
		Description: "Builds structures",
		Size:        24,
		Speed:       1.5,
		Color:       color.RGBA{200, 180, 50, 255}, // Yellow/gold
		Cost: map[resource.Type]float64{
			resource.Metal:  150,
			resource.Energy: 75,
		},
		BuildTime: 8.0,
		Health:    60,
		Damage:    0,
		Range:     0,
		FireRate:  0,
	},
}

func GetProducibleUnits(buildingType BuildingType) []*UnitDef {
	switch buildingType {
	case BuildingTankFactory:
		return []*UnitDef{
			UnitDefs[UnitTypeTank],
			UnitDefs[UnitTypeScout],
		}
	default:
		return nil
	}
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
}

var BuildingDefs = map[BuildingType]*BuildingDef{
	BuildingTankFactory: {
		Type:        BuildingTankFactory,
		Name:        "Tank Factory",
		Description: "Produces combat tanks",
		Size:        60,
		Color:       color.RGBA{100, 100, 100, 255}, // Gray
		Cost: map[resource.Type]float64{
			resource.Metal:  200,
			resource.Energy: 100,
		},
		EnergyConsumption: 5,
		BuildTime:         10,
	},
	BuildingSolarPanel: {
		Type:        BuildingSolarPanel,
		Name:        "Solar Panel",
		Description: "Generates energy",
		Size:        40,
		Color:       color.RGBA{50, 100, 200, 255}, // Blue
		Cost: map[resource.Type]float64{
			resource.Metal: 50,
		},
		EnergyProduction: 10,
		BuildTime:        5,
	},
	BuildingMetalExtractor: {
		Type:        BuildingMetalExtractor,
		Name:        "Metal Extractor",
		Description: "Extracts metal from the ground",
		Size:        35,
		Color:       color.RGBA{150, 150, 170, 255}, // Silver
		Cost: map[resource.Type]float64{
			resource.Metal:  50,
			resource.Energy: 25,
		},
		MetalProduction:   3,
		EnergyConsumption: 2,
		BuildTime:         5,
	},
	BuildingMetalStorage: {
		Type:        BuildingMetalStorage,
		Name:        "Metal Storage",
		Description: "Stores 500 metal",
		Size:        30,
		Color:       color.RGBA{120, 120, 140, 255}, // Dark silver
		Cost: map[resource.Type]float64{
			resource.Metal: 75,
		},
		MetalStorage: 500,
		BuildTime:    4,
	},
	BuildingEnergyStorage: {
		Type:        BuildingEnergyStorage,
		Name:        "Energy Storage",
		Description: "Stores 500 energy",
		Size:        30,
		Color:       color.RGBA{80, 120, 180, 255}, // Light blue
		Cost: map[resource.Type]float64{
			resource.Metal: 75,
		},
		EnergyStorage: 500,
		BuildTime:     4,
	},
	BuildingMetalStorageLarge: {
		Type:        BuildingMetalStorageLarge,
		Name:        "Large Metal Storage",
		Description: "Stores 2000 metal",
		Size:        50,
		Color:       color.RGBA{100, 100, 120, 255}, // Darker silver
		Cost: map[resource.Type]float64{
			resource.Metal:  200,
			resource.Energy: 50,
		},
		MetalStorage: 2000,
		BuildTime:    8,
	},
	BuildingEnergyStorageLarge: {
		Type:        BuildingEnergyStorageLarge,
		Name:        "Large Energy Storage",
		Description: "Stores 3000 energy",
		Size:        50,
		Color:       color.RGBA{60, 100, 160, 255}, // Darker blue
		Cost: map[resource.Type]float64{
			resource.Metal:  200,
			resource.Energy: 50,
		},
		EnergyStorage: 3000,
		BuildTime:     8,
	},
}

func GetBuildableDefs(unitType UnitType) []*BuildingDef {
	switch unitType {
	case UnitTypeConstructor:
		return []*BuildingDef{
			BuildingDefs[BuildingTankFactory],
			BuildingDefs[BuildingSolarPanel],
			BuildingDefs[BuildingMetalExtractor],
			BuildingDefs[BuildingMetalStorage],
			BuildingDefs[BuildingEnergyStorage],
			BuildingDefs[BuildingMetalStorageLarge],
			BuildingDefs[BuildingEnergyStorageLarge],
		}
	default:
		return nil
	}
}
