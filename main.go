package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	htmlFile := "growth-adventure.html"
	downloadImagesFromHTML(htmlFile, "src=\"")

	fmt.Println("END")
}

func downloadImagesFromHTML(htmlFile string, htmlSourceTag string) {
	_, ferr := os.Stat(htmlFile)
	if ferr != nil {
		if os.IsNotExist(ferr) {
			fmt.Printf("<!> File does not exist: %s\n", htmlFile)
		} else {
			fmt.Printf("<!> Failed to open html file: %s\n", htmlFile)
		}
		os.Exit(1)
	}
	outputDir, _ := strings.CutSuffix(htmlFile, ".html")
	createOutputDir(outputDir)

	htmlLines, err := readFileLines(htmlFile)
	if err != nil {
		fmt.Println("<!> Failed to read lines")
		os.Exit(1)
	}

	imgURLs, err := getImageSources(htmlLines, htmlSourceTag)
	if err != nil {
		fmt.Println("<!> Failed to get image sources")
		os.Exit(1)
	}

	for i := 1; i < len(imgURLs); i++ {
		fmt.Println(imgURLs[i])
		err := downloadImage(imgURLs[i], fmt.Sprintf("%s/%d.png", outputDir, i))
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}
}

func downloadImage(url string, destPath string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Download the image
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("<!> Failed to download content: %v", err)
	}
	defer resp.Body.Close()

	// Verify content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("<!> Not an image: %s", contentType)
	}

	// Create destination file
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("<!> Failed to create file: %v", err)
	}
	defer f.Close()

	// Copy the data
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("<!> Failed to save image: %v", err)
	}

	return nil
}

func readFileLines(filename string) ([]string, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("<!> Failed to open file: %v", err)
	}
	defer file.Close()

	var lines []string

	// Create scanner for reading line by line
	scanner := bufio.NewScanner(file)

	// Read each line and append to slice
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("<!> Failed to read file: %v", err)
	}

	return lines, nil
}

func getImageSources(htmlContent []string, element string) ([]string, error) {
	var imageSources []string

	// Parse the HTML content
	for i := range len(htmlContent) {
		if strings.Contains(htmlContent[i], element) {
			imageSources = append(imageSources, strings.Split(htmlContent[i], "\"")[1])
		}
	}

	return imageSources, nil
}

func createOutputDir(path string) {
	out := "output"
	path = fmt.Sprintf("%s/%s", out, path)
	_, terr := os.Stat(out)
	if os.IsNotExist(terr) {
		err := os.Mkdir(out, os.ModePerm)
		if err != nil {
			fmt.Println("<!> Failed to make output directory")
			os.Exit(1)
		}
	}
	_, derr := os.Stat(path)
	if os.IsNotExist(derr) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			fmt.Println("<!> Failed to make output directory")
			os.Exit(1)
		}
	} else {
		fmt.Printf("Directory already exist: %s\n", path)
	}
}

// GetHTML fetches and returns the HTML content of a webpage
// func getHTML(url string) (string, error) {
// 	// Create HTTP client with timeout
// 	client := &http.Client{
// 		Timeout: 30 * time.Second,
// 	}

// 	// Make HTTP GET request
// 	resp, err := client.Get(url)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to fetch page: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	// Read the response body
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read response body: %v", err)
// 	}

// 	return string(body), nil
// }
