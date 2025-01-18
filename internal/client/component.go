package client

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

type Component interface {
	Render()
	OnEvent(any)
	CanFocus() bool
	Focus(bool)
}

// Rect constrains a component from (X, Y) to (X+Width-1, Y+Height-1) included.
type Rect struct {
	Left   int
	Top    int
	Width  int
	Height int
}

func (r *Rect) Shrink(n int) *Rect {
	return &Rect{
		Left:   r.Left + n,
		Top:    r.Top + n,
		Width:  r.Width - 2*n,
		Height: r.Height - 2*n,
	}
}

type BaseComponent struct {
	screen     tcell.Screen
	bounds     *Rect
	drawCursor *Cursor
}

func NewBaseComponent(screen tcell.Screen, bounds *Rect) *BaseComponent {
	return &BaseComponent{
		screen:     screen,
		bounds:     bounds,
		drawCursor: NewCursor(bounds),
	}
}

func (b *BaseComponent) SetBounds(bounds *Rect) {
	b.bounds = bounds
	b.drawCursor.bounds = bounds
}

func (b *BaseComponent) PrintTextRightAlign(text string) {
	b.drawCursor.X = b.bounds.Left + b.bounds.Width - len(text)
	b.PrintText(text)
}

// PrintText prints text on the screen, after the cursor (default) or before.
func (b *BaseComponent) PrintText(text string) {
	b.PrintTextStyle(text, tcell.StyleDefault)
}

func (b *BaseComponent) PrintTextStyle(text string, style tcell.Style) {
	if b.drawCursor.IsOutOfBounds() {
		return
	}
	for _, ch := range text {
		b.screen.SetContent(b.drawCursor.X, b.drawCursor.Y, ch, nil, style)
		b.drawCursor.X++
		if b.drawCursor.X >= b.bounds.Left+b.bounds.Width {
			return
		}
	}
}

func (b *BaseComponent) PrintTextCentered(text string, style tcell.Style) {
	b.drawCursor.X = b.bounds.Left + (b.bounds.Width-len(text))/2
	b.PrintTextStyle(text, style)
}

func (b *BaseComponent) PrintBox(bounds *Rect, style tcell.Style) {
	b.drawCursor.MoveTo(bounds.Left, bounds.Top)
	b.PrintTextStyle("┌", style)
	b.PrintTextStyle(strings.Repeat("─", bounds.Width-2), style)
	b.PrintTextStyle("┐", style)
	b.drawCursor.MoveTo(bounds.Left, bounds.Top+bounds.Height-1)
	b.PrintTextStyle("└", style)
	b.PrintTextStyle(strings.Repeat("─", bounds.Width-2), style)
	b.PrintTextStyle("┘", style)
	for y := bounds.Top + 1; y < bounds.Top+bounds.Height-1; y++ {
		b.drawCursor.MoveTo(bounds.Left, y)
		b.PrintTextStyle("│", style)
		b.drawCursor.MoveTo(bounds.Left+bounds.Width-1, y)
		b.PrintTextStyle("│", style)
	}
}

type Cursor struct {
	bounds *Rect
	X      int
	Y      int
}

func NewCursor(bounds *Rect) *Cursor {
	return &Cursor{
		bounds: bounds,
		X:      bounds.Left,
		Y:      bounds.Top,
	}
}

func (p *Cursor) Reset() {
	p.X = p.bounds.Left
	p.Y = p.bounds.Top
}

func (p *Cursor) Newline() {
	p.X = p.bounds.Left
	p.Y++
	// TODO: Handle scrolling, do not write outside of bounds
}

func (p *Cursor) MoveTo(x, y int) {
	p.X = x
	p.Y = y
}

func (p *Cursor) IsOutOfBounds() bool {
	return p.X < p.bounds.Left ||
		p.X >= p.bounds.Left+p.bounds.Width ||
		p.Y < p.bounds.Top ||
		p.Y >= p.bounds.Top+p.bounds.Height
}
