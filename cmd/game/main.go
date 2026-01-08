package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/bklimczak/tanks/engine"
	"github.com/bklimczak/tanks/engine/ai"
	"github.com/bklimczak/tanks/engine/entity"
	"github.com/bklimczak/tanks/engine/fog"
	"github.com/bklimczak/tanks/engine/input"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/resource"
	"github.com/bklimczak/tanks/engine/terrain"
	"github.com/bklimczak/tanks/engine/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
)
const (
	unitSize         = 20.0
	selectionMargin  = 2.0
	tickRate         = 1.0 / 60.0
	worldMultiplier  = 5.0
	baseWidth        = 1280.0
	baseHeight       = 720.0
	worldWidth       = baseWidth * worldMultiplier
	worldHeight      = baseHeight * worldMultiplier
	minimapWidth     = 160.0
	minimapHeight    = 120.0
	minimapMargin    = 10.0
	buildingGridSize = terrain.TileSize
)

type Game struct {
	engine            *engine.Engine
	units             []*entity.Unit
	buildings         []*entity.Building
	wreckages         []*entity.Wreckage
	projectiles       []*entity.Projectile
	terrainMap        *terrain.Map
	fogOfWar          *fog.FogOfWar
	resourceBar       *ui.ResourceBar
	commandPanel      *ui.CommandPanel
	minimap           *ui.Minimap
	mainMenu          *ui.MainMenu
	enemyAI           *ai.EnemyAI
	tankSprite        *ebiten.Image
	tankFactorySprite *ebiten.Image
	scoutSprite       *ebiten.Image
	tileImages        map[color.RGBA]*ebiten.Image
	terrainCache      *ebiten.Image
	screenWidth       int
	screenHeight      int
	debugTerrainTime  time.Duration
	debugMinimapTime  time.Duration
	debugFogTiles     int
	nextUnitID        uint64
	nextBuildingID    uint64
	nextWreckageID    uint64
	nextProjectileID  uint64
	state             GameState
	placementMode     bool
	placementDef      *entity.BuildingDef
	placementValid    bool
	whitePixel        *ebiten.Image
}

