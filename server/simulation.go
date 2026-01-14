package server

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/bklimczak/tanks/engine/collision"
	"github.com/bklimczak/tanks/engine/entity"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/resource"
	"github.com/bklimczak/tanks/engine/terrain"
)

const (
	TickRate     = 1.0 / 60.0 // 60 FPS
	TickDuration = time.Second / 60
)

// Spawn positions for up to 4 players (corners of map)
var SpawnPositions = []emath.Vec2{
	{X: 400, Y: 300},   // Player 0: Top-left
	{X: 5800, Y: 3100}, // Player 1: Bottom-right
	{X: 5800, Y: 300},  // Player 2: Top-right
	{X: 400, Y: 3100},  // Player 3: Bottom-left
}

// PlayerSetup contains initial player configuration
type PlayerSetup struct {
	PlayerID string
	Name     string
	Slot     int // 0-3
}

// PlayerCommand represents a command from a player
type PlayerCommand struct {
	PlayerID string
	Slot     int
	Command  GameCommand
}

// Simulation runs the game logic on the server
type Simulation struct {
	units       []*entity.Unit
	buildings   []*entity.Building
	projectiles []*entity.Projectile
	terrainMap  *terrain.Map
	collision   *collision.System

	// Per-player state
	playerResources map[int]*resource.Manager // Slot -> Resources
	playerAlive     map[int]bool              // Slot -> Alive
	playerIDs       map[int]string            // Slot -> PlayerID
	playerNames     map[int]string            // Slot -> Name
	numPlayers      int

	// Entity ID generation
	nextUnitID       uint64
	nextBuildingID   uint64
	nextProjectileID uint64

	// Simulation state
	tick    uint64
	running bool

	// Command queue
	commandQueue chan PlayerCommand

	mu sync.RWMutex
}

// NewSimulation creates a new game simulation
func NewSimulation(players []PlayerSetup) *Simulation {
	// Create terrain (simple grass map for now)
	terrainMap := terrain.NewMap(6400, 3600)
	terrainMap.GenerateGrassOnly()

	// Place metal deposits in the center and near each spawn
	centerX := float64(terrainMap.PixelWidth) / 2
	centerY := float64(terrainMap.PixelHeight) / 2

	// Center deposits (contested)
	terrainMap.PlaceMetalDeposit(centerX, centerY)
	terrainMap.PlaceMetalDeposit(centerX+50, centerY)
	terrainMap.PlaceMetalDeposit(centerX-50, centerY)
	terrainMap.PlaceMetalDeposit(centerX, centerY+50)
	terrainMap.PlaceMetalDeposit(centerX, centerY-50)

	s := &Simulation{
		units:           make([]*entity.Unit, 0),
		buildings:       make([]*entity.Building, 0),
		projectiles:     make([]*entity.Projectile, 0),
		terrainMap:      terrainMap,
		collision:       collision.NewSystem(float64(terrainMap.PixelWidth), float64(terrainMap.PixelHeight)),
		playerResources: make(map[int]*resource.Manager),
		playerAlive:     make(map[int]bool),
		playerIDs:       make(map[int]string),
		playerNames:     make(map[int]string),
		numPlayers:      len(players),
		commandQueue:    make(chan PlayerCommand, 256),
	}

	s.collision.SetTerrain(terrainMap)

	// Spawn each player's base
	for _, setup := range players {
		s.spawnPlayerBase(setup)
	}

	return s
}

