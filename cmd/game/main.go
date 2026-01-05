package main

import (
	"image/color"
	"log"
	"math"

	"github.com/bklimczak/tanks/engine"
	"github.com/bklimczak/tanks/engine/entity"
	"github.com/bklimczak/tanks/engine/input"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/resource"
	"github.com/bklimczak/tanks/engine/terrain"
	"github.com/bklimczak/tanks/engine/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// GameState represents the current state of the game
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
)

const (
	unitSize        = 20.0
	selectionMargin = 2.0
	tickRate        = 1.0 / 60.0 // 60 FPS

	// World size is 5x the initial viewport
	worldMultiplier = 5.0
	baseWidth       = 1280.0
	baseHeight      = 720.0
	worldWidth      = baseWidth * worldMultiplier
	worldHeight     = baseHeight * worldMultiplier

	// Minimap dimensions
	minimapWidth  = 160.0
	minimapHeight = 120.0
	minimapMargin = 10.0

	// Building placement grid (matches terrain tile size)
	buildingGridSize = terrain.TileSize
)

type Game struct {
	engine       *engine.Engine
	units        []*entity.Unit
	buildings    []*entity.Building
	wreckages    []*entity.Wreckage
	projectiles  []*entity.Projectile
	terrainMap   *terrain.Map
	resourceBar  *ui.ResourceBar
	commandPanel *ui.CommandPanel
	minimap      *ui.Minimap
	mainMenu     *ui.MainMenu
	screenWidth  int
	screenHeight int
	nextUnitID   uint64
	nextBuildingID uint64
	nextWreckageID uint64
	nextProjectileID uint64

	// Game state
	state GameState

	// Building placement mode
	placementMode  bool
	placementDef   *entity.BuildingDef
	placementValid bool
}

func NewGame() *Game {
	resourceBar := ui.NewResourceBar(baseWidth)
	commandPanel := ui.NewCommandPanel(resourceBar.Height(), baseHeight)

	// Position minimap at bottom of command panel
	minimapY := baseHeight - minimapHeight - minimapMargin
	minimap := ui.NewMinimap(minimapMargin, minimapY, minimapWidth, minimapHeight)
	minimap.SetWorldSize(worldWidth, worldHeight)

	// Create main menu
	mainMenu := ui.NewMainMenu()

	// Load terrain from file, or generate and save if not found
	terrainMap, err := terrain.LoadMapFromFile("maps/default.yaml")
	if err != nil {
		// Generate a new map with a fixed seed for consistency
		terrainMap = terrain.NewMap(worldWidth, worldHeight)
		terrainMap.Generate(42) // Fixed seed for reproducible maps

		// Place a metal deposit near the starting area for the player
		terrainMap.PlaceMetalDeposit(400, 150)

		// Save the generated map for future editing
		if saveErr := terrain.SaveMapToFile(terrainMap, "maps/default.yaml", "Default Map", "Auto-generated default map", "System"); saveErr != nil {
			log.Printf("Warning: could not save map file: %v", saveErr)
		}
	}

	g := &Game{
		engine:       engine.New(worldWidth, worldHeight, baseWidth, baseHeight),
		terrainMap:   terrainMap,
		resourceBar:  resourceBar,
		commandPanel: commandPanel,
		minimap:      minimap,
		mainMenu:     mainMenu,
		state:        StateMenu, // Start in menu state
	}

	// Set terrain for collision system
	g.engine.Collision.SetTerrain(terrainMap)

	// Find a good starting position on grass for player
	startX, startY := g.findPassablePosition(300, 200)

	// Create player units
	g.units = append(g.units, entity.NewConstructor(g.nextUnitID, startX, startY, entity.FactionPlayer))
	g.nextUnitID++

	// Create some player tanks nearby
	for i := 0; i < 3; i++ {
		x, y := g.findPassablePosition(startX+float64(i+1)*40, startY+50)
		g.units = append(g.units, entity.NewTank(g.nextUnitID, x, y, entity.FactionPlayer))
		g.nextUnitID++
	}

	// Create a player scout
	x, y := g.findPassablePosition(startX+160, startY+50)
	g.units = append(g.units, entity.NewScout(g.nextUnitID, x, y, entity.FactionPlayer))
	g.nextUnitID++

	// Create enemy units in the bottom-right corner (static for now)
	enemyPositions := []emath.Vec2{
		{X: 3500, Y: 2650},
		{X: 3550, Y: 2700},
		{X: 3600, Y: 2650},
		{X: 3500, Y: 2750},
		{X: 3550, Y: 2800},
		{X: 3650, Y: 2700},
	}
	for _, pos := range enemyPositions {
		ex, ey := g.findPassablePosition(pos.X, pos.Y)
		g.units = append(g.units, entity.NewTank(g.nextUnitID, ex, ey, entity.FactionEnemy))
		g.nextUnitID++
	}

	// Add enemy scouts patrolling the land corridor (bottom of the map)
	enemyScoutPositions := []emath.Vec2{
		{X: 3000, Y: 2700},
		{X: 2500, Y: 2750},
		{X: 2000, Y: 2800},
	}
	for _, pos := range enemyScoutPositions {
		ex, ey := g.findPassablePosition(pos.X, pos.Y)
		g.units = append(g.units, entity.NewScout(g.nextUnitID, ex, ey, entity.FactionEnemy))
		g.nextUnitID++
	}

	return g
}

