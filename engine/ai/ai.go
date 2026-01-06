package ai

import (
	"github.com/bklimczak/tanks/engine/entity"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/resource"
	"math"
	"math/rand"
)

type AIState int

const (
	StateIdle AIState = iota
	StateDefending
	StateAttacking
	StateProducing
)

type EnemyAI struct {
	Faction         entity.Faction
	State           AIState
	Units           []*entity.Unit
	Buildings       []*entity.Building
	Resources       *resource.Manager
	BasePosition    emath.Vec2
	PlayerUnits     []*entity.Unit
	PlayerBuildings []*entity.Building
	decisionTimer   float64
	attackTimer     float64
	produceTimer    float64
	rallyPoint      emath.Vec2
	minArmySize     int
	maxArmySize     int
}

func NewEnemyAI(baseX, baseY float64) *EnemyAI {
	resources := resource.NewManager()
	resources.Get(resource.Metal).Add(500)
	resources.Get(resource.Energy).Add(500)
	resources.AddCapacity(resource.Metal, 1000)
	resources.AddCapacity(resource.Energy, 2000)
	resources.AddProduction(resource.Metal, 5)
	resources.AddProduction(resource.Energy, 15)

	return &EnemyAI{
		Faction:       entity.FactionEnemy,
		State:         StateProducing,
		Resources:     resources,
		BasePosition:  emath.Vec2{X: baseX, Y: baseY},
		rallyPoint:    emath.Vec2{X: baseX - 100, Y: baseY},
		minArmySize:   3,
		maxArmySize:   10,
		decisionTimer: 0,
		attackTimer:   0,
		produceTimer:  0,
	}
}

func (ai *EnemyAI) Update(dt float64, allUnits []*entity.Unit, allBuildings []*entity.Building) {
	ai.updateUnitLists(allUnits, allBuildings)
	ai.Resources.Update(dt)
	ai.decisionTimer += dt
	ai.attackTimer += dt
	ai.produceTimer += dt

	if ai.decisionTimer >= 2.0 {
		ai.makeDecision()
		ai.decisionTimer = 0
	}

	switch ai.State {
	case StateIdle:
		ai.doIdle()
	case StateDefending:
		ai.doDefend()
	case StateAttacking:
		ai.doAttack()
	case StateProducing:
		ai.doProduce()
	}

	ai.updateProduction(dt)
}

func (ai *EnemyAI) updateUnitLists(allUnits []*entity.Unit, allBuildings []*entity.Building) {
	ai.Units = nil
	ai.PlayerUnits = nil
	ai.Buildings = nil
	ai.PlayerBuildings = nil

	for _, u := range allUnits {
		if !u.Active {
			continue
		}
		if u.Faction == ai.Faction {
			ai.Units = append(ai.Units, u)
		} else if u.Faction == entity.FactionPlayer {
			ai.PlayerUnits = append(ai.PlayerUnits, u)
		}
	}

	for _, b := range allBuildings {
		if b.Faction == ai.Faction {
			ai.Buildings = append(ai.Buildings, b)
		} else if b.Faction == entity.FactionPlayer {
			ai.PlayerBuildings = append(ai.PlayerBuildings, b)
		}
	}
}

func (ai *EnemyAI) makeDecision() {
	armySize := len(ai.Units)
	threatNearBase := ai.detectThreatNearBase()

	if threatNearBase {
		ai.State = StateDefending
		return
	}

	if armySize >= ai.minArmySize && ai.attackTimer >= 30.0 {
		ai.State = StateAttacking
		return
	}

	if armySize < ai.maxArmySize {
		ai.State = StateProducing
		return
	}

	ai.State = StateIdle
}

func (ai *EnemyAI) detectThreatNearBase() bool {
	threatRange := 300.0
	for _, u := range ai.PlayerUnits {
		dist := u.Center().Distance(ai.BasePosition)
		if dist < threatRange {
			return true
		}
	}
	return false
}

func (ai *EnemyAI) doIdle() {
	for _, u := range ai.Units {
		if !u.HasTarget {
			idlePos := ai.rallyPoint.Add(emath.Vec2{
				X: (rand.Float64() - 0.5) * 100,
				Y: (rand.Float64() - 0.5) * 100,
			})
			u.SetTarget(idlePos)
		}
	}
}

func (ai *EnemyAI) doDefend() {
	var threat *entity.Unit
	minDist := math.MaxFloat64

	for _, u := range ai.PlayerUnits {
		dist := u.Center().Distance(ai.BasePosition)
		if dist < minDist {
			minDist = dist
			threat = u
		}
	}

	if threat == nil {
		ai.State = StateIdle
		return
	}

	for _, u := range ai.Units {
		if u.CanAttack() {
			dist := u.Center().Distance(threat.Center())
			if dist > u.Range {
				u.SetTarget(threat.Center())
			}
		}
	}
}

func (ai *EnemyAI) doAttack() {
	var target emath.Vec2
	hasTarget := false

	if len(ai.PlayerBuildings) > 0 {
		target = ai.PlayerBuildings[0].Center()
		hasTarget = true
	} else if len(ai.PlayerUnits) > 0 {
		target = ai.PlayerUnits[0].Center()
		hasTarget = true
	}

	if !hasTarget {
		ai.State = StateIdle
		ai.attackTimer = 0
		return
	}

	allAtTarget := true
	for _, u := range ai.Units {
		dist := u.Center().Distance(target)
		if dist > u.Range*0.8 {
			allAtTarget = false
			if !u.HasTarget || u.Target.Distance(target) > 50 {
				offset := emath.Vec2{
					X: (rand.Float64() - 0.5) * 50,
					Y: (rand.Float64() - 0.5) * 50,
				}
				u.SetTarget(target.Add(offset))
			}
		}
	}

	if allAtTarget && len(ai.PlayerUnits) == 0 && len(ai.PlayerBuildings) == 0 {
		ai.State = StateIdle
		ai.attackTimer = 0
	}
}

func (ai *EnemyAI) doProduce() {
	// Production is handled in updateProduction
}

func (ai *EnemyAI) updateProduction(dt float64) {
	for _, b := range ai.Buildings {
		if b.Type != entity.BuildingTankFactory || !b.Completed {
			continue
		}

		if !b.Producing && len(b.ProductionQueue) == 0 && ai.produceTimer >= 5.0 {
			if len(ai.Units) < ai.maxArmySize {
				if rand.Float64() < 0.7 {
					b.QueueProduction(entity.UnitDefs[entity.UnitTypeTank])
				} else {
					b.QueueProduction(entity.UnitDefs[entity.UnitTypeScout])
				}
				ai.produceTimer = 0
			}
		}

		if completedUnit := b.UpdateProduction(dt, ai.Resources); completedUnit != nil {
			// Unit will be spawned by the game's building update system
		}
	}
}

func (ai *EnemyAI) GetResources() *resource.Manager {
	return ai.Resources
}