// spawnPlayerBase creates starting units and buildings for a player
func (s *Simulation) spawnPlayerBase(setup PlayerSetup) {
	slot := setup.Slot
	spawn := SpawnPositions[slot]
	faction := slotToFaction(slot)

	// Store player info
	s.playerIDs[slot] = setup.PlayerID
	s.playerNames[slot] = setup.Name
	s.playerAlive[slot] = true

	// Initialize resources
	res := resource.NewManager()
	res.Get(resource.Metal).Current = 1000
	res.Get(resource.Metal).Capacity = 2000
	res.Get(resource.Energy).Current = 100
	res.Get(resource.Energy).Capacity = 200
	s.playerResources[slot] = res

	// Spawn Command Nexus
	nexusDef := entity.BuildingDefs[entity.BuildingCommandNexus]
	nexus := entity.NewBuilding(s.nextBuildingID, spawn.X, spawn.Y, nexusDef)
	nexus.Faction = faction
	nexus.Completed = true
	nexus.BuildProgress = 1.0
	s.buildings = append(s.buildings, nexus)
	s.nextBuildingID++
	s.applyBuildingEffects(slot, nexusDef)

	// Spawn 2 tanks
	tankDef := entity.UnitDefs[entity.UnitTypeTank]
	for i := 0; i < 2; i++ {
		tx := spawn.X + float64(i)*50 - 25
		ty := spawn.Y + nexusDef.Size + 40
		tank := entity.NewUnitFromDef(s.nextUnitID, tx, ty, tankDef, faction)
		s.units = append(s.units, tank)
		s.nextUnitID++
	}

	// Spawn 1 scout
	scoutDef := entity.UnitDefs[entity.UnitTypeScout]
	scout := entity.NewUnitFromDef(s.nextUnitID, spawn.X+80, spawn.Y+nexusDef.Size+40, scoutDef, faction)
	s.units = append(s.units, scout)
	s.nextUnitID++

	// Spawn Solar Array for starting energy production
	solarDef := entity.BuildingDefs[entity.BuildingSolarArray]
	solar := entity.NewBuilding(s.nextBuildingID, spawn.X+nexusDef.Size+20, spawn.Y, solarDef)
	solar.Faction = faction
	solar.Completed = true
	solar.BuildProgress = 1.0
	s.buildings = append(s.buildings, solar)
	s.nextBuildingID++
	s.applyBuildingEffects(slot, solarDef)

	// Place metal deposits near spawn
	metalX := spawn.X - 100
	metalY := spawn.Y - 100
	if metalX > 0 && metalY > 0 {
		s.terrainMap.PlaceMetalDeposit(metalX, metalY)
		s.terrainMap.PlaceMetalDeposit(metalX+50, metalY)
	}
}

// slotToFaction converts a player slot to a faction
func slotToFaction(slot int) entity.Faction {
	// Use different factions for each player
	// We'll extend the faction system, but for now use Player/Enemy
	switch slot {
	case 0:
		return entity.FactionPlayer
	case 1:
		return entity.FactionEnemy
	default:
		// For slots 2 and 3, we need to extend the faction system
		// For now, use a workaround with neutral
		return entity.Faction(slot + 10) // Custom faction IDs
	}
}

// factionToSlot converts a faction back to a slot
func factionToSlot(faction entity.Faction) int {
	switch faction {
	case entity.FactionPlayer:
		return 0
	case entity.FactionEnemy:
		return 1
	default:
		return int(faction) - 10
	}
}

// applyBuildingEffects applies resource production/consumption effects
func (s *Simulation) applyBuildingEffects(slot int, def *entity.BuildingDef) {
	res := s.playerResources[slot]
	if res == nil {
		return
	}

	if def.MetalProduction > 0 {
		res.AddProduction(resource.Metal, def.MetalProduction)
	}
	if def.EnergyProduction > 0 {
		res.AddProduction(resource.Energy, def.EnergyProduction)
	}
	if def.MetalConsumption > 0 {
		res.AddConsumption(resource.Metal, def.MetalConsumption)
	}
	if def.EnergyConsumption > 0 {
		res.AddConsumption(resource.Energy, def.EnergyConsumption)
	}
	if def.MetalStorage > 0 {
		res.AddCapacity(resource.Metal, def.MetalStorage)
	}
	if def.EnergyStorage > 0 {
		res.AddCapacity(resource.Energy, def.EnergyStorage)
	}
}

