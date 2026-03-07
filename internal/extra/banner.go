package extra

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/bdk38/HellPot/internal/config"
)

// ASCII art for HellPot banner
const hellpotArt = `   ▄█    █▄       ▄████████  ▄█        ▄█          ▄███████▄  ▄██████▄      ███
  ███    ███     ███    ███ ███       ███         ███    ███ ███    ███ ▀█████████▄
  ███    ███     ███    █▀  ███       ███         ███    ███ ███    ███    ▀███▀▀██
 ▄███▄▄▄▄███▄▄  ▄███▄▄▄     ███       ███         ███    ███ ███    ███     ███   ▀
▀▀███▀▀▀▀███▀  ▀▀███▀▀▀     ███       ███       ▀█████████▀  ███    ███     ███
  ███    ███     ███    █▄  ███       ███         ███        ███    ███     ███
  ███    ███     ███    ███ ███▌    ▄ ███▌    ▄   ███        ███    ███     ███
  ███    █▀      ██████████ █████▄▄██ █████▄▄██  ▄████▀       ▀██████▀     ▄████▀
                            ▀         ▀`

// ANSI color codes for the colorful effect
var colors = []string{
	"38;5;33",  // blue
	"38;5;39",  // bright blue
	"38;5;51",  // cyan
	"38;5;49",  // bright cyan
	"38;5;48",  // teal
	"38;5;84",  // green
	"38;5;118", // lime
	"38;5;154", // yellow-green
	"38;5;190", // yellow
	"38;5;226", // bright yellow
	"38;5;214", // orange
	"38;5;208", // bright orange
	"38;5;203", // red-orange
	"38;5;198", // pink
	"38;5;199", // magenta
	"38;5;171", // purple
	"38;5;141", // violet
}

// randomInt returns a random integer for color selection
func randomInt() int {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return 0
	}
	return int(binary.LittleEndian.Uint32(b))
}

// colorize adds random ANSI colors to the ASCII art
func colorize(art string) string {
	var result strings.Builder
	for _, char := range art {
		if char != ' ' && char != '\n' {
			color := colors[randomInt()%len(colors)]
			result.WriteString(fmt.Sprintf("\033[%sm%c\033[0m", color, char))
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// Banner prints the HellPot banner with version and URLs
func Banner() {
	// Windows and --nocolor mode: simple text output
	if runtime.GOOS == "windows" || config.NoColor {
		fmt.Fprintf(os.Stdout, "%s %s\n\n", config.Title, config.Version)
		return
	}

	// Colorful banner for Unix-like systems
	fmt.Println()
	fmt.Println(colorize(hellpotArt))
	fmt.Println()

	// Version
	fmt.Printf("                                    \033[38;5;46mv%s\033[0m\n\n", config.Version)

	// URLs with attribution
	fmt.Println("\033[38;5;240m[ \033[38;5;33mgithub.com/yunginnanet/HellPot\033[38;5;240m ]\033[0m (upstream)")
	fmt.Println("\033[38;5;240m[ \033[38;5;39mgithub.com/bdk38/HellPot\033[38;5;240m ]\033[0m (community fork)")
	fmt.Println()
}

