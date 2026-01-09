package render

import (
	"image"
	"image/color"
	"math"

	"github.com/bklimczak/tanks/engine/assets"
	"github.com/bklimczak/tanks/engine/entity"
	emath "github.com/bklimczak/tanks/engine/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type EntityRenderer struct {
	renderer   *Renderer
	assets     *assets.Manager
	whitePixel *ebiten.Image
}

func NewEntityRenderer(r *Renderer, a *assets.Manager) *EntityRenderer {
	return &EntityRenderer{
		renderer: r,
		assets:   a,
	}
}

func (er *EntityRenderer) DrawUnit(screen *ebiten.Image, u *entity.Unit, screenPos, screenCenter emath.Vec2, zoom float64) {
	scaledSize := u.Size.Mul(zoom)

	def := u.Def
	if def == nil {
		def = entity.UnitDefs[u.Type]
	}

	// Check if this unit has hull+gun sprites (tank-style rendering)
	if def != nil && def.HullSpritePath != "" {
		er.drawTank(screen, u, screenCenter, zoom)
		return
	}

	// Check for single sprite
	if def != nil && def.SpritePath != "" {
		sprite := er.assets.GetSprite(def.SpritePath)
		if sprite != nil {
			er.drawUnitSprite(screen, sprite, u, screenCenter, zoom)
			return
		}
	}

	er.drawUnitFallback(screen, u, screenPos, screenCenter, scaledSize, zoom)
}

func (er *EntityRenderer) drawUnitSprite(screen *ebiten.Image, sprite *ebiten.Image, u *entity.Unit, screenCenter emath.Vec2, zoom float64) {
	spriteW := float64(sprite.Bounds().Dx())
	spriteH := float64(sprite.Bounds().Dy())
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

	screen.DrawImage(sprite, op)
}

func (er *EntityRenderer) drawUnitFallback(screen *ebiten.Image, u *entity.Unit, screenPos, screenCenter emath.Vec2, scaledSize emath.Vec2, zoom float64) {
	screenBounds := emath.Rect{Pos: screenPos, Size: scaledSize}
	er.renderer.DrawRect(screen, screenBounds, u.Color)

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
	er.renderer.DrawLine(screen, emath.Vec2{X: baseX1, Y: baseY1}, emath.Vec2{X: tipX, Y: tipY}, 2, arrowColor)
	er.renderer.DrawLine(screen, emath.Vec2{X: baseX2, Y: baseY2}, emath.Vec2{X: tipX, Y: tipY}, 2, arrowColor)
	er.renderer.DrawLine(screen, emath.Vec2{X: baseX1, Y: baseY1}, emath.Vec2{X: baseX2, Y: baseY2}, 2, arrowColor)

	if u.Type == entity.UnitTypeConstructor || u.Type == entity.UnitTypeTechnician {
		plusOffset := arrowLength * 0.3
		plusX := screenCenter.X - math.Cos(u.Angle)*plusOffset
		plusY := screenCenter.Y - math.Sin(u.Angle)*plusOffset
		plusSize := 3.0 * zoom
		er.renderer.DrawLine(screen,
			emath.Vec2{X: plusX - plusSize, Y: plusY},
			emath.Vec2{X: plusX + plusSize, Y: plusY},
			2, color.RGBA{0, 0, 0, 200})
		er.renderer.DrawLine(screen,
			emath.Vec2{X: plusX, Y: plusY - plusSize},
			emath.Vec2{X: plusX, Y: plusY + plusSize},
			2, color.RGBA{0, 0, 0, 200})
	}
}

func (er *EntityRenderer) drawTank(screen *ebiten.Image, u *entity.Unit, screenCenter emath.Vec2, zoom float64) {
	def := u.Def
	if def == nil {
		def = entity.UnitDefs[u.Type]
	}

	// Try to use sprites if available
	var hullSprite, gunSprite *ebiten.Image
	if def != nil && def.HullSpritePath != "" {
		hullSprite = er.assets.GetSprite(def.HullSpritePath)
	}
	if def != nil && def.GunSpritePath != "" {
		gunSprite = er.assets.GetSprite(def.GunSpritePath)
	}

	if hullSprite != nil && gunSprite != nil {
		er.drawTankWithSprites(screen, u, hullSprite, gunSprite, screenCenter, zoom)
		return
	}

	// Fallback to procedural drawing
	er.drawTankFallback(screen, u, screenCenter, zoom)
}

