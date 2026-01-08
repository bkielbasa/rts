package save

import (
	"github.com/bklimczak/tanks/engine/entity"
	"github.com/bklimczak/tanks/engine/resource"
	"time"
)

const SaveVersion = 1

type SaveFile struct {
	Version   int          `yaml:"version"`
	Metadata  SaveMetadata `yaml:"metadata"`
	GameState GameState    `yaml:"game_state"`
}

type SaveMetadata struct {
	Name      string    `yaml:"name"`
	Timestamp time.Time `yaml:"timestamp"`
	MissionID string    `yaml:"mission_id,omitempty"`
	PlayTime  float64   `yaml:"play_time_seconds"`
}

type GameState struct {
	NextUnitID     uint64 `yaml:"next_unit_id"`
	NextBuildingID uint64 `yaml:"next_building_id"`
	NextWreckageID uint64 `yaml:"next_wreckage_id"`

	Resources    ResourcesState  `yaml:"resources"`
	Units        []UnitState     `yaml:"units"`
	Buildings    []BuildingState `yaml:"buildings"`
	Wreckages    []WreckageState `yaml:"wreckages"`
	FogOfWar     FogState        `yaml:"fog_of_war"`
	EnemyAI      AIState         `yaml:"enemy_ai"`
	MissionState MissionState    `yaml:"mission_state,omitempty"`
	CameraX      float64         `yaml:"camera_x"`
	CameraY      float64         `yaml:"camera_y"`
	Zoom         float64         `yaml:"zoom"`
	GameTime     float64         `yaml:"game_time"`
}

type ResourcesState struct {
	Credits ResourceState `yaml:"credits"`
	Energy  ResourceState `yaml:"energy"`
	Alloys  ResourceState `yaml:"alloys"`
}

type ResourceState struct {
	Current     float64 `yaml:"current"`
	Capacity    float64 `yaml:"capacity"`
	Production  float64 `yaml:"production"`
	Consumption float64 `yaml:"consumption"`
}

type UnitState struct {
	ID          uint64          `yaml:"id"`
	Type        entity.UnitType `yaml:"type"`
	Faction     entity.Faction  `yaml:"faction"`
	PosX        float64         `yaml:"pos_x"`
	PosY        float64         `yaml:"pos_y"`
	Health      float64         `yaml:"health"`
	Selected    bool            `yaml:"selected,omitempty"`
	Angle       float64         `yaml:"angle"`
	TurretAngle float64         `yaml:"turret_angle"`

	HasTarget bool    `yaml:"has_target,omitempty"`
	TargetX   float64 `yaml:"target_x,omitempty"`
	TargetY   float64 `yaml:"target_y,omitempty"`

	AttackTargetID         uint64 `yaml:"attack_target_id,omitempty"`
	BuildingAttackTargetID uint64 `yaml:"building_attack_target_id,omitempty"`

	HasBuildTask  bool                `yaml:"has_build_task,omitempty"`
	BuildDefType  entity.BuildingType `yaml:"build_def_type,omitempty"`
	BuildPosX     float64             `yaml:"build_pos_x,omitempty"`
	BuildPosY     float64             `yaml:"build_pos_y,omitempty"`
	BuildTargetID uint64              `yaml:"build_target_id,omitempty"`
	IsBuilding    bool                `yaml:"is_building,omitempty"`

	BuildQueue []BuildTaskState `yaml:"build_queue,omitempty"`

	RepairTargetID uint64  `yaml:"repair_target_id,omitempty"`
	FireCooldown   float64 `yaml:"fire_cooldown,omitempty"`
}

type BuildTaskState struct {
	DefType entity.BuildingType `yaml:"def_type"`
	PosX    float64             `yaml:"pos_x"`
	PosY    float64             `yaml:"pos_y"`
}