// findPassablePosition finds a nearby passable position
func (g *Game) findPassablePosition(x, y float64) (float64, float64) {
	bounds := emath.NewRect(x, y, unitSize, unitSize)

	// Check if current position is passable
	if g.terrainMap.IsPassable(bounds) {
		return x, y
	}

	// Search in expanding squares
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

	return x, y // Fallback
}

func (g *Game) Update() error {
	g.engine.Input.Update()
	inputState := g.engine.Input.State()

	// Update screen size - use monitor size when fullscreen
	if ebiten.IsFullscreen() {
		g.screenWidth, g.screenHeight = ebiten.Monitor().Size()
	} else {
		g.screenWidth, g.screenHeight = ebiten.WindowSize()
	}

	// Handle different game states
	switch g.state {
	case StateMenu:
		return g.updateMenu(inputState)
	case StatePlaying:
		return g.updatePlaying(inputState)
	}

	return nil
}

// updateMenu handles the main menu state
func (g *Game) updateMenu(inputState input.State) error {
	// Update menu size
	g.mainMenu.UpdateSize(float64(g.screenWidth), float64(g.screenHeight))

	// Handle ESC to quit from menu
	if inputState.EscapePressed {
		return ebiten.Termination
	}

	// Update hover based on mouse position
	g.mainMenu.UpdateHover(inputState.MousePos)

	// Handle mouse click
	if inputState.LeftJustPressed {
		clicked := g.mainMenu.HandleClick(inputState.MousePos)
		if clicked >= 0 {
			return g.handleMenuSelection(clicked)
		}
	}

	// Handle keyboard navigation
	selected := g.mainMenu.Update(inputState.MenuUp, inputState.MenuDown, inputState.EnterPressed)
	if selected >= 0 {
		return g.handleMenuSelection(selected)
	}

	return nil
}

// handleMenuSelection processes a menu selection
func (g *Game) handleMenuSelection(option ui.MenuOption) error {
	switch option {
	case ui.MenuOptionStartGame:
		g.state = StatePlaying
	case ui.MenuOptionExit:
		return ebiten.Termination
	}
	return nil
}

