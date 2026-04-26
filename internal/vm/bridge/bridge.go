// Package bridge adapts Go types to Lua tables and validates Lua output before
// it re-enters the engine. It does NOT execute scripts — that belongs in loader.
package bridge

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/mirageglobe/teio-senki/internal/models"
)

// PushGameState pushes a read-only GameState table onto the Lua stack.
func PushGameState(L *lua.LState, state models.GameState) {
	t := L.NewTable()
	L.SetField(t, "year", lua.LNumber(state.Year))
	L.SetField(t, "month", lua.LNumber(state.Month))
	L.SetField(t, "available_cp", lua.LNumber(state.AvailableCP))

	res := L.NewTable()
	L.SetField(res, "grain", lua.LNumber(state.Resources.Grain))
	L.SetField(res, "gold", lua.LNumber(state.Resources.Gold))
	L.SetField(t, "resources", res)

	L.Push(t)
}

// PullCommands reads a Lua table of command tables from the stack.
// Each command table must have: type (string), city (string), cost (number).
// Invalid entries are silently skipped to prevent Lua from crashing the engine.
func PullCommands(L *lua.LState, t *lua.LTable) []models.Command {
	var cmds []models.Command
	t.ForEach(func(_, v lua.LValue) {
		row, ok := v.(*lua.LTable)
		if !ok {
			return
		}
		cmdType, ok := L.GetField(row, "type").(lua.LString)
		if !ok || string(cmdType) == "" {
			return
		}
		city, _ := L.GetField(row, "city").(lua.LString)
		cost, _ := L.GetField(row, "cost").(lua.LNumber)
		cmds = append(cmds, models.Command{
			Type:   string(cmdType),
			Params: map[string]string{"city": string(city)},
			Cost:   int(cost),
		})
	})
	return cmds
}
