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
	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s [path(s) to mp3 file]\n", os.Args[0])
		return
	}

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
		if fileHasID3v2Tag(file_reader) == false {
			fmt.Fprintf(os.Stderr, "Unable to read ID3 tag: not present in file '%s'.\n", path)
			return
		}

		tag_header := getID3v2TagHeader(file_reader)

		// Update the reader so it will return EOF at the end of the tag.
		file_reader = bufio.NewReader(io.LimitReader(file_reader, int64(tag_header.Size)))

		var item *Item
		if tag_header.Version == 4 {
			item = v24GetManager(path, file_reader)
		} else if tag_header.Version == 3 {
			v23GetFrames(file_reader)
		} else if tag_header.Version == 2 {
			v22GetFrames(file_reader)
		} else {
			panic(fmt.Sprintf("Unrecognized ID3v2 version: %d", tag_header.Version))
		}
		makeItemTag(item, tag_header)

		printItemData(item)
	}
}

// fileHasID3v2Tag receives a Reader and returns a boolean indicating
// whether the file being read by that Reader contains an ID3 tag.
func fileHasID3v2Tag(reader *bufio.Reader) bool {
	// This check is very limited. It isn't necessary for the ID3 tag
	// to occur at the beginning of the file. They can also occur at
	// the end.  @TODO
	checkTag := func (bytes []byte) bool {
		return (bytes[0] == 'I' && bytes[1] == 'D' && bytes[2] == '3')
	}
	return areBytesOk(reader, 3, checkTag)
}

// getID3v2TagHeader receives a Reader and returns a struct containing
// the tag's header information.
// The tag's header will contain ten bytes:
// 0-2: the characters "ID3"
// 3-4: the version number
//      - first byte is the version's major number
//        So `04` here indicates v2.4
//      - second byte is the minor version
//        So `00` here indicates v2.4.0
// 5: flags: bitwise/boolean indicators of:
//    7th bit: Unsynchronisation
//    6th bit: Extended header
//    5th bit: Experimental indicator
//    4th bit: Footer present
//    Followed by four blank bits
// 6-9: size of the entire tag encoded in synchsafe integer
func getID3v2TagHeader(reader *bufio.Reader) ID3v2TagHeader {
	data := readBytes(reader, 10)

	header := ID3v2TagHeader{ }

	header.Version = int(data[3])
	header.MinorVersion = int(data[4])

	header.Unsynchronization = boolFromByte(data[5], 7)
	header.Extended = boolFromByte(data[5], 6)
	header.Experimental = boolFromByte(data[5], 5)
	header.Footer = boolFromByte(data[5], 4)

	header.Size = synchsafeBytesToInt(data[6:])

	return header
}

func makeItemTag(item *Item, header ID3v2TagHeader) {
	item.Tag.Header = header
	item.Tag.Frames = item.ReadFrames()
}

func printItemData(item *Item) {
	fmt.Printf("[%v]\n", item.Path)
	item.PrintFrames(item.Tag.Frames)
}
