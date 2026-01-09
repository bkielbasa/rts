package entity

import (
	"github.com/bklimczak/tanks/engine/resource"
	"image/color"
)

type UnitType int

const (
	UnitTypeBasic UnitType = iota
	// Tier 1 Infantry
	UnitTypeTrooper
	UnitTypeRocketMarine
	UnitTypeTechnician
	// Tier 1 Hover Vehicles
	UnitTypeReconSkimmer
	UnitTypeStriker
	UnitTypeCarrierAPC
	// Legacy types for compatibility
	UnitTypeConstructor
	UnitTypeTank
	UnitTypeTank2
	UnitTypeTank3
	UnitTypeTank4
	UnitTypeTank5
	UnitTypeTank6
	UnitTypeTank7
	UnitTypeTank8
	UnitTypeScout
	UnitTypeMechConstructor
)

func (t UnitType) String() string {
	switch t {
	case UnitTypeTrooper:
		return "Trooper"
	case UnitTypeRocketMarine:
		return "Rocket Marine"
	case UnitTypeTechnician:
		return "Technician"
	case UnitTypeReconSkimmer:
		return "Recon Skimmer"
	case UnitTypeStriker:
		return "Striker"
	case UnitTypeCarrierAPC:
		return "Carrier APC"
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
	Size        float64
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
	// === TIER 1 INFANTRY (Barracks) ===
	UnitTypeTrooper: {
		Type:        UnitTypeTrooper,
		Name:        "Trooper",
		Description: "Basic infantry with pulse rifle",
		Size:        12,
		Speed:       3.0,
		Color:       color.RGBA{100, 150, 200, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 80,
		},
		BuildTime:     3.0,
		Health:        40,
		VisionRange:   200,
		RotationSpeed: 0.2,
		IsInfantry:    true,
		Combat: &CombatDef{
			Damage:   10,
			Range:    120,
			FireRate: 2.0,
		},
	},
	UnitTypeRocketMarine: {
		Type:        UnitTypeRocketMarine,
		Name:        "Rocket Marine",
		Description: "Anti-vehicle infantry with guided rockets",
		Size:        14,
		Speed:       2.8,
		Color:       color.RGBA{200, 100, 100, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 150,
		},
		BuildTime:     5.0,
		Health:        35,
		VisionRange:   220,
		RotationSpeed: 0.18,
		IsInfantry:    true,
		Combat: &CombatDef{
			Damage:   30,
			Range:    180,
			FireRate: 0.8,
		},
	},
	UnitTypeTechnician: {
		Type:        UnitTypeTechnician,
		Name:        "Technician",
		Description: "Builds structures and repairs",
		Size:        14,
		Speed:       3.2,
		Color:       color.RGBA{200, 180, 50, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 180,
		},
		BuildTime:     6.0,
		Health:        25,
		VisionRange:   180,
		RotationSpeed: 0.2,
		IsInfantry:    true,
		Construction: &ConstructionDef{
			BuildableTypes: AllBuildableTypes,
			RepairRate:     15,
			RepairRange:    60,
			CanRepairUnits: true,
		},
	},

	// === TIER 1 HOVER VEHICLES (Hover Bay) ===
	UnitTypeReconSkimmer: {
		Type:        UnitTypeReconSkimmer,
		Name:        "Recon Skimmer",
		Description: "Fast scout with light laser",
		Size:        18,
		Speed:       6.0,
		Color:       color.RGBA{150, 200, 255, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 150,
		},
		BuildTime:     4.0,
		Health:        60,
		VisionRange:   400,
		RotationSpeed: 0.15,
		IsHoverUnit:   true,
		Combat: &CombatDef{
			Damage:   8,
			Range:    100,
			FireRate: 2.5,
		},
	},
	UnitTypeStriker: {
		Type:        UnitTypeStriker,
		Name:        "Striker",
		Description: "Versatile hover tank with twin autocannons",
		Size:        24,
		Speed:       4.0,
		Color:       color.RGBA{80, 140, 180, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 450,
			resource.Alloys:  10,
		},
		BuildTime:     6.0,
		Health:        180,
		VisionRange:   280,
		RotationSpeed: 0.08,
		IsHoverUnit:   true,
		Combat: &CombatDef{
			Damage:   15,
			Range:    140,
			FireRate: 3.0,
		},
	},
	UnitTypeCarrierAPC: {
		Type:        UnitTypeCarrierAPC,
		Name:        "Carrier APC",
		Description: "Armored transport for infantry",
		Size:        28,
		Speed:       3.5,
		Color:       color.RGBA{100, 120, 100, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 500,
			resource.Alloys:  10,
		},
		BuildTime:     7.0,
		Health:        220,
		VisionRange:   200,
		RotationSpeed: 0.06,
		IsHoverUnit:   true,
		Combat: &CombatDef{
			Damage:   10,
			Range:    80,
			FireRate: 1.5,
		},
	},

	// === LEGACY UNITS ===
	UnitTypeTank: {
		Type:        UnitTypeTank,
		Name:        "Tank",
		Description: "Heavy combat unit",
		Size:        40,
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
	UnitTypeTank2: {
		Type:          UnitTypeTank2,
		Name:          "Tank II",
		Description:   "Tank with dual cannon",
		Size:          40,
		Speed:         2.5,
		Color:         color.RGBA{80, 120, 80, 255},
		Health:        100,
		VisionRange:   250,
		RotationSpeed: 0.03,
		SpriteScale:   0.25,
		Combat: &CombatDef{
			Damage:   18,
			Range:    150,
			FireRate: 1.0,
		},
		TankRender: &TankRenderDef{
			HullSpritePath:      "units/color_a/Hull_01.png",
			GunSpritePath:       "units/color_a/Gun_02.png",
			TurretRotationSpeed: 0.08,
		},
	},
	UnitTypeTank3: {
		Type:          UnitTypeTank3,
		Name:          "Tank III",
		Description:   "Tank with heavy cannon",
		Size:          40,
		Speed:         2.3,
		Color:         color.RGBA{80, 120, 80, 255},
		Health:        120,
		VisionRange:   250,
		RotationSpeed: 0.03,
		SpriteScale:   0.25,
		Combat: &CombatDef{
			Damage:   22,
			Range:    160,
			FireRate: 0.8,
		},
		TankRender: &TankRenderDef{
			HullSpritePath:      "units/color_a/Hull_01.png",
			GunSpritePath:       "units/color_a/Gun_03.png",
			TurretRotationSpeed: 0.07,
			TurretOffsetY:       -30,
		},
	},
	UnitTypeTank4: {
		Type:          UnitTypeTank4,
		Name:          "Tank IV",
		Description:   "Tank with rapid fire cannon",
		Size:          40,
		Speed:         2.5,
		Color:         color.RGBA{80, 120, 80, 255},
		Health:        90,
		VisionRange:   250,
		RotationSpeed: 0.03,
		SpriteScale:   0.25,
		Combat: &CombatDef{
			Damage:   10,
			Range:    140,
			FireRate: 2.0,
		},
		TankRender: &TankRenderDef{
			HullSpritePath:      "units/color_a/Hull_01.png",
			GunSpritePath:       "units/color_a/Gun_04.png",
			TurretRotationSpeed: 0.10,
		},
	},
	UnitTypeTank5: {
		Type:          UnitTypeTank5,
		Name:          "Tank V",
		Description:   "Tank with missile launcher",
		Size:          40,
		Speed:         2.2,
		Color:         color.RGBA{80, 120, 80, 255},
		Health:        100,
		VisionRange:   280,
		RotationSpeed: 0.03,
		SpriteScale:   0.25,
		Combat: &CombatDef{
			Damage:   30,
			Range:    200,
			FireRate: 0.5,
		},
		TankRender: &TankRenderDef{
			HullSpritePath:      "units/color_a/Hull_01.png",
			GunSpritePath:       "units/color_a/Gun_05.png",
			TurretRotationSpeed: 0.06,
		},
	},
	UnitTypeTank6: {
		Type:          UnitTypeTank6,
		Name:          "Tank VI",
		Description:   "Tank with plasma cannon",
		Size:          40,
		Speed:         2.0,
		Color:         color.RGBA{80, 120, 80, 255},
		Health:        130,
		VisionRange:   250,
		RotationSpeed: 0.025,
		SpriteScale:   0.25,
		Combat: &CombatDef{
			Damage:   35,
			Range:    170,
			FireRate: 0.6,
		},
		TankRender: &TankRenderDef{
			HullSpritePath:      "units/color_a/Hull_01.png",
			GunSpritePath:       "units/color_a/Gun_06.png",
			TurretRotationSpeed: 0.05,
			TurretOffsetY:       -30,
		},
	},
	UnitTypeTank7: {
		Type:          UnitTypeTank7,
		Name:          "Tank VII",
		Description:   "Tank with artillery cannon",
		Size:          40,
		Speed:         1.8,
		Color:         color.RGBA{80, 120, 80, 255},
		Health:        110,
		VisionRange:   300,
		RotationSpeed: 0.02,
		SpriteScale:   0.25,
		Combat: &CombatDef{
			Damage:   40,
			Range:    250,
			FireRate: 0.4,
		},
		TankRender: &TankRenderDef{
			HullSpritePath:      "units/color_a/Hull_01.png",
			GunSpritePath:       "units/color_a/Gun_07.png",
			TurretRotationSpeed: 0.04,
		},
	},
	UnitTypeTank8: {
		Type:          UnitTypeTank8,
		Name:          "Tank VIII",
		Description:   "Tank with experimental weapon",
		Size:          40,
		Speed:         2.5,
		Color:         color.RGBA{80, 120, 80, 255},
		Health:        150,
		VisionRange:   250,
		RotationSpeed: 0.03,
		SpriteScale:   0.25,
		Combat: &CombatDef{
			Damage:   50,
			Range:    180,
			FireRate: 0.3,
		},
		TankRender: &TankRenderDef{
			HullSpritePath:      "units/color_a/Hull_01.png",
			GunSpritePath:       "units/color_a/Gun_08.png",
			TurretRotationSpeed: 0.08,
			TurretOffsetY:       -30,
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
	UnitTypeMechConstructor: {
		Type:        UnitTypeMechConstructor,
		Name:        "Mech Constructor",
		Description: "Builds structures and repairs units",
		Size:        28,
		Speed:       2.0,
		Color:       color.RGBA{180, 140, 200, 255},
		Cost: map[resource.Type]float64{
			resource.Credits: 200,
			resource.Energy:  100,
		},
		BuildTime:     10.0,
		Health:        80,
		VisionRange:   250,
		RotationSpeed: 0.08,
		Construction: &ConstructionDef{
			BuildableTypes: AllBuildableTypes,
			RepairRate:     20,
			RepairRange:    100,
			CanRepairUnits: true,
		},
	},
}

var BuildingDefs = map[BuildingType]*BuildingDef{
	// === TIER 1 - FOUNDATION ===
	BuildingCommandNexus: {
		Type:           BuildingCommandNexus,
		Name:           "Command Nexus",
		Description:    "Central command, builds technicians, provides radar",
		Size:           70,
		Color:          color.RGBA{60, 80, 120, 255},
		Cost:           map[resource.Type]float64{},
		EnergyConsumption: 20,
		BuildTime:      0,
		VisionRange:    350,
		Health:         1000,
		IsFactory:      true,
		ProducesUnits:  []UnitType{UnitTypeTechnician},
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
		Description: "Produces hover vehicles",
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
		ProducesUnits:     []UnitType{UnitTypeReconSkimmer, UnitTypeStriker, UnitTypeCarrierAPC},
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
		Description: "Produces mech constructors",
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
		ProducesUnits:     []UnitType{UnitTypeMechConstructor},
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
