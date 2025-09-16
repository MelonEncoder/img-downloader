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
	var htmlFile string = ""
	var filterText string = "src="
	var outputDir string = ""
	var urlPrefix string = ""

	if len(os.Args) < 2 {
		fmt.Println(
			"Image Downloader (v0.1.1)\n",
			"USAGE\n",
			"    imgdl [...flags]\n",
			"FLAGS\n",
			"    -f <input.html>\n",
			"    -t <filterText>\n",
			"    -o <outputDirectory>\n",
			"    -p <urlPrefix>")
		os.Exit(1)
	}
	for i := range os.Args {
		if i == 0 {
			i++
		}
		switch os.Args[i] {
		case "-f":
			i++
			if i >= len(os.Args) {
				fmt.Println("<!> Error: No file specified after -f flag")
				os.Exit(1)
			}
			htmlFile = os.Args[i]
			if !strings.Contains(htmlFile, ".html") {
				fmt.Printf("<!> Error: %s is not a .html file.\n", htmlFile)
				os.Exit(1)
			}
		case "-t":
			i++
			if i >= len(os.Args) {
				fmt.Println("<!> Error: filter text not specified after -t flag")
				os.Exit(1)
			}
			filterText = os.Args[i]
		case "-o":
			i++
			if i >= len(os.Args) {
				fmt.Println("<!> Error: No output directory specified after -o flag")
				os.Exit(1)
			}
			outputDir = os.Args[i]
		case "-p":
			i++
			if i >= len(os.Args) {
				fmt.Println("<!> Error: No url prefix specified after -p flag")
				os.Exit(1)
			}
			urlPrefix = os.Args[i]
		}
	}

	if len(htmlFile) == 0 {
		fmt.Println("<!> Error: No html file specified.")
		os.Exit(1)
	}

	if len(outputDir) == 0 {
		outputDir = htmlFile[:strings.LastIndex(htmlFile, ".")]
	}

	fmt.Printf("File: %s\nText: \"%s\"\nOutput: %s\nURLPrefix: %s\n\n", htmlFile, filterText, outputDir, urlPrefix)

	downloadImagesFromHTML(htmlFile, filterText, outputDir, urlPrefix)

	fmt.Println("END")
}

func downloadImagesFromHTML(htmlFile string, filterText string, outputDir string, urlPrefix string) {
	_, ferr := os.Stat(htmlFile)
	if ferr != nil {
		if os.IsNotExist(ferr) {
			fmt.Printf("<!> %s does not exist.\n", htmlFile)
		} else {
			fmt.Printf("<!> Failed to open %s.\n", htmlFile)
		}
		os.Exit(1)
	}
	createOutputDir(outputDir)

	htmlLines, err := getHTMLLines(htmlFile)
	if err != nil {
		fmt.Println("<!> Failed to read lines.")
		os.Exit(1)
	}

	imgURLs, err := getImageSources(htmlLines, filterText, urlPrefix)
	if err != nil {
		fmt.Println("<!> Failed to get image sources.")
		os.Exit(1)
	}

	for i := 1; i < len(imgURLs); i++ {
		err := downloadImage(imgURLs[i], outputDir, i)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}
}

func downloadImage(url string, outputDir string, index int) error {
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
		return fmt.Errorf("<!> Response type not an image: %s \nIndex: %d, URL: %s", contentType, index, url)
	}

	// Create output path
	fileExtension := url[strings.LastIndex(url, "."):]
	imgExts := []string{".png", ".apng", ".jpg", ".jpeg", ".gif", ".webp", ".avif"}
	if len(fileExtension) > 5 {
		for _, ext := range imgExts {
			if strings.Contains(url, ext) {
				fileExtension = ext
				break
			}
		}
	}
	outputPath := fmt.Sprintf("%s/%d%s", outputDir, index, fileExtension)

	// Create destination file
	fout, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("<!> Failed to create file: %v", err)
	}
	defer fout.Close()

	// Copy the data
	_, err = io.Copy(fout, resp.Body)
	if err != nil {
		return fmt.Errorf("<!> Failed to save image: %v", err)
	}

	return nil
}

func getHTMLLines(filename string) ([]string, error) {
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
		line := scanner.Text()
		lines = append(lines, line)
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("<!> Failed to read file: %v", err)
	}

	return lines, nil
}

func getImageSources(htmlContent []string, filterText string, urlPrefix string) ([]string, error) {
	var imageSources []string

	// Parse the HTML content
	for i := range len(htmlContent) {
		line := htmlContent[i]
		if strings.Contains(line, filterText) && strings.Contains(line, "/") {
			for strs := range strings.SplitSeq(line, "\"") {
				for str := range strings.SplitSeq(strs, " ") {
					if strings.Contains(str, "https://") {
						imageSources = append(imageSources, str)
						break
					} else if strings.Contains(str, "//") {
						tmpStr := "http:" + str
						imageSources = append(imageSources, tmpStr)
						break
					} else if strings.Contains(str, "/") {
						tmpStr := urlPrefix + str
						spaceIndex := strings.Index(str, "%20")
						if spaceIndex != -1 {
							s := strings.Split(tmpStr, "%20")
							tmpStr = s[0] + " " + s[1]
						}
						imageSources = append(imageSources, tmpStr)
						break
					}
				}
			}
		} else if strings.Contains(line, filterText) {
			nextLine := htmlContent[i+1]
			for strs := range strings.FieldsSeq(nextLine) {
				for str := range strings.SplitSeq(strs, "\"") {
					if strings.Contains(str, "https://") {
						imageSources = append(imageSources, str)
						break
					} else if strings.Contains(str, "//") {
						tmpStr := "http:" + str
						imageSources = append(imageSources, tmpStr)
						break
					} else if strings.Contains(str, "/") {
						tmpStr := urlPrefix + str
						spaceIndex := strings.Index(str, "%20")
						if spaceIndex != -1 {
							s := strings.Split(tmpStr, "%20")
							tmpStr = s[0] + " " + s[1]
						}
						imageSources = append(imageSources, tmpStr)
						break
					}
				}
			}
		}
	}

	return imageSources, nil
}

func createOutputDir(outputDir string) {
	_, derr := os.Stat(outputDir)
	if os.IsNotExist(derr) {
		err := os.Mkdir(outputDir, os.ModePerm)
		if err != nil {
			fmt.Println("<!> Failed to make output directory")
			os.Exit(1)
		}
	} else {
		fmt.Printf("Directory already exist: %s\n", outputDir)
	}
}
