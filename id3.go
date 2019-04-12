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

		tag_header, tag_data := getV2TagHeader(file_reader)

		// Update the reader so it will return EOF at the end of the tag.
		file_reader = bufio.NewReader(io.LimitReader(file_reader, int64(tag_header.Size)))

		var item *Item
		if tag_header.Version == 4 {
			item = v24GetManager(path, file_reader)
		} else if tag_header.Version == 3 {
			item = v23GetManager(path, file_reader)
		} else if tag_header.Version == 2 {
			item = v22GetManager(path, file_reader)
		} else {
			panic(fmt.Sprintf("Unrecognized ID3v2 version: %d", tag_header.Version))
		}
		makeItemTag(item, tag_header, tag_data)

		printItemData(item)
	}
}

func actOnStdin() {
	fmt.Println("Would act on stdin")
}

func fileHasV2Tag(reader *bufio.Reader) bool {
	// This check is very limited. It isn't strictly necessary for
	// the ID3 tag to occur at the beginning of the file.  @TODO
	checkTag := func (bytes []byte) bool {
		return (bytes[0] == 'I' && bytes[1] == 'D' && bytes[2] == '3')
	}
	return areBytesOk(reader, 3, checkTag)
}

// getV2TagHeader receives a Reader and returns a struct containing
// the tag's header information.
// The tag's header will contain ten bytes:
// 0-2: the characters "ID3"
// 3-4: the version number
//      - first byte is the version's major number
//        So `04` here indicates v2.4
//      - second byte is the minor version
//        So `00` here indicates v2.4.0
// 5: bitwise flags. See the `v2#FillHeader` functions.
//    Followed by four blank bits
// 6-9: size of the entire tag encoded in synchsafe integer
func getV2TagHeader(reader *bufio.Reader) (ID3v2TagHeader, []byte) {
	data := readBytes(reader, 10)

	header := ID3v2TagHeader{ }
	header.Version = int(data[3])
	header.MinorVersion = int(data[4])
	header.Size = synchsafeBytesToInt(data[6:])

	return header, data
}

func makeItemTag(item *Item, header ID3v2TagHeader, data []byte) {
	item.FillHeader(&header, data)
	item.Tag.Header = header
	item.Tag.Frames = item.ReadFrames()
}

func printItemData(item *Item) {
	fmt.Printf("[%v:%v]\n", item.Tag.Header.Version, item.Path)
	item.PrintFrames(item.Tag.Frames)
}

func printUsage() {
	fmt.Printf("Usage: %s [path(s) to mp3 file]\n", os.Args[0])
}
