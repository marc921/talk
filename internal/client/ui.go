package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/marc921/talk/internal/types/openapi"
)

type UI struct {
	drawer        *Drawer
	actions       <-chan Action
	storage       *Storage
	openapiClient *openapi.ClientWithResponses
}

type Mode string

const (
	ModeNormal Mode = "Normal"
	ModeInsert Mode = "Insert"
)

func NewUI(config *Config) (*UI, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("tcell.NewScreen: %w", err)
	}
	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("screen.Init: %w", err)
	}

	storage := NewStorage(config.HomeDir)

	openapiClient, err := openapi.NewClientWithResponses(
		config.Server.URL,
	)
	if err != nil {
		return nil, fmt.Errorf("openapi.NewClientWithResponses: %w", err)
	}

	actions := make(chan Action, 100)

	return &UI{
		drawer:        NewDrawer(screen, actions),
		actions:       actions,
		storage:       storage,
		openapiClient: openapiClient,
	}, nil
}

func (u *UI) Quit() {
	u.drawer.screen.Fini()
}

func (u *UI) Run(ctx context.Context) error {
	for {
		u.drawer.Draw()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case action := <-u.actions:
			err := action.Do(ctx, u)
			if err != nil {
				if errors.Is(err, ErrQuit) {
					return nil
				}
				return fmt.Errorf("action.Do(%s): %w", action.String(), err)
			}
		default:
			ev := u.drawer.screen.PollEvent()
			if ev == nil {
				// PollEvent _supposedly_ returns nil if screen is finalized
				return nil
			}
			// Fan out event to drawer
			u.drawer.OnEvent(ev)
			// Handle event
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyCtrlC:
					return nil
				// case tcell.KeyEscape:
				// 	u.mode = ModeNormal
				// 	u.page = PageHome
				// case tcell.KeyEnter:
				// 	if u.onEnter != nil {
				// 		u.currentErr = u.onEnter(u.buffer)
				// 		u.buffer = ""
				// 		u.onEnter = nil
				// 		u.mode = ModeNormal
				// 	}
				case tcell.KeyTab:

				case tcell.KeyRune:
					// if u.mode == ModeInsert {
					// 	u.buffer += string(ev.Rune())
					// 	continue
					// }
					switch ev.Rune() {
					// case 'q':
					// 	return nil
					}
				case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
					//
				}
			}
		}
	}
}

func (u *UI) CreateUser(ctx context.Context, username openapi.Username) (*User, error) {
	// Create user
	user, err := NewUser(
		username,
		u.storage,
		u.openapiClient,
	)
	if err != nil {
		return nil, fmt.Errorf("NewUser: %w", err)
	}
	// Register user on distant server, fail if already exists
	err = user.Register(ctx)
	if err != nil {
		return nil, fmt.Errorf("user.Register: %w", err)
	}
	return user, nil
}
