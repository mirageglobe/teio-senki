-- aggressive.lua: AI behaviour profile for aggressive sovereigns (e.g. Dong Zhuo, Lu Bu).
-- Function ai_decide(state) returns a list of command tables.
-- 'state' is a read-only GameState table pushed by bridge.PushGameState.

function ai_decide(state)
  local commands = {}

  -- aggressive profile: prioritise BUILD_AG in the first city if CP allows
  if state.available_cp >= 2 then
    table.insert(commands, {
      type = "BUILD_AG",
      city = "Luoyang",
      cost = 2,
    })
  end

  return commands
end
