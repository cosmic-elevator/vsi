package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bmaupin/go-epub"
)

func main() {
	numberOfProvidedArgs := len(os.Args)

	if numberOfProvidedArgs != 5 {
		log.Fatal("Four arguments required\nUsage: vsi <inputFolder> <outputFilename> <title> <author>")
	}

	inputFolder := os.Args[1]
	outputFilename := os.Args[2]
	title := os.Args[3]
	author := os.Args[4]

	err := createEpub(inputFolder, outputFilename, title, author)

	if err != nil {
		log.Fatal(err)
	}
}

// createEpub creates a new EPUB book from a directory of HTML documents
func createEpub(inputFolder string,
	outputFilename string,
	title string,
	author string) error {
	// Set up new EPUB with the correct metadata
	e := epub.NewEpub(title + ": A Very Short Introduction")
	e.SetAuthor(author)

	// Find all chapter files (ending in .html)
	allChapters, err := filepath.Glob(inputFolder + "/*.html")

	if err != nil {
		log.Fatal("Could not find HTML files in provided directory")
	}

	// Add each chapter to the EPUB
	for _, chapter := range allChapters {
		log.Printf("Formatting chapter %v\n", chapter)

		err = addChapter(e, chapter)

		if err != nil {
			return err
		}
	}

	// Write the EPUB to a file
	return e.Write(outputFilename)
}

// addChapter adds a chapter to an existing EPUB by loading a HTML file from the given filename
func addChapter(e *epub.Epub, chapterFileName string) error {
	// Since addChapter is given the filename of the HTML file, create the directory path for the corresponding files
	// e.g. CSS and images
	chapterFilesPath := strings.Replace(chapterFileName, ".html", "_files/", -1)

	// Find all CSS files loaded by the chapter
	allCss, err := filepath.Glob(chapterFilesPath + "*.css")

	if err != nil {
		return fmt.Errorf("could not read CSS files for chapter: %v", err)
	}

	// Read each CSS file and join into a combined file
	var joinedCss string
	for _, css := range allCss {
		cssData, err := ioutil.ReadFile(css)

		if err != nil {
			return fmt.Errorf("could not read file %v: %v", css, err)
		}

		joinedCss += string(cssData)
	}

	// Make and add the combined CSS file to the EPUB
	joinedCssFile, err := os.Create("joined-css.css")

	if err != nil {
		return fmt.Errorf("could not make joined CSS file: %v", err)
	}

	defer joinedCssFile.Close()

	_, err = joinedCssFile.WriteString(joinedCss)

	if err != nil {
		return fmt.Errorf("could not write CSS content in the file: %v", err)
	}

	epubCSSPath, err := e.AddCSS("joined-css.css", "")

	if err != nil {
		return fmt.Errorf("could not add css file into epub: %v", err)
	}



	// Open the chapter HTML file
	file, err := os.Open(chapterFileName)

	if err != nil {
		return fmt.Errorf("could not open chapter HTML file: %v", err)
	}

	// Create a document using goquery, which is used for parsing the HTML
	doc, err := goquery.NewDocumentFromReader(file)

	if err != nil {
		return fmt.Errorf("could not open chapter document with goquery: %v", err)
	}

	// Extract the chapter title from the document.
	var chapterTitle string
	doc.Find(".chapter-title-without-label").Each(func(i int, s *goquery.Selection) {
		// Clear span tag and get chapter title
		s.Find("span").Remove()
		chapterTitle = s.Text()

	})

	// fileMap associates original image file names (as downloaded from the web page) with generated image names
	// stored in the EPUB
	fileMap := make(map[string]string)

	// Extract the chapter content from the HTML (found under the div with class `chunkBody`)
	var readerError error
	doc.Find(".widget-items").Each(func(i int, s *goquery.Selection) {
		// Add chapter header
		s.PrependHtml("<h2>" + chapterTitle + "</h2>")

		// Read all HTML from the content div
		chapterHtml, err := s.Html()

		if err != nil {
			readerError = fmt.Errorf("failed to read HTML contents of chapterL %v", err)
			return
		}

		// Store all images in the EPUB
		s.Find("img").Each(func(i int, s *goquery.Selection) {
			imageName, exists := s.Attr("src")

			if !exists {
				return
			}

			// Create a full path to the image file in the local filesystem
			unescapedImageFileName := filepath.Dir(chapterFileName) + "/" + strings.Replace(html.UnescapeString(imageName), "%20", " ", -1)
			//fmt.Printf(unescapedImageFileName+"\n")
			//Add the image to the EPUB, getting a unique generated name
			imageEpubFilename, err := e.AddImage(unescapedImageFileName, "")

			if err != nil {
				readerError = fmt.Errorf("could not add image %v: %v", unescapedImageFileName, err)
				return
			}

			// Store the original filename and its corresponding generated filename in the file map
			fileMap[imageName] = imageEpubFilename
		})

		chapterHtml = html.UnescapeString(chapterHtml)

		// For each image in the document, replace original paths with the new intra-EPUB path
		for originalSrc, newSrc := range fileMap {
			chapterHtml = strings.Replace(chapterHtml, originalSrc, newSrc, -1)
		}

		// Run a series of deletions on the HTML:
		// - delete navigation links
		// - delete page references
		// - delete page number
		// - delete download all slides button
		// - delete download image button
		deleteNavRegex := regexp.MustCompile(`<ul class="div1-nav">.*<\/ul>`)
		deletePageRefRegex := regexp.MustCompile(`<span id="\w*" class="page-number-prefix js-page-prefix  ">p\. \d*<\/span>`)
		deletePageNumber := regexp.MustCompile(`<span class=" page-span".*</span>`)
		deletePageRefIconRegex := regexp.MustCompile(`<span title="\w*" class="printPageMark">↵<\/span>`)
		deleteImgButtonRegex := regexp.MustCompile(`<div class="ajax-articleAbstract-exclude-regex fig-orig original-slide figure-button-wrap">.*<\/div>`)
		deleteDownloadAllSlides := regexp.MustCompile(`<a id="lnkDownloadAllImages".*</a>`)

		chapterHtml = deleteNavRegex.ReplaceAllString(chapterHtml, "")
		chapterHtml = deletePageRefRegex.ReplaceAllString(chapterHtml, "")
		chapterHtml = deletePageRefIconRegex.ReplaceAllString(chapterHtml, "")
		chapterHtml = deleteImgButtonRegex.ReplaceAllString(chapterHtml, "")
		chapterHtml = deleteDownloadAllSlides.ReplaceAllString(chapterHtml, "")
		chapterHtml = deletePageNumber.ReplaceAllString(chapterHtml, "")

		// Create the filename for the chapter based on title
		chapterEpubFilename := strings.Replace(chapterTitle, "?", "", -1) + ".chapterHtml"

		// Add the chapter to the EPUB
		e.AddSection(chapterHtml, chapterTitle, chapterEpubFilename, epubCSSPath)
	})

	if readerError != nil {
		return fmt.Errorf("failed to parse chapter HTML: %v", readerError)
	}

	return nil
}