func (er *EntityRenderer) drawTankWithSprites(screen *ebiten.Image, u *entity.Unit, hullSprite, gunSprite *ebiten.Image, screenCenter emath.Vec2, zoom float64) {
	// Get sprite scale from unit definition
	spriteScale := 1.0
	def := u.Def
	if def == nil {
		def = entity.UnitDefs[u.Type]
	}
	if def != nil && def.SpriteScale > 0 {
		spriteScale = def.SpriteScale
	}
	totalScale := zoom * spriteScale

	// Draw hull (body) first
	hullW := float64(hullSprite.Bounds().Dx())
	hullH := float64(hullSprite.Bounds().Dy())

	hullOp := &ebiten.DrawImageOptions{}
	hullOp.GeoM.Translate(-hullW/2, -hullH/2)
	hullOp.GeoM.Rotate(u.Angle + math.Pi/2) // Sprite faces up, so add 90 degrees
	hullOp.GeoM.Scale(totalScale, totalScale)
	hullOp.GeoM.Translate(screenCenter.X, screenCenter.Y)

	if u.Faction == entity.FactionEnemy {
		hullOp.ColorScale.Scale(1.2, 0.6, 0.6, 1)
	}

	screen.DrawImage(hullSprite, hullOp)

	// Draw gun (turret) on top
	// The turret base (rotation point) is at pixel (47, 160) in the gun sprite
	// Gun pivot needs to be offset toward front of hull (negative Y in hull space before rotation)
	pivotX := 47.0
	pivotY := 160.0
	turretOffsetY := -30.0 // Base offset toward front of tank
	if def != nil && def.TurretOffsetY != 0 {
		turretOffsetY += def.TurretOffsetY
	}
	turretOffsetY *= spriteScale // Apply scale

	gunOp := &ebiten.DrawImageOptions{}
	// Move so pivot point is at origin, then rotate
	gunOp.GeoM.Translate(-pivotX, -pivotY)
	gunOp.GeoM.Rotate(u.TurretAngle + math.Pi/2)
	gunOp.GeoM.Scale(totalScale, totalScale)
	// Offset in hull's rotated space
	offsetX := turretOffsetY * math.Sin(u.Angle+math.Pi/2) * zoom
	offsetY := -turretOffsetY * math.Cos(u.Angle+math.Pi/2) * zoom
	gunOp.GeoM.Translate(screenCenter.X+offsetX, screenCenter.Y+offsetY)

	if u.Faction == entity.FactionEnemy {
		gunOp.ColorScale.Scale(1.2, 0.6, 0.6, 1)
	}

	screen.DrawImage(gunSprite, gunOp)
}

