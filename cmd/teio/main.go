package main

import (
	"fmt"
	"os"

	"github.com/mirageglobe/teio-senki/internal/engine/ledger"
	"github.com/mirageglobe/teio-senki/internal/ui/tui"
)

func main() {
	l := ledger.New()
	// data dir is relative to working directory; run from repo root via 'make tui'
	if err := l.LoadData("assets/data"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load data: %v\n", err)
		os.Exit(1)
	}

	if err := tui.Run(l); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}