func NewGame() *Game {
	resourceBar := ui.NewResourceBar(baseWidth)
	commandPanel := ui.NewCommandPanel(resourceBar.Height(), baseHeight)
	minimapY := baseHeight - minimapHeight - minimapMargin
	minimap := ui.NewMinimap(minimapMargin, minimapY, minimapWidth, minimapHeight)
	mainMenu := ui.NewMainMenu()
	tankSprite, _, err := ebitenutil.NewImageFromFile("assets/tank.png")
	if err != nil {
		log.Printf("Warning: could not load tank sprite: %v", err)
	}
	tankFactorySprite, _, err := ebitenutil.NewImageFromFile("assets/tank_factory.png")
	if err != nil {
		log.Printf("Warning: could not load tank factory sprite: %v", err)
	}
	scoutSprite, _, err := ebitenutil.NewImageFromFile("assets/scout.png")
	if err != nil {
		log.Printf("Warning: could not load scout sprite: %v", err)
	}
	terrainMap, err := terrain.LoadMapFromFile("maps/default.yaml")
	if err != nil {
		terrainMap = terrain.NewMap(worldWidth, worldHeight)
		terrainMap.Generate(42)
		terrainMap.PlaceMetalDeposit(400, 150)
		if saveErr := terrain.SaveMapToFile(terrainMap, "maps/default.yaml", "Default Map", "Auto-generated default map", "System"); saveErr != nil {
			log.Printf("Warning: could not save map file: %v", saveErr)
		}
	}
	actualWorldWidth := terrainMap.PixelWidth
	actualWorldHeight := terrainMap.PixelHeight
	minimap.SetWorldSize(actualWorldWidth, actualWorldHeight)
	g := &Game{
		engine:            engine.New(actualWorldWidth, actualWorldHeight, baseWidth, baseHeight),
		terrainMap:        terrainMap,
		fogOfWar:          fog.New(actualWorldWidth, actualWorldHeight, terrain.TileSize),
		resourceBar:       resourceBar,
		commandPanel:      commandPanel,
		minimap:           minimap,
		mainMenu:          mainMenu,
		tankSprite:        tankSprite,
		tankFactorySprite: tankFactorySprite,
		scoutSprite:       scoutSprite,
		state:             StateMenu,
	}
	g.engine.Collision.SetTerrain(terrainMap)
	startX, startY := g.findPassablePosition(300, 200)
	g.units = append(g.units, entity.NewConstructor(g.nextUnitID, startX, startY, entity.FactionPlayer))
	g.nextUnitID++
	for i := 0; i < 3; i++ {
		x, y := g.findPassablePosition(startX+float64(i+1)*40, startY+50)
		g.units = append(g.units, entity.NewTank(g.nextUnitID, x, y, entity.FactionPlayer))
		g.nextUnitID++
	}
	x, y := g.findPassablePosition(startX+160, startY+50)
	g.units = append(g.units, entity.NewScout(g.nextUnitID, x, y, entity.FactionPlayer))
	g.nextUnitID++
	// Enemy base position
	enemyBaseX, enemyBaseY := 3500.0, 2700.0
	g.enemyAI = ai.NewEnemyAI(enemyBaseX, enemyBaseY)

	// Create enemy tank factory
	factoryX, factoryY := g.findPassablePosition(enemyBaseX, enemyBaseY)
	enemyFactory := entity.NewBuilding(g.nextBuildingID, factoryX, factoryY, entity.BuildingDefs[entity.BuildingTankFactory])
	enemyFactory.Faction = entity.FactionEnemy
	enemyFactory.Color = entity.GetFactionTintedColor(enemyFactory.Def.Color, entity.FactionEnemy)
	enemyFactory.Completed = true
	enemyFactory.BuildProgress = 1.0
	g.buildings = append(g.buildings, enemyFactory)
	g.nextBuildingID++

	// Create initial enemy units
	for i := 0; i < 3; i++ {
		ex, ey := g.findPassablePosition(enemyBaseX-100+float64(i)*40, enemyBaseY+80)
		g.units = append(g.units, entity.NewTank(g.nextUnitID, ex, ey, entity.FactionEnemy))
		g.nextUnitID++
	}
	ex, ey := g.findPassablePosition(enemyBaseX-50, enemyBaseY+120)
	g.units = append(g.units, entity.NewScout(g.nextUnitID, ex, ey, entity.FactionEnemy))
	g.nextUnitID++

	return g
}
func (g *Game) findPassablePosition(x, y float64) (float64, float64) {
	bounds := emath.NewRect(x, y, unitSize, unitSize)
	if g.terrainMap.IsPassable(bounds) {
		return x, y
	}
	for radius := 1.0; radius < 500; radius += terrain.TileSize {
		for dx := -radius; dx <= radius; dx += terrain.TileSize {
			for dy := -radius; dy <= radius; dy += terrain.TileSize {
				testBounds := emath.NewRect(x+dx, y+dy, unitSize, unitSize)
				if g.terrainMap.IsPassable(testBounds) {
					return x + dx, y + dy
				}
			}
		}
	}
	return x, y
}
func (g *Game) Update() error {
	g.engine.Input.Update()
	inputState := g.engine.Input.State()
	if ebiten.IsFullscreen() {
		g.screenWidth, g.screenHeight = ebiten.Monitor().Size()
	} else {
		g.screenWidth, g.screenHeight = ebiten.WindowSize()
	}
	switch g.state {
	case StateMenu:
		return g.updateMenu(inputState)
	case StatePlaying:
		return g.updatePlaying(inputState)
	}
	return nil
}
func (g *Game) updateMenu(inputState input.State) error {
	g.mainMenu.UpdateSize(float64(g.screenWidth), float64(g.screenHeight))
	if inputState.EscapePressed {
		return ebiten.Termination
	}
	g.mainMenu.UpdateHover(inputState.MousePos)
	if inputState.LeftJustPressed {
		clicked := g.mainMenu.HandleClick(inputState.MousePos)
		if clicked >= 0 {
			return g.handleMenuSelection(clicked)
		}
	}
	selected := g.mainMenu.Update(inputState.MenuUp, inputState.MenuDown, inputState.EnterPressed)
	if selected >= 0 {
		return g.handleMenuSelection(selected)
	}
	return nil
}
func (g *Game) handleMenuSelection(option ui.MenuOption) error {
	switch option {
	case ui.MenuOptionStartGame:
		g.state = StatePlaying
	case ui.MenuOptionExit:
		return ebiten.Termination
	}
	return nil
}
func (g *Game) updatePlaying(inputState input.State) error {
	if inputState.EscapePressed {
		if g.placementMode {
			g.placementMode = false
			g.placementDef = nil
		} else {
			g.state = StateMenu
			return nil
		}
	}
	if inputState.BuildTankPressed {
		if factory := g.getSelectedFactory(); factory != nil {
			factory.QueueProduction(entity.UnitDefs[entity.UnitTypeTank])
		}
	}
	g.engine.UpdateViewportSize(float64(g.screenWidth), float64(g.screenHeight))
	g.resourceBar.UpdateWidth(float64(g.screenWidth))
	g.commandPanel.UpdateHeight(float64(g.screenHeight))
	minimapY := float64(g.screenHeight) - minimapHeight - minimapMargin
	g.minimap.SetPosition(minimapMargin, minimapY)
	g.engine.Resources.Update(tickRate)
	cam := g.engine.Camera
	topOffset := g.resourceBar.Height()
	leftOffset := 0.0
	if g.commandPanel.IsVisible() {
		leftOffset = g.commandPanel.Width()
	}
	cam.HandleEdgeScroll(inputState.MousePos.X, inputState.MousePos.Y, topOffset, leftOffset)
	cam.HandleKeyScroll(inputState.ScrollUp, inputState.ScrollDown, inputState.ScrollLeft, inputState.ScrollRight)
	// Handle mouse wheel zoom
	if inputState.MouseWheelY > 0 {
		cam.ZoomIn(inputState.MousePos)
	} else if inputState.MouseWheelY < 0 {
		cam.ZoomOut(inputState.MousePos)
	}
	if g.minimap.Contains(inputState.MousePos) {
		if inputState.LeftJustPressed || inputState.LeftPressed {
			worldPos := g.minimap.ScreenToWorld(inputState.MousePos)
			cam.MoveTo(worldPos)
		}
	} else if g.placementMode {
		worldPos := cam.ScreenToWorld(inputState.MousePos)
		g.placementValid = g.canPlaceBuilding(worldPos, g.placementDef)
		if inputState.RightJustPressed {
			g.placementMode = false
			g.placementDef = nil
		}
		if inputState.LeftJustPressed && g.placementValid {
			g.placeBuilding(worldPos, g.placementDef, inputState.ShiftHeld)
			// Keep placement mode active if shift is held for queue building
			if !inputState.ShiftHeld {
				g.placementMode = false
				g.placementDef = nil
			}
		}
	} else {
		factory := g.getSelectedFactory()
		if factory != nil {
			g.commandPanel.SetFactoryOptions(factory)
			g.commandPanel.UpdateQueueCounts()
		} else {
			g.commandPanel.SetBuildOptions(g.units)
		}
		if g.commandPanel.Contains(inputState.MousePos) {
			if inputState.LeftJustPressed {
				if clickedUnit := g.commandPanel.UpdateUnit(inputState.MousePos, true); clickedUnit != nil {
					if factory != nil {
						factory.QueueProduction(clickedUnit)
					}
				} else if clickedDef := g.commandPanel.Update(inputState.MousePos, true); clickedDef != nil {
					g.placementMode = true
					g.placementDef = clickedDef
				}
			} else if inputState.RightJustPressed {
				if clickedUnit := g.commandPanel.UpdateUnitRightClick(inputState.MousePos, true); clickedUnit != nil {
					if factory != nil {
						factory.RemoveFromQueue(clickedUnit.Type, g.engine.Resources)
					}
				}
			} else {
				g.commandPanel.Update(inputState.MousePos, false)
			}
		} else {
			g.handleSelection(inputState)
			if inputState.RightJustPressed {
				worldPos := cam.ScreenToWorld(inputState.MousePos)
				g.handleRightClickCommand(worldPos)
			}
		}
	}
	g.engine.Resources.ResetDrains()
	g.updateUnits()
	g.updateBuildings()
	g.updateAI()
	g.updateFogOfWar()
	return nil
}
func (g *Game) handleSelection(inputState input.State) {
	if inputState.MousePos.Y < g.resourceBar.Height() {
		return
	}
	cam := g.engine.Camera
	if inputState.LeftJustReleased {
		if g.engine.Input.State().IsDragging {
			screenBox := g.engine.Input.GetSelectionBox()
			worldBox := emath.Rect{
				Pos:  cam.ScreenToWorld(screenBox.Pos),
				Size: screenBox.Size,
			}
			g.selectUnitsInBox(worldBox, inputState.ShiftHeld)
		} else {
			worldPos := cam.ScreenToWorld(inputState.MousePos)
			g.selectUnitAt(worldPos, inputState.ShiftHeld)
		}
		g.engine.Input.ResetDrag()
	}
}
func (g *Game) selectUnitAt(worldPos emath.Vec2, additive bool) {
	if !additive {
		for _, u := range g.units {
			u.Selected = false
		}
		for _, b := range g.buildings {
			b.Selected = false
		}
	}
	for i := len(g.units) - 1; i >= 0; i-- {
		if g.units[i].Contains(worldPos) && g.units[i].Faction == entity.FactionPlayer {
			g.units[i].Selected = true
			return
		}
	}
	for i := len(g.buildings) - 1; i >= 0; i-- {
		if g.buildings[i].Contains(worldPos) && g.buildings[i].Faction == entity.FactionPlayer {
			g.buildings[i].Selected = true
			return
		}
	}
}
func (g *Game) selectUnitsInBox(worldBox emath.Rect, additive bool) {
	if !additive {
		for _, u := range g.units {
			u.Selected = false
		}
		for _, b := range g.buildings {
			b.Selected = false
		}
	}
	for _, u := range g.units {
		if worldBox.Contains(u.Center()) && u.Faction == entity.FactionPlayer {
			u.Selected = true
		}
	}
	for _, b := range g.buildings {
		if worldBox.Contains(b.Center()) && b.Faction == entity.FactionPlayer {
			b.Selected = true
		}
	}
}
func (g *Game) commandMoveSelected(worldTarget emath.Vec2) {
	selectedCount := 0
	for _, u := range g.units {
		if u.Selected {
			selectedCount++
		}
	}
	if selectedCount == 1 {
		for _, u := range g.units {
			if u.Selected {
				u.SetTarget(worldTarget)
				return
			}
		}
	}
	idx := 0
	for _, u := range g.units {
		if u.Selected {
			row := idx / 3
			col := idx % 3
			offsetX := float64(col-1) * (unitSize + 5)
			offsetY := float64(row) * (unitSize + 5)
			formationTarget := emath.Vec2{
				X: worldTarget.X + offsetX,
				Y: worldTarget.Y + offsetY,
			}
			u.SetTarget(formationTarget)
			idx++
		}
	}
}
func (g *Game) handleRightClickCommand(worldPos emath.Vec2) {
	// Check if clicking on a friendly damaged unit for repair
	var targetUnit *entity.Unit
	for _, u := range g.units {
		if u.Active && u.Faction == entity.FactionPlayer && u.Contains(worldPos) {
			if u.Health < u.MaxHealth {
				targetUnit = u
				break
			}
		}
	}

	// If we clicked on a damaged friendly unit, check if any selected unit can repair
	if targetUnit != nil {
		repairAssigned := false
		for _, u := range g.units {
			if u.Selected && u.CanRepair() && u != targetUnit {
				u.SetRepairTarget(targetUnit)
				u.SetTarget(targetUnit.Center())
				u.ClearBuildTask()
				repairAssigned = true
			}
		}
		if repairAssigned {
			return
		}
	}

	// Clear repair targets for selected units and move them
	for _, u := range g.units {
		if u.Selected {
			u.ClearRepairTarget()
		}
	}
	g.commandMoveSelected(worldPos)
}
func (g *Game) updateUnits() {
	for _, u := range g.units {
		if !u.Active {
			continue
		}
		if u.HasBuildTask {
			g.updateConstructorBuildTask(u)
		}
		if u.RepairTarget != nil {
			g.updateRepairTask(u)
		}
		if !u.HasTarget {
			continue
		}
		desiredPos := u.Update()
		obstacles := make([]emath.Rect, 0, len(g.units)+len(g.buildings)-1)
		for _, other := range g.units {
			if other.ID != u.ID && other.Active {
				obstacles = append(obstacles, other.Bounds())
			}
		}
		for _, b := range g.buildings {
			obstacles = append(obstacles, b.Bounds())
		}
		resolvedPos := g.engine.Collision.ResolveMovement(u.Bounds(), desiredPos, obstacles)

		// If stuck (resolved position is same as current), try avoidance steering
		if resolvedPos.DistanceSquared(u.Position) < 0.1 && u.HasTarget {
			resolvedPos = g.engine.Collision.CalculateAvoidanceDirection(u.Bounds(), u.Target, u.Speed, obstacles)
		}

		u.ApplyPosition(resolvedPos)
	}
	g.updateCombat()
	g.cleanupDeadUnits()
	g.cleanupDeadBuildings()
}
func (g *Game) updateCombat() {
	// Unit combat
	for _, u := range g.units {
		if !u.Active || !u.CanAttack() {
			continue
		}
		// Check if current targets are still valid (clear if dead or out of pursuit range)
		if u.AttackTarget != nil {
			if !u.AttackTarget.Active || !u.IsInPursuitRange(u.AttackTarget) {
				u.AttackTarget = nil
			}
		}
		if u.BuildingAttackTarget != nil {
			if !u.BuildingAttackTarget.Active || !u.IsBuildingInPursuitRange(u.BuildingAttackTarget) {
				u.BuildingAttackTarget = nil
			}
		}

		// Find new target if needed (only look within fire range for new targets)
		if !u.HasAnyAttackTarget() {
			var nearestEnemy *entity.Unit
			var nearestBuilding *entity.Building
			nearestUnitDist := u.Range + 1
			nearestBuildingDist := u.Range + 1

			// Find nearest enemy unit
			for _, other := range g.units {
				if other.Active && other.Faction != u.Faction {
					dist := u.Center().Distance(other.Center())
					if dist <= u.Range && dist < nearestUnitDist {
						nearestUnitDist = dist
						nearestEnemy = other
					}
				}
			}

			// Find nearest enemy building
			for _, b := range g.buildings {
				if b.Active && b.Faction != u.Faction {
					dist := u.Center().Distance(b.Center())
					if dist <= u.Range && dist < nearestBuildingDist {
						nearestBuildingDist = dist
						nearestBuilding = b
					}
				}
			}

			// Prioritize units over buildings
			if nearestEnemy != nil {
				u.SetAttackTarget(nearestEnemy)
			} else if nearestBuilding != nil {
				u.SetBuildingAttackTarget(nearestBuilding)
			}
		}

		// Pursue enemy if they're out of fire range but in pursuit range
		if u.AttackTarget != nil && u.AttackTarget.Active && !u.IsInRange(u.AttackTarget) {
			// Only pursue if unit doesn't have another movement target
			if !u.HasTarget {
				u.SetTarget(u.AttackTarget.Center())
			}
		} else if u.BuildingAttackTarget != nil && u.BuildingAttackTarget.Active && !u.IsBuildingInRange(u.BuildingAttackTarget) {
			if !u.HasTarget {
				u.SetTarget(u.BuildingAttackTarget.Center())
			}
		}

		if u.UpdateCombat(tickRate) {
			if u.AttackTarget != nil && u.AttackTarget.Active {
				projectile := entity.NewProjectile(g.nextProjectileID, u, u.AttackTarget)
				g.projectiles = append(g.projectiles, projectile)
				g.nextProjectileID++
			} else if u.BuildingAttackTarget != nil && u.BuildingAttackTarget.Active {
				projectile := entity.NewProjectileAtBuilding(g.nextProjectileID, u, u.BuildingAttackTarget)
				g.projectiles = append(g.projectiles, projectile)
				g.nextProjectileID++
			}
		}
	}

	// Building combat (laser towers, etc.)
	g.updateBuildingCombat()

	g.updateProjectiles()
}

