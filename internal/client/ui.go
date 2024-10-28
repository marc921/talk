package client

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/marc921/talk/internal/types/openapi"
)

type UI struct {
	screen        tcell.Screen
	openapiClient *openapi.ClientWithResponses
}

func NewUI(serverURL string) (*UI, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("tcell.NewScreen: %w", err)
	}
	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("screen.Init: %w", err)
	}

	openapiClient, err := openapi.NewClientWithResponses(
		serverURL,
	)
	if err != nil {
		return nil, fmt.Errorf("openapi.NewClientWithResponses: %w", err)
	}

	return &UI{
		screen:        screen,
		openapiClient: openapiClient,
	}, nil
}

func (u *UI) Close() {
	u.screen.Fini()
}

func (u *UI) Run(ctx context.Context) error {
	for {
		u.Draw()
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			switch ev := u.screen.PollEvent(); ev := ev.(type) {
			case nil:
				return nil
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyCtrlC:
					return nil
				case tcell.KeyRune:
					switch ev.Rune() {
					case 'q':
						return nil
					}
				case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
					//
				}
			}
		}
	}
}

func (u *UI) Draw() {
	u.screen.Clear()
	u.PrintHeader()
	u.screen.Show()
}

func (u *UI) PrintHeader() {
	width, _ := u.screen.Size()
	headerLeft := " Talk |"
	headerRight := "| [Q]uit "

	// Print the left part of the header
	for i, ch := range headerLeft {
		u.screen.SetContent(i+1, 0, ch, nil, tcell.StyleDefault)
	}

	// Print the right part of the header
	rightStart := width - len(headerRight)
	for i, ch := range headerRight {
		u.screen.SetContent(rightStart+i, 0, ch, nil, tcell.StyleDefault)
	}

	// Print the line separator
	for i := 0; i < width; i++ {
		u.screen.SetContent(i, 1, '-', nil, tcell.StyleDefault)
	}
}
