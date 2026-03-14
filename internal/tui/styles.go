package tui

import (
	"github.com/r3g/recurva/internal/tui/shared"
)

// Re-export styles from shared for backward compat
var (
	ColorPrimary   = shared.ColorPrimary
	ColorSecondary = shared.ColorSecondary
	ColorMuted     = shared.ColorMuted
	ColorAgain     = shared.ColorAgain
	ColorHard      = shared.ColorHard
	ColorGood      = shared.ColorGood
	ColorEasy      = shared.ColorEasy

	StyleTitle    = shared.StyleTitle
	StyleSubtle   = shared.StyleSubtle
	StyleCard     = shared.StyleCard
	StyleFront    = shared.StyleFront
	StyleBack     = shared.StyleBack
	StyleProgress = shared.StyleProgress
	StyleAgain    = shared.StyleAgain
	StyleHard     = shared.StyleHard
	StyleGood     = shared.StyleGood
	StyleEasy     = shared.StyleEasy
	StyleHelp     = shared.StyleHelp
	StyleSelected = shared.StyleSelected
)