// Run starts the game simulation loop
func (s *Simulation) Run(ctx context.Context, lobby *Lobby) {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	ticker := time.NewTicker(TickDuration)
	defer ticker.Stop()

	log.Printf("Simulation started for lobby")

	for {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
			log.Printf("Simulation stopped")
			return

		case <-ticker.C:
			s.mu.Lock()

			// Process commands
			s.processCommands()

			// Update game state
			s.updateResources()
			s.updateUnits()
			s.updateBuildings()
			s.updateCombat()
			s.updateProjectiles()
			s.cleanupDead()

			s.tick++

			// Check victory conditions
			finished, winnerSlot := s.checkVictory()

			// Get game state
			state := s.getGameState()

			s.mu.Unlock()

			// Broadcast state to all players
			lobby.BroadcastPayload(MsgGameState, state)

			// Handle game end
			if finished {
				winnerName := ""
				if name, ok := s.playerNames[winnerSlot]; ok {
					winnerName = name
				}
				lobby.BroadcastPayload(MsgGameEnd, GameEndPayload{
					WinnerSlot: winnerSlot,
					WinnerName: winnerName,
					Reason:     "last_standing",
				})
				lobby.Stop()
				return
			}
		}
	}
}

// EnqueueCommand adds a command to the processing queue
func (s *Simulation) EnqueueCommand(playerID string, slot int, cmd GameCommand) {
	select {
	case s.commandQueue <- PlayerCommand{PlayerID: playerID, Slot: slot, Command: cmd}:
	default:
		log.Printf("Command queue full, dropping command from player %s", playerID)
	}
}

// processCommands processes all queued commands
func (s *Simulation) processCommands() {
	for {
		select {
		case cmd := <-s.commandQueue:
			s.executeCommand(cmd)
		default:
			return
		}
	}
}

// executeCommand executes a single player command
func (s *Simulation) executeCommand(pc PlayerCommand) {
	slot := pc.Slot
	cmd := pc.Command
	faction := slotToFaction(slot)

	switch cmd.Type {
	case CmdMove:
		for _, unitID := range cmd.UnitIDs {
			if unit := s.getUnit(unitID); unit != nil && unit.Faction == faction {
				unit.SetTarget(emath.Vec2{X: cmd.TargetX, Y: cmd.TargetY})
				unit.ClearAttackTarget()
			}
		}

	case CmdAttack:
		target := s.getUnit(cmd.TargetID)
		targetBuilding := s.getBuilding(cmd.TargetID)

		for _, unitID := range cmd.UnitIDs {
			if unit := s.getUnit(unitID); unit != nil && unit.Faction == faction {
				if target != nil && target.Faction != faction {
					unit.SetAttackTarget(target)
				} else if targetBuilding != nil && targetBuilding.Faction != faction {
					unit.SetBuildingAttackTarget(targetBuilding)
				}
			}
		}

	case CmdStop:
		for _, unitID := range cmd.UnitIDs {
			if unit := s.getUnit(unitID); unit != nil && unit.Faction == faction {
				unit.ClearTarget()
				unit.ClearAttackTarget()
			}
		}

	case CmdPlaceBuilding:
		buildingType := entity.BuildingType(cmd.BuildingType)
		def := entity.BuildingDefs[buildingType]
		if def == nil {
			return
		}

		pos := emath.Vec2{X: cmd.TargetX, Y: cmd.TargetY}
		if s.canPlaceBuilding(pos, def) {
			res := s.playerResources[slot]
			if res != nil && res.CanAfford(def.Cost) {
				building := entity.NewBuildingUnderConstruction(s.nextBuildingID, pos.X, pos.Y, def)
				building.Faction = faction
				s.buildings = append(s.buildings, building)
				s.nextBuildingID++
			}
		}

	case CmdProduceUnit:
		building := s.getBuilding(cmd.BuildingID)
		if building == nil || building.Faction != faction || !building.CanProduce() {
			return
		}

		unitType := entity.UnitType(cmd.UnitType)
		unitDef := entity.UnitDefs[unitType]
		if unitDef == nil {
			return
		}

		building.QueueProduction(unitDef)

	case CmdCancelProduction:
		building := s.getBuilding(cmd.BuildingID)
		if building == nil || building.Faction != faction {
			return
		}

		unitType := entity.UnitType(cmd.UnitType)
		building.RemoveFromQueue(unitType, s.playerResources[slot])

	case CmdSetRallyPoint:
		building := s.getBuilding(cmd.BuildingID)
		if building == nil || building.Faction != faction {
			return
		}

		building.RallyPoint = emath.Vec2{X: cmd.TargetX, Y: cmd.TargetY}
		building.HasRallyPoint = true
	}
}

