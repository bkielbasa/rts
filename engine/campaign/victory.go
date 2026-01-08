package campaign

import (
	"github.com/bklimczak/tanks/engine/entity"
	emath "github.com/bklimczak/tanks/engine/math"
)

type GameStateReader interface {
	GetUnits() []*entity.Unit
	GetBuildings() []*entity.Building
	GetElapsedTime() float64
}

type VictoryCondition interface {
	Check(game GameStateReader) bool
	Progress(game GameStateReader) float64
	Description() string
}

type DefeatCondition interface {
	Check(game GameStateReader) bool
	Description() string
}

type DestroyAllEnemies struct{}

func (d *DestroyAllEnemies) Check(game GameStateReader) bool {
	for _, u := range game.GetUnits() {
		if u.Active && u.Faction == entity.FactionEnemy {
			return false
		}
	}
	for _, b := range game.GetBuildings() {
		if b.Faction == entity.FactionEnemy {
			return false
		}
	}
	return true
}

func (d *DestroyAllEnemies) Progress(game GameStateReader) float64 {
	total := 0
	destroyed := 0
	for _, u := range game.GetUnits() {
		if u.Faction == entity.FactionEnemy {
			total++
			if !u.Active {
				destroyed++
			}
		}
	}
	for _, b := range game.GetBuildings() {
		if b.Faction == entity.FactionEnemy {
			total++
		}
	}
	if total == 0 {
		return 1.0
	}
	return float64(destroyed) / float64(total)
}

func (d *DestroyAllEnemies) Description() string {
	return "Destroy all enemy units and buildings"
}

type SurviveForDuration struct {
	Duration float64
}

func (s *SurviveForDuration) Check(game GameStateReader) bool {
	return game.GetElapsedTime() >= s.Duration
}

func (s *SurviveForDuration) Progress(game GameStateReader) float64 {
	elapsed := game.GetElapsedTime()
	if elapsed >= s.Duration {
		return 1.0
	}
	return elapsed / s.Duration
}

func (s *SurviveForDuration) Description() string {
	minutes := int(s.Duration) / 60
	seconds := int(s.Duration) % 60
	if minutes > 0 {
		return "Survive for " + formatDuration(minutes, seconds)
	}
	return "Survive for " + formatDuration(0, seconds)
}

func formatDuration(minutes, seconds int) string {
	if minutes > 0 && seconds > 0 {
		return formatMinutes(minutes) + " " + formatSeconds(seconds)
	} else if minutes > 0 {
		return formatMinutes(minutes)
	}
	return formatSeconds(seconds)
}

func formatMinutes(m int) string {
	if m == 1 {
		return "1 minute"
	}
	return string(rune('0'+m/10)) + string(rune('0'+m%10)) + " minutes"
}

func formatSeconds(s int) string {
	if s == 1 {
		return "1 second"
	}
	return string(rune('0'+s/10)) + string(rune('0'+s%10)) + " seconds"
}

type ReachZone struct {
	Zone emath.Rect
}

func (r *ReachZone) Check(game GameStateReader) bool {
	for _, u := range game.GetUnits() {
		if u.Active && u.Faction == entity.FactionPlayer {
			if r.Zone.Contains(u.Center()) {
				return true
			}
		}
	}
	return false
}

func (r *ReachZone) Progress(game GameStateReader) float64 {
	if r.Check(game) {
		return 1.0
	}
	return 0.0
}

func (r *ReachZone) Description() string {
	return "Move a unit to the target zone"
}

type ReachZoneAll struct {
	Zone emath.Rect
}

func (r *ReachZoneAll) Check(game GameStateReader) bool {
	hasUnits := false
	for _, u := range game.GetUnits() {
		if u.Active && u.Faction == entity.FactionPlayer {
			hasUnits = true
			if !r.Zone.Contains(u.Center()) {
				return false
			}
		}
	}
	return hasUnits
}

func (r *ReachZoneAll) Progress(game GameStateReader) float64 {
	total := 0
	inZone := 0
	for _, u := range game.GetUnits() {
		if u.Active && u.Faction == entity.FactionPlayer {
			total++
			if r.Zone.Contains(u.Center()) {
				inZone++
			}
		}
	}
	if total == 0 {
		return 0.0
	}
	return float64(inZone) / float64(total)
}

func (r *ReachZoneAll) Description() string {
	return "Move all units to the target zone"
}

type DestroyCount struct {
	Count   int
	current int
}

func (d *DestroyCount) Check(game GameStateReader) bool {
	destroyed := 0
	for _, u := range game.GetUnits() {
		if u.Faction == entity.FactionEnemy && !u.Active {
			destroyed++
		}
	}
	return destroyed >= d.Count
}

