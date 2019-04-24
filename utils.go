package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf16"
)


const V2TAGHEADERSIZE = 10


// readV2TagHeader receives a Reader and returns a struct containing
// the tag's header information.
// The tag's header will contain ten bytes:
// 0-2: the characters "ID3"
// 3-4: the version number
//      - first byte is the version's major number
//        So `04` here indicates v2.4
//      - second byte is the minor version
//        So `00` here indicates v2.4.0
// 5: bitwise flags. See the `v2#FillTagHeader` functions.
//    Followed by four blank bits
// 6-9: size of the entire tag encoded in synchsafe integer
func readV2TagHeader(reader *bufio.Reader) (ID3v2TagHeader, []byte, error) {
	header := ID3v2TagHeader{ }
	var data []byte

	if fileHasV2Tag(reader) {
		data = readBytes(reader, V2TAGHEADERSIZE)
		header.Version = int(data[3])
		header.MinorVersion = int(data[4])
		header.Size = synchsafeBytesToInt(data[6:])
		return header, data, nil
	}

	return header, data, errors.New("File contains no ID3v2 tag.")
}

func fileHasV2Tag(reader *bufio.Reader) bool {
	// This check is very limited. It isn't strictly necessary for
	// the ID3 tag to occur at the beginning of the file.  @TODO
	checkTag := func (bytes []byte) bool {
		return (bytes[0] == 'I' && bytes[1] == 'D' && bytes[2] == '3')
	}
	return areBytesOk(reader, 3, checkTag)
}

func fillItemTag(item *Item, header ID3v2TagHeader, data []byte) {
	item.FillTagHeader(&header, data)
	item.Tag.Header = header
	item.Tag.Frames = item.ReadFrames()
}

// Frame IDs consist of three or four bytes, each in the range
// A-Z or 0-9.
func areBytesValidFrameId(bytes []byte) bool {
	for _, byte := range bytes {
		if ((byte < 'A' || byte > 'Z') && (byte < '0' || byte > '9')) {
			return false
		}
	}
	return true
}

// The `size` field of a header comprises the last four bytes. Each
// of those bytes uses only seven of the eight available bits (in
// effort to prevent the occurrence a sequence of twelve 1s, which
// is used as a "sync" signifier or chunk/field header in the body
// of the file's music data). This is called "synchsafe".
// So `synchsafeBytesToInt` receives a slice of bytes and returns
// the value encoded therein as an integer.
func synchsafeBytesToInt(data []byte) int {
	size := int(0)
	for i, b := range data {
		//    b: 0111 1111
		// 0x80: 1000 0000
		//    &: 0000 0000
		if (b & 0x80) > 0 {
			fmt.Println("Size byte had non-zero first bit")
		}

		shift := uint(len(data) - i - 1) * 7  // 21, 14, 7, 0
		size |= int(b & 0x7f) << shift
	}
	// fmt.Printf("WANT TO CONVERT %v\n", data)
	// synchsafeIntToBytes(size)
	return size
}

func synchsafeIntToBytes(size int) []byte {
	//fmt.Printf("CONVERTING %v\n", size)
	var bytes []byte
	swing := int(0)
	for n := 0 ; swing <= size ; n++ {
		swing = (0x7f << uint(7 * n))
		byte := uint8((size & swing) >> uint(7 * n))
		bytes = append(bytes, byte)
		//fmt.Printf("N: %v, SWING: %v, BYTE: %v, BYTES: %v\n", n, swing, byte, bytes)
	}
	for len(bytes) < 4 {
		bytes = append(bytes, uint8(0))
	}
	reverseByteSlice(bytes)
	//fmt.Printf("CONVERTED %v to %v\n", size, bytes)
	return bytes
}

func bytesToInt(bytes []byte) int {
	m := 0
	for i, b := range bytes {
		shift := uint(len(bytes) - i - 1) * 8  // 16, 8, 0
		m |= int(uint(b) << shift)
	}
	return m
}

// reverseByteSlice reverses a slice of bytes in place.
func reverseByteSlice(bytes []byte) {
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
}

func fileSize(file *os.File) int {
	stats, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't stat file '%v': %s\n", file, err)
	}
	return int(stats.Size())
}

func isFileEmpty(file *os.File) bool {
	return fileSize(file) == 0
}

// areBytesOk is a test runner. It receives a Reader, a number of
// bytes to read, and a test function to pass those bytes to. The
// test function must receive a slice of bytes and return a bool.
// areBytesOk will return the return of the test function.
func areBytesOk(reader *bufio.Reader, size int, test func([]byte) bool) bool {
	data, err := reader.Peek(size)
	if err != nil {
		return false
	}
	return test(data)
}

func readBytes(reader *bufio.Reader, c int) []byte {
	bytes := make([]byte, c)

	// Read could return fewer than c bytes, so if it does, then
	// `bytes` will contain 0-value bytes beyond the read bytes.
	// That could create ambiguity in the receiving function's
	// interpretation.  @TODO
	_, err := reader.Read(bytes)
	if ((err != nil) && (err != io.EOF)) {
		panic(err)
	}

	return bytes
}

