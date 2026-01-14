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
	"github.com/bklimczak/tanks/engine/assets"
	"github.com/bklimczak/tanks/engine/entity"
	"github.com/bklimczak/tanks/engine/fog"
	"github.com/bklimczak/tanks/engine/input"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/bklimczak/tanks/engine/network"
	"github.com/bklimczak/tanks/engine/render"
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
	StatePaused
	StateVictory
	StateDefeat
	StateMultiplayerLobby
	StateMultiplayerRoom
	StateMultiplayerPlaying
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
	engine             *engine.Engine
	units              []*entity.Unit
	buildings          []*entity.Building
	playerNexus        *entity.Building
	wreckages          []*entity.Wreckage
	projectiles        []*entity.Projectile
	terrainMap         *terrain.Map
	mapConfig          *terrain.MapConfig
	fogOfWar           *fog.FogOfWar
	resourceBar        *ui.ResourceBar
	commandPanel       *ui.CommandPanel
	minimap            *ui.Minimap
	mainMenu           *ui.MainMenu
	tooltip            *ui.Tooltip
	infoPanel          *ui.InfoPanel
	pauseMenu          *ui.PauseMenu
	lobbyBrowser       *ui.LobbyBrowser
	lobbyRoom          *ui.LobbyRoom
	networkClient      *network.Client
	enemyAI            *ai.EnemyAI
	assets             *assets.Manager
	entityRenderer     *render.EntityRenderer
	tankFactorySprite  *ebiten.Image
	tileImages         map[color.RGBA]*ebiten.Image
	grassTileScaled    *ebiten.Image
	terrainCache       *ebiten.Image
	screenWidth        int
	screenHeight       int
	debugTerrainTime   time.Duration
	debugMinimapTime   time.Duration
	debugFogTiles      int
	nextUnitID         uint64
	nextBuildingID     uint64
	nextWreckageID     uint64
	nextProjectileID   uint64
	state              GameState
	placementMode      bool
	placementDef       *entity.BuildingDef
	placementValid     bool
	elapsedTime        float64
	mpPlayerSlot       int
	mpIsReady          bool
	mpCameraPositioned bool
}

func NewGame() *Game {
	resourceBar := ui.NewResourceBar(baseWidth)
	commandPanel := ui.NewCommandPanel(resourceBar.Height(), baseHeight)
	minimapY := baseHeight - minimapHeight - minimapMargin
	minimap := ui.NewMinimap(minimapMargin, minimapY, minimapWidth, minimapHeight)
	mainMenu := ui.NewMainMenu()
	pauseMenu := ui.NewPauseMenu()

	assetManager := assets.NewManager("assets")
	terrain.LoadSprites()

	tankFactorySprite, _, err := ebitenutil.NewImageFromFile("assets/tank_factory.png")
	if err != nil {
		log.Printf("Warning: could not load tank factory sprite: %v", err)
	}

	// Load map configuration
	var mapConfig *terrain.MapConfig
	var terrainMap *terrain.Map

	mapConfig, err = terrain.LoadMapConfig("maps/skirmish.yaml")
	if err != nil {
		log.Printf("Map config not found, using default: %v", err)
		terrainMap = terrain.NewMap(worldWidth, worldHeight)
		terrainMap.Generate(42)
		terrainMap.PlaceMetalDeposit(400, 150)
	} else {
		terrainMap = mapConfig.ToMap()
	}

	actualWorldWidth := terrainMap.PixelWidth
	actualWorldHeight := terrainMap.PixelHeight
	minimap.SetWorldSize(actualWorldWidth, actualWorldHeight)
	tooltip := ui.NewTooltip()
	tooltip.SetScreenSize(baseWidth, baseHeight)
	infoPanel := ui.NewInfoPanel(baseWidth, baseHeight)

	eng := engine.New(actualWorldWidth, actualWorldHeight, baseWidth, baseHeight)
	entityRenderer := render.NewEntityRenderer(eng.Renderer, assetManager)

	lobbyBrowser := ui.NewLobbyBrowser()
	lobbyRoom := ui.NewLobbyRoom()

	g := &Game{
		engine:            eng,
		terrainMap:        terrainMap,
		mapConfig:         mapConfig,
		fogOfWar:          fog.New(actualWorldWidth, actualWorldHeight, terrain.TileSize),
		resourceBar:       resourceBar,
		commandPanel:      commandPanel,
		minimap:           minimap,
		mainMenu:          mainMenu,
		pauseMenu:         pauseMenu,
		lobbyBrowser:      lobbyBrowser,
		lobbyRoom:         lobbyRoom,
		tooltip:           tooltip,
		infoPanel:         infoPanel,
		assets:            assetManager,
		entityRenderer:    entityRenderer,
		tankFactorySprite: tankFactorySprite,
		state:             StateMenu,
		screenWidth:       int(baseWidth),
		screenHeight:      int(baseHeight),
		mpPlayerSlot:      -1,
	}
	g.engine.Collision.SetTerrain(terrainMap)

	// Load entities from map config or use legacy setup
	if mapConfig != nil {
		g.loadEntitiesFromConfig(mapConfig)
	} else {
		g.setupPlayerBase()
		g.setupEnemyBase()
	}

	return g
}

func (g *Game) setupPlayerBase() {
	startX, startY := g.findPassablePosition(300, 200)

	nexusDef := entity.BuildingDefs[entity.BuildingCommandNexus]
	commandNexus := entity.NewBuilding(g.nextBuildingID, startX, startY, nexusDef)
	commandNexus.Faction = entity.FactionPlayer
	commandNexus.Completed = true
	commandNexus.BuildProgress = 1.0
	g.buildings = append(g.buildings, commandNexus)
	g.playerNexus = commandNexus
	g.nextBuildingID++
	g.applyBuildingEffects(nexusDef)

	solarDef := entity.BuildingDefs[entity.BuildingSolarArray]
	solarX, solarY := g.findPassablePosition(startX+nexusDef.Size+20, startY)
	solarArray := entity.NewBuilding(g.nextBuildingID, solarX, solarY, solarDef)
	solarArray.Faction = entity.FactionPlayer
	solarArray.Completed = true
	solarArray.BuildProgress = 1.0
	g.buildings = append(g.buildings, solarArray)
	g.nextBuildingID++
	g.applyBuildingEffects(solarDef)

	tankDef := entity.UnitDefs[entity.UnitTypeTank]
	for i := 0; i < 2; i++ {
		tx, ty := g.findPassablePosition(startX+float64(i)*60, startY+nexusDef.Size+60)
		tank := entity.NewUnitFromDef(g.nextUnitID, tx, ty, tankDef, entity.FactionPlayer)
		g.units = append(g.units, tank)
		g.nextUnitID++
	}

	scoutDef := entity.UnitDefs[entity.UnitTypeScout]
	scoutX, scoutY := g.findPassablePosition(startX+100, startY+nexusDef.Size+60)
	scout := entity.NewUnitFromDef(g.nextUnitID, scoutX, scoutY, scoutDef, entity.FactionPlayer)
	g.units = append(g.units, scout)
	g.nextUnitID++
}

