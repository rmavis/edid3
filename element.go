package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)


func elementFromPath(file_name string) (*Element, error) {
	var element *Element

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
		element = v22MakeElement(path, file_reader)
	} else if tag_header.Version == 3 {
		element = v23MakeElement(path, file_reader)
	} else if tag_header.Version == 4 {
		element = v24MakeElement(path, file_reader)
	} else {
		return nil, errors.New(fmt.Sprintf("Unrecognized tag version (%d).", tag_header.Version))
	}
	fillElementTag(element, tag_header, header_data)

	return element, nil
}
