# gotesseract

* Searches through folders of images for <string>.
* Automatically optimizes threads
* Exponentially faster than `pytesseract`
* `-fast` option can resize images 
* Identifies total images in specific path
* Provides progress updates every 10 seconds 

~~~
~./gotesseract --help
Usage: ./tesseract-search -search <string> [-path <folder path>] [-out <results file>] [-fast]

Options:
  -search string        The string to search for in OCR-extracted text (required)
  -path   string        Path to the folder containing images (default: current directory)
  -out    string        File to save matching results (default: results.txt)
  -fast                 Enable fast mode (resize large images to 1200px width)
  -help, --help         Display this help message
~~~

Example usage: 

~~~
./gotesseract -search root -path ../Pictures/ -fast -out results.txt
Found 223 image(s) in '../Pictures/'
Using 12 workers
Found in: 2025-01-11_0-41-52.png | Line: ‘(root S helpdesk )-[/home/helpdesk]
Progress: 37/223 images processed
Progress: 45/223 images processed
Found in: 2025-01-11_0-51-35.png | Line: 1 hell unknown SSH root @ 10.0.0.1:42643 > 10.0.0.2:22 (10.0.7.32)
Progress: 57/223 images processed
Progress: 109/223 images processed
Progress: 128/223 images processed
Progress: 155/223 images processed
Progress: 181/223 images processed
Progress: 195/223 images processed
Progress: 218/223 images processed
Progress: 222/223 images processed
Progress: 222/223 images processed
Progress: 222/223 images processed
Progress: 222/223 images processed
Results saved to: results.txt
~~~