func (g *Game) setupEnemyBase() {
	enemyBaseX, enemyBaseY := 3500.0, 2700.0
	g.enemyAI = ai.NewEnemyAI(enemyBaseX, enemyBaseY)

	nexusDef := entity.BuildingDefs[entity.BuildingCommandNexus]
	enemyNexusX, enemyNexusY := g.findPassablePosition(enemyBaseX, enemyBaseY)
	enemyNexus := entity.NewBuilding(g.nextBuildingID, enemyNexusX, enemyNexusY, nexusDef)
	enemyNexus.Faction = entity.FactionEnemy
	enemyNexus.Color = entity.GetFactionTintedColor(nexusDef.Color, entity.FactionEnemy)
	enemyNexus.Completed = true
	enemyNexus.BuildProgress = 1.0
	g.buildings = append(g.buildings, enemyNexus)
	g.nextBuildingID++

	hoverBayDef := entity.BuildingDefs[entity.BuildingHoverBay]
	hoverBayX, hoverBayY := g.findPassablePosition(enemyBaseX+nexusDef.Size+20, enemyBaseY)
	enemyHoverBay := entity.NewBuilding(g.nextBuildingID, hoverBayX, hoverBayY, hoverBayDef)
	enemyHoverBay.Faction = entity.FactionEnemy
	enemyHoverBay.Color = entity.GetFactionTintedColor(hoverBayDef.Color, entity.FactionEnemy)
	enemyHoverBay.Completed = true
	enemyHoverBay.BuildProgress = 1.0
	g.buildings = append(g.buildings, enemyHoverBay)
	g.nextBuildingID++

	tankDef := entity.UnitDefs[entity.UnitTypeTank]
	for i := 0; i < 2; i++ {
		ex, ey := g.findPassablePosition(enemyBaseX-50+float64(i)*60, enemyBaseY+nexusDef.Size+40)
		tank := entity.NewUnitFromDef(g.nextUnitID, ex, ey, tankDef, entity.FactionEnemy)
		g.units = append(g.units, tank)
		g.nextUnitID++
	}

	scoutDef := entity.UnitDefs[entity.UnitTypeScout]
	enemyScoutX, enemyScoutY := g.findPassablePosition(enemyBaseX+50, enemyBaseY+nexusDef.Size+80)
	enemyScout := entity.NewUnitFromDef(g.nextUnitID, enemyScoutX, enemyScoutY, scoutDef, entity.FactionEnemy)
	g.units = append(g.units, enemyScout)
	g.nextUnitID++
}

func (g *Game) loadEntitiesFromConfig(config *terrain.MapConfig) {
	for _, factionConfig := range config.Factions {
		faction := g.getFactionFromConfig(factionConfig.Type)

		// Set starting resources for player
		if factionConfig.Type == "player" && factionConfig.Resources != nil {
			g.engine.Resources.Get(resource.Metal).Current = factionConfig.Resources.Metal
			g.engine.Resources.Get(resource.Energy).Current = factionConfig.Resources.Energy
		}

		// Load buildings
		for _, buildingConfig := range factionConfig.Buildings {
			building := g.createBuildingFromConfig(buildingConfig, faction)
			if building != nil {
				g.buildings = append(g.buildings, building)
				g.nextBuildingID++

				// Track player's Command Nexus
				if faction == entity.FactionPlayer && building.Type == entity.BuildingCommandNexus {
					g.playerNexus = building
				}

				// Apply building effects if completed
				if building.Completed {
					g.applyBuildingEffects(building.Def)
				}
			}
		}

		// Load units
		for _, unitConfig := range factionConfig.Units {
			count := unitConfig.Count
			if count <= 0 {
				count = 1
			}
			for i := 0; i < count; i++ {
				offsetX := float64(i%3) * 40
				offsetY := float64(i/3) * 40
				unit := g.createUnitFromConfig(unitConfig, faction, offsetX, offsetY)
				if unit != nil {
					g.units = append(g.units, unit)
					g.nextUnitID++
				}
			}
		}

		// Setup AI for enemy factions
		if factionConfig.Type == "ai" && len(factionConfig.Buildings) > 0 {
			// Find the Command Nexus position for AI base
			for _, b := range factionConfig.Buildings {
				if b.Type == "CommandNexus" {
					g.enemyAI = ai.NewEnemyAI(b.X, b.Y)
					break
				}
			}
		}
	}
}

func (g *Game) getFactionFromConfig(factionType string) entity.Faction {
	switch factionType {
	case "player":
		return entity.FactionPlayer
	case "ai":
		return entity.FactionEnemy
	default:
		return entity.FactionNeutral
	}
}

func (g *Game) createBuildingFromConfig(config terrain.BuildingConfig, faction entity.Faction) *entity.Building {
	buildingType := g.getBuildingTypeFromString(config.Type)
	def := entity.BuildingDefs[buildingType]
	if def == nil {
		log.Printf("Warning: unknown building type %s", config.Type)
		return nil
	}

	x, y := g.findPassablePosition(config.X, config.Y)
	building := entity.NewBuilding(g.nextBuildingID, x, y, def)
	building.Faction = faction

	if faction == entity.FactionEnemy {
		building.Color = entity.GetFactionTintedColor(def.Color, faction)
	}

	// Buildings from config default to completed unless specified otherwise
	completed := config.Completed
	if !config.Completed && config.Type != "" {
		completed = true // Default to completed for pre-placed buildings
	}
	if completed {
		building.Completed = true
		building.BuildProgress = 1.0
	}

	return building
}

