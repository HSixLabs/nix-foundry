/*
Package tui provides terminal user interface components for Nix Foundry.
*/
package tui

/*
ANSI color and style codes for terminal output.
These constants provide consistent terminal formatting across the application.

ColorReset: Resets all colors and styles
ColorCyan: Cyan text color
ColorBold: Bold text style
ColorGreen: Green text color
ColorYellow: Yellow text color
ColorRed: Red text color
ColorGrey: Grey text color
*/
const (
	ColorReset  = "\x1b[0m"
	ColorCyan   = "\x1b[36m"
	ColorBold   = "\x1b[1m"
	ColorGreen  = "\x1b[32m"
	ColorYellow = "\x1b[33m"
	ColorRed    = "\x1b[31m"
	ColorGrey   = "\x1b[2m"
)