func readString(reader *bufio.Reader, size int) string {
	return parseString(readBytes(reader, size))
}

// Parses a string from frame data. The first byte represents the encoding:
//   0x01  ISO-8859-1
//   0x02  UTF-16 w/ BOM
//   0x03  UTF-16BE w/o BOM
//   0x04  UTF-8
//
// Refer to section 4 of http://id3.org/id3v2.4.0-structure
func parseString(data []byte) string {
	var s string
	switch data[0] {
	case 0: // ISO-8859-1 text.
		s = ISO8859_1ToUTF8(data[1:])
		break
	case 1: // UTF-16 with BOM.
		s = string(utf16.Decode(toUTF16(data[1:])))
		break
	case 2: // UTF-16BE without BOM.
		panic("Unsupported text encoding UTF-16BE.")  // @TODO
	case 3: // UTF-8 text.
		s = string(data[1:])
		break
	default:
		// No encoding, assume ISO-8859-1 text.
		s = ISO8859_1ToUTF8(data)
	}
	return strings.TrimRight(s, "\u0000")
}

func ISO8859_1ToUTF8(data []byte) string {
	p := make([]rune, len(data))
	for i, b := range data {
		p[i] = rune(b)
	}
	return string(p)
}

func toUTF16(data []byte) []uint16 {
	if len(data) < 2 {
		panic("Sequence is too short too contain a UTF-16 BOM")
	}
	if len(data)%2 > 0 {
		// TODO: if this is UTF-16 BE then this is likely encoded wrong
		data = append(data, 0)
	}

	var shift0, shift1 uint
	if data[0] == 0xFF && data[1] == 0xFE {
		// UTF-16 LE
		shift0 = 0
		shift1 = 8
	} else if data[0] == 0xFE && data[1] == 0xFF {
		// UTF-16 BE
		shift0 = 8
		shift1 = 0
		panic("UTF-16 BE found!")
	} else {
		panic(fmt.Sprintf("Unrecognized UTF-16 BOM: 0x%02X%02X", data[0], data[1]))
	}

	s := make([]uint16, 0, len(data)/2)
	for i := 2; i < len(data); i += 2 {
		s = append(s, uint16(data[i])<<shift0|uint16(data[i+1])<<shift1)
	}
	return s
}

// isBitOn is a convenience function. It receives a byte and a
// number indicating a bit position in that byte. It returns a bool
// indicating the value of that bit (0 = false, 1 = true).
func isBitOn(byte byte, pos int) bool {
	if pos > 7 {
		return false
	}

	// Example:
	// Flag: 1000 0000
	//    1: 0000 0001
	// 1<<7: 1000 0000
	//  F&1: 1000 0000
	return (byte & (1 << uint(pos))) == 1
}

// Use makeMap to make a map from a slice of string tuples.
// The function parameter should return two strings. The first will
// be used as the map's keys, the second as the values.
func makeMap(parts [][2]string, pull func([2]string) (string, string)) map[string]string {
	_map := make(map[string]string)
	for _, part := range parts {
		key, val := pull(part)
		_map[key] = val
	}
	return _map
}

func makeTagFrame(reader *bufio.Reader, header ID3v2FrameHeader) ID3v2Frame {
	frame := ID3v2Frame{ }
	frame.Header = header
	frame.Body = readBytes(reader, header.Size)
	return frame
}

func makeFrameValidator(keys map[string]string, size int) func (frame ID3v2Frame) bool {
	check := func (frame ID3v2Frame) bool {
		if ((len(frame.Header.Id) == size) &&
			((frame.Header.Id[0:1] == "T") || (frame.Header.Id[0:1] == "W"))) {
			_, present := keys[frame.Header.Id]
			return present
		} else {
			return false
		}
	}
	return check
}

func printItemData(item *Item) {
	fmt.Printf("[%v:%v]\n", item.Tag.Header.Version, item.Path)
	item.PrintFrames(item.Tag.Frames)
}

// This isn't being used?  @TODO
func printTagHeader(head ID3v2TagHeader) {
	fmt.Println("Header Information:")
	fmt.Printf("Version: %v\n", head.Version)
	fmt.Printf("MinorVersion: %v\n", head.MinorVersion)
	fmt.Printf("Unsynchronization: %v\n", head.Unsynchronization)
	fmt.Printf("Extended: %v\n", head.Extended)
	fmt.Printf("Experimental: %v\n", head.Experimental)
	fmt.Printf("Footer: %v\n", head.Footer)
	fmt.Printf("Size: %v\n", head.Size)
	fmt.Println()
}

// This isn't being used?  @TODO
func printTagFrame(frame ID3v2Frame) {
	fmt.Println("Frame Information:")
	fmt.Printf("ID: %v\n", frame.Header.Id)
	fmt.Printf("Size: %v\n", frame.Header.Size)
	fmt.Printf("Flags: %v\n", frame.Header.Flags)
	fmt.Printf("Body: %v\n", frame.Body)
	fmt.Println()
}