func (g *Game) updateBuildingCombat() {
	for _, b := range g.buildings {
		if !b.Active || !b.CanAttack() {
			continue
		}

		// Update cooldown
		if b.FireCooldown > 0 {
			b.FireCooldown -= tickRate
		}

		// Check if current target is still valid
		if b.AttackTarget != nil && (!b.AttackTarget.Active || !b.IsInAttackRange(b.AttackTarget)) {
			b.AttackTarget = nil
		}

		// Find new target if needed
		if b.AttackTarget == nil {
			var nearestEnemy *entity.Unit
			nearestDist := b.Def.AttackRange + 1

			for _, u := range g.units {
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

		// Fire if ready and have target
		if b.FireCooldown <= 0 && b.AttackTarget != nil && b.AttackTarget.Active {
			// Check if we have enough energy
			energyRes := g.engine.Resources.Get(resource.Energy)
			if energyRes.Current >= b.Def.EnergyPerShot {
				energyRes.Spend(b.Def.EnergyPerShot)
				projectile := entity.NewProjectileFromBuilding(g.nextProjectileID, b, b.AttackTarget)
				g.projectiles = append(g.projectiles, projectile)
				g.nextProjectileID++
				b.FireCooldown = 1.0 / b.Def.FireRate
			}
		}
	}
}
func (g *Game) updateProjectiles() {
	aliveProjectiles := make([]*entity.Projectile, 0, len(g.projectiles))
	for _, p := range g.projectiles {
		if !p.Update(tickRate) {
			aliveProjectiles = append(aliveProjectiles, p)
		}
	}
	g.projectiles = aliveProjectiles
}
func (g *Game) cleanupDeadUnits() {
	aliveUnits := make([]*entity.Unit, 0, len(g.units))
	for _, u := range g.units {
		if u.Active {
			aliveUnits = append(aliveUnits, u)
		} else {
			wreckage := entity.NewWreckageFromUnit(g.nextWreckageID, u)
			g.wreckages = append(g.wreckages, wreckage)
			g.nextWreckageID++
		}
	}
	g.units = aliveUnits
}

func (g *Game) cleanupDeadBuildings() {
	aliveBuildings := make([]*entity.Building, 0, len(g.buildings))
	for _, b := range g.buildings {
		if b.Active {
			aliveBuildings = append(aliveBuildings, b)
		} else {
			wreckage := entity.NewWreckageFromBuilding(g.nextWreckageID, b)
			g.wreckages = append(g.wreckages, wreckage)
			g.nextWreckageID++
		}
	}
	g.buildings = aliveBuildings
}
func (g *Game) updateConstructorBuildTask(u *entity.Unit) {
	if u.HasTarget {
		return
	}
	if !u.IsNearBuildSite() {
		// Target a position just below (south of) the building site, so constructor doesn't end up inside
		buildSiteTarget := emath.Vec2{
			X: u.BuildPos.X + u.BuildDef.Size/2,
			Y: u.BuildPos.Y + u.BuildDef.Size + u.Size.Y/2 + 5, // Below the building
		}
		u.SetTarget(buildSiteTarget)
		return
	}
	if !u.IsBuilding {
		building := entity.NewBuildingUnderConstruction(
			g.nextBuildingID,
			u.BuildPos.X,
			u.BuildPos.Y,
			u.BuildDef,
		)
		g.buildings = append(g.buildings, building)
		g.nextBuildingID++
		u.BuildTarget = building
		u.IsBuilding = true
	}
	if u.BuildTarget != nil {
		if u.BuildTarget.UpdateConstruction(tickRate, g.engine.Resources) {
			g.applyBuildingEffects(u.BuildTarget.Def)
			u.ClearBuildTask()
		}
	}
}

// Repair cost per health point
const repairMetalCostPerHP = 0.5
const repairEnergyCostPerHP = 0.25

func (g *Game) updateRepairTask(u *entity.Unit) {
	target := u.RepairTarget

	// Check if target is still valid
	if target == nil || !target.Active || target.Health >= target.MaxHealth {
		u.ClearRepairTarget()
		return
	}

	// Move towards target if not in range
	if !u.IsInRepairRange(target) {
		if !u.HasTarget {
			u.SetTarget(target.Center())
		}
		return
	}

	// In range - perform repair
	u.ClearTarget() // Stop moving

	// Calculate repair amount for this tick
	repairAmount := u.RepairRate * tickRate

	// Don't over-repair
	healthNeeded := target.MaxHealth - target.Health
	if repairAmount > healthNeeded {
		repairAmount = healthNeeded
	}

	// Calculate resource cost
	metalCost := repairAmount * repairMetalCostPerHP
	energyCost := repairAmount * repairEnergyCostPerHP

	// Check if we have enough resources
	resources := g.engine.Resources
	metalRes := resources.Get(resource.Metal)
	energyRes := resources.Get(resource.Energy)
	if metalRes.Current < metalCost || energyRes.Current < energyCost {
		return // Not enough resources - wait
	}

	// Consume resources and repair
	metalRes.Spend(metalCost)
	energyRes.Spend(energyCost)
	target.Health += repairAmount

	// Check if fully repaired
	if target.Health >= target.MaxHealth {
		target.Health = target.MaxHealth
		u.ClearRepairTarget()
	}
}
func (g *Game) applyBuildingEffects(def *entity.BuildingDef) {
	resources := g.engine.Resources
	if def.MetalProduction > 0 {
		resources.AddProduction(resource.Metal, def.MetalProduction)
	}
	if def.EnergyProduction > 0 {
		resources.AddProduction(resource.Energy, def.EnergyProduction)
	}
	if def.MetalConsumption > 0 {
		resources.AddConsumption(resource.Metal, def.MetalConsumption)
	}
	if def.EnergyConsumption > 0 {
		resources.AddConsumption(resource.Energy, def.EnergyConsumption)
	}
	if def.MetalStorage > 0 {
		resources.AddCapacity(resource.Metal, def.MetalStorage)
	}
	if def.EnergyStorage > 0 {
		resources.AddCapacity(resource.Energy, def.EnergyStorage)
	}
}
func (g *Game) updateBuildings() {
	for _, b := range g.buildings {
		if completedUnit := b.UpdateProduction(tickRate, g.engine.Resources); completedUnit != nil {
			spawnPos := b.GetSpawnPoint()
			unit := entity.NewUnitFromDef(g.nextUnitID, spawnPos.X, spawnPos.Y, completedUnit, b.Faction)
			g.units = append(g.units, unit)
			g.nextUnitID++
			if b.HasRallyPoint {
				unit.SetTarget(b.RallyPoint)
			}
		}
	}
}
func (g *Game) updateAI() {
	if g.enemyAI == nil {
		return
	}
	g.enemyAI.Update(tickRate, g.units, g.buildings)
}
func (g *Game) updateFogOfWar() {
	g.fogOfWar.ClearVisibility()
	for _, u := range g.units {
		if u.Active && u.Faction == entity.FactionPlayer {
			center := u.Center()
			g.fogOfWar.RevealCircle(center.X, center.Y, u.VisionRange)
		}
	}
	for _, b := range g.buildings {
		if b.Completed && b.Faction == entity.FactionPlayer {
			center := b.Center()
			g.fogOfWar.RevealCircle(center.X, center.Y, b.Def.VisionRange)
		}
	}
}
func (g *Game) getSelectedFactory() *entity.Building {
	for _, b := range g.buildings {
		if b.Selected && b.CanProduce() {
			return b
		}
	}
	return nil
}
func snapToGrid(pos emath.Vec2) emath.Vec2 {
	return emath.Vec2{
		X: math.Floor(pos.X/buildingGridSize) * buildingGridSize,
		Y: math.Floor(pos.Y/buildingGridSize) * buildingGridSize,
	}
}
func (g *Game) canPlaceBuilding(worldPos emath.Vec2, def *entity.BuildingDef) bool {
	buildingPos := snapToGrid(worldPos)
	bounds := emath.NewRect(buildingPos.X, buildingPos.Y, def.Size, def.Size)
	if !g.terrainMap.IsBuildable(bounds) {
		return false
	}
	if def.Type == entity.BuildingMetalExtractor {
		if !g.hasMetal(bounds) {
			return false
		}
	}
	for _, u := range g.units {
		if bounds.Intersects(u.Bounds()) {
			return false
		}
	}
	for _, b := range g.buildings {
		if bounds.Intersects(b.Bounds()) {
			return false
		}
	}
	return true
}
func (g *Game) hasMetal(bounds emath.Rect) bool {
	startX, startY := g.terrainMap.GetTileCoords(bounds.Pos.X, bounds.Pos.Y)
	endX, endY := g.terrainMap.GetTileCoords(bounds.Pos.X+bounds.Size.X, bounds.Pos.Y+bounds.Size.Y)
	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			if x >= 0 && x < g.terrainMap.Width && y >= 0 && y < g.terrainMap.Height {
				if g.terrainMap.Tiles[y][x].HasMetal {
					return true
				}
			}
		}
	}
	return false
}
func (g *Game) placeBuilding(worldPos emath.Vec2, def *entity.BuildingDef, queue bool) {
	var constructor *entity.Unit
	for _, u := range g.units {
		if u.Selected && u.CanBuild() {
			constructor = u
			break
		}
	}
	if constructor == nil {
		return
	}
	buildingPos := snapToGrid(worldPos)
	if queue {
		constructor.QueueBuildTask(def, buildingPos)
	} else {
		constructor.SetBuildTask(def, buildingPos)
	}
}
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateMenu:
		g.mainMenu.Draw(screen)
		return
	case StatePlaying:
		g.drawPlaying(screen)
	}
}
func (g *Game) drawPlaying(screen *ebiten.Image) {
	r := g.engine.Renderer
	cam := g.engine.Camera
	r.Clear(screen)
	terrainStart := time.Now()
	g.drawTerrain(screen)
	g.debugTerrainTime = time.Since(terrainStart)
	for _, w := range g.wreckages {
		if cam.IsVisible(w.Bounds()) {
			g.drawWreckage(screen, w)
		}
	}
	for _, b := range g.buildings {
		if cam.IsVisible(b.Bounds()) {
			if b.Faction == entity.FactionPlayer || g.fogOfWar.IsVisible(b.Bounds()) {
				g.drawBuilding(screen, b)
			}
		}
	}
	for _, u := range g.units {
		if cam.IsVisible(u.Bounds()) {
			if u.Faction == entity.FactionPlayer || g.fogOfWar.IsVisible(u.Bounds()) {
				g.drawUnit(screen, u)
			}
		}
	}
	for _, p := range g.projectiles {
		if cam.IsVisible(p.Bounds()) && g.fogOfWar.IsVisible(p.Bounds()) {
			g.drawProjectile(screen, p)
		}
	}
	if g.placementMode && g.placementDef != nil {
		g.drawPlacementPreview(screen)
	}
	if g.engine.Input.State().IsDragging {
		box := g.engine.Input.GetSelectionBox()
		r.DrawRectOutline(screen, box, 1, color.RGBA{0, 255, 0, 255})
	}
	g.resourceBar.Draw(screen, g.engine.Resources)
	g.commandPanel.Draw(screen, g.engine.Resources)
	minimapEntities := make([]ui.MinimapEntity, 0, len(g.units)+len(g.buildings))
	for _, u := range g.units {
		if u.Faction == entity.FactionPlayer || g.fogOfWar.IsVisible(u.Bounds()) {
			minimapEntities = append(minimapEntities, ui.MinimapEntity{
				Position: u.Position,
				Size:     u.Size,
				Color:    u.Color,
			})
		}
	}
	for _, b := range g.buildings {
		if b.Faction == entity.FactionPlayer || g.fogOfWar.IsVisible(b.Bounds()) {
			minimapEntities = append(minimapEntities, ui.MinimapEntity{
				Position: b.Position,
				Size:     b.Size,
				Color:    b.Color,
			})
		}
	}
	minimapStart := time.Now()
	g.minimap.Draw(screen, cam, g.terrainMap, g.fogOfWar, minimapEntities)
	g.debugMinimapTime = time.Since(minimapStart)
	instructionX := int(g.commandPanel.Width()) + 10
	instructions := "WASD/Arrows: Scroll | Left Click: Select | Right Click: Move | ESC: Menu"
	if g.placementMode {
		instructions = "Left Click: Place | Shift+Click: Queue Multiple | Right Click/ESC: Cancel"
	} else if factory := g.getSelectedFactory(); factory != nil {
		factoryName := factory.Def.Name
		if factory.Producing {
			instructions = fmt.Sprintf("%s selected - Building unit...", factoryName)
		} else {
			instructions = fmt.Sprintf("%s selected - Click unit to build", factoryName)
		}
	}
	r.DrawTextAt(screen, instructions, instructionX, int(g.resourceBar.Height())+5)
	fpsText := fmt.Sprintf("FPS: %.1f  TPS: %.1f  Units: %d  Buildings: %d  Projectiles: %d",
		ebiten.ActualFPS(), ebiten.ActualTPS(), len(g.units), len(g.buildings), len(g.projectiles))
	ebitenutil.DebugPrintAt(screen, fpsText, 10, int(baseHeight)-20)
}
func (g *Game) getTileImage(c color.RGBA) *ebiten.Image {
	if g.tileImages == nil {
		g.tileImages = make(map[color.RGBA]*ebiten.Image)
	}
	if img, ok := g.tileImages[c]; ok {
		return img
	}
	img := ebiten.NewImage(terrain.TileSize, terrain.TileSize)
	img.Fill(c)
	g.tileImages[c] = img
	return img
}

