package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/nfnt/resize"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Actions []Action `yaml:"actions,omitempty"`
}

type Action struct {
	Name        string `yaml:"name,omitempty"`
	Source      string `yaml:"source,omitempty"`
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

	nWorkers := runtime.NumCPU()
	queue := make(chan ProcessImageAction, nWorkers)

	ctx, done := context.WithCancel(context.Background())
	log.Println("detected %d cpu", nWorkers)
	for i := 0; i < nWorkers; i++ {
		go StartWorker(ctx, queue)
	}

	for _, action := range config.Actions {
		ProcessAction(ctx, queue, action)
	}

	done()

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
	return fmt.Sprintf("%s: %dx%d (%s)", i.Path, i.X, i.Y, i.Size)
}

type ActionResult struct {
	Originals    []ImageData
	Destinations []ImageData
}
type ProcessImageActionResult struct {
	Original    ImageData
	Destination ImageData
}

type ProcessImageAction struct {
	Action          Action
	destinationName string
	sourceImage     string

	result chan ProcessImageActionResult
}

func StartWorker(ctx context.Context, queue <-chan ProcessImageAction) {
	log.Printf("Starting worker...")
	for {
		select {
		case work := <-queue:
			originalImage, processedImage, err := ProcessImage(work.Action, work.destinationName, work.sourceImage)
			if err != nil {
				log.Printf("Failed to process image %s, %s", work.sourceImage, err)
			}
			work.result <- ProcessImageActionResult{
				Original:    originalImage,
				Destination: processedImage,
			}
		case <-ctx.Done():
			log.Print("Stop worker because context is done")
			return
		}
	}
}

func ProcessAction(ctx context.Context, queue chan<- ProcessImageAction, action Action) {
	log.Printf("Running action %s", action.Name)

	if strings.HasPrefix(action.Source, "http") {
		r, err := http.Get(action.Source)
		if err != nil {
			log.Printf("Failed to fetch image %s, %s", action.Name, err)
			return
		}
		if r.StatusCode != http.StatusOK {
			log.Printf("Unexpected status code %s, %s", action.Name, r.Status)
			return
		}
		defer r.Body.Close()
		tmp, err := os.CreateTemp("", "download_*.jpg")
		if err != nil {
			log.Printf("Cloudn't create a tmp file, %s, %s", action.Name, err)
			return
		}
		io.Copy(tmp, r.Body)
		tmp.Close()

		results := make(chan ProcessImageActionResult, 1)
		go func() {
			queue <- ProcessImageAction{
				destinationName: fmt.Sprintf("%s.jpg", action.Name),
				sourceImage:     tmp.Name(),
				result:          results,
				Action:          action,
			}
		}()

		select {
		case result := <-results:
			log.Printf("Original %s, processed %s, Saved %.0f%%", result.Original, result.Destination, 100-100*float64(result.Destination.Size)/float64(result.Original.Size))
		case <-ctx.Done():
			log.Printf("Context cancelled before all images were processed")
		}

		err = os.Remove(tmp.Name())
		if err != nil {
			log.Printf("Warning: failed to remove tmp file %s, %s", tmp.Name(), err)
		}

		return
	}

	imageGlob, err := filepath.Glob(action.Source)
	if err != nil {
		log.Printf("Action glob is invalid, %s ,%s", action.Source, err)
		return
	}
	if len(imageGlob) == 0 {
		log.Printf("No images found for %s", action.Source)
	}
	results := make(chan ProcessImageActionResult)
	go func() {
		for i, sourceImage := range imageGlob {
			queue <- ProcessImageAction{
				destinationName: fmt.Sprintf("%s_%d.jpg", action.Name, i),
				sourceImage:     sourceImage,
				result:          results,
				Action:          action,
			}
		}
	}()

	result := ActionResult{
		Originals:    []ImageData{},
		Destinations: []ImageData{},
	}
	for i := 0; i < len(imageGlob); i++ {
		select {
		case actionResult := <-results:
			result.Originals = append(result.Originals, actionResult.Original)
			result.Destinations = append(result.Destinations, actionResult.Destination)
		case <-ctx.Done():
			log.Printf("Context cancelled before all images were processed")
			return
		}
	}

	sort.Slice(result.Originals, func(i, j int) bool { return result.Originals[i].Path < result.Originals[j].Path })
	sort.Slice(result.Destinations, func(i, j int) bool { return result.Destinations[i].Path < result.Destinations[j].Path })

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

	resizedImage := resize.Resize(uint(maxWidth), uint(Y), source, resize.NearestNeighbor)
	return resizedImage, nil
}

func ProcessImage(action Action, destinationName string, sourceImage string) (original ImageData, destination ImageData, _ error) {
	log.Printf("Processing image %s", sourceImage)
	original.Path = sourceImage

	destination.Path = filepath.Join(action.Destination, destinationName)

	destinationFile, err := os.Create(destination.Path)
	if err != nil {
		return original, destination, fmt.Errorf("Couldn't create destination file, %s,%w", destination.Path, err)
	}
	defer destinationFile.Close()

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