func (g *Game) createUnitFromConfig(config terrain.UnitConfig, faction entity.Faction, offsetX, offsetY float64) *entity.Unit {
	var unitDef *entity.UnitDef

	// Check for custom tank configuration
	if config.Type == "Tank" && config.Color != "" && config.Hull > 0 && config.Gun > 0 {
		unitDef = entity.CreateTankDef(config.Color, config.Hull, config.Gun)
	} else {
		unitType := g.getUnitTypeFromString(config.Type)
		unitDef = entity.UnitDefs[unitType]
	}

	if unitDef == nil {
		log.Printf("Warning: unknown unit type %s", config.Type)
		return nil
	}

	x, y := g.findPassablePosition(config.X+offsetX, config.Y+offsetY)
	return entity.NewUnitFromDef(g.nextUnitID, x, y, unitDef, faction)
}

func (g *Game) getBuildingTypeFromString(typeName string) entity.BuildingType {
	switch typeName {
	case "CommandNexus":
		return entity.BuildingCommandNexus
	case "SolarArray":
		return entity.BuildingSolarArray
	case "SolarPanel":
		return entity.BuildingSolarPanel
	case "FusionReactor":
		return entity.BuildingFusionReactor
	case "MetalExtractor", "OreExtractor":
		return entity.BuildingMetalExtractor
	case "AlloyFoundry":
		return entity.BuildingAlloyFoundry
	case "TanksFactory", "VehicleFactory", "TankFactory":
		return entity.BuildingTanksFactory
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
	case "LaserTower":
		return entity.BuildingLaserTower
	default:
		return entity.BuildingCommandNexus
	}
}

func (g *Game) getUnitTypeFromString(typeName string) entity.UnitType {
	switch typeName {
	case "Tank":
		return entity.UnitTypeTank
	case "Scout":
		return entity.UnitTypeScout
	case "HeavyTank":
		return entity.UnitTypeHeavyTank
	case "LightTank":
		return entity.UnitTypeLightTank
	case "Artillery":
		return entity.UnitTypeArtillery
	case "RocketTank":
		return entity.UnitTypeRocketTank
	case "FlameTank":
		return entity.UnitTypeFlameTank
	case "AAVehicle":
		return entity.UnitTypeAAVehicle
	case "Constructor":
		return entity.UnitTypeConstructor
	default:
		return entity.UnitTypeTank
	}
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
	switch g.state {
	case StateMenu:
		return g.updateMenu(inputState)
	case StatePlaying:
		return g.updatePlaying(inputState)
	case StatePaused:
		return g.updatePaused(inputState)
	case StateVictory, StateDefeat:
		return g.updateEndScreen(inputState)
	case StateMultiplayerLobby:
		return g.updateMultiplayerLobby(inputState)
	case StateMultiplayerRoom:
		return g.updateMultiplayerRoom(inputState)
	case StateMultiplayerPlaying:
		return g.updateMultiplayerPlaying(inputState)
	}
	return nil
}
func (g *Game) updateEndScreen(inputState input.State) error {
	if inputState.EscapePressed {
		return ebiten.Termination
	}
	if inputState.EnterPressed {
		g.resetGame()
		g.state = StateMenu
	}
	return nil
}

func (g *Game) resetGame() {
	g.units = nil
	g.buildings = nil
	g.playerNexus = nil
	g.wreckages = nil
	g.projectiles = nil
	g.nextUnitID = 0
	g.nextBuildingID = 0
	g.nextWreckageID = 0
	g.nextProjectileID = 0
	g.placementMode = false
	g.placementDef = nil
	g.terrainCache = nil
	g.enemyAI = nil

	g.engine.Resources = resource.NewManager()

	actualWorldWidth := g.terrainMap.PixelWidth
	actualWorldHeight := g.terrainMap.PixelHeight
	g.fogOfWar = fog.New(actualWorldWidth, actualWorldHeight, terrain.TileSize)

	// Load entities from map config or use legacy setup
	if g.mapConfig != nil {
		g.loadEntitiesFromConfig(g.mapConfig)
	} else {
		g.setupPlayerBase()
		g.setupEnemyBase()
	}
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
	case ui.MenuOptionSkirmish:
		g.resetGame()
		g.state = StatePlaying
	case ui.MenuOptionMultiplayer:
		g.enterMultiplayerLobby()
	case ui.MenuOptionExit:
		return ebiten.Termination
	}
	return nil
}

func (g *Game) enterMultiplayerLobby() {
	g.state = StateMultiplayerLobby
	g.lobbyBrowser.SetConnecting(true)
	g.lobbyBrowser.ClearError()

	// Create network client if not exists
	if g.networkClient == nil {
		g.networkClient = network.NewClient("Player")
	}

	// Connect to server in background
	go func() {
		err := g.networkClient.Connect("localhost:8080")
		if err != nil {
			g.lobbyBrowser.SetError(err.Error())
			g.lobbyBrowser.SetConnecting(false)
			return
		}
		g.lobbyBrowser.SetConnecting(false)
		g.networkClient.RequestLobbyList()
	}()
}

func (g *Game) updateMultiplayerLobby(inputState input.State) error {
	g.lobbyBrowser.UpdateSize(float64(g.screenWidth), float64(g.screenHeight))

	if inputState.EscapePressed {
		if g.networkClient != nil {
			g.networkClient.Disconnect()
		}
		g.state = StateMenu
		return nil
	}

	// Update lobby list from network client
	if g.networkClient != nil && g.networkClient.IsConnected() {
		lobbies := g.networkClient.GetLobbies()
		uiLobbies := make([]ui.LobbyInfo, len(lobbies))
		for i, l := range lobbies {
			uiLobbies[i] = ui.LobbyInfo{
				ID:          l.ID,
				Name:        l.Name,
				PlayerCount: len(l.Players),
				MaxPlayers:  l.MaxPlayers,
				State:       l.State,
			}
		}
		g.lobbyBrowser.SetLobbies(uiLobbies)

		// Check for errors
		if err := g.networkClient.GetLastError(); err != "" {
			g.lobbyBrowser.SetError(err)
			g.networkClient.ClearError()
		}

		// Check if we joined a lobby
		if g.networkClient.InLobby() {
			g.state = StateMultiplayerRoom
			return nil
		}
	}

	g.lobbyBrowser.UpdateHover(inputState.MousePos)

	if inputState.LeftJustPressed {
		action := g.lobbyBrowser.HandleClick(inputState.MousePos)
		return g.handleLobbyBrowserAction(action)
	}

	action := g.lobbyBrowser.Update(inputState.MenuUp, inputState.MenuDown, inputState.EnterPressed)
	return g.handleLobbyBrowserAction(action)
}