func (g *Game) drawTerrain(screen *ebiten.Image) {
	// Cache base terrain (without fog) - only build once
	if g.terrainCache == nil {
		worldW := g.terrainMap.Width * terrain.TileSize
		worldH := g.terrainMap.Height * terrain.TileSize
		g.terrainCache = ebiten.NewImage(worldW, worldH)

		for y := 0; y < g.terrainMap.Height; y++ {
			for x := 0; x < g.terrainMap.Width; x++ {
				tile := g.terrainMap.Tiles[y][x]
				screenX := float64(x) * terrain.TileSize
				screenY := float64(y) * terrain.TileSize

				tileColor := terrain.TileColorVariation(tile.Type, x, y).(color.RGBA)
				tileImg := g.getTileImage(tileColor)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(screenX, screenY)
				g.terrainCache.DrawImage(tileImg, op)

				// Draw metal deposits on base terrain
				if tile.HasMetal {
					centerX := screenX + terrain.TileSize/2
					centerY := screenY + terrain.TileSize/2
					vector.DrawFilledCircle(g.terrainCache, float32(centerX), float32(centerY), 8, color.RGBA{180, 180, 200, 255}, false)
					vector.DrawFilledCircle(g.terrainCache, float32(centerX), float32(centerY), 5, color.RGBA{120, 120, 140, 255}, false)
				}
			}
		}
	}

	cam := g.engine.Camera
	zoom := cam.GetZoom()

	// Draw entire terrain in ONE draw call (visible portion via SubImage)
	// When zoomed out, we need to see more world; when zoomed in, see less
	visibleWorldW := cam.ViewportSize.X / zoom
	visibleWorldH := cam.ViewportSize.Y / zoom

	viewX := int(cam.Position.X)
	viewY := int(cam.Position.Y)
	viewW := int(visibleWorldW) + terrain.TileSize
	viewH := int(visibleWorldH) + terrain.TileSize

	// Clamp to terrain bounds
	if viewX < 0 {
		viewX = 0
	}
	if viewY < 0 {
		viewY = 0
	}
	maxX := g.terrainMap.Width * terrain.TileSize
	maxY := g.terrainMap.Height * terrain.TileSize
	if viewX+viewW > maxX {
		viewW = maxX - viewX
	}
	if viewY+viewH > maxY {
		viewH = maxY - viewY
	}

	srcRect := image.Rect(viewX, viewY, viewX+viewW, viewY+viewH)
	op := &ebiten.DrawImageOptions{}
	// First translate to align with camera, then scale for zoom
	op.GeoM.Translate(-cam.Position.X+float64(viewX), -cam.Position.Y+float64(viewY))
	op.GeoM.Scale(zoom, zoom)
	screen.DrawImage(g.terrainCache.SubImage(srcRect).(*ebiten.Image), op)

	// Draw fog overlay on top (only visible tiles)
	startX, startY, endX, endY := g.terrainMap.GetVisibleTiles(
		cam.Position.X, cam.Position.Y,
		visibleWorldW, visibleWorldH,
	)

	blackImg := g.getTileImage(color.RGBA{0, 0, 0, 255})
	darkImg := g.getTileImage(color.RGBA{0, 0, 0, 153}) // 60% darkness for explored

	fogTileCount := 0
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			fogState := g.fogOfWar.GetTileStateAt(x, y)
			if fogState == fog.Visible {
				continue // No overlay needed
			}

			fogTileCount++
			worldX := float64(x) * terrain.TileSize
			worldY := float64(y) * terrain.TileSize
			screenPos := cam.WorldToScreen(emath.Vec2{X: worldX, Y: worldY})

			fogOp := &ebiten.DrawImageOptions{}
			fogOp.GeoM.Scale(zoom, zoom)
			fogOp.GeoM.Translate(screenPos.X, screenPos.Y)

			if fogState == fog.Unexplored {
				screen.DrawImage(blackImg, fogOp)
			} else {
				screen.DrawImage(darkImg, fogOp)
			}
		}
	}
	g.debugFogTiles = fogTileCount
}
func (g *Game) drawUnit(screen *ebiten.Image, u *entity.Unit) {
	r := g.engine.Renderer
	cam := g.engine.Camera
	zoom := cam.GetZoom()
	screenPos := cam.WorldToScreen(u.Position)
	scaledSize := u.Size.Mul(zoom)
	screenBounds := emath.Rect{Pos: screenPos, Size: scaledSize}
	if u.Selected {
		selectionRect := emath.Rect{
			Pos:  screenPos.Sub(emath.Vec2{X: selectionMargin * zoom, Y: selectionMargin * zoom}),
			Size: scaledSize.Add(emath.Vec2{X: selectionMargin * 2 * zoom, Y: selectionMargin * 2 * zoom}),
		}
		r.DrawRect(screen, selectionRect, color.RGBA{0, 255, 0, 128})
	}
	screenCenter := cam.WorldToScreen(u.Center())
	if u.Type == entity.UnitTypeTank {
		// Draw tank as square body with rotating turret
		g.drawTank(screen, u, screenCenter)
	} else if u.Type == entity.UnitTypeScout && g.scoutSprite != nil {
		spriteW := float64(g.scoutSprite.Bounds().Dx())
		spriteH := float64(g.scoutSprite.Bounds().Dy())
		targetSize := 48.0 * zoom
		scaleX := targetSize / spriteW
		scaleY := targetSize / spriteH
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-spriteW/2, -spriteH/2)
		op.GeoM.Rotate(u.Angle + math.Pi/2)
		op.GeoM.Scale(scaleX, scaleY)
		op.GeoM.Translate(screenCenter.X, screenCenter.Y)
		if u.Faction == entity.FactionEnemy {
			op.ColorScale.Scale(1.2, 0.6, 0.6, 1)
		}
		screen.DrawImage(g.scoutSprite, op)
	} else {
		r.DrawRect(screen, screenBounds, u.Color)
		arrowLength := scaledSize.X * 0.5
		arrowWidth := scaledSize.X * 0.25
		tipX := screenCenter.X + math.Cos(u.Angle)*arrowLength
		tipY := screenCenter.Y + math.Sin(u.Angle)*arrowLength
		perpAngle := u.Angle + math.Pi/2
		baseX1 := screenCenter.X + math.Cos(perpAngle)*arrowWidth
		baseY1 := screenCenter.Y + math.Sin(perpAngle)*arrowWidth
		baseX2 := screenCenter.X - math.Cos(perpAngle)*arrowWidth
		baseY2 := screenCenter.Y - math.Sin(perpAngle)*arrowWidth
		arrowColor := color.RGBA{0, 0, 0, 200}
		r.DrawLine(screen, emath.Vec2{X: baseX1, Y: baseY1}, emath.Vec2{X: tipX, Y: tipY}, 2, arrowColor)
		r.DrawLine(screen, emath.Vec2{X: baseX2, Y: baseY2}, emath.Vec2{X: tipX, Y: tipY}, 2, arrowColor)
		r.DrawLine(screen, emath.Vec2{X: baseX1, Y: baseY1}, emath.Vec2{X: baseX2, Y: baseY2}, 2, arrowColor)
		if u.Type == entity.UnitTypeConstructor {
			plusOffset := arrowLength * 0.3
			plusX := screenCenter.X - math.Cos(u.Angle)*plusOffset
			plusY := screenCenter.Y - math.Sin(u.Angle)*plusOffset
			plusSize := 3.0 * zoom
			r.DrawLine(screen,
				emath.Vec2{X: plusX - plusSize, Y: plusY},
				emath.Vec2{X: plusX + plusSize, Y: plusY},
				2, color.RGBA{0, 0, 0, 200})
			r.DrawLine(screen,
				emath.Vec2{X: plusX, Y: plusY - plusSize},
				emath.Vec2{X: plusX, Y: plusY + plusSize},
				2, color.RGBA{0, 0, 0, 200})
		}
	}
	if u.HasTarget && u.Selected {
		screenTarget := cam.WorldToScreen(u.Target)
		r.DrawCircle(screen, screenTarget, float32(4*zoom), color.RGBA{0, 255, 0, 200})
	}
	if u.HasBuildTask && u.Selected {
		buildCenter := emath.Vec2{
			X: u.BuildPos.X + u.BuildDef.Size/2,
			Y: u.BuildPos.Y + u.BuildDef.Size/2,
		}
		screenBuildCenter := cam.WorldToScreen(buildCenter)
		r.DrawLine(screen, screenCenter, screenBuildCenter, 1, color.RGBA{255, 200, 50, 150})
		if !u.IsBuilding {
			buildScreenPos := cam.WorldToScreen(u.BuildPos)
			buildRect := emath.Rect{
				Pos:  buildScreenPos,
				Size: emath.Vec2{X: u.BuildDef.Size * zoom, Y: u.BuildDef.Size * zoom},
			}
			r.DrawRectOutline(screen, buildRect, 2, color.RGBA{255, 200, 50, 200})
		}
	}
	if u.Health < u.MaxHealth || u.Selected {
		barWidth := scaledSize.X
		barHeight := 4.0 * zoom
		barX := screenPos.X
		barY := screenPos.Y - barHeight - 2*zoom
		bgRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth, Y: barHeight},
		}
		r.DrawRect(screen, bgRect, color.RGBA{40, 40, 40, 200})
		healthRatio := u.HealthRatio()
		healthRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth * healthRatio, Y: barHeight},
		}
		var healthColor color.RGBA
		if healthRatio > 0.6 {
			healthColor = color.RGBA{0, 200, 0, 255}
		} else if healthRatio > 0.3 {
			healthColor = color.RGBA{200, 200, 0, 255}
		} else {
			healthColor = color.RGBA{200, 0, 0, 255}
		}
		r.DrawRect(screen, healthRect, healthColor)
	}
	if u.AttackTarget != nil && u.AttackTarget.Active && u.Selected {
		targetCenter := cam.WorldToScreen(u.AttackTarget.Center())
		r.DrawLine(screen, screenCenter, targetCenter, 1, color.RGBA{255, 0, 0, 150})
	}
}