// getUnit finds a unit by ID
func (s *Simulation) getUnit(id uint64) *entity.Unit {
	for _, u := range s.units {
		if u.ID == id && u.Active {
			return u
		}
	}
	return nil
}

// getBuilding finds a building by ID
func (s *Simulation) getBuilding(id uint64) *entity.Building {
	for _, b := range s.buildings {
		if b.ID == id && b.Active {
			return b
		}
	}
	return nil
}

// canPlaceBuilding checks if a building can be placed at the given position
func (s *Simulation) canPlaceBuilding(pos emath.Vec2, def *entity.BuildingDef) bool {
	bounds := emath.NewRect(pos.X, pos.Y, def.Size, def.Size)

	if !s.terrainMap.IsBuildable(bounds) {
		return false
	}

	// Check for metal deposit requirement
	if def.RequiresDeposit {
		if !s.hasMetal(bounds) {
			return false
		}
	}

	// Check unit collisions
	for _, u := range s.units {
		if u.Active && bounds.Intersects(u.Bounds()) {
			return false
		}
	}

	// Check building collisions
	for _, b := range s.buildings {
		if b.Active && bounds.Intersects(b.Bounds()) {
			return false
		}
	}

	return true
}

// hasMetal checks if a location has metal deposits
func (s *Simulation) hasMetal(bounds emath.Rect) bool {
	startX, startY := s.terrainMap.GetTileCoords(bounds.Pos.X, bounds.Pos.Y)
	endX, endY := s.terrainMap.GetTileCoords(bounds.Pos.X+bounds.Size.X, bounds.Pos.Y+bounds.Size.Y)

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			if x >= 0 && x < s.terrainMap.Width && y >= 0 && y < s.terrainMap.Height {
				if s.terrainMap.Tiles[y][x].HasMetal {
					return true
				}
			}
		}
	}
	return false
}

// updateResources updates resource production for all players
func (s *Simulation) updateResources() {
	for slot, res := range s.playerResources {
		if s.playerAlive[slot] {
			res.Update(TickRate)
		}
	}
}

// updateUnits updates all unit positions and states
func (s *Simulation) updateUnits() {
	for _, u := range s.units {
		if !u.Active || !u.HasTarget {
			continue
		}

		desiredPos := u.Update()

		// Gather obstacles
		obstacles := make([]emath.Rect, 0, len(s.units)+len(s.buildings)-1)
		for _, other := range s.units {
			if other.ID != u.ID && other.Active {
				obstacles = append(obstacles, other.Bounds())
			}
		}
		for _, b := range s.buildings {
			if b.Active {
				obstacles = append(obstacles, b.Bounds())
			}
		}

		// Resolve movement
		resolvedPos := s.collision.ResolveMovement(u.Bounds(), desiredPos, obstacles)
		u.ApplyPosition(resolvedPos)
	}
}

// updateBuildings updates building construction and production
func (s *Simulation) updateBuildings() {
	for _, b := range s.buildings {
		if !b.Active {
			continue
		}

		slot := factionToSlot(b.Faction)
		res := s.playerResources[slot]

		// Construction
		if !b.Completed {
			if b.UpdateConstruction(TickRate, res) {
				s.applyBuildingEffects(slot, b.Def)
			}
		}

		// Production
		if completedUnit := b.UpdateProduction(TickRate, res); completedUnit != nil {
			spawnPos := b.GetSpawnPoint()
			unit := entity.NewUnitFromDef(s.nextUnitID, spawnPos.X, spawnPos.Y, completedUnit, b.Faction)
			s.units = append(s.units, unit)
			s.nextUnitID++

			if b.HasRallyPoint {
				unit.SetTarget(b.RallyPoint)
			}
		}
	}
}