func (g *Game) handleLobbyBrowserAction(action ui.LobbyBrowserAction) error {
	switch action {
	case ui.LobbyActionBack:
		if g.networkClient != nil {
			g.networkClient.Disconnect()
		}
		g.state = StateMenu
	case ui.LobbyActionRefresh:
		if g.networkClient != nil && g.networkClient.IsConnected() {
			g.networkClient.RequestLobbyList()
		}
	case ui.LobbyActionCreate:
		if g.networkClient != nil && g.networkClient.IsConnected() {
			g.networkClient.CreateLobby("New Game", 4)
		}
	case ui.LobbyActionJoin:
		if g.networkClient != nil && g.networkClient.IsConnected() {
			if lobby := g.lobbyBrowser.GetSelectedLobby(); lobby != nil {
				g.networkClient.JoinLobby(lobby.ID)
			}
		}
	}
	return nil
}

func (g *Game) updateMultiplayerRoom(inputState input.State) error {
	g.lobbyRoom.UpdateSize(float64(g.screenWidth), float64(g.screenHeight))

	if inputState.EscapePressed {
		if g.networkClient != nil {
			g.networkClient.LeaveLobby()
		}
		g.state = StateMultiplayerLobby
		return nil
	}

	// Update room state from network client
	if g.networkClient != nil && g.networkClient.IsConnected() {
		lobby := g.networkClient.GetCurrentLobby()
		if lobby != nil {
			isHost := g.networkClient.IsHost()
			g.lobbyRoom.SetLobby(lobby.ID, lobby.Name, lobby.MaxPlayers, isHost)

			players := make([]ui.PlayerSlot, len(lobby.Players))
			playerID := g.networkClient.GetPlayerID()
			allReady := len(lobby.Players) >= 2
			for i, p := range lobby.Players {
				players[i] = ui.PlayerSlot{
					Name:   p.Name,
					Ready:  p.Ready,
					IsHost: p.ID == lobby.HostID,
					IsYou:  p.ID == playerID,
				}
				if !p.Ready {
					allReady = false
				}
			}
			g.lobbyRoom.SetPlayers(players)
			g.lobbyRoom.SetCanStart(allReady && isHost)
			g.mpPlayerSlot = g.networkClient.GetYourSlot()
		}

		// Check if game started
		if g.networkClient.IsGameStarted() {
			g.initMultiplayerTerrain()
			g.state = StateMultiplayerPlaying
			g.mpCameraPositioned = false
			return nil
		}

		// Check if we left the lobby
		if !g.networkClient.InLobby() {
			g.state = StateMultiplayerLobby
			return nil
		}
	}

	if inputState.LeftJustPressed {
		action := g.lobbyRoom.HandleClick(inputState.MousePos)
		return g.handleLobbyRoomAction(action)
	}

	action := g.lobbyRoom.Update(inputState.EnterPressed)
	return g.handleLobbyRoomAction(action)
}

func (g *Game) handleLobbyRoomAction(action ui.LobbyRoomAction) error {
	switch action {
	case ui.LobbyRoomActionLeave:
		if g.networkClient != nil {
			g.networkClient.LeaveLobby()
		}
		g.state = StateMultiplayerLobby
	case ui.LobbyRoomActionReady:
		if g.networkClient != nil {
			g.mpIsReady = !g.mpIsReady
			g.networkClient.SetReady(g.mpIsReady)
			g.lobbyRoom.SetReady(g.mpIsReady)
		}
	case ui.LobbyRoomActionStart:
		if g.networkClient != nil {
			g.networkClient.StartGame()
		}
	}
	return nil
}

