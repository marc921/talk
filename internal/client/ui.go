package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/marc921/talk/internal/types/openapi"
)

type UI struct {
	screen        tcell.Screen
	cursor        *Position
	mode          Mode
	buffer        string
	onEnter       func(string) error
	currentUser   *User
	currentErr    error
	storage       *Storage
	openapiClient *openapi.ClientWithResponses
}

type Mode string

const (
	ModeNormal Mode = "Normal"
	ModeInsert Mode = "Insert"
)

type Position struct {
	X int
	Y int
}

func (p *Position) Reset() {
	p.X = 0
	p.Y = 0
}

func (p *Position) Newline() {
	p.X = 0
	p.Y++
}

type Side string

const (
	Left  Side = "left"
	Right Side = "right"
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

	return &UI{
		screen:        screen,
		cursor:        &Position{X: 0, Y: 0},
		mode:          ModeNormal,
		storage:       storage,
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
				case tcell.KeyCtrlC:
					return nil
				case tcell.KeyEscape:
					u.mode = ModeNormal
				case tcell.KeyEnter:
					if u.onEnter != nil {
						u.currentErr = u.onEnter(u.buffer)
						u.buffer = ""
						u.onEnter = nil
						u.mode = ModeNormal
					}
				case tcell.KeyRune:
					if u.mode == ModeInsert {
						u.buffer += string(ev.Rune())
						continue
					}
					switch ev.Rune() {
					case 'q':
						return nil
					case 'u':
						u.NewUser(ctx)
					}
				case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
					//
				}
			}
		}
	}
}

func (u *UI) NewUser(ctx context.Context) {
	u.mode = ModeInsert
	u.onEnter = func(username string) error {
		// Create user
		privKey, err := u.storage.GetOrCreatePrivateKey(username)
		if err != nil {
			return fmt.Errorf("GetOrCreatePrivateKey: %w", err)
		}
		user := NewUser(
			openapi.Username(username),
			privKey,
			NewClient(u.openapiClient, openapi.Username(username)),
		)
		// Register user on distant server, fail if already exists
		err = user.Register(ctx)
		if err != nil {
			return fmt.Errorf("user.Register: %w", err)
		}

		// Switch to user screen
		u.currentUser = user
		return nil
	}
}

func (u *UI) Draw() {
	u.screen.Clear()
	u.PrintHeader()
	u.PrintUser()
	u.PrintActions()
	u.screen.Show()
}

// PrintText prints text on the screen, after the cursor (default) or before.
func (u *UI) PrintText(text string) {
	for _, ch := range text {
		u.screen.SetContent(u.cursor.X, u.cursor.Y, ch, nil, tcell.StyleDefault)
		u.cursor.X++
	}
}

func (u *UI) PrintTextBefore(text string) {
	u.cursor.X -= len(text)
	u.PrintText(text)
}

func (u *UI) PrintHeader() {
	u.cursor.Reset()
	width, _ := u.screen.Size()

	// Print the left part of the header
	u.PrintText(" Talk | Mode: " + string(u.mode) + " | ")

	// Print the right part of the header
	u.cursor.X = width
	u.PrintTextBefore("| [Q]uit ")

	// Print the line separator
	u.cursor.Newline()
	u.PrintText(strings.Repeat("-", width))
	u.cursor.Newline()
	if u.currentErr != nil {
		u.PrintText("ERROR: " + u.currentErr.Error())
		u.cursor.Newline()
		u.PrintText(strings.Repeat("-", width))
		u.cursor.Newline()
	}
}

func (u *UI) PrintUser() {
	if u.currentUser == nil {
		return
	}
	u.cursor.Newline()
	u.PrintText("User: " + u.currentUser.name)
	u.cursor.Newline()
}

func (u *UI) PrintActions() {
	u.PrintText("Actions:")
	u.cursor.Newline()
	u.PrintText(" - Create new [U]ser")
	if u.buffer != "" {
		u.cursor.Newline()
		u.PrintText("Username: " + u.buffer)
	}
}
