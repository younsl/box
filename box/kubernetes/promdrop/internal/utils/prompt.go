package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// GetUserConfirmation prompts the user for confirmation (y/n).
func GetUserConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Proceed with YAML generation? (y/n): ")
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" {
			return true
		}
		if input == "n" {
			return false
		}
		fmt.Println("[Error] Please enter 'y' or 'n'.")
	}
}
