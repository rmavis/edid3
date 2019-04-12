package main

import (
	"bufio"
	//"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)


func main() {
	has_args := len(os.Args) > 0
	has_data := !isFileEmpty(os.Stdin)

	if (!(has_args || has_data)) {
		printUsage()
		return
	}

	if has_args {
		actOnArgs(os.Args)
	}

	if has_data {
		actOnStdin()
	}
}

func actOnArgs(args []string) {
	for _, arg := range os.Args[1:] {
		path, err := filepath.Abs(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Path '%s' appears to be invalid: %s\n", path, err)
			return
		}

		handle, err := os.Open(path)
		defer handle.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open file '%s': %s\n", path, err)
			return
		}

		file_reader := bufio.NewReader(handle)
		if fileHasV2Tag(file_reader) == false {
			fmt.Fprintf(os.Stderr, "Unable to read ID3 tag: not present in file '%s'.\n", path)
			return
		}

		tag_header, tag_data := readV2TagHeader(file_reader)

		// Update the reader so it will return EOF at the end of the tag.
		file_reader = bufio.NewReader(io.LimitReader(file_reader, int64(tag_header.Size)))

		var item *Item
		if tag_header.Version == 4 {
			item = v24MakeItem(path, file_reader)
		} else if tag_header.Version == 3 {
			item = v23MakeItem(path, file_reader)
		} else if tag_header.Version == 2 {
			item = v22MakeItem(path, file_reader)
		} else {
			panic(fmt.Sprintf("Unrecognized ID3v2 version: %d", tag_header.Version))
		}
		fillItemTag(item, tag_header, tag_data)

		printItemData(item)
	}
}

func actOnStdin() {
	fmt.Println("Would act on stdin")
}

func printUsage() {
	fmt.Printf("Usage: %s [path(s) to mp3 file]\n", os.Args[0])
}