// updatePlaying handles the main gameplay state
func (g *Game) updatePlaying(inputState input.State) error {
	// ESC to cancel placement mode or go back to menu
	if inputState.EscapePressed {
		if g.placementMode {
			g.placementMode = false
			g.placementDef = nil
		} else {
			g.state = StateMenu
			return nil
		}
	}

	// T key to build tank from selected factory (legacy shortcut)
	if inputState.BuildTankPressed {
		if factory := g.getSelectedFactory(); factory != nil {
			factory.QueueProduction(entity.UnitDefs[entity.UnitTypeTank])
		}
	}

	g.engine.UpdateViewportSize(float64(g.screenWidth), float64(g.screenHeight))
	g.resourceBar.UpdateWidth(float64(g.screenWidth))
	g.commandPanel.UpdateHeight(float64(g.screenHeight))

	// Update minimap position (bottom-left of screen)
	minimapY := float64(g.screenHeight) - minimapHeight - minimapMargin
	g.minimap.SetPosition(minimapMargin, minimapY)

	// Update resources
	g.engine.Resources.Update(tickRate)

	// Handle camera scrolling
	cam := g.engine.Camera
	topOffset := g.resourceBar.Height()
	leftOffset := 0.0
	if g.commandPanel.IsVisible() {
		leftOffset = g.commandPanel.Width()
	}

	// Edge scrolling
	cam.HandleEdgeScroll(inputState.MousePos.X, inputState.MousePos.Y, topOffset, leftOffset)

	// Keyboard scrolling
	cam.HandleKeyScroll(inputState.ScrollUp, inputState.ScrollDown, inputState.ScrollLeft, inputState.ScrollRight)

	// Handle minimap clicks
	if g.minimap.Contains(inputState.MousePos) {
		if inputState.LeftJustPressed || inputState.LeftPressed {
			// Click on minimap moves camera
			worldPos := g.minimap.ScreenToWorld(inputState.MousePos)
			cam.MoveTo(worldPos)
		}
		// Don't process other clicks if on minimap
	} else if g.placementMode {
		// Building placement mode
		worldPos := cam.ScreenToWorld(inputState.MousePos)
		g.placementValid = g.canPlaceBuilding(worldPos, g.placementDef)

		// Right-click cancels placement
		if inputState.RightJustPressed {
			g.placementMode = false
			g.placementDef = nil
		}

		// Left-click places building if valid
		if inputState.LeftJustPressed && g.placementValid {
			g.placeBuilding(worldPos, g.placementDef)
			g.placementMode = false
			g.placementDef = nil
		}
	} else {
		// Update command panel based on selection
		factory := g.getSelectedFactory()
		if factory != nil {
			// Factory selected - show unit production options
			g.commandPanel.SetFactoryOptions(factory)
			g.commandPanel.UpdateQueueCounts()
		} else {
			// Check for constructor selection
			g.commandPanel.SetBuildOptions(g.units)
		}

		// Handle command panel clicks
		if g.commandPanel.Contains(inputState.MousePos) {
			if inputState.LeftJustPressed {
				// Check for unit production click
				if clickedUnit := g.commandPanel.UpdateUnit(inputState.MousePos, true); clickedUnit != nil {
					// Queue unit production
					if factory != nil {
						factory.QueueProduction(clickedUnit)
					}
				} else if clickedDef := g.commandPanel.Update(inputState.MousePos, true); clickedDef != nil {
					// Enter building placement mode
					g.placementMode = true
					g.placementDef = clickedDef
				}
			} else {
				g.commandPanel.Update(inputState.MousePos, false)
			}
		} else {
			// Handle selection (ignore if clicking on UI)
			g.handleSelection(inputState)

			// Handle movement command
			if inputState.RightJustPressed {
				// Convert screen position to world position
				worldPos := cam.ScreenToWorld(inputState.MousePos)
				g.commandMoveSelected(worldPos)
			}
		}
	}

	// Update units with collision detection
	g.updateUnits()

	// Update buildings (production)
	g.updateBuildings()

	return nil
}

