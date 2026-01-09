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

// CombatDef contains combat-related stats for units that can attack
type CombatDef struct {
	Damage   float64
	Range    float64
	FireRate float64
}

// ConstructionDef contains construction/repair capabilities
type ConstructionDef struct {
	BuildableTypes []BuildingType
	RepairRate     float64
	RepairRange    float64
	CanRepairUnits bool
}

// TankRenderDef contains tank-specific rendering with hull and turret
type TankRenderDef struct {
	HullSpritePath      string
	GunSpritePath       string
	TurretRotationSpeed float64
	TurretOffsetY       float64
}

type UnitDef struct {
	// Core - always present
	Type        UnitType
	Name        string
	Description string
	Size        float64 // Square size (legacy, used if Width/Height not set)
	Width       float64 // Explicit width (takes precedence over Size)
	Height      float64 // Explicit height (takes precedence over Size)
	Speed       float64
	Color       color.Color
	Cost        map[resource.Type]float64
	BuildTime   float64
	Health      float64
	VisionRange float64

	// Movement
	RotationSpeed float64
	IsHoverUnit   bool
	IsInfantry    bool

	// Optional capabilities - nil means unit doesn't have this capability
	Combat       *CombatDef       // nil for non-combat units
	Construction *ConstructionDef // nil for non-constructor units
	TankRender   *TankRenderDef   // nil for non-tank units

	// Simple sprite (used when TankRender is nil)
	SpritePath  string
	SpriteScale float64
}

// CanAttack returns true if unit can deal damage
func (d *UnitDef) CanAttack() bool {
	return d.Combat != nil && d.Combat.Damage > 0
}

// CanConstruct returns true if unit can build structures
func (d *UnitDef) CanConstruct() bool {
	return d.Construction != nil && len(d.Construction.BuildableTypes) > 0
}

// CanRepairUnits returns true if unit can repair other units
func (d *UnitDef) CanRepairUnits() bool {
	return d.Construction != nil && d.Construction.CanRepairUnits
}

// HasTurret returns true if unit has a rotating turret
func (d *UnitDef) HasTurret() bool {
	return d.TankRender != nil
}

// GetDamage returns damage or 0 if non-combat unit
func (d *UnitDef) GetDamage() float64 {
	if d.Combat == nil {
		return 0
	}
	return d.Combat.Damage
}

// GetRange returns attack range or 0 if non-combat unit
func (d *UnitDef) GetRange() float64 {
	if d.Combat == nil {
		return 0
	}
	return d.Combat.Range
}

// GetFireRate returns fire rate or 0 if non-combat unit
func (d *UnitDef) GetFireRate() float64 {
	if d.Combat == nil {
		return 0
	}
	return d.Combat.FireRate
}

// GetBuildableTypes returns buildable buildings or nil
func (d *UnitDef) GetBuildableTypes() []BuildingType {
	if d.Construction == nil {
		return nil
	}
	return d.Construction.BuildableTypes
}

// GetRepairRate returns repair rate or 0 if unit can't repair
func (d *UnitDef) GetRepairRate() float64 {
	if d.Construction == nil {
		return 0
	}
	return d.Construction.RepairRate
}

// GetRepairRange returns repair range or 0 if unit can't repair
func (d *UnitDef) GetRepairRange() float64 {
	if d.Construction == nil {
		return 0
	}
	return d.Construction.RepairRange
}

// GetTurretRotationSpeed returns turret speed or 0 if no turret
func (d *UnitDef) GetTurretRotationSpeed() float64 {
	if d.TankRender == nil {
		return 0
	}
	return d.TankRender.TurretRotationSpeed
}

// GetWidth returns the actual width of the unit
func (d *UnitDef) GetWidth() float64 {
	if d.Width > 0 {
		return d.Width
	}
	return d.Size
}

// GetHeight returns the actual height of the unit
func (d *UnitDef) GetHeight() float64 {
	if d.Height > 0 {
		return d.Height
	}
	return d.Size
}

type BuildingType int

