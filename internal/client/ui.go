package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/marc921/talk/internal/types/openapi"
)

type UI struct {
	drawer        *Drawer
	actions       chan Action
	db            *sql.DB
	openapiClient *openapi.ClientWithResponses
	config        *Config
}

type Mode string

const (
	ModeNormal Mode = "Normal"
	ModeInsert Mode = "Insert"
)

var UISingleton *UI

func InitUI(
	config *Config,
	openapiClient *openapi.ClientWithResponses,
	db *sql.DB,
) error {
	UISingleton = &UI{
		actions:       make(chan Action, 100),
		db:            db,
		openapiClient: openapiClient,
		config:        config,
	}

	drawer, err := NewDrawer()
	if err != nil {
		return fmt.Errorf("NewDrawer: %w", err)
	}
	UISingleton.drawer = drawer

	return nil
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
				u.drawer.OnEvent(
					&EventSetError{
						err: fmt.Errorf("action.Do(%s): %w", action.String(), err),
					},
				)
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
				}
			}
		}
	}
}