// updateCombat handles combat between units
func (s *Simulation) updateCombat() {
	for _, u := range s.units {
		if !u.Active || !u.CanAttack() {
			continue
		}

		// Clear invalid targets
		if u.AttackTarget != nil && !u.AttackTarget.Active {
			u.AttackTarget = nil
		}
		if u.BuildingAttackTarget != nil && !u.BuildingAttackTarget.Active {
			u.BuildingAttackTarget = nil
		}

		// Auto-acquire targets
		if !u.HasAnyAttackTarget() {
			var nearestEnemy *entity.Unit
			var nearestBuilding *entity.Building
			nearestUnitDist := u.Range + 1
			nearestBuildingDist := u.Range + 1

			for _, other := range s.units {
				if other.Active && other.Faction != u.Faction {
					dist := u.Center().Distance(other.Center())
					if dist <= u.Range && dist < nearestUnitDist {
						nearestUnitDist = dist
						nearestEnemy = other
					}
				}
			}

			for _, b := range s.buildings {
				if b.Active && b.Faction != u.Faction {
					dist := u.Center().Distance(b.Center())
					if dist <= u.Range && dist < nearestBuildingDist {
						nearestBuildingDist = dist
						nearestBuilding = b
					}
				}
			}

			if nearestEnemy != nil {
				u.SetAttackTarget(nearestEnemy)
			} else if nearestBuilding != nil {
				u.SetBuildingAttackTarget(nearestBuilding)
			}
		}

		// Fire projectile
		if u.UpdateCombat(TickRate) {
			if u.AttackTarget != nil && u.AttackTarget.Active {
				projectile := entity.NewProjectile(s.nextProjectileID, u, u.AttackTarget)
				s.projectiles = append(s.projectiles, projectile)
				s.nextProjectileID++
			} else if u.BuildingAttackTarget != nil && u.BuildingAttackTarget.Active {
				projectile := entity.NewProjectileAtBuilding(s.nextProjectileID, u, u.BuildingAttackTarget)
				s.projectiles = append(s.projectiles, projectile)
				s.nextProjectileID++
			}
		}
	}

	// Building combat
	for _, b := range s.buildings {
		if !b.Active || !b.CanAttack() {
			continue
		}

		if b.FireCooldown > 0 {
			b.FireCooldown -= TickRate
		}

		if b.AttackTarget != nil && (!b.AttackTarget.Active || !b.IsInAttackRange(b.AttackTarget)) {
			b.AttackTarget = nil
		}

		if b.AttackTarget == nil {
			var nearestEnemy *entity.Unit
			nearestDist := b.Def.AttackRange + 1

			for _, u := range s.units {
				if u.Active && u.Faction != b.Faction {
					dist := b.Center().Distance(u.Center())
					if dist <= b.Def.AttackRange && dist < nearestDist {
						nearestDist = dist
						nearestEnemy = u
					}
				}
			}

			if nearestEnemy != nil {
				b.SetAttackTarget(nearestEnemy)
			}
		}

		if b.FireCooldown <= 0 && b.AttackTarget != nil && b.AttackTarget.Active {
			slot := factionToSlot(b.Faction)
			res := s.playerResources[slot]
			energyRes := res.Get(resource.Energy)

			if energyRes.Current >= b.Def.EnergyPerShot {
				energyRes.Spend(b.Def.EnergyPerShot)
				projectile := entity.NewProjectileFromBuilding(s.nextProjectileID, b, b.AttackTarget)
				s.projectiles = append(s.projectiles, projectile)
				s.nextProjectileID++
				b.FireCooldown = 1.0 / b.Def.FireRate
			}
		}
	}
}

// updateProjectiles updates projectile movement
func (s *Simulation) updateProjectiles() {
	alive := make([]*entity.Projectile, 0, len(s.projectiles))
	for _, p := range s.projectiles {
		if !p.Update(TickRate) {
			alive = append(alive, p)
		}
	}
	s.projectiles = alive
}

// cleanupDead removes dead units and buildings
func (s *Simulation) cleanupDead() {
	alive := make([]*entity.Unit, 0, len(s.units))
	for _, u := range s.units {
		if u.Active {
			alive = append(alive, u)
		}
	}
	s.units = alive

	aliveBuildings := make([]*entity.Building, 0, len(s.buildings))
	for _, b := range s.buildings {
		if b.Active {
			aliveBuildings = append(aliveBuildings, b)
		}
	}
	s.buildings = aliveBuildings
}

