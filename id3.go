package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)


// Full reference: http://id3.org/id3v2.4.0-structure


/*

TODO
- It isn't necessary for the ID3 tag to occur at the beginning of the
  file -- they can also occur at the end, or presumably anywhere else.
- What about pulling/scanning for a tag from the end of the file? A
  tag with a footer must appear at the end of a file.
- In `readBytes`
- In `v24GetFrames`
- What about being a little more fault-tolerant?
  Would that involve a lot of work? I'm slightly concerned
  about the position of the reader and the contents of the
  file. The reader relies on the contents of the file being
  in the right/required/specified place, assuming all is as
  specified in the spec. What if there are extraneous bytes?
  What if the `size` value in the header is wrong?

*/



type ID3v2Tag struct {
	Header ID3v2TagHeader
	Frames []ID3v2Frame
}

type ID3v2TagHeader struct {
	Version           int
	MinorVersion      int
	Unsynchronization bool
	Extended          bool
	Experimental      bool
	Footer            bool
	Size              int
}

type ID3v2Frame struct {
	Header ID3v2FrameHeader
	Body   []byte
}

type ID3v2FrameHeader struct {
	Id    string
	Size  int
	Flags []byte
}



func main() {
	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s [path(s) to mp3 file]\n", os.Args[0])
		return
	}

	for _, arg := range os.Args[1:] {
		fmt.Println(arg)

		handle, err := os.Open(arg)
		defer handle.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open %s: %s\n", arg, err)
			return
		}
		reader := bufio.NewReader(handle)

		tag_header, err := getID3v2TagHeader(reader)
		if err != nil {
			fmt.Println(err)
			return
		}
		printHeader(tag_header)

		// Update the reader so it will return EOF at the end of the tag.
		reader = bufio.NewReader(io.LimitReader(reader, int64(tag_header.Size)))

		if tag_header.Version == 4 {
			frames := v24GetFrames(reader)
			v24PrintFrames(frames)
			// for _, frame := range frames {
			// 	printFrame(frame)
			// }
		} else if tag_header.Version == 3 {
			v23GetFrames(reader)
		} else if tag_header.Version == 2 {
			v22GetFrames(reader)
		} else {
			panic(fmt.Sprintf("Unrecognized ID3v2 version: %d", tag_header.Version))
		}
	}
}



/*
For a full reference on ID3v2 tags: http://id3.org/id3v2.4.0-structure

Quick reference:
A "tag" consists of
- a header
- one or more "frames", each of which comprise a key-value pair,
  the keys being part of the pre-defined set, the values being
  character strings
- padding or a footer

The tag's header will contain ten bytes:
0-2: the characters "ID3"
3-4: the version number
     - first byte is the version's major number
       So `04` here indicates v2.4
     - second byte is the minor version
       So `00` here indicates v2.4.0
5: flags: bitwise/boolean indicators of:
   7th bit: Unsynchronisation
   6th bit: Extended header
   5th bit: Experimental indicator
   4th bit: Footer present
   Followed by four blank bits
6-9: size of the entire tag
   
*/

func getID3v2TagHeader(reader *bufio.Reader) (ID3v2TagHeader, error) {
	header := ID3v2TagHeader{ }

	// If the first three bytes aren't ID3, there's no need to proceed.
	checkTag := func (bytes []byte) bool {
		return (bytes[0] == 'I' && bytes[1] == 'D' && bytes[2] == '3')
	}
	if (!areBytesOk(reader, 3, checkTag)) {
		return header, errors.New("Unable to read ID3 tags: invalid tag format.")
	}

	// It isn't necessary for the ID3 tag to occur at the beginning
	// of the file -- they can also occur at the end.  @TODO
	data := readBytes(reader, 10)

	header.Version = int(data[3])
	header.MinorVersion = int(data[4])
	// Flag: 1000 0000
	//    1: 0000 0001
	// 1<<7: 1000 0000
	//  F&1: 1000 0000
	header.Unsynchronization = (data[5] & (1 << 7)) != 0
	header.Extended = (data[5] & (1 << 6)) != 0
	header.Experimental = (data[5] & (1 << 5)) != 0
	header.Footer = (data[5] & (1 << 4)) != 0
	header.Size = calcSynchsafe(data[6:])

	return header, nil
}





/*
The shape of a frame is:
- header (10 bytes, like the tag header).
  0-3: Four characters ID'ing the frame
  4-7: Four bytes indicating the size (synchsafe)
  8-9: Two bytes containing flags
- body
  The default character encoding is ISO-8859-1, but others are
  allowed if a flag is set that specifies the encoding:
     $00   ISO-8859-1 [ISO-8859-1]. Terminated with $00.
     $01   UTF-16 [UTF-16] encoded Unicode [UNICODE] with BOM. All
           strings in the same frame SHALL have the same byteorder.
           Terminated with $00 00.
     $02   UTF-16BE [UTF-16] encoded Unicode [UNICODE] without BOM.
           Terminated with $00 00.
     $03   UTF-8 [UTF-8] encoded Unicode [UNICODE]. Terminated with $00.

So the process of getting a frame's data is much like getting a tag's
header data: read and parse the first ten bytes, read from there up
to the specified size, decode the results.

*/




/*
FOOTER:
  +-----------------------------+
  |      Header (10 bytes)      |
  +-----------------------------+
  |       Extended Header       |
  | (variable length, OPTIONAL) |
  +-----------------------------+
  |   Frames (variable length)  |
  +-----------------------------+
  |           Padding           |
  | (variable length, OPTIONAL) |
  +-----------------------------+
  | Footer (10 bytes, OPTIONAL) |
  +-----------------------------+

The footer is a copy of the header, but with a different
   identifier.

     ID3v2 identifier           "3DI"
     ID3v2 version              $04 00
     ID3v2 flags                %abcd0000
     ID3v2 size             4 * %0xxxxxxx
*/




/*

420
101
(4 x 1) + (2 x 0) + 1 = 5
110 = 6
111
(4x1)+(2x1)+1 = 7

0xAC -> 172
16 0
 A C
A = 10
C = 13
(16 x 10) + 12 = 172

*/



/*

SYNCHSAFE INTEGER CALCULATION

Pass 1:
size: 00000000 00000000 00000000 00000000
shift: 21
b: 01010101
size: 
  01010101
& 01111111
= 01010101
  00000000 00000000 00000000 00000000
|                            01010101
= 00000000 00000000 00000000 01010101
!=00001010 10100000 00000000 00000000

Pass 2:
size: 00001010 10100000 00000000 00000000
shift: 14
b: 01010101
size:
  00001010 10100000 00000000 00000000
             010101 01000000 00000000
!=00001010 10110101 01000000 00000000

*/


/*
PROCESS FOR READING FRAMES:
- Get header, which will contain size information
- Get slice of bytes from end of header through point indicated by size
  (size - 10)

*/