func (g *Game) updateMultiplayerPlaying(inputState input.State) error {
	if inputState.EscapePressed {
		if g.placementMode {
			g.placementMode = false
			g.placementDef = nil
		} else {
			if g.networkClient != nil {
				g.networkClient.LeaveLobby()
				g.networkClient.ResetGameState()
			}
			g.state = StateMenu
			return nil
		}
	}

	// Update game state from server
	if g.networkClient != nil && g.networkClient.IsConnected() {
		gameState := g.networkClient.GetGameState()
		if gameState != nil {
			g.updateFromServerState(gameState)
			g.updateFogOfWar()

			// Position camera on player's base once we have units
			if !g.mpCameraPositioned && len(g.buildings) > 0 {
				g.positionCameraOnPlayerBase()
				g.mpCameraPositioned = true
			}
		}

		// Check for game end
		if g.networkClient.IsGameEnded() {
			endInfo := g.networkClient.GetGameEndInfo()
			if endInfo != nil {
				if endInfo.WinnerSlot == g.mpPlayerSlot {
					g.state = StateVictory
				} else {
					g.state = StateDefeat
				}
			}
			return nil
		}
	}

	// Handle camera and input
	g.engine.UpdateViewportSize(float64(g.screenWidth), float64(g.screenHeight))
	g.resourceBar.UpdateWidth(float64(g.screenWidth))
	g.commandPanel.UpdateHeight(float64(g.screenHeight))

	cam := g.engine.Camera
	topOffset := g.resourceBar.Height()
	leftOffset := 0.0
	if g.commandPanel.IsVisible() {
		leftOffset = g.commandPanel.Width()
	}
	cam.HandleEdgeScroll(inputState.MousePos.X, inputState.MousePos.Y, topOffset, leftOffset)
	cam.HandleKeyScroll(inputState.ScrollUp, inputState.ScrollDown, inputState.ScrollLeft, inputState.ScrollRight)

	if inputState.MouseWheelY != 0 {
		if g.commandPanel.Contains(inputState.MousePos) {
			g.commandPanel.HandleScroll(inputState.MouseWheelY)
		} else if inputState.MouseWheelY > 0 {
			cam.ZoomIn(inputState.MousePos)
		} else {
			cam.ZoomOut(inputState.MousePos)
		}
	}

	// Handle minimap clicks
	if g.minimap.Contains(inputState.MousePos) {
		if inputState.LeftJustPressed || inputState.LeftPressed {
			worldPos := g.minimap.ScreenToWorld(inputState.MousePos)
			cam.MoveTo(worldPos)
		}
		return nil
	}

	// Handle building placement mode
	if g.placementMode {
		worldPos := cam.ScreenToWorld(inputState.MousePos)
		g.placementValid = g.canPlaceBuilding(worldPos, g.placementDef)
		if inputState.RightJustPressed {
			g.placementMode = false
			g.placementDef = nil
		}
		if inputState.LeftJustPressed && g.placementValid {
			if g.networkClient != nil {
				g.networkClient.SendPlaceBuildingCommand(int(g.placementDef.Type), worldPos.X, worldPos.Y)
			}
			if !inputState.ShiftHeld {
				g.placementMode = false
				g.placementDef = nil
			}
		}
		return nil
	}

	// Update info panel - do this early so it persists across interactions
	selectedBuilding := g.getSelectedBuilding()
	if selectedBuilding != nil {
		g.infoPanel.SetBuilding(selectedBuilding)
	} else {
		g.infoPanel.Hide()
	}

	// Update command panel based on selection
	factory := g.getSelectedFactory()
	buildingWithStructures := g.getSelectedBuildingWithStructures()
	if factory != nil {
		g.commandPanel.SetFactoryOptions(factory)
		g.commandPanel.UpdateQueueCounts()
	} else if buildingWithStructures != nil {
		g.commandPanel.SetBuildingBuildOptions(buildingWithStructures)
	} else {
		g.commandPanel.SetVisible(false)
	}

	// Handle command panel interactions
	if g.commandPanel.Contains(inputState.MousePos) {
		if inputState.LeftJustPressed {
			if clickedUnit := g.commandPanel.UpdateUnit(inputState.MousePos, true); clickedUnit != nil {
				if factory != nil && g.networkClient != nil {
					g.networkClient.SendProduceUnitCommand(factory.ID, int(clickedUnit.Type))
				}
			} else if clickedDef := g.commandPanel.Update(inputState.MousePos, true); clickedDef != nil {
				g.placementMode = true
				g.placementDef = clickedDef
			}
		} else if inputState.RightJustPressed {
			if clickedUnit := g.commandPanel.UpdateUnitRightClick(inputState.MousePos, true); clickedUnit != nil {
				if factory != nil && g.networkClient != nil {
					g.networkClient.SendCancelProductionCommand(factory.ID, int(clickedUnit.Type))
				}
			}
		} else {
			g.commandPanel.Update(inputState.MousePos, false)
		}

		// Handle tooltips
		if hoveredBuilding := g.commandPanel.GetHoveredBuilding(inputState.MousePos); hoveredBuilding != nil {
			bounds := g.commandPanel.GetHoveredButtonBounds(inputState.MousePos)
			if bounds != nil {
				g.tooltip.ShowBuilding(hoveredBuilding, bounds.Pos.X+bounds.Size.X+5, bounds.Pos.Y)
			}
		} else if hoveredUnit := g.commandPanel.GetHoveredUnit(inputState.MousePos); hoveredUnit != nil {
			bounds := g.commandPanel.GetHoveredButtonBounds(inputState.MousePos)
			if bounds != nil {
				g.tooltip.ShowUnit(hoveredUnit, bounds.Pos.X+bounds.Size.X+5, bounds.Pos.Y)
			}
		} else {
			g.tooltip.Hide()
		}
		return nil
	}

	g.tooltip.Hide()

	// Handle unit selection and commands
	g.handleMultiplayerSelection(inputState)
	if inputState.RightJustPressed {
		worldPos := cam.ScreenToWorld(inputState.MousePos)
		g.handleMultiplayerCommand(worldPos)
	}

	return nil
}

func (g *Game) updateFromServerState(state *network.GameStatePayload) {
	// Preserve selected unit IDs before rebuilding
	selectedUnitIDs := make(map[uint64]bool)
	for _, u := range g.units {
		if u.Selected {
			selectedUnitIDs[u.ID] = true
		}
	}

	// Preserve selected building IDs before rebuilding
	selectedBuildingIDs := make(map[uint64]bool)
	for _, b := range g.buildings {
		if b.Selected {
			selectedBuildingIDs[b.ID] = true
		}
	}

	// Clear and rebuild units from server state
	g.units = make([]*entity.Unit, 0, len(state.Units))
	for _, u := range state.Units {
		unitType := entity.UnitType(u.Type)
		unitDef := entity.UnitDefs[unitType]
		if unitDef == nil {
			continue
		}
		faction := g.getFactionFromSlot(u.OwnerSlot)
		unit := entity.NewUnitFromDef(u.ID, u.X, u.Y, unitDef, faction)
		unit.Angle = u.Angle
		unit.Health = u.Health
		unit.MaxHealth = u.MaxHealth
		unit.Selected = selectedUnitIDs[u.ID]
		g.units = append(g.units, unit)
	}

	// Clear and rebuild buildings from server state
	g.buildings = make([]*entity.Building, 0, len(state.Buildings))
	for _, b := range state.Buildings {
		buildingType := entity.BuildingType(b.Type)
		buildingDef := entity.BuildingDefs[buildingType]
		if buildingDef == nil {
			continue
		}
		faction := g.getFactionFromSlot(b.OwnerSlot)
		building := entity.NewBuilding(b.ID, b.X, b.Y, buildingDef)
		building.Faction = faction
		building.Health = b.Health
		building.MaxHealth = b.MaxHealth
		building.Completed = b.Completed
		building.BuildProgress = b.BuildProgress
		building.Selected = selectedBuildingIDs[b.ID]
		if faction != entity.FactionPlayer {
			building.Color = entity.GetFactionTintedColor(buildingDef.Color, faction)
		}
		g.buildings = append(g.buildings, building)
	}

	// Clear and rebuild projectiles from server state
	g.projectiles = make([]*entity.Projectile, 0, len(state.Projectiles))
	for _, p := range state.Projectiles {
		faction := g.getFactionFromSlot(p.OwnerSlot)
		projectile := &entity.Projectile{
			Entity: entity.Entity{
				ID:       p.ID,
				Position: emath.Vec2{X: p.X, Y: p.Y},
				Size:     emath.Vec2{X: 4, Y: 4},
				Color:    color.RGBA{255, 200, 50, 255},
				Active:   true,
				Faction:  faction,
			},
		}
		g.projectiles = append(g.projectiles, projectile)
	}

	// Update resources for our player
	for _, p := range state.Players {
		if p.Slot == g.mpPlayerSlot {
			g.engine.Resources.Get(resource.Metal).Current = p.Resources.Metal
			g.engine.Resources.Get(resource.Metal).Capacity = p.Resources.MetalCap
			g.engine.Resources.Get(resource.Energy).Current = p.Resources.Energy
			g.engine.Resources.Get(resource.Energy).Capacity = p.Resources.EnergyCap
			break
		}
	}
}

