package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func downloadImage(url string, destPath string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Download the image
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	// Verify content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("not an image: %s", contentType)
	}

	// Create destination file
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	// Copy the data
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}

	return nil
}

// GetHTML fetches and returns the HTML content of a webpage
func getHTML(url string) (string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make HTTP GET request
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}

func readFileLines(filename string) ([]string, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
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
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return lines, nil
}

func getImageSources(htmlContent []string, element string) ([]string, error) {
	var imageSources []string

	// Parse the HTML content
	for i := 0; i < len(htmlContent); i++ {
		if strings.Contains(htmlContent[i], element) {
			imageSources = append(imageSources, strings.Split(htmlContent[i], "\"")[1])
		}
	}

	// find the index of all <img> tags

	return imageSources, nil
}

func createOutputDir(path string) error {
	err := os.Mkdir(path, os.ModePerm)
	return err
}

func main() {
	htmlFileName := "grow-wish"

	err := createOutputDir(htmlFileName)
	if err != nil {
		fmt.Println("<!> Failed to make output directory")
		os.Exit(1)
	}

	htmlLines, err := readFileLines(fmt.Sprintf("%s.html", htmlFileName))
	if err != nil {
		fmt.Println("<!> Failed to read lines")
	}

	imgURLs, err := getImageSources(htmlLines, "href=\"")
	if err != nil {
		fmt.Println("<!> Failed to get image sources")
	}

	for i := 1; i < len(imgURLs); i++ {
		err := downloadImage(imgURLs[i], fmt.Sprintf("%s/%d.png", htmlFileName, i))
		if err != nil {
			fmt.Printf("<!> Failed to download image %d\n", i)
		}
	}

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Image downloaded successfully!")
}
