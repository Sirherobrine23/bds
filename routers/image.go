package routers

import (
	"fmt"
	"image"
	"image/color"
	"image/png" // Image format encoder (part of the standard library image functionality)
	"io"
	"math/rand" // Needed for generating random numbers
	"os"        // Needed for file operations (creating/writing the image file)
	"time"      // Needed to seed the random number generator
)

func RandomPng(outFile io.Writer) {
	// --- Configuration ---
	width := 256
	height := 256

	// --- Image Creation ---
	// Create a new RGBA image. RGBA stands for Red, Green, Blue, Alpha.
	// image.Rect defines the rectangle area of the image, starting from (0,0)
	// up to (width, height).
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// --- Randomization Setup ---
	// Seed the random number generator using the current time.
	// This ensures you get different random images each time you run the program.
	// Using rand.NewSource and rand.New is the modern way, preferred over the global rand.Seed.
	randomSource := rand.New(rand.NewSource(time.Now().UnixNano()))

	// --- Pixel Generation ---
	// Get the boundaries of the image.
	bounds := img.Bounds()

	// Iterate over every pixel coordinate (x, y).
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Generate random 8-bit values (0-255) for Red, Green, and Blue.
			// We use the specific randomSource we created.
			r := uint8(randomSource.Intn(256)) // Intn(256) gives a number from 0 to 255
			g := uint8(randomSource.Intn(256))
			b := uint8(randomSource.Intn(256))
			// Set Alpha to 255 (fully opaque). You could make this random too:
			// a := uint8(randomSource.Intn(256))
			a := uint8(255)

			// Create a color.RGBA struct representing the color for this pixel.
			pixelColor := color.RGBA{R: r, G: g, B: b, A: a}

			// Set the pixel at coordinates (x, y) to the generated random color.
			img.Set(x, y, pixelColor)
		}
	}

	// Encode the generated image data into PNG format and write it to the file.
	// You could use other encoders like jpeg.Encode (from "image/jpeg") if needed.
	err := png.Encode(outFile, img)
	if err != nil {
		// Handle error during image encoding
		fmt.Printf("Error encoding image to PNG: %v\n", err)
		os.Exit(1) // Exit if encoding fails
	}
}
