package client

import "github.com/gdamore/tcell/v2"

type Component interface {
	Render()
	OnEvent(any)
	CanFocus() bool
	Focus(bool)
}

// Rect constrains a component from (X, Y) to (X+Width-1, Y+Height-1) included.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

type BaseComponent struct {
	screen     tcell.Screen
	rect       *Rect
	drawCursor *Position
	actions    chan<- Action
}

func NewBaseComponent(screen tcell.Screen, rect *Rect, actions chan<- Action) *BaseComponent {
	return &BaseComponent{
		screen:     screen,
		rect:       rect,
		drawCursor: &Position{X: rect.X, Y: rect.Y},
		actions:    actions,
	}
}

func (b *BaseComponent) PrintTextBefore(text string) {
	b.drawCursor.X -= len(text)
	b.PrintText(text)
}

// PrintText prints text on the screen, after the cursor (default) or before.
func (b *BaseComponent) PrintText(text string) {
	b.PrintTextStyle(text, tcell.StyleDefault)
}

func (b *BaseComponent) PrintTextStyle(text string, style tcell.Style) {
	for _, ch := range text {
		b.screen.SetContent(b.drawCursor.X, b.drawCursor.Y, ch, nil, style)
		b.drawCursor.X++
		if b.drawCursor.X >= b.rect.X+b.rect.Width {
			return
		}
	}
}

type Position struct {
	X int
	Y int
}

func (p *Position) Reset(rect *Rect) {
	p.X = rect.X
	p.Y = rect.Y
}

func (p *Position) Newline() {
	p.X = 0
	p.Y++
}

func (p *Position) Move(x, y int) {
	p.X = x
	p.Y = y
}
