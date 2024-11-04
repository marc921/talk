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
	TabMessages
	
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
				&Rect{Left: 0, Top: 0, Width: width, Height: 1},
				actions,
			)),
			NewSeparatorLine(NewBaseComponent(
				screen,
				&Rect{Left: 0, Top: 1, Width: width, Height: 1},
				actions,
			)),
			NewUsersTab(NewBaseComponent(
				screen,
				&Rect{Left: 0, Top: 2, Width: leftSideWidth, Height: 10},
				actions,
			)),
			NewSeparatorLine(NewBaseComponent(
				screen,
				&Rect{Left: 0, Top: 12, Width: leftSideWidth, Height: 1},
				actions,
			)),
			NewConversationsTab(NewBaseComponent(
				screen,
				&Rect{Left: 0, Top: 13, Width: leftSideWidth, Height: height - 13},
				actions,
			)),
			NewSeparatorLine(NewBaseComponent(
				screen,
				&Rect{Left: leftSideWidth, Top: 2, Width: 1, Height: height - 2},
				actions,
			)),
			NewMessagesTab(NewBaseComponent(
				screen,
				&Rect{Left: leftSideWidth + 1, Top: 2, Width: width - leftSideWidth - 1, Height: height - 2},
				actions,
			)),
			NewErrorModal(NewBaseComponent(
				screen,
				&Rect{Left: 0, Top: 0, Width: width, Height: height},
				actions,
			)),
		},
	}
}

func (d *Drawer) OnEvent(event any) {
	switch event := event.(type) {
	case *EventSwitchTab:
		d.focusedTab = &event.tabIndex
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