func (g *Game) drawTank(screen *ebiten.Image, u *entity.Unit, screenCenter emath.Vec2) {
	zoom := g.engine.Camera.GetZoom()

	// Tank body size (scaled by zoom)
	bodySize := u.Size.X * 1.2 * zoom
	halfBody := bodySize / 2

	// Turret size (smaller than body)
	turretSize := bodySize * 0.5
	halfTurret := turretSize / 2

	// Barrel length and width
	barrelLength := bodySize * 0.7
	barrelWidth := turretSize * 0.25

	// Colors based on faction
	var bodyColor, turretColor, barrelColor color.RGBA
	if u.Faction == entity.FactionEnemy {
		bodyColor = color.RGBA{140, 60, 60, 255}   // Dark red
		turretColor = color.RGBA{180, 80, 80, 255} // Lighter red
		barrelColor = color.RGBA{100, 40, 40, 255} // Darker red
	} else {
		bodyColor = color.RGBA{60, 100, 60, 255}   // Dark green
		turretColor = color.RGBA{80, 140, 80, 255} // Lighter green
		barrelColor = color.RGBA{40, 70, 40, 255}  // Darker green
	}

	// Draw tank body (rotated square)
	g.drawRotatedRect(screen, screenCenter, bodySize, bodySize, u.Angle, bodyColor)

	// Draw outline for body
	g.drawRotatedRectOutline(screen, screenCenter, bodySize, bodySize, u.Angle, color.RGBA{30, 30, 30, 255})

	// Draw barrel first (so turret overlaps it at the base)
	barrelStartX := screenCenter.X + math.Cos(u.TurretAngle)*halfTurret*0.3
	barrelStartY := screenCenter.Y + math.Sin(u.TurretAngle)*halfTurret*0.3
	barrelEndX := screenCenter.X + math.Cos(u.TurretAngle)*barrelLength
	barrelEndY := screenCenter.Y + math.Sin(u.TurretAngle)*barrelLength

	// Draw barrel as thick line
	vector.StrokeLine(screen,
		float32(barrelStartX), float32(barrelStartY),
		float32(barrelEndX), float32(barrelEndY),
		float32(barrelWidth), barrelColor, false)

	// Draw turret (rotated square, centered on tank)
	g.drawRotatedRect(screen, screenCenter, turretSize, turretSize, u.TurretAngle, turretColor)

	// Draw outline for turret
	g.drawRotatedRectOutline(screen, screenCenter, turretSize, turretSize, u.TurretAngle, color.RGBA{30, 30, 30, 255})

	// Draw small direction indicator on body front
	frontX := screenCenter.X + math.Cos(u.Angle)*halfBody*0.6
	frontY := screenCenter.Y + math.Sin(u.Angle)*halfBody*0.6
	vector.DrawFilledCircle(screen, float32(frontX), float32(frontY), float32(2*zoom), color.RGBA{200, 200, 200, 200}, false)
}