func (er *EntityRenderer) drawTankFallback(screen *ebiten.Image, u *entity.Unit, screenCenter emath.Vec2, zoom float64) {
	bodySize := u.Size.X * 1.2 * zoom
	halfBody := bodySize / 2
	turretSize := bodySize * 0.5
	halfTurret := turretSize / 2
	barrelLength := bodySize * 0.7
	barrelWidth := turretSize * 0.25

	var bodyColor, turretColor, barrelColor color.RGBA
	if u.Faction == entity.FactionEnemy {
		bodyColor = color.RGBA{140, 60, 60, 255}
		turretColor = color.RGBA{180, 80, 80, 255}
		barrelColor = color.RGBA{100, 40, 40, 255}
	} else {
		bodyColor = color.RGBA{60, 100, 60, 255}
		turretColor = color.RGBA{80, 140, 80, 255}
		barrelColor = color.RGBA{40, 70, 40, 255}
	}

	er.drawRotatedRect(screen, screenCenter, bodySize, bodySize, u.Angle, bodyColor)
	er.drawRotatedRectOutline(screen, screenCenter, bodySize, bodySize, u.Angle, color.RGBA{30, 30, 30, 255})

	barrelStartX := screenCenter.X + math.Cos(u.TurretAngle)*halfTurret*0.3
	barrelStartY := screenCenter.Y + math.Sin(u.TurretAngle)*halfTurret*0.3
	barrelEndX := screenCenter.X + math.Cos(u.TurretAngle)*barrelLength
	barrelEndY := screenCenter.Y + math.Sin(u.TurretAngle)*barrelLength

	vector.StrokeLine(screen,
		float32(barrelStartX), float32(barrelStartY),
		float32(barrelEndX), float32(barrelEndY),
		float32(barrelWidth), barrelColor, false)

	er.drawRotatedRect(screen, screenCenter, turretSize, turretSize, u.TurretAngle, turretColor)
	er.drawRotatedRectOutline(screen, screenCenter, turretSize, turretSize, u.TurretAngle, color.RGBA{30, 30, 30, 255})

	frontX := screenCenter.X + math.Cos(u.Angle)*halfBody*0.6
	frontY := screenCenter.Y + math.Sin(u.Angle)*halfBody*0.6
	vector.DrawFilledCircle(screen, float32(frontX), float32(frontY), float32(2*zoom), color.RGBA{200, 200, 200, 200}, false)
}

func (er *EntityRenderer) drawRotatedRect(screen *ebiten.Image, center emath.Vec2, width, height, angle float64, c color.RGBA) {
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

	vs := []ebiten.Vertex{
		{DstX: float32(rotated[0].X), DstY: float32(rotated[0].Y), ColorR: float32(c.R) / 255, ColorG: float32(c.G) / 255, ColorB: float32(c.B) / 255, ColorA: float32(c.A) / 255},
		{DstX: float32(rotated[1].X), DstY: float32(rotated[1].Y), ColorR: float32(c.R) / 255, ColorG: float32(c.G) / 255, ColorB: float32(c.B) / 255, ColorA: float32(c.A) / 255},
		{DstX: float32(rotated[2].X), DstY: float32(rotated[2].Y), ColorR: float32(c.R) / 255, ColorG: float32(c.G) / 255, ColorB: float32(c.B) / 255, ColorA: float32(c.A) / 255},
		{DstX: float32(rotated[3].X), DstY: float32(rotated[3].Y), ColorR: float32(c.R) / 255, ColorG: float32(c.G) / 255, ColorB: float32(c.B) / 255, ColorA: float32(c.A) / 255},
	}
	indices := []uint16{0, 1, 2, 0, 2, 3}

	screen.DrawTriangles(vs, indices, er.getWhitePixel(), &ebiten.DrawTrianglesOptions{})
}

func (er *EntityRenderer) drawRotatedRectOutline(screen *ebiten.Image, center emath.Vec2, width, height, angle float64, c color.RGBA) {
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

	for i := range 4 {
		next := (i + 1) % 4
		vector.StrokeLine(screen,
			float32(rotated[i].X), float32(rotated[i].Y),
			float32(rotated[next].X), float32(rotated[next].Y),
			1, c, false)
	}
}

func (er *EntityRenderer) getWhitePixel() *ebiten.Image {
	if er.whitePixel == nil {
		er.whitePixel = ebiten.NewImage(1, 1)
		er.whitePixel.Fill(color.White)
	}
	return er.whitePixel
}

func (er *EntityRenderer) DrawBuilding(screen *ebiten.Image, b *entity.Building, screenPos emath.Vec2, zoom float64) {
	scaledSize := b.Size.Mul(zoom)
	screenBounds := emath.Rect{Pos: screenPos, Size: scaledSize}
	screenCenter := emath.Vec2{X: screenPos.X + scaledSize.X/2, Y: screenPos.Y + scaledSize.Y/2}

	if !b.Completed {
		er.drawBuildingConstruction(screen, b, screenPos, scaledSize, zoom)
		return
	}

	if b.Def != nil && b.Def.SpritePath != "" {
		sprite := er.assets.GetSprite(b.Def.SpritePath)
		if sprite != nil {
			er.drawBuildingSprite(screen, sprite, b, screenCenter, zoom)
			return
		}
	}

	er.renderer.DrawRect(screen, screenBounds, b.Color)
}

