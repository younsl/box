package main

import (
	_ "time/tzdata" // Embed timezone data

	"github.com/younsl/box/tools/ol/internal/cmd"
)

func main() {
	cmd.Execute()
}