func (g *Game) getFactionFromSlot(slot int) entity.Faction {
	if slot == g.mpPlayerSlot {
		return entity.FactionPlayer
	}
	return entity.FactionEnemy
}

func (g *Game) positionCameraOnPlayerBase() {
	// Find a player-owned building (preferably Command Nexus) to center camera on
	for _, b := range g.buildings {
		if b.Faction == entity.FactionPlayer {
			g.engine.Camera.MoveTo(b.Center())
			return
		}
	}
	// Fallback to first player unit
	for _, u := range g.units {
		if u.Faction == entity.FactionPlayer {
			g.engine.Camera.MoveTo(u.Center())
			return
		}
	}
}

func (g *Game) initMultiplayerTerrain() {
	// Create a fresh terrain map matching server dimensions
	g.terrainMap = terrain.NewMap(6400, 3600)
	g.terrainMap.GenerateGrassOnly()

	// Place metal deposits matching server layout
	centerX := float64(g.terrainMap.PixelWidth) / 2
	centerY := float64(g.terrainMap.PixelHeight) / 2

	// Center deposits (contested)
	g.terrainMap.PlaceMetalDeposit(centerX, centerY)
	g.terrainMap.PlaceMetalDeposit(centerX+50, centerY)
	g.terrainMap.PlaceMetalDeposit(centerX-50, centerY)
	g.terrainMap.PlaceMetalDeposit(centerX, centerY+50)
	g.terrainMap.PlaceMetalDeposit(centerX, centerY-50)

	// Spawn positions matching server
	spawnPositions := []emath.Vec2{
		{X: 400, Y: 300},
		{X: 5800, Y: 3100},
		{X: 5800, Y: 300},
		{X: 400, Y: 3100},
	}

	// Place metal deposits near each spawn
	for _, spawn := range spawnPositions {
		metalX := spawn.X - 100
		metalY := spawn.Y - 100
		if metalX > 0 && metalY > 0 {
			g.terrainMap.PlaceMetalDeposit(metalX, metalY)
			g.terrainMap.PlaceMetalDeposit(metalX+50, metalY)
		}
	}

	// Reset terrain cache so it gets rebuilt with new terrain
	g.terrainCache = nil

	// Reset fog of war for new terrain
	g.fogOfWar = fog.New(g.terrainMap.PixelWidth, g.terrainMap.PixelHeight, terrain.TileSize)
}