type BuildingState struct {
	ID       uint64              `yaml:"id"`
	Type     entity.BuildingType `yaml:"type"`
	Faction  entity.Faction      `yaml:"faction"`
	PosX     float64             `yaml:"pos_x"`
	PosY     float64             `yaml:"pos_y"`
	Health   float64             `yaml:"health"`
	Selected bool                `yaml:"selected,omitempty"`

	Completed     bool    `yaml:"completed"`
	BuildProgress float64 `yaml:"build_progress"`
	MetalSpent    float64 `yaml:"metal_spent,omitempty"`
	EnergySpent   float64 `yaml:"energy_spent,omitempty"`

	Producing             bool              `yaml:"producing,omitempty"`
	ProductionProgress    float64           `yaml:"production_progress,omitempty"`
	CurrentProductionType entity.UnitType   `yaml:"current_production_type,omitempty"`
	ProductionMetalSpent  float64           `yaml:"production_metal_spent,omitempty"`
	ProductionEnergySpent float64           `yaml:"production_energy_spent,omitempty"`
	ProductionQueue       []entity.UnitType `yaml:"production_queue,omitempty"`

	RallyPointX   float64 `yaml:"rally_point_x,omitempty"`
	RallyPointY   float64 `yaml:"rally_point_y,omitempty"`
	HasRallyPoint bool    `yaml:"has_rally_point,omitempty"`

	AttackTargetID uint64  `yaml:"attack_target_id,omitempty"`
	FireCooldown   float64 `yaml:"fire_cooldown,omitempty"`
}

type WreckageState struct {
	ID         uint64  `yaml:"id"`
	PosX       float64 `yaml:"pos_x"`
	PosY       float64 `yaml:"pos_y"`
	SizeX      float64 `yaml:"size_x"`
	SizeY      float64 `yaml:"size_y"`
	MetalValue float64 `yaml:"metal_value"`
}

type FogState struct {
	Width    int       `yaml:"width"`
	Height   int       `yaml:"height"`
	TileSize float64   `yaml:"tile_size"`
	Tiles    [][]int8  `yaml:"tiles"`
}

type AIState struct {
	State         int            `yaml:"state"`
	BasePositionX float64        `yaml:"base_position_x"`
	BasePositionY float64        `yaml:"base_position_y"`
	RallyPointX   float64        `yaml:"rally_point_x"`
	RallyPointY   float64        `yaml:"rally_point_y"`
	MinArmySize   int            `yaml:"min_army_size"`
	MaxArmySize   int            `yaml:"max_army_size"`
	DecisionTimer float64        `yaml:"decision_timer"`
	AttackTimer   float64        `yaml:"attack_timer"`
	ProduceTimer  float64        `yaml:"produce_timer"`
	Resources     ResourcesState `yaml:"resources"`
}

type MissionState struct {
	MissionID      string   `yaml:"mission_id,omitempty"`
	Elapsed        float64  `yaml:"elapsed,omitempty"`
	ObjectivesDone []string `yaml:"objectives_done,omitempty"`
}

func NewResourcesStateFromManager(m *resource.Manager) ResourcesState {
	return ResourcesState{
		Credits: ResourceState{
			Current:     m.Get(resource.Credits).Current,
			Capacity:    m.Get(resource.Credits).Capacity,
			Production:  m.Get(resource.Credits).Production,
			Consumption: m.Get(resource.Credits).Consumption,
		},
		Energy: ResourceState{
			Current:     m.Get(resource.Energy).Current,
			Capacity:    m.Get(resource.Energy).Capacity,
			Production:  m.Get(resource.Energy).Production,
			Consumption: m.Get(resource.Energy).Consumption,
		},
		Alloys: ResourceState{
			Current:     m.Get(resource.Alloys).Current,
			Capacity:    m.Get(resource.Alloys).Capacity,
			Production:  m.Get(resource.Alloys).Production,
			Consumption: m.Get(resource.Alloys).Consumption,
		},
	}
}

func (rs *ResourcesState) ApplyToManager(m *resource.Manager) {
	credits := m.Get(resource.Credits)
	credits.Current = rs.Credits.Current
	credits.Capacity = rs.Credits.Capacity
	credits.Production = rs.Credits.Production
	credits.Consumption = rs.Credits.Consumption

	energy := m.Get(resource.Energy)
	energy.Current = rs.Energy.Current
	energy.Capacity = rs.Energy.Capacity
	energy.Production = rs.Energy.Production
	energy.Consumption = rs.Energy.Consumption

	alloys := m.Get(resource.Alloys)
	alloys.Current = rs.Alloys.Current
	alloys.Capacity = rs.Alloys.Capacity
	alloys.Production = rs.Alloys.Production
	alloys.Consumption = rs.Alloys.Consumption
}