func (er *EntityRenderer) drawBuildingSprite(screen *ebiten.Image, sprite *ebiten.Image, b *entity.Building, screenCenter emath.Vec2, zoom float64) {
	def := b.Def

	spriteSheetW := float64(sprite.Bounds().Dx())
	spriteSheetH := float64(sprite.Bounds().Dy())

	// Determine frame dimensions
	frameW := spriteSheetW
	frameH := spriteSheetH

	// Check if this is an animated sprite (has frame height specified)
	numFrames := 1
	if def.SpriteHeight > 0 {
		frameH = def.SpriteHeight
		numFrames = int(spriteSheetH / frameH)
	}
	if def.SpriteWidth > 0 {
		frameW = def.SpriteWidth
	}

	// Calculate target size using the building's actual dimensions
	targetW := def.GetWidth() * zoom
	targetH := def.GetHeight() * zoom

	scaleX := targetW / frameW
	scaleY := targetH / frameH

	// Get the correct frame from the sprite sheet
	var frameSprite *ebiten.Image
	if numFrames > 1 {
		frameIndex := b.AnimationFrame % numFrames
		frameY := frameIndex * int(frameH)
		frameRect := image.Rect(0, frameY, int(frameW), frameY+int(frameH))
		frameSprite = sprite.SubImage(frameRect).(*ebiten.Image)
	} else {
		frameSprite = sprite
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-frameW/2, -frameH/2)
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(screenCenter.X, screenCenter.Y)

	if b.Faction == entity.FactionEnemy {
		op.ColorScale.Scale(1.2, 0.6, 0.6, 1)
	}

	screen.DrawImage(frameSprite, op)
}

func (er *EntityRenderer) drawBuildingConstruction(screen *ebiten.Image, b *entity.Building, screenPos emath.Vec2, scaledSize emath.Vec2, zoom float64) {
	screenBounds := emath.Rect{Pos: screenPos, Size: scaledSize}

	rgba := b.Color.(color.RGBA)
	constructionColor := color.RGBA{
		R: uint8(float64(rgba.R) * 0.5),
		G: uint8(float64(rgba.G) * 0.5),
		B: uint8(float64(rgba.B) * 0.5),
		A: 180,
	}
	er.renderer.DrawRect(screen, screenBounds, constructionColor)

	scaffoldColor := color.RGBA{100, 80, 50, 150}
	lineSpacing := 10.0 * zoom
	for i := 0.0; i < scaledSize.X; i += lineSpacing {
		er.renderer.DrawLine(screen,
			emath.Vec2{X: screenPos.X + i, Y: screenPos.Y},
			emath.Vec2{X: screenPos.X, Y: screenPos.Y + i},
			1, scaffoldColor)
	}
}

func (er *EntityRenderer) DrawBuildingBorder(screen *ebiten.Image, b *entity.Building, screenPos emath.Vec2, zoom float64) {
	scaledSize := b.Size.Mul(zoom)
	screenBounds := emath.Rect{Pos: screenPos, Size: scaledSize}

	if b.Def != nil && b.Def.SpritePath != "" && b.Completed {
		sprite := er.assets.GetSprite(b.Def.SpritePath)
		if sprite != nil {
			return
		}
	}

	borderColor := color.RGBA{60, 60, 60, 255}
	if !b.Completed {
		borderColor = color.RGBA{200, 150, 50, 255}
	}
	er.renderer.DrawRectOutline(screen, screenBounds, 2, borderColor)
}

func (er *EntityRenderer) NeedsBorder(b *entity.Building) bool {
	if b.Def != nil && b.Def.SpritePath != "" && b.Completed {
		sprite := er.assets.GetSprite(b.Def.SpritePath)
		if sprite != nil {
			return false
		}
	}
	return true
}
