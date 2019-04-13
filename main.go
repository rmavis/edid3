package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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

	reader := bufio.NewReader(os.Stdin)
	var line string
	var char Rune
	current, max := 0, fileSize(os.Stdin)
	for current < max {
		char = reader.ReadRune()
		if char == "\n" {

		} else if char == '[' {
			line, err = reader.ReadString(']')
			if err != nil {
				// This should be more graceful  @TODO
				panic(fmt.Sprintf(os.Stderr, "Error reading for ']': %s\n", err))
			}
			path := strings.TrimSuffix(line, ']')
			// Make token: type: 'path', value: path
		}

		current++  // CAREFUL
	}


	// The `[^/]*` allows for information to be present before the path.
	// The path must be an absolute path.
	file_regex := regexp.MustCompile(`\[[^/]*([^\]]+)\]`)
	frame_regex := regexp.MustCompile(`^([^:]+)[ ]*:[ ]*(.+)$`)
	for reader.Scan() {
		line := scanner.Text()
		//fmt.Printf("GOT LINE '%v'\n", line)

		if ((len(line) == 0) ||
			(line[0] == '#')) {
			continue
		}

		var match []string
		if file_regex.MatchString(line) {
			match = file_regex.FindStringSubmatch(line)
			path := match[1]
			fmt.Printf("GOT FILE PATH '%v'\n", path)
		} else if frame_regex.MatchString(line) {
			match = frame_regex.FindStringSubmatch(line)
			name := match[1]
			data := match[2]
			fmt.Printf("GOT FRAME WITH NAME '%v' AND DATA '%v'\n", name, data)
		} else {
			fmt.Printf("CAN'T HANDLE LINE '%v'\n", line)
		}

	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}

func printUsage(program_name string) {
	fmt.Printf("Usage: %s [path(s) to mp3 file]\n", program_name)
}
