package main

import (
	"github.com/bklimczak/tanks/engine"
	"github.com/bklimczak/tanks/engine/entity"
	"github.com/bklimczak/tanks/engine/input"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/resource"
	"github.com/bklimczak/tanks/engine/terrain"
	"github.com/bklimczak/tanks/engine/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"log"
	"math"
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
	engine           *engine.Engine
	units            []*entity.Unit
	buildings        []*entity.Building
	wreckages        []*entity.Wreckage
	projectiles      []*entity.Projectile
	terrainMap       *terrain.Map
	resourceBar      *ui.ResourceBar
	commandPanel     *ui.CommandPanel
	minimap          *ui.Minimap
	mainMenu         *ui.MainMenu
	screenWidth      int
	screenHeight     int
	nextUnitID       uint64
	nextBuildingID   uint64
	nextWreckageID   uint64
	nextProjectileID uint64
	state            GameState
	placementMode    bool
	placementDef     *entity.BuildingDef
	placementValid   bool
}

func NewGame() *Game {
	resourceBar := ui.NewResourceBar(baseWidth)
	commandPanel := ui.NewCommandPanel(resourceBar.Height(), baseHeight)
	minimapY := baseHeight - minimapHeight - minimapMargin
	minimap := ui.NewMinimap(minimapMargin, minimapY, minimapWidth, minimapHeight)
	minimap.SetWorldSize(worldWidth, worldHeight)
	mainMenu := ui.NewMainMenu()
	terrainMap, err := terrain.LoadMapFromFile("maps/default.yaml")
	if err != nil {
		terrainMap = terrain.NewMap(worldWidth, worldHeight)
		terrainMap.Generate(42)
		terrainMap.PlaceMetalDeposit(400, 150)
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
		state:        StateMenu,
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
			g.placeBuilding(worldPos, g.placementDef)
			g.placementMode = false
			g.placementDef = nil
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
				g.commandMoveSelected(worldPos)
			}
		}
	}
	g.engine.Resources.ResetDrains()
	g.updateUnits()
	g.updateBuildings()
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
func (g *Game) updateUnits() {
	for _, u := range g.units {
		if !u.Active {
			continue
		}
		if u.HasBuildTask {
			g.updateConstructorBuildTask(u)
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
		u.ApplyPosition(resolvedPos)
	}
	g.updateCombat()
	g.cleanupDeadUnits()
}
func (g *Game) updateCombat() {
	for _, u := range g.units {
		if !u.Active || !u.CanAttack() {
			continue
		}
		if u.AttackTarget == nil || !u.AttackTarget.Active {
			u.ClearAttackTarget()
			var nearestEnemy *entity.Unit
			nearestDist := u.Range + 1
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
		if u.UpdateCombat(tickRate) {
			if u.AttackTarget != nil && u.AttackTarget.Active {
				projectile := entity.NewProjectile(g.nextProjectileID, u, u.AttackTarget)
				g.projectiles = append(g.projectiles, projectile)
				g.nextProjectileID++
			}
		}
	}
	g.updateProjectiles()
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
func (g *Game) updateConstructorBuildTask(u *entity.Unit) {
	if u.HasTarget {
		return
	}
	if !u.IsNearBuildSite() {
		buildSiteTarget := emath.Vec2{
			X: u.BuildPos.X - u.Size.X - 5,
			Y: u.BuildPos.Y + u.BuildDef.Size/2,
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
func (g *Game) getSelectedFactory() *entity.Building {
	for _, b := range g.buildings {
		if b.Selected && b.Type == entity.BuildingTankFactory {
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
func (g *Game) placeBuilding(worldPos emath.Vec2, def *entity.BuildingDef) {
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
	constructor.SetBuildTask(def, buildingPos)
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
	g.drawTerrain(screen)
	for _, w := range g.wreckages {
		if cam.IsVisible(w.Bounds()) {
			g.drawWreckage(screen, w)
		}
	}
	for _, b := range g.buildings {
		if cam.IsVisible(b.Bounds()) {
			g.drawBuilding(screen, b)
		}
	}
	for _, u := range g.units {
		if cam.IsVisible(u.Bounds()) {
			g.drawUnit(screen, u)
		}
	}
	for _, p := range g.projectiles {
		if cam.IsVisible(p.Bounds()) {
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
	startX, startY, endX, endY := g.terrainMap.GetVisibleTiles(
		cam.Position.X, cam.Position.Y,
		cam.ViewportSize.X, cam.ViewportSize.Y,
	)
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			tile := g.terrainMap.Tiles[y][x]
			tileColor := terrain.TileColorVariation(tile.Type, x, y)
			worldX := float64(x) * terrain.TileSize
			worldY := float64(y) * terrain.TileSize
			screenPos := cam.WorldToScreen(emath.Vec2{X: worldX, Y: worldY})
			tileRect := emath.Rect{
				Pos:  screenPos,
				Size: emath.Vec2{X: terrain.TileSize + 1, Y: terrain.TileSize + 1},
			}
			r.DrawRect(screen, tileRect, tileColor)
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
	screenPos := cam.WorldToScreen(u.Position)
	screenBounds := emath.Rect{Pos: screenPos, Size: u.Size}
	if u.Selected {
		selectionRect := emath.Rect{
			Pos:  screenPos.Sub(emath.Vec2{X: selectionMargin, Y: selectionMargin}),
			Size: u.Size.Add(emath.Vec2{X: selectionMargin * 2, Y: selectionMargin * 2}),
		}
		r.DrawRect(screen, selectionRect, color.RGBA{0, 255, 0, 128})
	}
	r.DrawRect(screen, screenBounds, u.Color)
	screenCenter := cam.WorldToScreen(u.Center())
	arrowLength := u.Size.X * 0.5
	arrowWidth := u.Size.X * 0.25
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
		r.DrawLine(screen,
			emath.Vec2{X: plusX - 3, Y: plusY},
			emath.Vec2{X: plusX + 3, Y: plusY},
			2, color.RGBA{0, 0, 0, 200})
		r.DrawLine(screen,
			emath.Vec2{X: plusX, Y: plusY - 3},
			emath.Vec2{X: plusX, Y: plusY + 3},
			2, color.RGBA{0, 0, 0, 200})
	}
	if u.HasTarget && u.Selected {
		screenTarget := cam.WorldToScreen(u.Target)
		r.DrawCircle(screen, screenTarget, 4, color.RGBA{0, 255, 0, 200})
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
				Size: emath.Vec2{X: u.BuildDef.Size, Y: u.BuildDef.Size},
			}
			r.DrawRectOutline(screen, buildRect, 2, color.RGBA{255, 200, 50, 200})
		}
	}
	if u.Health < u.MaxHealth || u.Selected {
		barWidth := u.Size.X
		barHeight := 4.0
		barX := screenPos.X
		barY := screenPos.Y - barHeight - 2
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
func (g *Game) drawProjectile(screen *ebiten.Image, p *entity.Projectile) {
	r := g.engine.Renderer
	cam := g.engine.Camera
	screenPos := cam.WorldToScreen(p.Position)
	screenBounds := emath.Rect{Pos: screenPos, Size: p.Size}
	r.DrawRect(screen, screenBounds, p.Color)
	trailEnd := p.Position.Sub(p.Direction.Mul(8))
	screenTrailEnd := cam.WorldToScreen(trailEnd)
	r.DrawLine(screen, cam.WorldToScreen(p.Center()), screenTrailEnd, 2, color.RGBA{255, 150, 0, 150})
}
func (g *Game) drawWreckage(screen *ebiten.Image, w *entity.Wreckage) {
	r := g.engine.Renderer
	cam := g.engine.Camera
	screenPos := cam.WorldToScreen(w.Position)
	screenBounds := emath.Rect{Pos: screenPos, Size: w.Size}
	r.DrawRect(screen, screenBounds, w.Color)
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
	screenPos := cam.WorldToScreen(b.Position)
	screenBounds := emath.Rect{Pos: screenPos, Size: b.Size}
	if b.Selected {
		selectionRect := emath.Rect{
			Pos:  screenPos.Sub(emath.Vec2{X: selectionMargin, Y: selectionMargin}),
			Size: b.Size.Add(emath.Vec2{X: selectionMargin * 2, Y: selectionMargin * 2}),
		}
		r.DrawRect(screen, selectionRect, color.RGBA{0, 255, 0, 128})
	}
	if b.Completed {
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
		lineSpacing := 10.0
		for i := 0.0; i < b.Size.X; i += lineSpacing {
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
	r.DrawRectOutline(screen, screenBounds, 2, borderColor)
	if !b.Completed {
		barWidth := b.Size.X - 10
		barHeight := 8.0
		barX := screenPos.X + 5
		barY := screenPos.Y + b.Size.Y/2 - barHeight/2
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
		barWidth := b.Size.X - 10
		barHeight := 6.0
		barX := screenPos.X + 5
		barY := screenPos.Y + b.Size.Y - 12
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
		r.DrawCircle(screen, rallyScreen, 5, color.RGBA{255, 255, 0, 200})
		buildingCenter := cam.WorldToScreen(b.Center())
		r.DrawLine(screen, buildingCenter, rallyScreen, 1, color.RGBA{255, 255, 0, 100})
	}
}
func (g *Game) drawPlacementPreview(screen *ebiten.Image) {
	r := g.engine.Renderer
	cam := g.engine.Camera
	mousePos := g.engine.Input.State().MousePos
	worldPos := cam.ScreenToWorld(mousePos)
	buildingPos := snapToGrid(worldPos)
	screenPos := cam.WorldToScreen(buildingPos)
	screenBounds := emath.Rect{
		Pos:  screenPos,
		Size: emath.Vec2{X: g.placementDef.Size, Y: g.placementDef.Size},
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
