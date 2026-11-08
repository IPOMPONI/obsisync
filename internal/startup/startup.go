package startup

import (
	"fmt"
	"obsisync/internal/config"
	"strings"
)

const (
	reset = "\033[0m"
	cyan  = "\033[36m"
	white = "\033[37m"
)

func buildBanner() string {
	var b strings.Builder
	b.Grow(720)

	left := []string{
		"      __                     ",
		"     /\\ \\              __    ",
		"  ___\\ \\ \\____    ____/\\_\\   ",
		" / __`\\ \\ '__`\\  /',__\\/\\ \\  ",
		"/\\ \\L\\ \\ \\ \\L\\ \\/\\__, `\\ \\ \\ ",
		"\\ \\____/\\ \\_,__/\\/\\____/\\ \\_\\",
		" \\/___/  \\/___/  \\/___/  \\/_/",
	}

	right := []string{
		"                                ",
		"                                ",
		"  ____  __  __    ___     ___   ",
		" /',__\\/\\ \\/\\ \\ /' _ `\\  /'___\\ ",
		"/\\__, `\\ \\ \\_\\ \\/\\ \\/\\ \\/\\ \\__/ ",
		"\\/\\____/\\`____  \\ \\_\\ \\_\\ \\____\\",
		" \\/___/  `/___/> \\/_/\\/_/\\/____/",
		"                                         /\\___/              ",
		"                                         \\/__/               ",
	}

	for i := range right {
		b.WriteString(white)
		if i < 7 {
			b.WriteString(left[i])
		}
		b.WriteString(reset)
		b.WriteString(cyan)
		b.WriteString(right[i])
		b.WriteString(reset)
		b.WriteByte('\n')
	}

	return b.String()
}

func PrintBanner() {
	banner := buildBanner()
	fmt.Printf("%s", banner)
	printLine()
}

func PrintInfo(cfg config.Config) {
	fmt.Printf("Server will use: %s:%d\n\nVaults:\n", cfg.Server.Host, cfg.Server.Port)
	for vault := range cfg.Vaults {
		fmt.Printf("\t%s\n", vault)
	}
	printLine()
}

func printLine() {
	for range 61 {
		fmt.Print("_")
	}
	fmt.Printf("\n\n")
}