const (
	// Tier 1 - Foundation
	BuildingCommandNexus BuildingType = iota
	BuildingSolarArray
	BuildingFusionReactor
	BuildingOreExtractor
	BuildingAlloyFoundry
	BuildingVehicleFactory
	BuildingHoverBay
	BuildingDataUplink
	BuildingWall
	BuildingAutocannonTurret
	BuildingMissileBattery
	// Legacy types
	BuildingTankFactory
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
	case BuildingCommandNexus:
		return "Command Nexus"
	case BuildingSolarArray:
		return "Solar Array"
	case BuildingFusionReactor:
		return "Fusion Reactor"
	case BuildingOreExtractor:
		return "Ore Extractor"
	case BuildingAlloyFoundry:
		return "Alloy Foundry"
	case BuildingVehicleFactory:
		return "Vehicle Factory"
	case BuildingHoverBay:
		return "Hover Bay"
	case BuildingDataUplink:
		return "Data Uplink"
	case BuildingWall:
		return "Wall"
	case BuildingAutocannonTurret:
		return "Autocannon Turret"
	case BuildingMissileBattery:
		return "Missile Battery"
	case BuildingTankFactory:
		return "Tank Factory"
	case BuildingSolarPanel:
		return "Solar Panel"
	case BuildingMetalExtractor:
		return "Credit Extractor"
	case BuildingMetalStorage:
		return "Credit Storage"
	case BuildingEnergyStorage:
		return "Energy Storage"
	case BuildingMetalStorageLarge:
		return "Large Credit Storage"
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
	Size              float64 // For square buildings (legacy/convenience)
	Width             float64 // Explicit width (takes precedence over Size)
	Height            float64 // Explicit height (takes precedence over Size)
	Color             color.Color
	Cost              map[resource.Type]float64
	CreditsProduction float64
	EnergyProduction  float64
	AlloysProduction  float64
	CreditsConsumption float64
	EnergyConsumption float64
	AlloysConsumption float64
	CreditsStorage    float64
	EnergyStorage     float64
	AlloysStorage     float64
	BuildTime         float64
	VisionRange       float64
	Health            float64

	IsFactory       bool
	ProducesUnits   []UnitType
	RequiresDeposit bool

	CanAttack     bool
	Damage        float64
	AttackRange   float64
	FireRate      float64
	EnergyPerShot float64

	AntiAir    bool
	AntiGround bool

	SpritePath     string
	SpriteWidth    float64 // Width of single frame (0 = use Size)
	SpriteHeight   float64 // Height of single frame (0 = use Size, frames calculated from sprite height / SpriteHeight)
	AnimationSpeed float64 // Frames per second (0 = static)
}

// GetWidth returns the actual width of the building
func (d *BuildingDef) GetWidth() float64 {
	if d.Width > 0 {
		return d.Width
	}
	return d.Size
}

// GetHeight returns the actual height of the building
func (d *BuildingDef) GetHeight() float64 {
	if d.Height > 0 {
		return d.Height
	}
	return d.Size
}

var AllBuildableTypes = []BuildingType{
	BuildingCommandNexus,
	BuildingSolarArray,
	BuildingFusionReactor,
	BuildingOreExtractor,
	BuildingAlloyFoundry,
	BuildingVehicleFactory,
	BuildingHoverBay,
	BuildingDataUplink,
	BuildingWall,
	BuildingAutocannonTurret,
	BuildingMissileBattery,
}

var UnitDefs = map[UnitType]*UnitDef{
	UnitTypeTank: {
		Type:        UnitTypeTank,
		Name:        "Tank",
		Description: "Heavy combat unit",
		Width:       55,
		Height:      40,
		Speed:       2.5,
		Color:       color.RGBA{80, 120, 80, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 100,
			resource.Energy:  50,
		},
		BuildTime:     5.0,
		Health:        100,
		VisionRange:   250,
		RotationSpeed: 0.03,
		SpriteScale:   0.25,
		Combat: &CombatDef{
			Damage:   15,
			Range:    150,
			FireRate: 1.0,
		},
		TankRender: &TankRenderDef{
			HullSpritePath:      "units/color_a/Hull_01.png",
			GunSpritePath:       "units/color_a/Gun_01.png",
			TurretRotationSpeed: 0.08,
		},
	},
	UnitTypeScout: {
		Type:        UnitTypeScout,
		Name:        "Scout",
		Description: "Fast recon unit, light weapons",
		Size:        16,
		Speed:       5.0,
		Color:       color.RGBA{180, 180, 100, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 30,
			resource.Energy:  20,
		},
		BuildTime:     2.0,
		Health:        40,
		VisionRange:   400,
		RotationSpeed: 0.2,
		SpritePath:    "scout.png",
		Combat: &CombatDef{
			Damage:   5,
			Range:    100,
			FireRate: 2.0,
		},
	},
	UnitTypeConstructor: {
		Type:        UnitTypeConstructor,
		Name:        "Constructor",
		Description: "Builds structures",
		Size:        24,
		Speed:       1.5,
		Color:       color.RGBA{200, 180, 50, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 150,
			resource.Energy:  75,
		},
		BuildTime:     8.0,
		Health:        60,
		VisionRange:   200,
		RotationSpeed: 0.1,
		Construction: &ConstructionDef{
			BuildableTypes: AllBuildableTypes,
		},
	},
}

