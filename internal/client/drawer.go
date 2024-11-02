package client

import (
	"github.com/gdamore/tcell/v2"
)

type Drawer struct {
	screen     tcell.Screen
	actions    chan<- Action
	components []Component
	focusedTab *TabIndex
}

type TabIndex int

const (
	TabUsers TabIndex = iota
	TabConversations
	// TabMessages
	TabCount // Keep this last
)

func NewDrawer(screen tcell.Screen, actions chan<- Action) *Drawer {
	width, height := screen.Size()
	leftSideWidth := 30
	return &Drawer{
		screen: screen,
		components: []Component{
			NewHeader(NewBaseComponent(
				screen,
				&Rect{X: 0, Y: 0, Width: width, Height: 1},
				actions,
			)),
			NewSeparatorLine(NewBaseComponent(
				screen,
				&Rect{X: 0, Y: 1, Width: width, Height: 1},
				actions,
			)),
			NewUsersTab(NewBaseComponent(
				screen,
				&Rect{X: 0, Y: 2, Width: leftSideWidth, Height: 10},
				actions,
			)),
			NewSeparatorLine(NewBaseComponent(
				screen,
				&Rect{X: 0, Y: 12, Width: leftSideWidth, Height: 1},
				actions,
			)),
			NewConversationsTab(NewBaseComponent(
				screen,
				&Rect{X: 0, Y: 13, Width: leftSideWidth, Height: height - 13},
				actions,
			)),
			NewSeparatorLine(NewBaseComponent(
				screen,
				&Rect{X: leftSideWidth, Y: 2, Width: 1, Height: height - 2},
				actions,
			)),
		},
	}
}

func (d *Drawer) OnEvent(event any) {
	switch event := event.(type) {
	case *tcell.EventKey:
		switch event.Key() {
		case tcell.KeyTab:
			// Handle tab switching
			if d.focusedTab == nil {
				d.focusedTab = new(TabIndex)
				*d.focusedTab = TabUsers
			} else {
				*d.focusedTab = (*d.focusedTab + 1) % TabCount
			}
		case tcell.KeyEscape:
			d.focusedTab = nil
		}
	}

	focusIndex := 0
	for _, component := range d.components {
		if component.CanFocus() {
			component.Focus(d.focusedTab != nil && focusIndex == int(*d.focusedTab))
			focusIndex++
		}
		component.OnEvent(event)
	}
	d.Draw()
}

func (d *Drawer) Draw() {
	d.screen.Clear()
	for _, component := range d.components {
		component.Render()
	}
	d.screen.Show()
}