func (g *Game) drawRotatedRect(screen *ebiten.Image, center emath.Vec2, width, height, angle float64, c color.RGBA) {
	halfW := width / 2
	halfH := height / 2

	// Calculate corner offsets
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	// Four corners relative to center, then rotated
	corners := [4]emath.Vec2{
		{X: -halfW, Y: -halfH},
		{X: halfW, Y: -halfH},
		{X: halfW, Y: halfH},
		{X: -halfW, Y: halfH},
	}

	// Rotate and translate corners
	var rotated [4]emath.Vec2
	for i, corner := range corners {
		rotated[i] = emath.Vec2{
			X: center.X + corner.X*cos - corner.Y*sin,
			Y: center.Y + corner.X*sin + corner.Y*cos,
		}
	}

	// Draw as two triangles
	vs := []ebiten.Vertex{
		{DstX: float32(rotated[0].X), DstY: float32(rotated[0].Y), ColorR: float32(c.R) / 255, ColorG: float32(c.G) / 255, ColorB: float32(c.B) / 255, ColorA: float32(c.A) / 255},
		{DstX: float32(rotated[1].X), DstY: float32(rotated[1].Y), ColorR: float32(c.R) / 255, ColorG: float32(c.G) / 255, ColorB: float32(c.B) / 255, ColorA: float32(c.A) / 255},
		{DstX: float32(rotated[2].X), DstY: float32(rotated[2].Y), ColorR: float32(c.R) / 255, ColorG: float32(c.G) / 255, ColorB: float32(c.B) / 255, ColorA: float32(c.A) / 255},
		{DstX: float32(rotated[3].X), DstY: float32(rotated[3].Y), ColorR: float32(c.R) / 255, ColorG: float32(c.G) / 255, ColorB: float32(c.B) / 255, ColorA: float32(c.A) / 255},
	}
	indices := []uint16{0, 1, 2, 0, 2, 3}

	screen.DrawTriangles(vs, indices, g.getWhitePixel(), &ebiten.DrawTrianglesOptions{})
}

