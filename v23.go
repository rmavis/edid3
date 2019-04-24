package main

import (
	"bufio"
	"fmt"
)

// http://id3.org/id3v2.3.0


const V23TAGIDSIZE int = 4
const V23TAGSIZESIZE int = 4
const V23TAGFLAGSSIZE int = 2


func v23MakeItem(path string, reader *bufio.Reader) *Item {
	item := Item{ }
	item.Path = path
	item.FillTagHeader = v23FillTagHeader
	item.ReadFrames = func () []ID3v2Frame {
		return v23ReadFrames(reader)
	}
	item.PrintFrames = v23PrintFrames
	return &item
}

func v23FillTagHeader(header *ID3v2TagHeader, data []byte) {
	header.Unsynchronization = isBitOn(data[5], 7)
	header.Extended = isBitOn(data[5], 6)
	header.Experimental = isBitOn(data[5], 5)
}

func v23ReadFrames(reader *bufio.Reader) []ID3v2Frame {
	var frames []ID3v2Frame
	for areBytesOk(reader, V23TAGIDSIZE, areBytesValidFrameId) {
		header := v23ReadFrameHeader(reader)
		frames = append(frames, makeTagFrame(reader, header))
	}
	return frames
}

func v23ReadFrameHeader(reader *bufio.Reader) ID3v2FrameHeader {
	header := ID3v2FrameHeader{ }
	header.Id = string(readBytes(reader, V23TAGIDSIZE))
	header.Size = bytesToInt(readBytes(reader, V23TAGSIZESIZE))
	// Need to parse and appropriately handle these flags  @TODO
	header.Flags = readBytes(reader, V23TAGFLAGSSIZE)
	return header
}

func v23PrintFrames(frames []ID3v2Frame) {
	pull := func (part [2]string) (string, string) {
		return part[0], part[1]
	}

	keys := v23MakeFrameMap(pull)

	for _, frame := range frames {
		if v23IsFrameEditable(keys, frame) {
			fmt.Printf("%v: %v\n", keys[frame.Header.Id], parseString(frame.Body))
		}// else {
		// 	fmt.Printf("Frame is not text frame (%v)\n", frame.Header.Id)
		// }
	}
}

// This function could be replaced with `makeFrameValidator`
func v23IsFrameEditable(keys map[string]string, frame ID3v2Frame) bool {
	if ((len(frame.Header.Id) == V23TAGIDSIZE) &&
		((frame.Header.Id[0:1] == "T") || (frame.Header.Id[0:1] == "W"))) {
		_, present := keys[frame.Header.Id]
		return present
	} else {
		return false
	}
}

func v23MakeFrameMap(pull func([2]string) (string, string)) map[string]string {
	parts := [...][2]string{
		[2]string{"AENC", "Audio encryption"},
		[2]string{"APIC", "Attached picture"},
		[2]string{"COMM", "Comments"},
		[2]string{"COMR", "Commercial frame"},
		[2]string{"ENCR", "Encryption method registration"},
		[2]string{"EQUA", "Equalization"},
		[2]string{"ETCO", "Event timing codes"},
		[2]string{"GEOB", "General encapsulated object"},
		[2]string{"GRID", "Group identification registration"},
		[2]string{"IPLS", "Involved people list"},
		[2]string{"LINK", "Linked information"},
		[2]string{"MCDI", "Music CD identifier"},
		[2]string{"MLLT", "MPEG location lookup table"},
		[2]string{"OWNE", "Ownership frame"},
		[2]string{"PRIV", "Private frame"},
		[2]string{"PCNT", "Play counter"},
		[2]string{"POPM", "Popularimeter"},
		[2]string{"POSS", "Position synchronisation frame"},
		[2]string{"RBUF", "Recommended buffer size"},
		[2]string{"RVAD", "Relative volume adjustment"},
		[2]string{"RVRB", "Reverb"},
		[2]string{"SYLT", "Synchronized lyric/text"},
		[2]string{"SYTC", "Synchronized tempo codes"},

		// Text frames.
		[2]string{"TALB", "Album/Movie/Show title"},
		[2]string{"TBPM", "BPM (beats per minute)"},
		[2]string{"TCOM", "Composer"},
		[2]string{"TCON", "Content type"},
		[2]string{"TCOP", "Copyright message"},
		[2]string{"TDAT", "Date"},
		[2]string{"TDLY", "Playlist delay"},
		[2]string{"TENC", "Encoded by"},
		[2]string{"TEXT", "Lyricist/Text writer"},
		[2]string{"TFLT", "File type"},
		[2]string{"TIME", "Time"},
		[2]string{"TIT1", "Content group description"},
		[2]string{"TIT2", "Title/songname/content description"},
		[2]string{"TIT3", "Subtitle/Description refinement"},
		[2]string{"TKEY", "Initial key"},
		[2]string{"TLAN", "Language(s)"},
		[2]string{"TLEN", "Length"},
		[2]string{"TMED", "Media type"},
		[2]string{"TOAL", "Original album/movie/show title"},
		[2]string{"TOFN", "Original filename"},
		[2]string{"TOLY", "Original lyricist(s)/text writer(s)"},
		[2]string{"TOPE", "Original artist(s)/performer(s)"},
		[2]string{"TORY", "Original release year"},
		[2]string{"TOWN", "File owner/licensee"},
		[2]string{"TPE1", "Lead performer(s)/Soloist(s)"},
		[2]string{"TPE2", "Band/orchestra/accompaniment"},
		[2]string{"TPE3", "Conductor/performer refinement"},
		[2]string{"TPE4", "Interpreted, remixed, or otherwise modified by"},
		[2]string{"TPOS", "Part of a set"},
		[2]string{"TPUB", "Publisher"},
		[2]string{"TRCK", "Track number/Position in set"},
		[2]string{"TRDA", "Recording dates"},
		[2]string{"TRSN", "Internet radio station name"},
		[2]string{"TRSO", "Internet radio station owner"},
		[2]string{"TSIZ", "Size"},
		[2]string{"TSRC", "ISRC (international standard recording code)"},
		[2]string{"TSSE", "Software/Hardware and settings used for encoding"},
		[2]string{"TYER", "Year"},
		[2]string{"TXXX", "User defined text information frame"},  // special

		[2]string{"UFID", "Unique file identifier"},
		[2]string{"USER", "Terms of use"},
		[2]string{"USLT", "Unsychronized lyric/text transcription"},

		// URLs.
		[2]string{"WCOM", "Commercial information"},
		[2]string{"WCOP", "Copyright/Legal information"},
		[2]string{"WOAF", "Official audio file webpage"},
		[2]string{"WOAR", "Official artist/performer webpage"},
		[2]string{"WOAS", "Official audio source webpage"},
		[2]string{"WORS", "Official internet radio station homepage"},
		[2]string{"WPAY", "Payment"},
		[2]string{"WPUB", "Publishers official webpage"},
		[2]string{"WXXX", "User defined URL link frame"},
	}

	return makeMap(parts[:], pull)
}
