// Package loader manages the Lua VM lifecycle, script discovery, and hot-reload.
// Scripts embedded via //go:embed are the fallback when no live file is found.
package loader

import (
	"fmt"
	"os"

	lua "github.com/yuin/gopher-lua"
)

// VM wraps a gopher-lua state and handles script execution.
type VM struct {
	L *lua.LState
}

// New creates a new Lua VM instance.
func New() *VM {
	return &VM{L: lua.NewState()}
}

// Close releases the Lua VM.
func (v *VM) Close() {
	v.L.Close()
}

// ExecFile loads and executes a Lua script file.
func (v *VM) ExecFile(path string) error {
	return v.L.DoFile(path)
}

// ExecString runs a Lua script from a string (used for embedded scripts).
func (v *VM) ExecString(src string) error {
	return v.L.DoString(src)
}

// ExecEmbed runs a Lua script from embedded bytes.
func (v *VM) ExecEmbed(src []byte) error {
	return v.ExecString(string(src))
}

// CallFunc calls a named Lua function with a single table argument already on
// the stack. Returns an error if the function does not exist or panics.
func (v *VM) CallFunc(name string, nArgs, nRet int) error {
	fn := v.L.GetGlobal(name)
	if fn == lua.LNil {
		return fmt.Errorf("lua function %q not found", name)
	}
	return v.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    nRet,
		Protect: true,
	})
}

// WatchAndReload watches a file for changes and re-executes it in the VM.
// Runs in a goroutine; errors are printed to stderr and do not crash the game.
func WatchAndReload(v *VM, path string) {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[vm] watch: %v\n", err)
		return
	}
	lastMod := info.ModTime()

	go func() {
		for {
			current, err := os.Stat(path)
			if err != nil {
				continue
			}
			if current.ModTime().After(lastMod) {
				lastMod = current.ModTime()
				if err := v.ExecFile(path); err != nil {
					fmt.Fprintf(os.Stderr, "[vm] reload %s: %v\n", path, err)
				} else {
					fmt.Printf("[vm] reloaded %s\n", path)
				}
			}
		}
	}()
}