func (g *Game) handleSelection(inputState input.State) {
	// Ignore clicks on the resource bar
	if inputState.MousePos.Y < g.resourceBar.Height() {
		return
	}

	cam := g.engine.Camera

	if inputState.LeftJustReleased {
		if g.engine.Input.State().IsDragging {
			// Box selection - convert screen box to world box
			screenBox := g.engine.Input.GetSelectionBox()
			worldBox := emath.Rect{
				Pos:  cam.ScreenToWorld(screenBox.Pos),
				Size: screenBox.Size, // Size stays the same
			}
			g.selectUnitsInBox(worldBox, inputState.ShiftHeld)
		} else {
			// Single click selection - convert to world position
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

	// Check units first (top-most first) - only select player units
	for i := len(g.units) - 1; i >= 0; i-- {
		if g.units[i].Contains(worldPos) && g.units[i].Faction == entity.FactionPlayer {
			g.units[i].Selected = true
			return
		}
	}

	// Check buildings - only player buildings
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

	// Select units in box - only player units
	for _, u := range g.units {
		if worldBox.Contains(u.Center()) && u.Faction == entity.FactionPlayer {
			u.Selected = true
		}
	}

	// Select buildings in box - only player buildings
	for _, b := range g.buildings {
		if worldBox.Contains(b.Center()) && b.Faction == entity.FactionPlayer {
			b.Selected = true
		}
	}
}

func (g *Game) commandMoveSelected(worldTarget emath.Vec2) {
	// Count selected units for formation
	selectedCount := 0
	for _, u := range g.units {
		if u.Selected {
			selectedCount++
		}
	}

	// Single unit - no formation offset
	if selectedCount == 1 {
		for _, u := range g.units {
			if u.Selected {
				u.SetTarget(worldTarget)
				return
			}
		}
	}

	// Multiple units - assign formation positions centered on click point
	idx := 0
	for _, u := range g.units {
		if u.Selected {
			// Grid formation around click point
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

func (g *Game) updateUnits() {
	// First pass: movement and build tasks
	for _, u := range g.units {
		if !u.Active {
			continue
		}

		// Handle constructor build tasks
		if u.HasBuildTask {
			g.updateConstructorBuildTask(u)
		}

		if !u.HasTarget {
			continue
		}

		// Get desired position from unit
		desiredPos := u.Update()

		// Build obstacle list (other active units and buildings)
		obstacles := make([]emath.Rect, 0, len(g.units)+len(g.buildings)-1)
		for _, other := range g.units {
			if other.ID != u.ID && other.Active {
				obstacles = append(obstacles, other.Bounds())
			}
		}
		for _, b := range g.buildings {
			obstacles = append(obstacles, b.Bounds())
		}

		// Resolve collisions and apply position
		resolvedPos := g.engine.Collision.ResolveMovement(u.Bounds(), desiredPos, obstacles)
		u.ApplyPosition(resolvedPos)
	}

	// Second pass: combat
	g.updateCombat()

	// Third pass: cleanup dead units
	g.cleanupDeadUnits()
}

// updateCombat handles unit combat logic
func (g *Game) updateCombat() {
	for _, u := range g.units {
		if !u.Active || !u.CanAttack() {
			continue
		}

		// Auto-acquire targets if no current target or target is dead
		if u.AttackTarget == nil || !u.AttackTarget.Active {
			u.ClearAttackTarget()
			// Find nearest enemy in range
			var nearestEnemy *entity.Unit
			nearestDist := u.Range + 1 // Start beyond range

			for _, other := range g.units {
				if other.Active && other.Faction != u.Faction {
					dist := u.Center().Distance(other.Center())
					if dist <= u.Range && dist < nearestDist {
						nearestDist = dist
						nearestEnemy = other
					}
				}
			}

			if nearestEnemy != nil {
				u.SetAttackTarget(nearestEnemy)
			}
		}

		// Process combat
		if u.UpdateCombat(tickRate) {
			// Unit fired, spawn projectile
			if u.AttackTarget != nil && u.AttackTarget.Active {
				projectile := entity.NewProjectile(g.nextProjectileID, u, u.AttackTarget)
				g.projectiles = append(g.projectiles, projectile)
				g.nextProjectileID++
			}
		}
	}

	// Update projectiles
	g.updateProjectiles()
}

// updateProjectiles moves projectiles and handles collisions
func (g *Game) updateProjectiles() {
	aliveProjectiles := make([]*entity.Projectile, 0, len(g.projectiles))

	for _, p := range g.projectiles {
		if !p.Update(tickRate) {
			aliveProjectiles = append(aliveProjectiles, p)
		}
	}

	g.projectiles = aliveProjectiles
}

// cleanupDeadUnits removes dead units and creates wreckages
func (g *Game) cleanupDeadUnits() {
	aliveUnits := make([]*entity.Unit, 0, len(g.units))

	for _, u := range g.units {
		if u.Active {
			aliveUnits = append(aliveUnits, u)
		} else {
			// Create wreckage
			wreckage := entity.NewWreckageFromUnit(g.nextWreckageID, u)
			g.wreckages = append(g.wreckages, wreckage)
			g.nextWreckageID++
		}
	}

	g.units = aliveUnits
}

// updateConstructorBuildTask handles the build task for a constructor unit
func (g *Game) updateConstructorBuildTask(u *entity.Unit) {
	// If unit is moving, wait until it arrives
	if u.HasTarget {
		return
	}

	// Check if unit is near the build site
	if !u.IsNearBuildSite() {
		// Try to move closer
		buildSiteTarget := emath.Vec2{
			X: u.BuildPos.X - u.Size.X - 5,
			Y: u.BuildPos.Y + u.BuildDef.Size/2,
		}
		u.SetTarget(buildSiteTarget)
		return
	}

	// Unit is at the build site, start or continue building
	if !u.IsBuilding {
		// Create the building under construction
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

	// Progress construction
	if u.BuildTarget != nil {
		if u.BuildTarget.UpdateConstruction(tickRate) {
			// Building complete - apply its effects
			g.applyBuildingEffects(u.BuildTarget.Def)
			u.ClearBuildTask()
		}
	}
}

// applyBuildingEffects applies a building's production, consumption, and storage to resources
func (g *Game) applyBuildingEffects(def *entity.BuildingDef) {
	resources := g.engine.Resources

	// Add production
	if def.MetalProduction > 0 {
		resources.AddProduction(resource.Metal, def.MetalProduction)
	}
	if def.EnergyProduction > 0 {
		resources.AddProduction(resource.Energy, def.EnergyProduction)
	}

	// Add consumption
	if def.MetalConsumption > 0 {
		resources.AddConsumption(resource.Metal, def.MetalConsumption)
	}
	if def.EnergyConsumption > 0 {
		resources.AddConsumption(resource.Energy, def.EnergyConsumption)
	}

	// Add storage capacity
	if def.MetalStorage > 0 {
		resources.AddCapacity(resource.Metal, def.MetalStorage)
	}
	if def.EnergyStorage > 0 {
		resources.AddCapacity(resource.Energy, def.EnergyStorage)
	}
}

func (g *Game) updateBuildings() {
	for _, b := range g.buildings {
		if completedUnit := b.UpdateProduction(tickRate); completedUnit != nil {
			// Unit finished, spawn it
			spawnPos := b.GetSpawnPoint()
			unit := entity.NewUnitFromDef(g.nextUnitID, spawnPos.X, spawnPos.Y, completedUnit, b.Faction)
			g.units = append(g.units, unit)
			g.nextUnitID++

			// Send unit to rally point
			if b.HasRallyPoint {
				unit.SetTarget(b.RallyPoint)
			}
		}
	}
}

// getSelectedFactory returns the selected tank factory, or nil if none selected
func (g *Game) getSelectedFactory() *entity.Building {
	for _, b := range g.buildings {
		if b.Selected && b.Type == entity.BuildingTankFactory {
			return b
		}
	}
	return nil
}

// snapToGrid snaps a position to the building placement grid
func snapToGrid(pos emath.Vec2) emath.Vec2 {
	return emath.Vec2{
		X: math.Floor(pos.X/buildingGridSize) * buildingGridSize,
		Y: math.Floor(pos.Y/buildingGridSize) * buildingGridSize,
	}
}

// canPlaceBuilding checks if a building can be placed at the given world position
func (g *Game) canPlaceBuilding(worldPos emath.Vec2, def *entity.BuildingDef) bool {
	// Snap to grid (building top-left corner)
	buildingPos := snapToGrid(worldPos)
	bounds := emath.NewRect(buildingPos.X, buildingPos.Y, def.Size, def.Size)

	// Check terrain - must be buildable
	if !g.terrainMap.IsBuildable(bounds) {
		return false
	}

	// Metal Extractor must be placed on metal deposits
	if def.Type == entity.BuildingMetalExtractor {
		if !g.hasMetal(bounds) {
			return false
		}
	}

	// Check collision with units
	for _, u := range g.units {
		if bounds.Intersects(u.Bounds()) {
			return false
		}
	}

	// Check collision with other buildings
	for _, b := range g.buildings {
		if bounds.Intersects(b.Bounds()) {
			return false
		}
	}

	return true
}

// hasMetal checks if the given area contains a metal deposit
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

// placeBuilding assigns a build task to the selected constructor
func (g *Game) placeBuilding(worldPos emath.Vec2, def *entity.BuildingDef) {
	// Find selected constructor
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

	// Snap to grid
	buildingPos := snapToGrid(worldPos)

	// Assign build task to constructor
	constructor.SetBuildTask(def, buildingPos)
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Handle different game states
	switch g.state {
	case StateMenu:
		g.mainMenu.Draw(screen)
		return
	case StatePlaying:
		g.drawPlaying(screen)
	}
}

// drawPlaying renders the main gameplay
func (g *Game) drawPlaying(screen *ebiten.Image) {
	r := g.engine.Renderer
	cam := g.engine.Camera

	// Clear screen
	r.Clear(screen)

	// Draw terrain
	g.drawTerrain(screen)

	// Draw wreckages (below buildings and units)
	for _, w := range g.wreckages {
		if cam.IsVisible(w.Bounds()) {
			g.drawWreckage(screen, w)
		}
	}

	// Draw buildings
	for _, b := range g.buildings {
		if cam.IsVisible(b.Bounds()) {
			g.drawBuilding(screen, b)
		}
	}

	// Draw units (in world coordinates, offset by camera)
	for _, u := range g.units {
		// Only draw if visible
		if cam.IsVisible(u.Bounds()) {
			g.drawUnit(screen, u)
		}
	}

	// Draw projectiles (above units)
	for _, p := range g.projectiles {
		if cam.IsVisible(p.Bounds()) {
			g.drawProjectile(screen, p)
		}
	}

	// Draw building placement preview
	if g.placementMode && g.placementDef != nil {
		g.drawPlacementPreview(screen)
	}

	// Draw selection box if dragging (in screen coordinates)
	if g.engine.Input.State().IsDragging {
		box := g.engine.Input.GetSelectionBox()
		r.DrawRectOutline(screen, box, 1, color.RGBA{0, 255, 0, 255})
	}

	// Draw UI elements (on top of everything, in screen coordinates)
	g.resourceBar.Draw(screen, g.engine.Resources)
	g.commandPanel.Draw(screen, g.engine.Resources)

	// Draw minimap
	minimapEntities := make([]ui.MinimapEntity, 0, len(g.units)+len(g.buildings))
	for _, u := range g.units {
		minimapEntities = append(minimapEntities, ui.MinimapEntity{
			Position: u.Position,
			Size:     u.Size,
			Color:    u.Color,
		})
	}
	for _, b := range g.buildings {
		minimapEntities = append(minimapEntities, ui.MinimapEntity{
			Position: b.Position,
			Size:     b.Size,
			Color:    b.Color,
		})
	}
	g.minimap.Draw(screen, cam, g.terrainMap, minimapEntities)

	// Draw instructions below resource bar
	instructionX := int(g.commandPanel.Width()) + 10
	instructions := "WASD/Arrows: Scroll | Left Click: Select | Right Click: Move | ESC: Menu"
	if g.placementMode {
		instructions = "Left Click: Place Building | Right Click/ESC: Cancel"
	} else if factory := g.getSelectedFactory(); factory != nil {
		if factory.Producing {
			instructions = "Tank Factory selected - Building tank..."
		} else {
			instructions = "Tank Factory selected - Press T to build Tank"
		}
	}
	r.DrawTextAt(screen, instructions, instructionX, int(g.resourceBar.Height())+5)
}

func (g *Game) drawTerrain(screen *ebiten.Image) {
	r := g.engine.Renderer
	cam := g.engine.Camera

	// Get visible tile range
	startX, startY, endX, endY := g.terrainMap.GetVisibleTiles(
		cam.Position.X, cam.Position.Y,
		cam.ViewportSize.X, cam.ViewportSize.Y,
	)

	// Draw each visible tile
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			tile := g.terrainMap.Tiles[y][x]
			tileColor := terrain.TileColorVariation(tile.Type, x, y)

			// Calculate screen position
			worldX := float64(x) * terrain.TileSize
			worldY := float64(y) * terrain.TileSize
			screenPos := cam.WorldToScreen(emath.Vec2{X: worldX, Y: worldY})

			tileRect := emath.Rect{
				Pos:  screenPos,
				Size: emath.Vec2{X: terrain.TileSize + 1, Y: terrain.TileSize + 1}, // +1 to avoid gaps
			}

			r.DrawRect(screen, tileRect, tileColor)

			// Draw metal deposit indicator
			if tile.HasMetal {
				centerX := screenPos.X + terrain.TileSize/2
				centerY := screenPos.Y + terrain.TileSize/2
				r.DrawCircle(screen, emath.Vec2{X: centerX, Y: centerY}, 8, color.RGBA{180, 180, 200, 255})
				r.DrawCircle(screen, emath.Vec2{X: centerX, Y: centerY}, 5, color.RGBA{120, 120, 140, 255})
			}
		}
	}
}

func (g *Game) drawUnit(screen *ebiten.Image, u *entity.Unit) {
	r := g.engine.Renderer
	cam := g.engine.Camera

	// Convert world position to screen position
	screenPos := cam.WorldToScreen(u.Position)
	screenBounds := emath.Rect{Pos: screenPos, Size: u.Size}

	// Selection highlight
	if u.Selected {
		selectionRect := emath.Rect{
			Pos:  screenPos.Sub(emath.Vec2{X: selectionMargin, Y: selectionMargin}),
			Size: u.Size.Add(emath.Vec2{X: selectionMargin * 2, Y: selectionMargin * 2}),
		}
		r.DrawRect(screen, selectionRect, color.RGBA{0, 255, 0, 128})
	}

	// Unit body
	r.DrawRect(screen, screenBounds, u.Color)

	// Draw front arrow indicator showing unit direction
	screenCenter := cam.WorldToScreen(u.Center())
	arrowLength := u.Size.X * 0.5
	arrowWidth := u.Size.X * 0.25

	// Calculate arrow tip position (front of unit)
	tipX := screenCenter.X + math.Cos(u.Angle)*arrowLength
	tipY := screenCenter.Y + math.Sin(u.Angle)*arrowLength

	// Calculate arrow base corners (perpendicular to direction)
	perpAngle := u.Angle + math.Pi/2
	baseX1 := screenCenter.X + math.Cos(perpAngle)*arrowWidth
	baseY1 := screenCenter.Y + math.Sin(perpAngle)*arrowWidth
	baseX2 := screenCenter.X - math.Cos(perpAngle)*arrowWidth
	baseY2 := screenCenter.Y - math.Sin(perpAngle)*arrowWidth

	// Draw arrow as three lines forming a triangle
	arrowColor := color.RGBA{0, 0, 0, 200}
	r.DrawLine(screen, emath.Vec2{X: baseX1, Y: baseY1}, emath.Vec2{X: tipX, Y: tipY}, 2, arrowColor)
	r.DrawLine(screen, emath.Vec2{X: baseX2, Y: baseY2}, emath.Vec2{X: tipX, Y: tipY}, 2, arrowColor)
	r.DrawLine(screen, emath.Vec2{X: baseX1, Y: baseY1}, emath.Vec2{X: baseX2, Y: baseY2}, 2, arrowColor)

	// Draw unit type indicator for constructors (plus sign)
	if u.Type == entity.UnitTypeConstructor {
		// Draw smaller plus sign offset from center
		plusOffset := arrowLength * 0.3
		plusX := screenCenter.X - math.Cos(u.Angle)*plusOffset
		plusY := screenCenter.Y - math.Sin(u.Angle)*plusOffset
		r.DrawLine(screen,
			emath.Vec2{X: plusX - 3, Y: plusY},
			emath.Vec2{X: plusX + 3, Y: plusY},
			2, color.RGBA{0, 0, 0, 200})
		r.DrawLine(screen,
			emath.Vec2{X: plusX, Y: plusY - 3},
			emath.Vec2{X: plusX, Y: plusY + 3},
			2, color.RGBA{0, 0, 0, 200})
	}

	// Move target indicator
	if u.HasTarget && u.Selected {
		screenTarget := cam.WorldToScreen(u.Target)
		r.DrawCircle(screen, screenTarget, 4, color.RGBA{0, 255, 0, 200})
	}

	// Build task indicator for constructors
	if u.HasBuildTask && u.Selected {
		// Draw line from unit to build site
		buildCenter := emath.Vec2{
			X: u.BuildPos.X + u.BuildDef.Size/2,
			Y: u.BuildPos.Y + u.BuildDef.Size/2,
		}
		screenBuildCenter := cam.WorldToScreen(buildCenter)
		r.DrawLine(screen, screenCenter, screenBuildCenter, 1, color.RGBA{255, 200, 50, 150})

		// Draw build site outline if not yet building
		if !u.IsBuilding {
			buildScreenPos := cam.WorldToScreen(u.BuildPos)
			buildRect := emath.Rect{
				Pos:  buildScreenPos,
				Size: emath.Vec2{X: u.BuildDef.Size, Y: u.BuildDef.Size},
			}
			r.DrawRectOutline(screen, buildRect, 2, color.RGBA{255, 200, 50, 200})
		}
	}

	// Health bar (always show if damaged, or when selected)
	if u.Health < u.MaxHealth || u.Selected {
		barWidth := u.Size.X
		barHeight := 4.0
		barX := screenPos.X
		barY := screenPos.Y - barHeight - 2

		// Background
		bgRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth, Y: barHeight},
		}
		r.DrawRect(screen, bgRect, color.RGBA{40, 40, 40, 200})

		// Health fill
		healthRatio := u.HealthRatio()
		healthRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth * healthRatio, Y: barHeight},
		}
		// Color based on health: green -> yellow -> red
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

	// Attack target indicator (line to target when attacking)
	if u.AttackTarget != nil && u.AttackTarget.Active && u.Selected {
		targetCenter := cam.WorldToScreen(u.AttackTarget.Center())
		r.DrawLine(screen, screenCenter, targetCenter, 1, color.RGBA{255, 0, 0, 150})
	}
}

// drawProjectile draws a projectile
func (g *Game) drawProjectile(screen *ebiten.Image, p *entity.Projectile) {
	r := g.engine.Renderer
	cam := g.engine.Camera

	// Convert world position to screen position
	screenPos := cam.WorldToScreen(p.Position)
	screenBounds := emath.Rect{Pos: screenPos, Size: p.Size}

	// Draw projectile as a bright circle/rectangle
	r.DrawRect(screen, screenBounds, p.Color)

	// Add a small trail effect (line behind the projectile)
	trailEnd := p.Position.Sub(p.Direction.Mul(8))
	screenTrailEnd := cam.WorldToScreen(trailEnd)
	r.DrawLine(screen, cam.WorldToScreen(p.Center()), screenTrailEnd, 2, color.RGBA{255, 150, 0, 150})
}

// drawWreckage draws a destroyed unit wreckage
func (g *Game) drawWreckage(screen *ebiten.Image, w *entity.Wreckage) {
	r := g.engine.Renderer
	cam := g.engine.Camera

	// Convert world position to screen position
	screenPos := cam.WorldToScreen(w.Position)
	screenBounds := emath.Rect{Pos: screenPos, Size: w.Size}

	// Draw wreckage as dark rectangle
	r.DrawRect(screen, screenBounds, w.Color)

	// Draw some "debris" lines
	r.DrawLine(screen,
		emath.Vec2{X: screenPos.X, Y: screenPos.Y},
		emath.Vec2{X: screenPos.X + w.Size.X, Y: screenPos.Y + w.Size.Y},
		1, color.RGBA{50, 50, 50, 255})
	r.DrawLine(screen,
		emath.Vec2{X: screenPos.X + w.Size.X, Y: screenPos.Y},
		emath.Vec2{X: screenPos.X, Y: screenPos.Y + w.Size.Y},
		1, color.RGBA{50, 50, 50, 255})
}

func (g *Game) drawBuilding(screen *ebiten.Image, b *entity.Building) {
	r := g.engine.Renderer
	cam := g.engine.Camera

	// Convert world position to screen position
	screenPos := cam.WorldToScreen(b.Position)
	screenBounds := emath.Rect{Pos: screenPos, Size: b.Size}

	// Selection highlight
	if b.Selected {
		selectionRect := emath.Rect{
			Pos:  screenPos.Sub(emath.Vec2{X: selectionMargin, Y: selectionMargin}),
			Size: b.Size.Add(emath.Vec2{X: selectionMargin * 2, Y: selectionMargin * 2}),
		}
		r.DrawRect(screen, selectionRect, color.RGBA{0, 255, 0, 128})
	}

	// Building body - show semi-transparent if under construction
	if b.Completed {
		r.DrawRect(screen, screenBounds, b.Color)
	} else {
		// Draw a darker, semi-transparent version while under construction
		rgba := b.Color.(color.RGBA)
		constructionColor := color.RGBA{
			R: uint8(float64(rgba.R) * 0.5),
			G: uint8(float64(rgba.G) * 0.5),
			B: uint8(float64(rgba.B) * 0.5),
			A: 180,
		}
		r.DrawRect(screen, screenBounds, constructionColor)

		// Draw construction scaffold pattern
		scaffoldColor := color.RGBA{100, 80, 50, 150}
		lineSpacing := 10.0
		for i := 0.0; i < b.Size.X; i += lineSpacing {
			// Diagonal lines
			r.DrawLine(screen,
				emath.Vec2{X: screenPos.X + i, Y: screenPos.Y},
				emath.Vec2{X: screenPos.X, Y: screenPos.Y + i},
				1, scaffoldColor)
		}
	}

	// Border
	borderColor := color.RGBA{60, 60, 60, 255}
	if !b.Completed {
		borderColor = color.RGBA{200, 150, 50, 255} // Yellow border for construction
	}
	r.DrawRectOutline(screen, screenBounds, 2, borderColor)

	// Construction progress bar (building being constructed)
	if !b.Completed {
		barWidth := b.Size.X - 10
		barHeight := 8.0
		barX := screenPos.X + 5
		barY := screenPos.Y + b.Size.Y/2 - barHeight/2

		// Background
		bgRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth, Y: barHeight},
		}
		r.DrawRect(screen, bgRect, color.RGBA{40, 40, 40, 220})

		// Progress
		progressRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth * b.BuildProgress, Y: barHeight},
		}
		r.DrawRect(screen, progressRect, color.RGBA{255, 200, 50, 255}) // Yellow/gold for construction
	}

	// Production progress bar for tank factories (only when complete and producing)
	if b.Type == entity.BuildingTankFactory && b.Completed && b.Producing {
		barWidth := b.Size.X - 10
		barHeight := 6.0
		barX := screenPos.X + 5
		barY := screenPos.Y + b.Size.Y - 12

		// Background
		bgRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth, Y: barHeight},
		}
		r.DrawRect(screen, bgRect, color.RGBA{40, 40, 40, 200})

		// Progress
		progressRect := emath.Rect{
			Pos:  emath.Vec2{X: barX, Y: barY},
			Size: emath.Vec2{X: barWidth * b.ProductionProgress, Y: barHeight},
		}
		r.DrawRect(screen, progressRect, color.RGBA{0, 200, 0, 255})
	}

	// Rally point indicator for selected factories
	if b.Selected && b.Type == entity.BuildingTankFactory && b.HasRallyPoint && b.Completed {
		rallyScreen := cam.WorldToScreen(b.RallyPoint)
		r.DrawCircle(screen, rallyScreen, 5, color.RGBA{255, 255, 0, 200})

		// Line from building to rally point
		buildingCenter := cam.WorldToScreen(b.Center())
		r.DrawLine(screen, buildingCenter, rallyScreen, 1, color.RGBA{255, 255, 0, 100})
	}
}

func (g *Game) drawPlacementPreview(screen *ebiten.Image) {
	r := g.engine.Renderer
	cam := g.engine.Camera
	mousePos := g.engine.Input.State().MousePos

	// Convert screen to world position and snap to grid
	worldPos := cam.ScreenToWorld(mousePos)
	buildingPos := snapToGrid(worldPos)

	// Convert back to screen for drawing
	screenPos := cam.WorldToScreen(buildingPos)
	screenBounds := emath.Rect{
		Pos:  screenPos,
		Size: emath.Vec2{X: g.placementDef.Size, Y: g.placementDef.Size},
	}

	// Choose color based on validity
	var previewColor color.Color
	var borderColor color.Color
	if g.placementValid {
		// Semi-transparent green
		previewColor = color.RGBA{0, 200, 0, 100}
		borderColor = color.RGBA{0, 255, 0, 200}
	} else {
		// Semi-transparent red
		previewColor = color.RGBA{200, 0, 0, 100}
		borderColor = color.RGBA{255, 0, 0, 200}
	}

	// Draw preview
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
