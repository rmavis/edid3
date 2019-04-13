package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)


func main() {
	// The first argument is the program name.
	has_args := len(os.Args) > 1
	has_data := !isFileEmpty(os.Stdin)

	if (!(has_args || has_data)) {
		printUsage(os.Args[0])
		return
	}

	if has_args {
		actOnArgs(os.Args[1:])
	}

	if has_data {
		actOnStdin()
	}
}

func actOnArgs(args []string) {
	x := len(args) - 1
	for _, arg := range args {
		if (arg[0] == '-') {
			fmt.Printf("FLAG ARGUMENT '%v'\n", arg)
		} else {
			item, err := itemFromFile(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				continue
			}

			printItemData(item)
			if x > 0 {
				fmt.Println()
			}
		}
		x--
	}
}

func itemFromFile(file_name string) (*Item, error) {
	var item *Item

	path, err := filepath.Abs(file_name)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("File '%s' appears not to exist (%v).", path, err))
	}

	handle, err := os.Open(path)
	defer handle.Close()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't open file '%s' (%s).", path, err))
	}

	file_reader := bufio.NewReader(handle)

	tag_header, header_data, err := readV2TagHeader(file_reader)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ID3 tag not present in file '%s'.\n", path))
	}

	// Update the reader so it will return EOF at the end of the tag.
	file_reader = bufio.NewReader(io.LimitReader(file_reader, int64(tag_header.Size)))

	if tag_header.Version == 2 {
		item = v22MakeItem(path, file_reader)
	} else if tag_header.Version == 3 {
		item = v23MakeItem(path, file_reader)
	} else if tag_header.Version == 4 {
		item = v24MakeItem(path, file_reader)
	} else {
		return nil, errors.New(fmt.Sprintf("Unrecognized tag version (%d).", tag_header.Version))
	}
	fillItemTag(item, tag_header, header_data)

	return item, nil
}

func actOnStdin() {
	fmt.Println("Would act on stdin")
}

func printUsage(program_name string) {
	fmt.Printf("Usage: %s [path(s) to mp3 file]\n", program_name)
}
