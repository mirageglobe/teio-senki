package tui

type scenario struct {
	name     string
	epoch    string
	year     int
	desc     string
	unlocked bool
}

var allScenarios = []scenario{
	{
		name:     "Yellow Turban Rebellion",
		epoch:    "AD 184",
		year:     184,
		desc:     "The fall of the Han; Zhang Jiao's Yellow Turban uprising sweeps the empire.",
		unlocked: false,
	},
	{
		name:     "Dong Zhuo's Rise",
		epoch:    "AD 189",
		year:     189,
		desc:     "Chaos grips the capital. Dong Zhuo seizes the emperor. The coalition of lords prepares to strike.",
		unlocked: true,
	},
	{
		name:     "Rivalry of Lords",
		epoch:    "AD 194",
		year:     194,
		desc:     "Dong Zhuo is dead. Independent lords vie for supremacy across a fractured empire.",
		unlocked: false,
	},
	{
		name:     "Battle of Guandu",
		epoch:    "AD 200",
		year:     200,
		desc:     "Cao Cao and Yuan Shao clash in the decisive battle for the north.",
		unlocked: false,
	},
	{
		name:     "Three Kingdoms",
		epoch:    "AD 220",
		year:     220,
		desc:     "Wei, Shu, and Wu stand divided. The age of the Three Kingdoms begins.",
		unlocked: false,
	},
}
