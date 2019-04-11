package main

import (
	"bufio"
	"fmt"
)


func v24GetManager(path string, reader *bufio.Reader) *Item {
	item := Item{ }
	item.Path = path
	item.ReadFrames = func () []ID3v2Frame {
		return v24ReadFrames(reader)
	}
	item.PrintFrames = v24PrintFrames
	return &item
}

func v24ReadFrames(reader *bufio.Reader) []ID3v2Frame {
	checkId := func (bytes []byte) bool {
		for _, byte := range bytes {
			if ((byte < 'A' || byte > 'Z') && (byte < '0' || byte > '9')) {
				return false
			}
		}
		return true
	}

	var frames []ID3v2Frame
	for areBytesOk(reader, 4, checkId) {
		header := ID3v2FrameHeader{ }
		header.Id = string(readBytes(reader, 4))
		header.Size = synchsafeBytesToInt(readBytes(reader, 4))
		// Need to parse and appropriately handle these flags  @TODO
		header.Flags = readBytes(reader, 2)

		frame := ID3v2Frame{ }
		frame.Header = header
		frame.Body = readBytes(reader, header.Size)

		frames = append(frames, frame)
	}

	return frames
}

func v24PrintFrames(frames []ID3v2Frame) {
	pull := func (part [2]string) (string, string) {
		return part[0], part[1]
	}

	keys := v24MakeFrameMap(pull)

	for _, frame := range frames {
		if frame.Header.Id[0:1] == "T" {
			fmt.Printf("%v: %v\n", keys[frame.Header.Id], parseString(frame.Body))
		} else {
			fmt.Printf("Frame is not text frame (%v)\n", frame.Header.Id)
		}
	}
}

// Full reference: http://id3.org/id3v2.4.0-frames
func v24MakeFrameMap(pull func([2]string) (string, string)) map[string]string {
	parts := [...][2]string{
		[2]string{"AENC", "Audio encryption"},
		[2]string{"APIC", "Attached picture"},
		[2]string{"ASPI", "Audio seek point index"},
		[2]string{"COMM", "Comments"},
		[2]string{"COMR", "Commercial frame"},
		[2]string{"ENCR", "Encryption method registration"},
		[2]string{"EQU2", "Equalisation (2)"},
		[2]string{"ETCO", "Event timing codes"},
		[2]string{"GEOB", "General encapsulated object"},
		[2]string{"GRID", "Group identification registration"},
		[2]string{"LINK", "Linked information"},
		[2]string{"MCDI", "Music CD identifier"},
		[2]string{"MLLT", "MPEG location lookup table"},
		[2]string{"OWNE", "Ownership frame"},
		[2]string{"PRIV", "Private frame"},
		[2]string{"PCNT", "Play counter"},
		[2]string{"POPM", "Popularimeter"},
		[2]string{"POSS", "Position synchronisation frame"},
		[2]string{"RBUF", "Recommended buffer size"},
		[2]string{"RVA2", "Relative volume adjustment (2)"},
		[2]string{"RVRB", "Reverb"},
		[2]string{"SEEK", "Seek frame"},
		[2]string{"SIGN", "Signature frame"},
		[2]string{"SYLT", "Synchronised lyric/text"},
		[2]string{"SYTC", "Synchronised tempo codes"},
		[2]string{"TALB", "Album/Movie/Show title"},
		[2]string{"TBPM", "BPM (beats per minute)"},
		[2]string{"TCOM", "Composer"},
		[2]string{"TCON", "Content type"},
		[2]string{"TCOP", "Copyright message"},
		[2]string{"TDEN", "Encoding time"},
		[2]string{"TDLY", "Playlist delay"},
		[2]string{"TDOR", "Original release time"},
		[2]string{"TDRC", "Recording time"},
		[2]string{"TDRL", "Release time"},
		[2]string{"TDTG", "Tagging time"},
		[2]string{"TENC", "Encoded by"},
		[2]string{"TEXT", "Lyricist/Text writer"},
		[2]string{"TFLT", "File type"},
		[2]string{"TIPL", "Involved people list"},
		[2]string{"TIT1", "Content group description"},
		[2]string{"TIT2", "Title/songname/content description"},
		[2]string{"TIT3", "Subtitle/Description refinement"},
		[2]string{"TKEY", "Initial key"},
		[2]string{"TLAN", "Language(s)"},
		[2]string{"TLEN", "Length"},
		[2]string{"TMCL", "Musician credits list"},
		[2]string{"TMED", "Media type"},
		[2]string{"TMOO", "Mood"},
		[2]string{"TOAL", "Original album/movie/show title"},
		[2]string{"TOFN", "Original filename"},
		[2]string{"TOLY", "Original lyricist(s)/text writer(s)"},
		[2]string{"TOPE", "Original artist(s)/performer(s)"},
		[2]string{"TOWN", "File owner/licensee"},
		[2]string{"TPE1", "Lead performer(s)/Soloist(s)"},
		[2]string{"TPE2", "Band/orchestra/accompaniment"},
		[2]string{"TPE3", "Conductor/performer refinement"},
		[2]string{"TPE4", "Interpreted, remixed, or otherwise modified by"},
		[2]string{"TPOS", "Part of a set"},
		[2]string{"TPRO", "Produced notice"},
		[2]string{"TPUB", "Publisher"},
		[2]string{"TRCK", "Track number/Position in set"},
		[2]string{"TRSN", "Internet radio station name"},
		[2]string{"TRSO", "Internet radio station owner"},
		[2]string{"TSOA", "Album sort order"},
		[2]string{"TSOP", "Performer sort order"},
		[2]string{"TSOT", "Title sort order"},
		[2]string{"TSRC", "ISRC (international standard recording code)"},
		[2]string{"TSSE", "Software/Hardware and settings used for encoding"},
		[2]string{"TSST", "Set subtitle"},
		[2]string{"TXXX", "User defined text information frame"},
		[2]string{"UFID", "Unique file identifier"},
		[2]string{"USER", "Terms of use"},
		[2]string{"USLT", "Unsynchronised lyric/text transcription"},
		[2]string{"WCOM", "Commercial information"},
		[2]string{"WCOP", "Copyright/Legal information"},
		[2]string{"WOAF", "Official audio file webpage"},
		[2]string{"WOAR", "Official artist/performer webpage"},
		[2]string{"WOAS", "Official audio source webpage"},
		[2]string{"WORS", "Official Internet radio station homepage"},
		[2]string{"WPAY", "Payment"},
		[2]string{"WPUB", "Publishers official webpage"},
		[2]string{"WXXX", "User defined URL link frame"},
	}

	return makeMap(parts[:], pull)
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
