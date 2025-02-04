# VSI

### This project is a fork of kdelwat's project, I fixed some bugs and made some changes to work properly on newer vsi page.

*Very Short Introduction* is a series of books from Oxford University Press that are available via a subscription service. However, since these books are only available in HTML, this program converts them to Kindle-compatible ebooks for easier reading.

**You must have a valid VSI subscription through your institution to use this program. Generated books are for personal use only.**

## Installation
1. `go install github.com/cosmic-elevator/vsi@latest`

## Usage

1. Create a new directory for the book you are creating (e.g. "VSI Weather")
2. Download each chapter of the book as HTML (complete webpage). This step can't be automated as a subscription log-in is required. You should add a number sequentially to preserve the order of your files. Your directory should look like this:

```
1. Introduction - Very Short Introductions_files
1. Introduction.html
2. The international evolution of accounting - Very Short Introductions_files
2. The international evolution of accounting.html
3. The fundamentals of financial accounting - Very Short Introductions_files
3. The fundamentals of financial accounting.html
4. Financial reports of listed companies - Very Short Introductions_files
4. Financial reports of listed companies.html
5. International differences and standardization - Very Short Introductions_files
5. International differences and standardization.html
6. Regulation and audit - Very Short Introductions_files
6. Regulation and audit.html
7. Internal decision-making  costs and volumes - Very Short Introductions_files
7. Internal decision-making  costs and volumes.html
8. Accounting as control - Very Short Introductions_files
8. Accounting as control.html
9. Epilogue - Very Short Introductions_files
9. Epilogue.html
```

3. Run the program: `vsi <inputFolder> <outputFilename> <title> <author>`. This will generate an EPUB at the path given. For the above directory, you would run `vsi "VSI accounting" accounting.epub "Accounting" "Christopher Nobes"`.
4. (optional) To load onto a Kindle device, use [Calibre](https://calibre-ebook.com/).
