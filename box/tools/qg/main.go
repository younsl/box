package main

import (
	"fmt"
	"image/png"
	"os"
	"regexp"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/spf13/cobra"
)

const (
	version = "0.1.1"
)

func main() {
	var width int
	var height int
	var filename string
	var quiet bool

	var rootCmd = &cobra.Command{
		Use:     "qg [flags] <url>",
		Short:   "QR code generator",
		Long:    "A simple QR code generator that creates a QR code from a given URL.",
		Example: "qg --quiet --width 200 --height 200 --filename qrcode.png https://github.com/",
		Version: version,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			address := args[0]

			matched, err := regexp.MatchString(`^https?://`, address)
			if err != nil || !matched {
				fmt.Println("Invalid URL. The URL must start with http:// or https://. Please check the URL and try again.")
				os.Exit(1)
			}

			qrCode, err := qr.Encode(address, qr.L, qr.Auto)
			if err != nil {
				panic(err)
			}

			qrCode, err = barcode.Scale(qrCode, width, height)
			if err != nil {
				panic(err)
			}

			file, err := os.Create(filename)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			if err := png.Encode(file, qrCode); err != nil {
				panic(err)
			}

			if !quiet {
				fmt.Printf("QR code saved as %s.\n", filename)
				fmt.Printf("Address: %s. Size: %dx%d\n", address, width, height)
			}
		},
	}

	rootCmd.Flags().StringVarP(&filename, "filename", "f", "qrcode.png", "Output filename for the QR code")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress output messages")
	rootCmd.Flags().IntVar(&height, "height", 100, "Height of the QR code in pixels")
	rootCmd.Flags().IntVar(&width, "width", 100, "Width of the QR code in pixels")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
