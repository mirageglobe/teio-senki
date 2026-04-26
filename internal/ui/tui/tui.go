// Package tui owns the terminal UI frontend. It is a dumb view: it calls engine
// methods and renders text. It does NOT contain game logic.
package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mirageglobe/teio-senki/internal/engine/sovereign"
	"github.com/mirageglobe/teio-senki/internal/models"
)

// Run starts the terminal game loop.
func Run(eng *sovereign.Engine) error {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("帝王战纪：三国录 — Sovereign Record [TUI v0]")
	fmt.Println("commands: start, settle, build <city> ag|com, quit")

	for {
		state := eng.GetState()
		printStatus(state)
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)

		switch parts[0] {
		case "start":
			eng.StartTurn()
			printLastLogs(eng.GetState(), 4)

		case "settle":
			entries := eng.SettleTurn()
			for _, e := range entries {
				fmt.Printf("  [%s] %s\n", e.Type, e.Description)
			}

		case "build":
			if len(parts) < 3 {
				fmt.Println("  usage: build <city> ag|com")
				continue
			}
			city := parts[1]
			cmdType := map[string]string{"ag": "BUILD_AG", "com": "BUILD_COM"}[parts[2]]
			if cmdType == "" {
				fmt.Println("  unknown build type, use ag or com")
				continue
			}
			ok := eng.QueueCommand(models.Command{
				Type:   cmdType,
				Params: map[string]string{"city": city},
				Cost:   2,
			})
			if ok {
				fmt.Printf("  queued %s on %s\n", cmdType, city)
			} else {
				fmt.Println("  not enough CP")
			}

		case "quit", "q":
			fmt.Println("goodbye.")
			return nil

		default:
			fmt.Printf("  unknown command: %s\n", parts[0])
		}
	}
	return scanner.Err()
}

func printStatus(s models.GameState) {
	fmt.Printf("\n[%d.%d]  grain:%d  gold:%d  CP:%d\n",
		s.Year, s.Month, s.Resources.Grain, s.Resources.Gold, s.AvailableCP)
}

func printLastLogs(s models.GameState, n int) {
	start := max(0, len(s.Logs)-n)
	for _, e := range s.Logs[start:] {
		fmt.Printf("  [%s] %s\n", e.Type, e.Description)
	}
}