func (g *Game) handleMultiplayerSelection(inputState input.State) {
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

func (g *Game) handleMultiplayerCommand(worldPos emath.Vec2) {
	// Send move command to server for selected units
	if g.networkClient == nil || !g.networkClient.IsConnected() {
		return
	}

	var selectedIDs []uint64
	for _, u := range g.units {
		if u.Selected && u.Faction == entity.FactionPlayer {
			selectedIDs = append(selectedIDs, u.ID)
		}
	}

	if len(selectedIDs) > 0 {
		g.networkClient.SendMoveCommand(selectedIDs, worldPos.X, worldPos.Y)
	}
}

func (g *Game) updatePaused(inputState input.State) error {
	g.pauseMenu.UpdateSize(float64(g.screenWidth), float64(g.screenHeight))
	g.pauseMenu.UpdateHover(inputState.MousePos)

	if inputState.EscapePressed {
		g.state = StatePlaying
		return nil
	}

	if inputState.LeftJustPressed {
		if option := g.pauseMenu.HandleClick(inputState.MousePos); option >= 0 {
			return g.handlePauseMenuSelection(option)
		}
	}

	option := g.pauseMenu.Update(
		inputState.MenuUp,
		inputState.MenuDown,
		inputState.EnterPressed,
	)

	if option >= 0 {
		return g.handlePauseMenuSelection(option)
	}

	return nil
}

func (g *Game) handlePauseMenuSelection(option ui.PauseMenuOption) error {
	switch option {
	case ui.PauseOptionResume:
		g.state = StatePlaying
	case ui.PauseOptionMainMenu:
		g.state = StateMenu
	case ui.PauseOptionQuit:
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
			g.state = StatePaused
			return nil
		}
	}
	if inputState.BuildTankPressed {
		if factory := g.getSelectedFactory(); factory != nil {
			factory.QueueProduction(entity.UnitDefs[entity.UnitTypeTank])
		}
	}
	g.elapsedTime += tickRate
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
	// Handle mouse wheel - scroll command panel if hovering over it, otherwise zoom camera
	if inputState.MouseWheelY != 0 {
		if g.commandPanel.Contains(inputState.MousePos) {
			g.commandPanel.HandleScroll(inputState.MouseWheelY)
		} else if inputState.MouseWheelY > 0 {
			cam.ZoomIn(inputState.MousePos)
		} else {
			cam.ZoomOut(inputState.MousePos)
		}
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
		buildingWithStructures := g.getSelectedBuildingWithStructures()
		if factory != nil {
			g.commandPanel.SetFactoryOptions(factory)
			g.commandPanel.UpdateQueueCounts()
		} else if buildingWithStructures != nil {
			g.commandPanel.SetBuildingBuildOptions(buildingWithStructures)
		} else {
			g.commandPanel.SetVisible(false)
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
			if hoveredBuilding := g.commandPanel.GetHoveredBuilding(inputState.MousePos); hoveredBuilding != nil {
				bounds := g.commandPanel.GetHoveredButtonBounds(inputState.MousePos)
				if bounds != nil {
					g.tooltip.ShowBuilding(hoveredBuilding, bounds.Pos.X+bounds.Size.X+5, bounds.Pos.Y)
				}
			} else if hoveredUnit := g.commandPanel.GetHoveredUnit(inputState.MousePos); hoveredUnit != nil {
				bounds := g.commandPanel.GetHoveredButtonBounds(inputState.MousePos)
				if bounds != nil {
					g.tooltip.ShowUnit(hoveredUnit, bounds.Pos.X+bounds.Size.X+5, bounds.Pos.Y)
				}
			} else {
				g.tooltip.Hide()
			}
		} else {
			g.tooltip.Hide()
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
	g.updateInfoPanel()
	g.checkVictoryConditions()
	return nil
}

func (g *Game) updateInfoPanel() {
	selectedBuilding := g.getSelectedBuilding()
	if selectedBuilding != nil {
		g.infoPanel.SetBuilding(selectedBuilding)
	} else {
		g.infoPanel.Hide()
	}
}

func (g *Game) checkVictoryConditions() {
	// Victory: destroy all enemy units and buildings
	// Defeat: lose all player units and buildings
	enemyUnits := 0
	enemyBuildings := 0
	playerUnits := 0
	playerBuildings := 0

	for _, u := range g.units {
		if !u.Active {
			continue
		}
		if u.Faction == entity.FactionEnemy {
			enemyUnits++
		} else if u.Faction == entity.FactionPlayer {
			playerUnits++
		}
	}

	for _, b := range g.buildings {
		if b.Faction == entity.FactionEnemy {
			enemyBuildings++
		} else if b.Faction == entity.FactionPlayer {
			playerBuildings++
		}
	}

	if enemyUnits == 0 && enemyBuildings == 0 {
		g.state = StateVictory
	} else if playerUnits == 0 && playerBuildings == 0 {
		g.state = StateDefeat
	}
}

// GameStateReader interface implementation for victory/defeat condition checks
func (g *Game) GetUnits() []*entity.Unit {
	return g.units
}

func (g *Game) GetBuildings() []*entity.Building {
	return g.buildings
}

func (g *Game) GetElapsedTime() float64 {
	return g.elapsedTime
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
			// Check if player's Command Nexus was destroyed
			if b == g.playerNexus {
				g.state = StateDefeat
			}

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
		b.UpdateAnimation(tickRate)

		// Auto-construction for buildings under construction
		if !b.Completed {
			if b.UpdateConstruction(tickRate, g.engine.Resources) {
				g.applyBuildingEffects(b.Def)
			}
		}

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

func (g *Game) getSelectedBuilding() *entity.Building {
	for _, b := range g.buildings {
		if b.Selected && b.Faction == entity.FactionPlayer {
			return b
		}
	}
	return nil
}

func (g *Game) getSelectedBuildingWithStructures() *entity.Building {
	for _, b := range g.buildings {
		if b.Selected && b.Faction == entity.FactionPlayer && b.Def != nil && len(b.Def.BuildableStructures) > 0 {
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
	buildingPos := snapToGrid(worldPos)
	building := entity.NewBuildingUnderConstruction(g.nextBuildingID, buildingPos.X, buildingPos.Y, def)
	building.Faction = entity.FactionPlayer
	g.buildings = append(g.buildings, building)
	g.nextBuildingID++
}
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateMenu:
		g.mainMenu.Draw(screen)
		return
	case StatePlaying:
		g.drawPlaying(screen)
	case StatePaused:
		g.drawPlaying(screen)
		g.pauseMenu.Draw(screen)
	case StateVictory:
		g.drawPlaying(screen)
		g.drawEndScreen(screen, "VICTORY!", color.RGBA{0, 200, 0, 255})
	case StateDefeat:
		g.drawPlaying(screen)
		g.drawEndScreen(screen, "DEFEAT", color.RGBA{200, 0, 0, 255})
	case StateMultiplayerLobby:
		g.lobbyBrowser.Draw(screen)
	case StateMultiplayerRoom:
		g.lobbyRoom.Draw(screen)
	case StateMultiplayerPlaying:
		g.drawMultiplayerPlaying(screen)
	}
}

func (g *Game) drawEndScreen(screen *ebiten.Image, text string, textColor color.RGBA) {
	screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()

	overlayColor := color.RGBA{0, 0, 0, 180}
	vector.FillRect(screen, 0, 0, float32(screenW), float32(screenH), overlayColor, false)

	boxWidth := 400.0
	boxHeight := 200.0
	boxX := (float64(screenW) - boxWidth) / 2
	boxY := (float64(screenH) - boxHeight) / 2

	boxColor := color.RGBA{30, 30, 40, 240}
	borderColor := color.RGBA{80, 80, 100, 255}
	vector.FillRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), boxColor, false)
	vector.StrokeRect(screen, float32(boxX), float32(boxY), float32(boxWidth), float32(boxHeight), 2, borderColor, false)

	textX := int(boxX) + int(boxWidth)/2 - len(text)*4
	textY := int(boxY) + 60
	ebitenutil.DebugPrintAt(screen, text, textX, textY)

	vector.FillRect(screen, float32(boxX)+20, float32(boxY)+20, 10, 10, textColor, false)
	vector.FillRect(screen, float32(boxX)+float32(boxWidth)-30, float32(boxY)+20, 10, 10, textColor, false)
	vector.FillRect(screen, float32(boxX)+20, float32(boxY)+float32(boxHeight)-30, 10, 10, textColor, false)
	vector.FillRect(screen, float32(boxX)+float32(boxWidth)-30, float32(boxY)+float32(boxHeight)-30, 10, 10, textColor, false)

	instruction := "Press ENTER to return to menu or ESC to quit"
	instrX := int(boxX) + int(boxWidth)/2 - len(instruction)*3
	instrY := int(boxY) + 120
	ebitenutil.DebugPrintAt(screen, instruction, instrX, instrY)
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
		if cam.IsVisible(p.Bounds()) {
			if p.Faction == entity.FactionPlayer || g.fogOfWar.IsVisible(p.Bounds()) {
				g.drawProjectile(screen, p)
			}
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
	viewportBounds := cam.GetViewportBounds()
	debugText := fmt.Sprintf("Cam: pos=(%.0f,%.0f) viewport=(%.0f,%.0f) zoom=%.2f visible=(%.0f,%.0f)",
		cam.Position.X, cam.Position.Y, cam.ViewportSize.X, cam.ViewportSize.Y, cam.Zoom,
		viewportBounds.Size.X, viewportBounds.Size.Y)
	ebitenutil.DebugPrintAt(screen, debugText, 10, int(baseHeight)-40)
	mmBounds := g.minimap.Bounds()
	mmWorld := g.minimap.WorldSize()
	// Calculate expected minimap viewport size
	scaleX := mmBounds.Size.X / mmWorld.X
	scaleY := mmBounds.Size.Y / mmWorld.Y
	expectedVpW := viewportBounds.Size.X * scaleX
	expectedVpH := viewportBounds.Size.Y * scaleY
	debugText2 := fmt.Sprintf("Minimap: bounds=(%.0f,%.0f) worldSize=(%.0f,%.0f) vpRect=(%.1f,%.1f)",
		mmBounds.Size.X, mmBounds.Size.Y, mmWorld.X, mmWorld.Y, expectedVpW, expectedVpH)
	ebitenutil.DebugPrintAt(screen, debugText2, 10, int(baseHeight)-60)
	if g.terrainCache != nil {
		tcBounds := g.terrainCache.Bounds()
		debugText3 := fmt.Sprintf("TerrainCache: size=(%d,%d) terrain=(%d,%d tiles)",
			tcBounds.Dx(), tcBounds.Dy(), g.terrainMap.Width, g.terrainMap.Height)
		ebitenutil.DebugPrintAt(screen, debugText3, 10, int(baseHeight)-80)
	}
	// Debug minimap viewport position
	debugText4 := fmt.Sprintf("MM pos=(%.0f,%.0f) camWorldPos=(%.0f,%.0f)",
		mmBounds.Pos.X, mmBounds.Pos.Y, cam.Position.X, cam.Position.Y)
	ebitenutil.DebugPrintAt(screen, debugText4, 10, int(baseHeight)-100)
	g.infoPanel.Draw(screen)
	g.tooltip.Draw(screen)
}

func (g *Game) drawMultiplayerPlaying(screen *ebiten.Image) {
	r := g.engine.Renderer
	cam := g.engine.Camera
	r.Clear(screen)

	g.drawTerrain(screen)

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

	// Draw building placement preview
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
	g.minimap.Draw(screen, cam, g.terrainMap, g.fogOfWar, minimapEntities)

	instructionX := int(g.commandPanel.Width()) + 10
	if !g.commandPanel.IsVisible() {
		instructionX = 10
	}
	instructions := "MULTIPLAYER | WASD/Arrows: Scroll | Left Click: Select | Right Click: Move | ESC: Leave"
	r.DrawTextAt(screen, instructions, instructionX, int(g.resourceBar.Height())+5)

	fpsText := fmt.Sprintf("FPS: %.1f  Units: %d  Buildings: %d  Slot: %d",
		ebiten.ActualFPS(), len(g.units), len(g.buildings), g.mpPlayerSlot)
	ebitenutil.DebugPrintAt(screen, fpsText, 10, int(baseHeight)-20)

	g.infoPanel.Draw(screen)
	g.tooltip.Draw(screen)
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
	// Cache scaled grass tile if not done yet
	if g.grassTileScaled == nil && terrain.SpritesLoaded() {
		grassTile := terrain.GetGrassTile()
		if grassTile != nil {
			scale := float64(terrain.TileSize) / float64(terrain.TilesetTileSize)
			g.grassTileScaled = ebiten.NewImage(terrain.TileSize, terrain.TileSize)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			g.grassTileScaled.DrawImage(grassTile, op)
		}
	}

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

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(screenX, screenY)

				// Use grass sprite for grass tiles, fallback to color for others
				if tile.Type == terrain.TileGrass && g.grassTileScaled != nil {
					g.terrainCache.DrawImage(g.grassTileScaled, op)
				} else {
					tileColor := terrain.TileColorVariation(tile.Type, x, y).(color.RGBA)
					tileImg := g.getTileImage(tileColor)
					g.terrainCache.DrawImage(tileImg, op)
				}

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
	screenCenter := cam.WorldToScreen(u.Center())
	if u.Selected {
		selectionWidth := scaledSize.X + selectionMargin*2*zoom
		selectionHeight := scaledSize.Y + selectionMargin*2*zoom
		r.DrawRotatedRect(screen, screenCenter, selectionWidth, selectionHeight, u.Angle, color.RGBA{0, 255, 0, 128})
	}
	g.entityRenderer.DrawUnit(screen, u, screenPos, screenCenter, zoom)
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

	// Handle tank factory sprite separately (legacy, not migrated to SpritePath yet)
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
	} else {
		g.entityRenderer.DrawBuilding(screen, b, screenPos, zoom)
	}

	// Draw border for buildings without sprites
	if g.entityRenderer.NeedsBorder(b) && (b.Type != entity.BuildingTankFactory || !b.Completed || g.tankFactorySprite == nil) {
		borderColor := color.RGBA{60, 60, 60, 255}
		if !b.Completed {
			borderColor = color.RGBA{200, 150, 50, 255}
		}
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
	if g.screenWidth != outsideWidth || g.screenHeight != outsideHeight {
		g.screenWidth = outsideWidth
		g.screenHeight = outsideHeight
		g.tooltip.SetScreenSize(float64(outsideWidth), float64(outsideHeight))
		g.infoPanel.UpdatePosition(float64(outsideWidth), float64(outsideHeight))
	}
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