var BuildingDefs = map[BuildingType]*BuildingDef{
	// === TIER 1 - FOUNDATION ===
	BuildingCommandNexus: {
		Type:           BuildingCommandNexus,
		Name:           "Command Nexus",
		Description:    "Central command, builds constructors, provides radar",
		Size:           70,
		Color:          color.RGBA{60, 80, 120, 255},
		Cost:           map[resource.Type]float64{},
		EnergyConsumption: 20,
		BuildTime:      0,
		VisionRange:    350,
		Health:         1000,
		IsFactory:      true,
		ProducesUnits:  []UnitType{UnitTypeConstructor},
		CreditsStorage: 500,
		EnergyStorage:  100,
	},
	BuildingSolarArray: {
		Type:             BuildingSolarArray,
		Name:             "Solar Array",
		Description:      "Basic power generation",
		Size:             128,
		Color:            color.RGBA{50, 120, 200, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 400,
		},
		EnergyProduction: 40,
		BuildTime:        8,
		VisionRange:      100,
		Health:           150,
		SpritePath:       "buildings/solar_array.png",
	},
	BuildingFusionReactor: {
		Type:             BuildingFusionReactor,
		Name:             "Fusion Reactor",
		Description:      "Reliable high-output power",
		Size:             45,
		Color:            color.RGBA{100, 180, 255, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 800,
		},
		EnergyProduction: 80,
		BuildTime:        15,
		VisionRange:      120,
		Health:           300,
	},
	BuildingOreExtractor: {
		Type:        BuildingOreExtractor,
		Name:        "Ore Extractor",
		Description: "Harvests ore for credits (place on deposit)",
		Size:        40,
		Color:       color.RGBA{180, 150, 100, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 600,
		},
		CreditsProduction: 15,
		EnergyConsumption: 5,
		BuildTime:         12,
		VisionRange:       100,
		Health:            250,
		RequiresDeposit:   true,
	},
	BuildingAlloyFoundry: {
		Type:        BuildingAlloyFoundry,
		Name:        "Alloy Foundry",
		Description: "Processes ore into alloys",
		Size:        50,
		Color:       color.RGBA{140, 100, 160, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 700,
		},
		AlloysProduction:   5,
		CreditsConsumption: 3,
		EnergyConsumption:  15,
		BuildTime:          18,
		VisionRange:        100,
		Health:             400,
		AlloysStorage:      200,
	},
	BuildingVehicleFactory: {
		Type:        BuildingVehicleFactory,
		Name:        "Vehicle Factory",
		Description: "Produces combat vehicles",
		Width:       128,
		Height:      64,
		Color:       color.RGBA{120, 100, 80, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 600,
		},
		EnergyConsumption: 15,
		BuildTime:         15,
		VisionRange:       150,
		Health:            500,
		IsFactory:         true,
		ProducesUnits:     []UnitType{UnitTypeTank, UnitTypeScout},
		SpritePath:        "buildings/factory.png",
		SpriteWidth:       128,
		SpriteHeight:      64,
		AnimationSpeed:    4,
	},
	BuildingHoverBay: {
		Type:        BuildingHoverBay,
		Name:        "Hover Bay",
		Description: "Produces scout vehicles",
		Size:        60,
		Color:       color.RGBA{80, 100, 140, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 900,
		},
		EnergyConsumption: 15,
		BuildTime:         18,
		VisionRange:       150,
		Health:            500,
		IsFactory:         true,
		ProducesUnits:     []UnitType{UnitTypeScout},
	},
	BuildingDataUplink: {
		Type:        BuildingDataUplink,
		Name:        "Data Uplink",
		Description: "Extends radar range significantly",
		Size:        30,
		Color:       color.RGBA{100, 200, 200, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 600,
		},
		EnergyConsumption: 10,
		BuildTime:         10,
		VisionRange:       500,
		Health:            150,
	},
	BuildingWall: {
		Type:        BuildingWall,
		Name:        "Wall",
		Description: "Basic defensive barrier",
		Size:        20,
		Color:       color.RGBA{80, 80, 90, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 50,
		},
		BuildTime:   2,
		VisionRange: 50,
		Health:      500,
	},
	BuildingAutocannonTurret: {
		Type:        BuildingAutocannonTurret,
		Name:        "Autocannon Turret",
		Description: "Anti-ground defense",
		Size:        25,
		Color:       color.RGBA{150, 100, 80, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 350,
		},
		EnergyConsumption: 5,
		BuildTime:         8,
		VisionRange:       250,
		Health:            300,
		CanAttack:         true,
		Damage:            12,
		AttackRange:       200,
		FireRate:          3.0,
		EnergyPerShot:     2,
		AntiGround:        true,
	},
	BuildingMissileBattery: {
		Type:        BuildingMissileBattery,
		Name:        "Missile Battery",
		Description: "Anti-air defense",
		Size:        30,
		Color:       color.RGBA{180, 80, 80, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 450,
		},
		EnergyConsumption: 8,
		BuildTime:         10,
		VisionRange:       300,
		Health:            250,
		CanAttack:         true,
		Damage:            25,
		AttackRange:       280,
		FireRate:          1.5,
		EnergyPerShot:     5,
		AntiAir:           true,
		AntiGround:        true,
	},

	// === LEGACY BUILDINGS ===
	BuildingTankFactory: {
		Type:        BuildingTankFactory,
		Name:        "Tank Factory",
		Description: "Produces combat tanks",
		Size:        60,
		Color:       color.RGBA{100, 100, 100, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 200,
			resource.Energy:  100,
		},
		EnergyConsumption: 5,
		BuildTime:         10,
		VisionRange:       150,
		Health:            500,
		IsFactory:         true,
		ProducesUnits:     []UnitType{UnitTypeTank, UnitTypeScout},
	},
	BuildingSolarPanel: {
		Type:             BuildingSolarPanel,
		Name:             "Solar Panel",
		Description:      "Generates energy",
		Size:             40,
		Color:            color.RGBA{50, 100, 200, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 50,
		},
		EnergyProduction: 10,
		BuildTime:        5,
		VisionRange:      100,
		Health:           150,
	},
	BuildingMetalExtractor: {
		Type:        BuildingMetalExtractor,
		Name:        "Credit Extractor",
		Description: "Extracts credits from deposits",
		Size:        35,
		Color:       color.RGBA{150, 150, 170, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 50,
			resource.Energy:  25,
		},
		CreditsProduction: 3,
		EnergyConsumption: 2,
		BuildTime:         5,
		VisionRange:       100,
		Health:            200,
		RequiresDeposit:   true,
	},
	BuildingMetalStorage: {
		Type:        BuildingMetalStorage,
		Name:        "Credit Storage",
		Description: "Stores 500 credits",
		Size:        30,
		Color:       color.RGBA{120, 120, 140, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 75,
		},
		CreditsStorage: 500,
		BuildTime:      4,
		VisionRange:    80,
		Health:         300,
	},
	BuildingEnergyStorage: {
		Type:        BuildingEnergyStorage,
		Name:        "Energy Storage",
		Description: "Stores 500 energy",
		Size:        30,
		Color:       color.RGBA{80, 120, 180, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 75,
		},
		EnergyStorage: 500,
		BuildTime:     4,
		VisionRange:   80,
		Health:        300,
	},
	BuildingMetalStorageLarge: {
		Type:        BuildingMetalStorageLarge,
		Name:        "Large Credit Storage",
		Description: "Stores 2000 credits",
		Size:        50,
		Color:       color.RGBA{100, 100, 120, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 200,
			resource.Energy:  50,
		},
		CreditsStorage: 2000,
		BuildTime:      8,
		VisionRange:    100,
		Health:         600,
	},
	BuildingEnergyStorageLarge: {
		Type:        BuildingEnergyStorageLarge,
		Name:        "Large Energy Storage",
		Description: "Stores 3000 energy",
		Size:        50,
		Color:       color.RGBA{60, 100, 160, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 200,
			resource.Energy:  50,
		},
		EnergyStorage: 3000,
		BuildTime:     8,
		VisionRange:   100,
		Health:        600,
	},
	BuildingMechFactory: {
		Type:        BuildingMechFactory,
		Name:        "Mech Factory",
		Description: "Produces constructors",
		Size:        70,
		Color:       color.RGBA{120, 90, 140, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 300,
			resource.Energy:  150,
		},
		EnergyConsumption: 8,
		BuildTime:         15,
		VisionRange:       150,
		Health:            600,
		IsFactory:         true,
		ProducesUnits:     []UnitType{UnitTypeConstructor},
	},
	BuildingLaserTower: {
		Type:        BuildingLaserTower,
		Name:        "Laser Tower",
		Description: "Defensive laser turret",
		Size:        30,
		Color:       color.RGBA{200, 50, 50, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 150,
			resource.Energy:  100,
		},
		BuildTime:     6,
		VisionRange:   300,
		Health:        250,
		CanAttack:     true,
		Damage:        25,
		AttackRange:   250,
		FireRate:      2.0,
		EnergyPerShot: 5,
		AntiGround:    true,
		AntiAir:       true,
	},
}

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

func GetBuildableDefs(unitType UnitType) []*BuildingDef {
	def := UnitDefs[unitType]
	if def == nil || !def.CanConstruct() {
		return nil
	}
	buildableTypes := def.GetBuildableTypes()
	if len(buildableTypes) == 0 {
		return nil
	}
	result := make([]*BuildingDef, 0, len(buildableTypes))
	for _, buildingType := range buildableTypes {
		if buildingDef := BuildingDefs[buildingType]; buildingDef != nil {
			result = append(result, buildingDef)
		}
	}
	return result
}
