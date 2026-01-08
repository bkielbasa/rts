package main

import (
	"github.com/bklimczak/tanks/engine/ai"
	"github.com/bklimczak/tanks/engine/entity"
	"github.com/bklimczak/tanks/engine/fog"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/save"
)

func (g *Game) ToSaveState() *save.GameState {
	state := &save.GameState{
		NextUnitID:     g.nextUnitID,
		NextBuildingID: g.nextBuildingID,
		NextWreckageID: g.nextWreckageID,
		Resources:      save.NewResourcesStateFromManager(g.engine.Resources),
		CameraX:        g.engine.Camera.Position.X,
		CameraY:        g.engine.Camera.Position.Y,
		Zoom:           g.engine.Camera.GetZoom(),
	}

	state.Units = make([]save.UnitState, 0, len(g.units))
	for _, u := range g.units {
		if !u.Active {
			continue
		}
		us := save.UnitState{
			ID:          u.ID,
			Type:        u.Type,
			Faction:     u.Faction,
			PosX:        u.Position.X,
			PosY:        u.Position.Y,
			Health:      u.Health,
			Selected:    u.Selected,
			Angle:       u.Angle,
			TurretAngle: u.TurretAngle,
			HasTarget:   u.HasTarget,
			TargetX:     u.Target.X,
			TargetY:     u.Target.Y,
			FireCooldown: u.FireCooldown,
		}

		if u.AttackTarget != nil && u.AttackTarget.Active {
			us.AttackTargetID = u.AttackTarget.ID
		}
		if u.BuildingAttackTarget != nil && u.BuildingAttackTarget.Active {
			us.BuildingAttackTargetID = u.BuildingAttackTarget.ID
		}

		us.HasBuildTask = u.HasBuildTask
		if u.BuildDef != nil {
			us.BuildDefType = u.BuildDef.Type
		}
		us.BuildPosX = u.BuildPos.X
		us.BuildPosY = u.BuildPos.Y
		if u.BuildTarget != nil {
			us.BuildTargetID = u.BuildTarget.ID
		}
		us.IsBuilding = u.IsBuilding

		if len(u.BuildQueue) > 0 {
			us.BuildQueue = make([]save.BuildTaskState, len(u.BuildQueue))
			for i, task := range u.BuildQueue {
				us.BuildQueue[i] = save.BuildTaskState{
					DefType: task.Def.Type,
					PosX:    task.Pos.X,
					PosY:    task.Pos.Y,
				}
			}
		}

		if u.RepairTarget != nil && u.RepairTarget.Active {
			us.RepairTargetID = u.RepairTarget.ID
		}

		state.Units = append(state.Units, us)
	}

	state.Buildings = make([]save.BuildingState, 0, len(g.buildings))
	for _, b := range g.buildings {
		bs := save.BuildingState{
			ID:            b.ID,
			Type:          b.Type,
			Faction:       b.Faction,
			PosX:          b.Position.X,
			PosY:          b.Position.Y,
			Health:        b.Health,
			Selected:      b.Selected,
			Completed:     b.Completed,
			BuildProgress: b.BuildProgress,
			MetalSpent:    b.MetalSpent,
			EnergySpent:   b.EnergySpent,
			RallyPointX:   b.RallyPoint.X,
			RallyPointY:   b.RallyPoint.Y,
			HasRallyPoint: b.HasRallyPoint,
			FireCooldown:  b.FireCooldown,
		}

		if b.AttackTarget != nil && b.AttackTarget.Active {
			bs.AttackTargetID = b.AttackTarget.ID
		}

		bs.Producing = b.Producing
		bs.ProductionProgress = b.ProductionProgress
		bs.ProductionMetalSpent = b.ProductionMetalSpent
		bs.ProductionEnergySpent = b.ProductionEnergySpent
		if b.CurrentProduction != nil {
			bs.CurrentProductionType = b.CurrentProduction.Type
		}
		if len(b.ProductionQueue) > 0 {
			bs.ProductionQueue = make([]entity.UnitType, len(b.ProductionQueue))
			for i, def := range b.ProductionQueue {
				bs.ProductionQueue[i] = def.Type
			}
		}

		state.Buildings = append(state.Buildings, bs)
	}

	state.Wreckages = make([]save.WreckageState, 0, len(g.wreckages))
	for _, w := range g.wreckages {
		if !w.Active {
			continue
		}
		state.Wreckages = append(state.Wreckages, save.WreckageState{
			ID:         w.ID,
			PosX:       w.Position.X,
			PosY:       w.Position.Y,
			SizeX:      w.Size.X,
			SizeY:      w.Size.Y,
			MetalValue: w.MetalValue,
		})
	}

	state.FogOfWar = save.FogState{
		Width:    g.fogOfWar.Width,
		Height:   g.fogOfWar.Height,
		TileSize: g.fogOfWar.TileSize,
		Tiles:    make([][]int8, g.fogOfWar.Height),
	}
	for y := 0; y < g.fogOfWar.Height; y++ {
		state.FogOfWar.Tiles[y] = make([]int8, g.fogOfWar.Width)
		for x := 0; x < g.fogOfWar.Width; x++ {
			state.FogOfWar.Tiles[y][x] = int8(g.fogOfWar.Tiles[y][x])
		}
	}

	if g.enemyAI != nil {
		basePos := g.enemyAI.GetBasePosition()
		rallyPoint := g.enemyAI.GetRallyPoint()
		decision, attack, produce := g.enemyAI.GetTimers()
		minArmy, maxArmy := g.enemyAI.GetArmySizes()

		state.EnemyAI = save.AIState{
			State:         int(g.enemyAI.GetState()),
			BasePositionX: basePos.X,
			BasePositionY: basePos.Y,
			RallyPointX:   rallyPoint.X,
			RallyPointY:   rallyPoint.Y,
			MinArmySize:   minArmy,
			MaxArmySize:   maxArmy,
			DecisionTimer: decision,
			AttackTimer:   attack,
			ProduceTimer:  produce,
			Resources:     save.NewResourcesStateFromManager(g.enemyAI.GetResources()),
		}
	}

	return state
}

