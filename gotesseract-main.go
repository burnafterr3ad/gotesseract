package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/nfnt/resize"
	"github.com/otiai10/gosseract/v2"
)

func resizeImageIfEnabled(imagePath string, fast bool) (string, error) {
	if !fast {
		return imagePath, nil
	}
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	if width > 1200 {
		resized := resize.Resize(1200, 0, img, resize.Lanczos3)
		var buf bytes.Buffer
		switch format {
		case "jpeg":
			err = jpeg.Encode(&buf, resized, nil)
		case "png":
			err = png.Encode(&buf, resized)
		default:
			return "", fmt.Errorf("unsupported image format: %s", format)
		}
		if err != nil {
			return "", err
		}
		tempPath := filepath.Join(os.TempDir(), filepath.Base(imagePath))
		err = os.WriteFile(tempPath, buf.Bytes(), 0644)
		if err != nil {
			return "", err
		}
		return tempPath, nil
	}
	return imagePath, nil
}

func worker(id int, jobs <-chan string, searchString string, fast bool, mu *sync.Mutex, counter *int, counterMu *sync.Mutex, wg *sync.WaitGroup, results *[]string, resultsMu *sync.Mutex) {
	defer wg.Done()
	client := gosseract.NewClient()
	defer client.Close()

	for imagePath := range jobs {
		processedPath, err := resizeImageIfEnabled(imagePath, fast)
		if err != nil {
			log.Printf("[Worker %d] Failed to process image %s: %v", id, imagePath, err)
			continue
		}
		client.SetImage(processedPath)
		text, err := client.Text()
		if err != nil {
			log.Printf("[Worker %d] Error processing %s: %v", id, processedPath, err)
			continue
		}
		for _, line := range strings.Split(text, "\n") {
			if strings.Contains(line, searchString) {
				result := fmt.Sprintf("Found in: %s | Line: %s", filepath.Base(imagePath), line)
				mu.Lock()
				fmt.Println(result)
				mu.Unlock()
				resultsMu.Lock()
				*results = append(*results, result)
				resultsMu.Unlock()
			}
		}
		counterMu.Lock()
		*counter++
		counterMu.Unlock()
	}
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

func main() {
	searchString := flag.String("search", "", "The string to search for in OCR-extracted text (required)")
	folderPath := flag.String("path", ".", "Path to the folder containing images")
	outputFile := flag.String("out", "results.txt", "File to save matching results")
	fastMode := flag.Bool("fast", false, "Enable fast mode (resize large images to 1200px width)")
	verbose := flag.Bool("help", false, "Display help information")
	flag.Parse()

	// Manual support for --help
	for _, arg := range os.Args {
		if arg == "--help" {
			*verbose = true
		}
	}

	if *verbose || *searchString == "" {
		fmt.Println("Usage: ./tesseract-search -search <string> [-path <folder path>] [-out <results file>] [-fast]")
		fmt.Println("\nOptions:")
		fmt.Println("  -search string\tThe string to search for in OCR-extracted text (required)")
		fmt.Println("  -path   string\tPath to the folder containing images (default: current directory)")
		fmt.Println("  -out    string\tFile to save matching results (default: results.txt)")
		fmt.Println("  -fast          \tEnable fast mode (resize large images to 1200px width)")
		fmt.Println("  -help, --help  \tDisplay this help message")
		return
	}

	files, err := os.ReadDir(*folderPath)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	imageFiles := []string{}
	for _, file := range files {
		if !file.IsDir() && isImageFile(file.Name()) {
			imageFiles = append(imageFiles, filepath.Join(*folderPath, file.Name()))
		}
	}

	totalImages := len(imageFiles)
	fmt.Printf("Found %d image(s) in '%s'\n", totalImages, *folderPath)

	jobs := make(chan string, totalImages)
	var mu sync.Mutex
	var counter int
	var counterMu sync.Mutex
	var results []string
	var resultsMu sync.Mutex

	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	fmt.Printf("Using %d workers\n", numWorkers)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, *searchString, *fastMode, &mu, &counter, &counterMu, &wg, &results, &resultsMu)
	}

	// Progress reporter
	go func() {
		for {
			time.Sleep(10 * time.Second)
			counterMu.Lock()
			fmt.Printf("Progress: %d/%d images processed\n", counter, totalImages)
			counterMu.Unlock()
			if counter >= totalImages {
				return
			}
		}
	}()

	for _, imagePath := range imageFiles {
		jobs <- imagePath
	}
	close(jobs)

	wg.Wait()

	// Write results to file
	if len(results) > 0 {
		err := os.WriteFile(*outputFile, []byte(strings.Join(results, "\n")), 0644)
		if err != nil {
			log.Fatalf("Failed to write results to file: %v", err)
		}
		fmt.Printf("Results saved to: %s\n", *outputFile)
	} else {
		fmt.Println("No matches found.")
	}
}
