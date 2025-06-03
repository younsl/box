package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/younsl/box/box/tools/cryptopunk-generator/pkg/assets"
	"github.com/younsl/box/box/tools/cryptopunk-generator/pkg/config"
	"github.com/younsl/box/box/tools/cryptopunk-generator/pkg/generator"
)

var race string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cryptopunk-generator",
	Short: "A simple cryptopunk generator",
	Long:  `Generates unique cryptopunk images based on various attributes.`,
	Example: `  # Generate a default set of cryptopunks
  cryptopunk-generator

  # Generate only female punks
  cryptopunk-generator --race female

  # Generate only ape punks
  cryptopunk-generator --race ape`,
	Run: func(cmd *cobra.Command, args []string) {
		// 입력된 race 값이 유효한지 확인
		validRaces := []string{"", "ape", "alien", "female", "male", "zombie"} // 빈 문자열은 전체 생성을 의미
		isValidRace := false
		for _, r := range validRaces {
			if race == r {
				isValidRace = true
				break
			}
		}
		if !isValidRace {
			log.Fatalf("Invalid race value: %s. Allowed values are: ape, alien, female, male, zombie", race)
		}

		// 출력 디렉토리 준비
		if _, err := os.Stat(config.OutputPath); os.IsNotExist(err) {
			err := os.MkdirAll(config.OutputPath, 0755)
			if err != nil {
				log.Fatalf("Failed to create output directory %s: %v", config.OutputPath, err)
			}
		}

		// 에셋 로드
		if err := assets.LoadAssets(race); err != nil {
			log.Fatalf("Failed to load assets: %v", err)
		}

		if race != "" {
			fmt.Printf("Generating only %s punks.\n", race)
		}

		fmt.Printf("Generating %d punks...\n", config.NumPunksToGenerate)
		for i := 0; i < config.NumPunksToGenerate; i++ {
			if err := generator.GeneratePunk(i+1, race); err != nil {
				log.Printf("Failed to generate punk %d: %v", i+1, err)
			}
		}

		fmt.Println("Done!")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&race, "race", "r", "", "Generate only punks of a specific race (e.g., ape, alien, female, male, zombie)")
}