func (g *Game) drawRotatedRectOutline(screen *ebiten.Image, center emath.Vec2, width, height, angle float64, c color.RGBA) {
	halfW := width / 2
	halfH := height / 2

	cos := math.Cos(angle)
	sin := math.Sin(angle)

	corners := [4]emath.Vec2{
		{X: -halfW, Y: -halfH},
		{X: halfW, Y: -halfH},
		{X: halfW, Y: halfH},
		{X: -halfW, Y: halfH},
	}

	var rotated [4]emath.Vec2
	for i, corner := range corners {
		rotated[i] = emath.Vec2{
			X: center.X + corner.X*cos - corner.Y*sin,
			Y: center.Y + corner.X*sin + corner.Y*cos,
		}
	}

	// Draw lines between corners
	for i := range 4 {
		next := (i + 1) % 4
		vector.StrokeLine(screen,
			float32(rotated[i].X), float32(rotated[i].Y),
			float32(rotated[next].X), float32(rotated[next].Y),
			1, c, false)
	}
}

func (g *Game) getWhitePixel() *ebiten.Image {
	if g.whitePixel == nil {
		g.whitePixel = ebiten.NewImage(1, 1)
		g.whitePixel.Fill(color.White)
	}
	return g.whitePixel
}

func (g *Game) drawProjectile(screen *ebiten.Image, p *entity.Projectile) {
	r := g.engine.Renderer
	cam := g.engine.Camera
	zoom := cam.GetZoom()
	screenPos := cam.WorldToScreen(p.Position)
	scaledSize := p.Size.Mul(zoom)
	screenBounds := emath.Rect{Pos: screenPos, Size: scaledSize}
	r.DrawRect(screen, screenBounds, p.Color)
	trailEnd := p.Position.Sub(p.Direction.Mul(8))
	screenTrailEnd := cam.WorldToScreen(trailEnd)
	r.DrawLine(screen, cam.WorldToScreen(p.Center()), screenTrailEnd, 2, color.RGBA{255, 150, 0, 150})
}
func (g *Game) drawWreckage(screen *ebiten.Image, w *entity.Wreckage) {
	r := g.engine.Renderer
	cam := g.engine.Camera
	zoom := cam.GetZoom()
	screenPos := cam.WorldToScreen(w.Position)
	scaledSize := w.Size.Mul(zoom)
	screenBounds := emath.Rect{Pos: screenPos, Size: scaledSize}
	r.DrawRect(screen, screenBounds, w.Color)
	r.DrawLine(screen,
		emath.Vec2{X: screenPos.X, Y: screenPos.Y},
		emath.Vec2{X: screenPos.X + scaledSize.X, Y: screenPos.Y + scaledSize.Y},
		1, color.RGBA{50, 50, 50, 255})
	r.DrawLine(screen,
		emath.Vec2{X: screenPos.X + scaledSize.X, Y: screenPos.Y},
		emath.Vec2{X: screenPos.X, Y: screenPos.Y + scaledSize.Y},
		1, color.RGBA{50, 50, 50, 255})
}
func (g *Game) drawBuilding(screen *ebiten.Image, b *entity.Building) {
	r := g.engine.Renderer
	cam := g.engine.Camera
	zoom := cam.GetZoom()
	screenPos := cam.WorldToScreen(b.Position)
	scaledSize := b.Size.Mul(zoom)
	screenBounds := emath.Rect{Pos: screenPos, Size: scaledSize}
	if b.Selected {
		selectionRect := emath.Rect{
			Pos:  screenPos.Sub(emath.Vec2{X: selectionMargin * zoom, Y: selectionMargin * zoom}),
			Size: scaledSize.Add(emath.Vec2{X: selectionMargin * 2 * zoom, Y: selectionMargin * 2 * zoom}),
		}
		r.DrawRect(screen, selectionRect, color.RGBA{0, 255, 0, 128})
	}
	screenCenter := emath.Vec2{X: screenPos.X + scaledSize.X/2, Y: screenPos.Y + scaledSize.Y/2}
	if b.Type == entity.BuildingTankFactory && g.tankFactorySprite != nil && b.Completed {
		spriteW := float64(g.tankFactorySprite.Bounds().Dx())
		spriteH := float64(g.tankFactorySprite.Bounds().Dy())
		targetSize := 128.0 * zoom
		scaleX := targetSize / spriteW
		scaleY := targetSize / spriteH
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-spriteW/2, -spriteH/2)
		op.GeoM.Scale(scaleX, scaleY)
		op.GeoM.Translate(screenCenter.X, screenCenter.Y)
		screen.DrawImage(g.tankFactorySprite, op)
	} else if b.Completed {
		r.DrawRect(screen, screenBounds, b.Color)
	} else {
		rgba := b.Color.(color.RGBA)
		constructionColor := color.RGBA{
			R: uint8(float64(rgba.R) * 0.5),
			G: uint8(float64(rgba.G) * 0.5),
			B: uint8(float64(rgba.B) * 0.5),
			A: 180,
		}
		r.DrawRect(screen, screenBounds, constructionColor)
		scaffoldColor := color.RGBA{100, 80, 50, 150}
		lineSpacing := 10.0 * zoom
		for i := 0.0; i < scaledSize.X; i += lineSpacing {
			r.DrawLine(screen,
				emath.Vec2{X: screenPos.X + i, Y: screenPos.Y},
				emath.Vec2{X: screenPos.X, Y: screenPos.Y + i},
				1, scaffoldColor)
		}
	}
	borderColor := color.RGBA{60, 60, 60, 255}
	if !b.Completed {
		borderColor = color.RGBA{200, 150, 50, 255}
	}
	if b.Type != entity.BuildingTankFactory || !b.Completed || g.tankFactorySprite == nil {
		r.DrawRectOutline(screen, screenBounds, 2, borderColor)
	}
	if !b.Completed {
		barWidth := scaledSize.X - 10*zoom
		barHeight := 8.0 * zoom
		barX := screenPos.X + 5*zoom
		barY := screenPos.Y + scaledSize.Y/2 - barHeight/2
		bgRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth, Y: barHeight},
		}
		r.DrawRect(screen, bgRect, color.RGBA{40, 40, 40, 220})
		progressRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth * b.BuildProgress, Y: barHeight},
		}
		r.DrawRect(screen, progressRect, color.RGBA{255, 200, 50, 255})
	}
	if b.Type == entity.BuildingTankFactory && b.Completed && b.Producing {
		barWidth := scaledSize.X - 10*zoom
		barHeight := 6.0 * zoom
		barX := screenPos.X + 5*zoom
		barY := screenPos.Y + scaledSize.Y - 12*zoom
		bgRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth, Y: barHeight},
		}
		r.DrawRect(screen, bgRect, color.RGBA{40, 40, 40, 200})
		progressRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth * b.ProductionProgress, Y: barHeight},
		}
		r.DrawRect(screen, progressRect, color.RGBA{0, 200, 0, 255})
	}
	if b.Selected && b.Type == entity.BuildingTankFactory && b.HasRallyPoint && b.Completed {
		rallyScreen := cam.WorldToScreen(b.RallyPoint)
		r.DrawCircle(screen, rallyScreen, float32(5*zoom), color.RGBA{255, 255, 0, 200})
		buildingCenter := cam.WorldToScreen(b.Center())
		r.DrawLine(screen, buildingCenter, rallyScreen, 1, color.RGBA{255, 255, 0, 100})
	}
}
func (g *Game) drawPlacementPreview(screen *ebiten.Image) {
	r := g.engine.Renderer
	cam := g.engine.Camera
	zoom := cam.GetZoom()
	mousePos := g.engine.Input.State().MousePos
	worldPos := cam.ScreenToWorld(mousePos)
	buildingPos := snapToGrid(worldPos)
	screenPos := cam.WorldToScreen(buildingPos)
	scaledSize := g.placementDef.Size * zoom
	screenBounds := emath.Rect{
		Pos:  screenPos,
		Size: emath.Vec2{X: scaledSize, Y: scaledSize},
	}
	var previewColor color.Color
	var borderColor color.Color
	if g.placementValid {
		previewColor = color.RGBA{0, 200, 0, 100}
		borderColor = color.RGBA{0, 255, 0, 200}
	} else {
		previewColor = color.RGBA{200, 0, 0, 100}
		borderColor = color.RGBA{255, 0, 0, 200}
	}
	r.DrawRect(screen, screenBounds, previewColor)
	r.DrawRectOutline(screen, screenBounds, 2, borderColor)
}
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
func main() {
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Tanks RTS")
	ebiten.SetFullscreen(true)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(NewGame()); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