// checkVictory checks if the game has ended
func (s *Simulation) checkVictory() (finished bool, winnerSlot int) {
	// Count alive players
	alivePlayers := make(map[int]bool)

	for _, u := range s.units {
		if u.Active {
			slot := factionToSlot(u.Faction)
			alivePlayers[slot] = true
		}
	}

	for _, b := range s.buildings {
		if b.Active {
			slot := factionToSlot(b.Faction)
			alivePlayers[slot] = true
		}
	}

	// Update player alive status
	for slot := 0; slot < s.numPlayers; slot++ {
		s.playerAlive[slot] = alivePlayers[slot]
	}

	// Last player standing wins
	if len(alivePlayers) == 1 {
		for slot := range alivePlayers {
			return true, slot
		}
	}

	// No players alive = draw
	if len(alivePlayers) == 0 {
		return true, -1
	}

	return false, -1
}

// getGameState returns the current game state for broadcasting
func (s *Simulation) getGameState() GameStatePayload {
	players := make([]PlayerGameState, 0, s.numPlayers)
	for slot := 0; slot < s.numPlayers; slot++ {
		res := s.playerResources[slot]
		var resState ResourceStateNet
		if res != nil {
			metal := res.Get(resource.Metal)
			energy := res.Get(resource.Energy)
			resState = ResourceStateNet{
				Metal:      metal.Current,
				MetalCap:   metal.Capacity,
				MetalProd:  metal.NetFlow(),
				Energy:     energy.Current,
				EnergyCap:  energy.Capacity,
				EnergyProd: energy.NetFlow(),
			}
		}

		players = append(players, PlayerGameState{
			Slot:      slot,
			Name:      s.playerNames[slot],
			Alive:     s.playerAlive[slot],
			Resources: resState,
		})
	}

	units := make([]UnitState, 0, len(s.units))
	for _, u := range s.units {
		if !u.Active {
			continue
		}
		units = append(units, UnitState{
			ID:          u.ID,
			Type:        int(u.Type),
			OwnerSlot:   factionToSlot(u.Faction),
			PosX:        u.Position.X,
			PosY:        u.Position.Y,
			Health:      u.Health,
			MaxHealth:   u.MaxHealth,
			Angle:       u.Angle,
			TurretAngle: u.TurretAngle,
			HasTarget:   u.HasTarget,
			TargetX:     u.Target.X,
			TargetY:     u.Target.Y,
		})
	}

	buildings := make([]BuildingState, 0, len(s.buildings))
	for _, b := range s.buildings {
		if !b.Active {
			continue
		}

		var prodType int
		if b.CurrentProduction != nil {
			prodType = int(b.CurrentProduction.Type)
		}

		buildings = append(buildings, BuildingState{
			ID:            b.ID,
			Type:          int(b.Type),
			OwnerSlot:     factionToSlot(b.Faction),
			PosX:          b.Position.X,
			PosY:          b.Position.Y,
			Health:        b.Health,
			MaxHealth:     b.MaxHealth,
			Completed:     b.Completed,
			BuildProgress: b.BuildProgress,
			Producing:     b.Producing,
			ProdProgress:  b.ProductionProgress,
			ProdType:      prodType,
		})
	}

	projectiles := make([]ProjectileState, 0, len(s.projectiles))
	for _, p := range s.projectiles {
		if !p.Active {
			continue
		}

		var targetX, targetY float64
		if p.Target != nil {
			targetX = p.Target.Position.X
			targetY = p.Target.Position.Y
		} else if p.BuildingTarget != nil {
			targetX = p.BuildingTarget.Position.X
			targetY = p.BuildingTarget.Position.Y
		}

		projectiles = append(projectiles, ProjectileState{
			ID:        p.ID,
			OwnerSlot: factionToSlot(p.Faction),
			PosX:      p.Position.X,
			PosY:      p.Position.Y,
			TargetX:   targetX,
			TargetY:   targetY,
		})
	}

	return GameStatePayload{
		Tick:        s.tick,
		Players:     players,
		Units:       units,
		Buildings:   buildings,
		Projectiles: projectiles,
	}
}