func (d *DestroyCount) Progress(game GameStateReader) float64 {
	destroyed := 0
	for _, u := range game.GetUnits() {
		if u.Faction == entity.FactionEnemy && !u.Active {
			destroyed++
		}
	}
	if destroyed >= d.Count {
		return 1.0
	}
	return float64(destroyed) / float64(d.Count)
}

func (d *DestroyCount) Description() string {
	return "Destroy enemy units"
}

type BuildStructure struct {
	BuildingType entity.BuildingType
	Count        int
}

func (b *BuildStructure) Check(game GameStateReader) bool {
	count := 0
	for _, building := range game.GetBuildings() {
		if building.Faction == entity.FactionPlayer && building.Type == b.BuildingType && building.Completed {
			count++
		}
	}
	return count >= b.Count
}

func (b *BuildStructure) Progress(game GameStateReader) float64 {
	count := 0
	for _, building := range game.GetBuildings() {
		if building.Faction == entity.FactionPlayer && building.Type == b.BuildingType && building.Completed {
			count++
		}
	}
	if count >= b.Count {
		return 1.0
	}
	return float64(count) / float64(b.Count)
}

func (b *BuildStructure) Description() string {
	def := entity.BuildingDefs[b.BuildingType]
	if def != nil {
		return "Build " + def.Name
	}
	return "Build required structure"
}

type TrainUnits struct {
	Count int
}

func (t *TrainUnits) Check(game GameStateReader) bool {
	count := 0
	for _, u := range game.GetUnits() {
		if u.Active && u.Faction == entity.FactionPlayer {
			count++
		}
	}
	return count >= t.Count
}

func (t *TrainUnits) Progress(game GameStateReader) float64 {
	count := 0
	for _, u := range game.GetUnits() {
		if u.Active && u.Faction == entity.FactionPlayer {
			count++
		}
	}
	if count >= t.Count {
		return 1.0
	}
	return float64(count) / float64(t.Count)
}

func (t *TrainUnits) Description() string {
	return "Train units"
}

type LoseAllUnits struct{}

func (l *LoseAllUnits) Check(game GameStateReader) bool {
	for _, u := range game.GetUnits() {
		if u.Active && u.Faction == entity.FactionPlayer {
			return false
		}
	}
	return true
}

func (l *LoseAllUnits) Description() string {
	return "All units destroyed"
}

type LoseAllBuildings struct{}

func (l *LoseAllBuildings) Check(game GameStateReader) bool {
	for _, b := range game.GetBuildings() {
		if b.Faction == entity.FactionPlayer {
			return false
		}
	}
	return true
}

func (l *LoseAllBuildings) Description() string {
	return "All buildings destroyed"
}

type LoseAll struct{}

func (l *LoseAll) Check(game GameStateReader) bool {
	hasUnits := false
	hasBuildings := false
	for _, u := range game.GetUnits() {
		if u.Active && u.Faction == entity.FactionPlayer {
			hasUnits = true
			break
		}
	}
	for _, b := range game.GetBuildings() {
		if b.Faction == entity.FactionPlayer {
			hasBuildings = true
			break
		}
	}
	return !hasUnits && !hasBuildings
}

func (l *LoseAll) Description() string {
	return "All forces destroyed"
}

func CreateVictoryCondition(def VictoryConditionDef) VictoryCondition {
	switch def.Type {
	case "destroy_all_enemies":
		return &DestroyAllEnemies{}
	case "survive_duration":
		return &SurviveForDuration{Duration: def.Duration}
	case "reach_zone":
		if def.Zone != nil {
			return &ReachZone{Zone: def.Zone.ToRect()}
		}
	case "reach_zone_all":
		if def.Zone != nil {
			return &ReachZoneAll{Zone: def.Zone.ToRect()}
		}
	case "destroy_count":
		return &DestroyCount{Count: def.Count}
	case "build_structure":
		return &BuildStructure{
			BuildingType: (&Mission{}).GetBuildingType(def.Target),
			Count:        max(1, def.Count),
		}
	case "train_units":
		return &TrainUnits{Count: def.Count}
	}
	return &DestroyAllEnemies{}
}

func CreateDefeatCondition(def DefeatConditionDef) DefeatCondition {
	switch def.Type {
	case "lose_all_units":
		return &LoseAllUnits{}
	case "lose_all_buildings":
		return &LoseAllBuildings{}
	case "lose_all":
		return &LoseAll{}
	}
	return &LoseAll{}
}
