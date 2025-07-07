package constant

import "github.com/abhissng/neuron/utils/types"

const (
	RedirectToURL = "Redirecting to URL"
	ResetColor    = "\033[0m"  // Reset color
	RedColor      = "\033[31m" // Red (Error)
	YellowColor   = "\033[33m" // Yellow (Warn)
	GreenColor    = "\033[32m" // Green (Info)
	BlueColor     = "\033[34m" // Blue (Debug)
	CyanColor     = "\033[36m" // Cyan
	MagentaColor  = "\033[35m" // Magenta
	WhiteColor    = "\033[37m" // White
)

// Supported log modes
const (
	INFO  types.LogMode = "info"
	WARN  types.LogMode = "warn"
	ERROR types.LogMode = "error"
	DEBUG types.LogMode = "debug"
	FATAL types.LogMode = "fatal"
)
