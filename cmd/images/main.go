package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Actions []Action `yaml:"actions,omitempty"`
}

type Action struct {
	Name        string `yaml:"name,omitempty"`
	Glob        string `yaml:"glob,omitempty"`
	Destination string `yaml:"destination,omitempty"`
	MaxWidth    int    `yaml:"maxWidth,omitempty"`
	Quality     int    `yaml:"quality,omitempty"`
}

func main() {
	configPath := flag.String("config", "image-processing.yaml", "config file location")
	flag.Parse()

	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("Couldn't load config file, %s", err)
	}
	defer configFile.Close()

	config := Config{}
	err = yaml.NewDecoder(configFile).Decode(&config)
	if err != nil {
		log.Fatalf("Couldn't read the config file, %s", err)
	}

	for _, action := range config.Actions {
		ProcessAction(action)
	}

}

type Bytes int64

func (b Bytes) String() string {
	return ByteCountIEC(int64(b))
}

type ImageData struct {
	X, Y int
	Size Bytes
	Path string
}

func (i ImageData) String() string {
	return fmt.Sprintf("%dx%d (%s)", i.X, i.Y, i.Size)
}

type ActionResult struct {
	Originals    []ImageData
	Destinations []ImageData
}

func ProcessAction(action Action) {
	log.Printf("Running action %s", action.Name)
	imageGlob, err := filepath.Glob(action.Glob)
	if err != nil {
		log.Printf("Action glob is invalid, %s ,%s", action.Glob, err)
		return
	}
	if len(imageGlob) == 0 {
		log.Printf("No images found for %s", action.Glob)
	}
	result := ActionResult{
		Originals:    []ImageData{},
		Destinations: []ImageData{},
	}

	for i, sourceImage := range imageGlob {
		originalImage, processedImage, err := ProcessImage(action, i, sourceImage)
		if err != nil {
			log.Printf("Failed to process image %s, %s", sourceImage, err)
		}
		result.Originals = append(result.Originals, originalImage)
		result.Destinations = append(result.Destinations, processedImage)
	}

	var totalOriginalSize Bytes
	var totalProcessedSize Bytes
	for i := range result.Originals {

		totalOriginalSize += result.Originals[i].Size
		totalProcessedSize += result.Destinations[i].Size
		log.Printf("Original %s, processed %s, Saved %.0f%%", result.Originals[i], result.Destinations[i], 100-100*float64(result.Destinations[i].Size)/float64(result.Originals[i].Size))
	}

	log.Printf("Total Results: Original %s, processed %s, Saved %.0f%%", totalOriginalSize, totalProcessedSize, 100-100*float64(totalProcessedSize)/float64(totalOriginalSize))

}

func ResizeImage(source image.Image, maxWidth int) (image.Image, error) {
	if source.Bounds().Dx() < maxWidth {
		log.Printf("Image is already smaller than %d, just copy the file", maxWidth)
		return source, nil
	}

	Y := (source.Bounds().Dy() * maxWidth) / source.Bounds().Dx()

	log.Printf("Resizing image to %dx%d", maxWidth, Y)
	resizedImage := resize.Resize(uint(maxWidth), uint(Y), source, resize.Lanczos2)
	return resizedImage, nil
}

func ProcessImage(action Action, i int, sourceImage string) (original ImageData, destination ImageData, _ error) {
	original.Path = sourceImage

	betterName := fmt.Sprintf("images_%d.jpg", i)
	destination.Path = filepath.Join(action.Destination, betterName)

	destinationFile, err := os.Create(destination.Path)
	if err != nil {
		return original, destination, fmt.Errorf("Couldn't create destination file, %s,%w", destination.Path, err)
	}
	defer destinationFile.Close()

	log.Printf("Processing image %s", sourceImage)
	sourceImageFile, err := os.Open(sourceImage)
	if err != nil {
		return original, destination, fmt.Errorf("Couldn't open source image, %s ,%w", sourceImage, err)
	}
	defer sourceImageFile.Close()

	sourceDecoded, _, err := image.Decode(sourceImageFile)
	if err != nil {
		return original, destination, fmt.Errorf("Failed to decode image, %w", err)
	}

	original.X = sourceDecoded.Bounds().Dx()
	original.Y = sourceDecoded.Bounds().Dy()
	original.Size = getFileSize(original.Path)

	resizedImage, err := ResizeImage(sourceDecoded, action.MaxWidth)
	if err != nil {
		return original, destination, fmt.Errorf("Couldn't resize image %s, %w", sourceImage, err)
	}

	quality := 80
	if action.Quality != 0 {
		quality = action.Quality
	}

	err = jpeg.Encode(destinationFile, resizedImage, &jpeg.Options{
		Quality: quality,
	})
	if err != nil {
		return original, destination, fmt.Errorf("Failed to encode image, %w", err)
	}

	// close the file before getting metrics
	destinationFile.Close()
	destination.X = resizedImage.Bounds().Dx()
	destination.Y = resizedImage.Bounds().Dy()
	destination.Size = getFileSize(destination.Path)

	return original, destination, nil
}

func getFileSize(path string) Bytes {
	s, err := os.Stat(path)
	if err != nil {
		log.Printf("Couldn't get file stat, %s, %s", path, err)
		return 0
	}
	return Bytes(s.Size())
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
