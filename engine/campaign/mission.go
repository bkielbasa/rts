package campaign

import (
	"fmt"
	"os"

	"github.com/bklimczak/tanks/engine/entity"
	emath "github.com/bklimczak/tanks/engine/math"
	"gopkg.in/yaml.v3"
)

type Mission struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	MapPath     string `yaml:"map"`

	PlayerStart *FactionStart `yaml:"player_start"`
	EnemyStart  *FactionStart `yaml:"enemy_start,omitempty"`

	VictoryConditions []VictoryConditionDef `yaml:"victory_conditions"`
	DefeatConditions  []DefeatConditionDef  `yaml:"defeat_conditions"`

	Restrictions *MissionRestrictions `yaml:"restrictions,omitempty"`
	Briefing     *MissionBriefing     `yaml:"briefing"`
	Difficulty   string               `yaml:"difficulty,omitempty"`
}

type FactionStart struct {
	Position  Position        `yaml:"position"`
	Resources ResourcesConfig `yaml:"resources"`
	Units     []UnitPlacement `yaml:"units,omitempty"`
	Buildings []BuildingPlacement `yaml:"buildings,omitempty"`
}

type Position struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
}

func (p Position) ToVec2() emath.Vec2 {
	return emath.Vec2{X: p.X, Y: p.Y}
}

type ResourcesConfig struct {
	Credits float64 `yaml:"credits"`
	Energy  float64 `yaml:"energy"`
	Alloys  float64 `yaml:"alloys"`
}

type UnitPlacement struct {
	Type     string   `yaml:"type"`
	Position Position `yaml:"position"`
	Count    int      `yaml:"count,omitempty"`
	// Tank customization (only used when Type is "Tank")
	Color string `yaml:"color,omitempty"` // color_a, color_b, color_c, color_d
	Hull  int    `yaml:"hull,omitempty"`  // 1-8
	Gun   int    `yaml:"gun,omitempty"`   // 1, 2, 4, 5, 7
}

type BuildingPlacement struct {
	Type      string   `yaml:"type"`
	Position  Position `yaml:"position"`
	Completed bool     `yaml:"completed,omitempty"`
}

type VictoryConditionDef struct {
	Type     string  `yaml:"type"`
	Zone     *Zone   `yaml:"zone,omitempty"`
	Duration float64 `yaml:"duration,omitempty"`
	Count    int     `yaml:"count,omitempty"`
	Target   string  `yaml:"target,omitempty"`
}

type DefeatConditionDef struct {
	Type     string  `yaml:"type"`
	Duration float64 `yaml:"duration,omitempty"`
}

type Zone struct {
	X      float64 `yaml:"x"`
	Y      float64 `yaml:"y"`
	Width  float64 `yaml:"width"`
	Height float64 `yaml:"height"`
}

func (z *Zone) ToRect() emath.Rect {
	return emath.NewRect(z.X, z.Y, z.Width, z.Height)
}

type MissionRestrictions struct {
	CanBuild       bool     `yaml:"can_build"`
	CanProduce     bool     `yaml:"can_produce"`
	AllowedUnits   []string `yaml:"allowed_units,omitempty"`
	AllowedBuildings []string `yaml:"allowed_buildings,omitempty"`
}

type MissionBriefing struct {
	Title      string   `yaml:"title"`
	Objectives []string `yaml:"objectives"`
	Background string   `yaml:"background"`
}

func LoadMission(path string) (*Mission, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read mission file: %w", err)
	}

	var mission Mission
	if err := yaml.Unmarshal(data, &mission); err != nil {
		return nil, fmt.Errorf("failed to parse mission file: %w", err)
	}

	if mission.ID == "" {
		return nil, fmt.Errorf("mission ID is required")
	}
	if mission.Name == "" {
		return nil, fmt.Errorf("mission name is required")
	}
	if mission.MapPath == "" {
		return nil, fmt.Errorf("mission map path is required")
	}

	return &mission, nil
}

func (m *Mission) GetUnitType(typeName string) entity.UnitType {
	switch typeName {
	case "Tank":
		return entity.UnitTypeTank
	case "Scout":
		return entity.UnitTypeScout
	case "Constructor":
		return entity.UnitTypeConstructor
	default:
		return entity.UnitTypeBasic
	}
}

// GetUnitDef returns the unit definition for a placement, handling tank customization
func (up *UnitPlacement) GetUnitDef() *entity.UnitDef {
	switch up.Type {
	case "Tank":
		// If custom tank parameters are set, create a custom tank def
		if up.Color != "" && up.Hull > 0 && up.Gun > 0 {
			return entity.CreateTankDef(up.Color, up.Hull, up.Gun)
		}
		return entity.UnitDefs[entity.UnitTypeTank]
	case "Scout":
		return entity.UnitDefs[entity.UnitTypeScout]
	case "Constructor":
		return entity.UnitDefs[entity.UnitTypeConstructor]
	default:
		return nil
	}
}

func (m *Mission) GetBuildingType(typeName string) entity.BuildingType {
	switch typeName {
	case "CommandNexus":
		return entity.BuildingCommandNexus
	case "SolarArray":
		return entity.BuildingSolarArray
	case "FusionReactor":
		return entity.BuildingFusionReactor
	case "OreExtractor":
		return entity.BuildingOreExtractor
	case "AlloyFoundry":
		return entity.BuildingAlloyFoundry
	case "VehicleFactory", "Barracks":
		return entity.BuildingVehicleFactory
	case "HoverBay":
		return entity.BuildingHoverBay
	case "DataUplink":
		return entity.BuildingDataUplink
	case "Wall":
		return entity.BuildingWall
	case "AutocannonTurret":
		return entity.BuildingAutocannonTurret
	case "MissileBattery":
		return entity.BuildingMissileBattery
	case "TankFactory":
		return entity.BuildingTankFactory
	case "SolarPanel":
		return entity.BuildingSolarPanel
	case "LaserTower":
		return entity.BuildingLaserTower
	default:
		return entity.BuildingCommandNexus
	}
}

func (m *Mission) HasRestrictions() bool {
	return m.Restrictions != nil
}

func (m *Mission) CanBuild() bool {
	if m.Restrictions == nil {
		return true
	}
	return m.Restrictions.CanBuild
}

func (m *Mission) CanProduce() bool {
	if m.Restrictions == nil {
		return true
	}
	return m.Restrictions.CanProduce
}