func (g *Game) LoadFromSaveState(state *save.GameState) {
	g.units = nil
	g.buildings = nil
	g.wreckages = nil
	g.projectiles = nil

	g.nextUnitID = state.NextUnitID
	g.nextBuildingID = state.NextBuildingID
	g.nextWreckageID = state.NextWreckageID

	state.Resources.ApplyToManager(g.engine.Resources)

	g.engine.Camera.Position.X = state.CameraX
	g.engine.Camera.Position.Y = state.CameraY
	g.engine.Camera.SetZoom(state.Zoom)

	unitMap := make(map[uint64]*entity.Unit)
	buildingMap := make(map[uint64]*entity.Building)

	for _, us := range state.Units {
		def := entity.UnitDefs[us.Type]
		if def == nil {
			continue
		}
		u := entity.NewUnitFromDef(us.ID, us.PosX, us.PosY, def, us.Faction)
		u.Health = us.Health
		u.Selected = us.Selected
		u.Angle = us.Angle
		u.TurretAngle = us.TurretAngle
		u.HasTarget = us.HasTarget
		u.Target = emath.Vec2{X: us.TargetX, Y: us.TargetY}
		u.FireCooldown = us.FireCooldown
		u.HasBuildTask = us.HasBuildTask
		if us.HasBuildTask {
			u.BuildDef = entity.BuildingDefs[us.BuildDefType]
		}
		u.BuildPos = emath.Vec2{X: us.BuildPosX, Y: us.BuildPosY}
		u.IsBuilding = us.IsBuilding

		if len(us.BuildQueue) > 0 {
			u.BuildQueue = make([]entity.BuildTask, len(us.BuildQueue))
			for i, task := range us.BuildQueue {
				u.BuildQueue[i] = entity.BuildTask{
					Def: entity.BuildingDefs[task.DefType],
					Pos: emath.Vec2{X: task.PosX, Y: task.PosY},
				}
			}
		}

		g.units = append(g.units, u)
		unitMap[u.ID] = u
	}

	for _, bs := range state.Buildings {
		def := entity.BuildingDefs[bs.Type]
		if def == nil {
			continue
		}
		var b *entity.Building
		if bs.Completed {
			b = entity.NewBuilding(bs.ID, bs.PosX, bs.PosY, def)
		} else {
			b = entity.NewBuildingUnderConstruction(bs.ID, bs.PosX, bs.PosY, def)
			b.BuildProgress = bs.BuildProgress
			b.MetalSpent = bs.MetalSpent
			b.EnergySpent = bs.EnergySpent
		}
		b.Faction = bs.Faction
		b.Health = bs.Health
		b.Selected = bs.Selected
		b.RallyPoint = emath.Vec2{X: bs.RallyPointX, Y: bs.RallyPointY}
		b.HasRallyPoint = bs.HasRallyPoint
		b.FireCooldown = bs.FireCooldown

		b.Producing = bs.Producing
		b.ProductionProgress = bs.ProductionProgress
		b.ProductionMetalSpent = bs.ProductionMetalSpent
		b.ProductionEnergySpent = bs.ProductionEnergySpent
		if bs.Producing {
			b.CurrentProduction = entity.UnitDefs[bs.CurrentProductionType]
			if b.CurrentProduction != nil {
				b.ProductionTime = b.CurrentProduction.BuildTime
			}
		}
		if len(bs.ProductionQueue) > 0 {
			b.ProductionQueue = make([]*entity.UnitDef, len(bs.ProductionQueue))
			for i, unitType := range bs.ProductionQueue {
				b.ProductionQueue[i] = entity.UnitDefs[unitType]
			}
		}

		if bs.Faction == entity.FactionEnemy {
			b.Color = entity.GetFactionTintedColor(def.Color, entity.FactionEnemy)
		}

		g.buildings = append(g.buildings, b)
		buildingMap[b.ID] = b
	}

	for i, us := range state.Units {
		if us.AttackTargetID != 0 {
			if target, ok := unitMap[us.AttackTargetID]; ok {
				g.units[i].AttackTarget = target
			}
		}
		if us.BuildingAttackTargetID != 0 {
			if target, ok := buildingMap[us.BuildingAttackTargetID]; ok {
				g.units[i].BuildingAttackTarget = target
			}
		}
		if us.BuildTargetID != 0 {
			if target, ok := buildingMap[us.BuildTargetID]; ok {
				g.units[i].BuildTarget = target
			}
		}
		if us.RepairTargetID != 0 {
			if target, ok := unitMap[us.RepairTargetID]; ok {
				g.units[i].RepairTarget = target
			}
		}
	}

	for i, bs := range state.Buildings {
		if bs.AttackTargetID != 0 {
			if target, ok := unitMap[bs.AttackTargetID]; ok {
				g.buildings[i].AttackTarget = target
			}
		}
	}

	for _, ws := range state.Wreckages {
		w := &entity.Wreckage{
			Entity: entity.Entity{
				ID:       ws.ID,
				Position: emath.Vec2{X: ws.PosX, Y: ws.PosY},
				Size:     emath.Vec2{X: ws.SizeX, Y: ws.SizeY},
				Color:    entity.WreckageColor,
				Active:   true,
				Faction:  entity.FactionNeutral,
			},
			MetalValue: ws.MetalValue,
		}
		g.wreckages = append(g.wreckages, w)
	}

	if state.FogOfWar.Width > 0 && state.FogOfWar.Height > 0 {
		g.fogOfWar.Width = state.FogOfWar.Width
		g.fogOfWar.Height = state.FogOfWar.Height
		g.fogOfWar.TileSize = state.FogOfWar.TileSize
		g.fogOfWar.Tiles = make([][]fog.TileState, state.FogOfWar.Height)
		for y := 0; y < state.FogOfWar.Height; y++ {
			g.fogOfWar.Tiles[y] = make([]fog.TileState, state.FogOfWar.Width)
			for x := 0; x < state.FogOfWar.Width; x++ {
				g.fogOfWar.Tiles[y][x] = fog.TileState(state.FogOfWar.Tiles[y][x])
			}
		}
	}

	if state.EnemyAI.BasePositionX != 0 || state.EnemyAI.BasePositionY != 0 {
		g.enemyAI = ai.NewEnemyAI(state.EnemyAI.BasePositionX, state.EnemyAI.BasePositionY)
		g.enemyAI.SetState(ai.AIState(state.EnemyAI.State))
		g.enemyAI.SetRallyPoint(emath.Vec2{X: state.EnemyAI.RallyPointX, Y: state.EnemyAI.RallyPointY})
		g.enemyAI.SetTimers(state.EnemyAI.DecisionTimer, state.EnemyAI.AttackTimer, state.EnemyAI.ProduceTimer)
		g.enemyAI.SetArmySizes(state.EnemyAI.MinArmySize, state.EnemyAI.MaxArmySize)
		state.EnemyAI.Resources.ApplyToManager(g.enemyAI.GetResources())
	}

	g.placementMode = false
	g.placementDef = nil
	g.terrainCache = nil
}
