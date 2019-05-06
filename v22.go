package main

import (
	"bufio"
	"fmt"
)

// http://id3.org/id3v2-00


const V22TAGIDSIZE int = 3
const V22TAGSIZESIZE int = 3


func v22MakeElement(path string, reader *bufio.Reader) *Element {
	element := Element{ }
	element.Path = path
	element.FillTagHeader = v22FillTagHeader
	element.ReadFrames = func () []ID3v2Frame {
		return v22ReadFrames(reader)
	}
	element.PrintFrames = v22PrintFrames
	//element.IsFrameEditable = makeFrameValidator(V22TAGIDSIZE)
	return &element
}

func v22FillTagHeader(header *ID3v2TagHeader, data []byte) {
	header.Unsynchronization = isBitOn(data[5], 7)
	header.Compression = isBitOn(data[5], 6)
}

func v22ReadFrames(reader *bufio.Reader) []ID3v2Frame {
	var frames []ID3v2Frame
	for areBytesOk(reader, V22TAGIDSIZE, areBytesValidFrameId) {
		header := v22ReadFrameHeader(reader)
		frames = append(frames, makeTagFrame(reader, header))
	}
	return frames
}

func v22ReadFrameHeader(reader *bufio.Reader) ID3v2FrameHeader {
	header := ID3v2FrameHeader{ }
	header.Id = string(readBytes(reader, V22TAGIDSIZE))
	header.Size = bytesToInt(readBytes(reader, V22TAGSIZESIZE))
	return header
}

func v22PrintFrames(frames []ID3v2Frame) {
	pull := func (part [2]string) (string, string) {
		return part[0], part[1]
	}

	keys := v22MakeFrameMap(pull)

	for _, frame := range frames {
		if v22IsFrameEditable(keys, frame) {
			fmt.Printf("%v: %v\n", keys[frame.Header.Id], parseString(frame.Body))
		}//  else {
		// 	fmt.Printf("Frame is not text frame (%v)\n", frame.Header.Id)
		// }
	}
}

// This function could be replaced with `makeFrameValidator`
func v22IsFrameEditable(keys map[string]string, frame ID3v2Frame) bool {
	if ((len(frame.Header.Id) == V22TAGIDSIZE) &&
		((frame.Header.Id[0:1] == "T") || (frame.Header.Id[0:1] == "W"))) {
		_, present := keys[frame.Header.Id]
		return present
	} else {
		return false
	}
}

// Full reference: http://id3.org/id3v2-00
func v22MakeFrameMap(pull func([2]string) (string, string)) map[string]string {
	parts := [...][2]string{
		[2]string{"BUF", "Recommended buffer size"},
		[2]string{"CNT", "Play counter"},
		[2]string{"COM", "Comments"},
		[2]string{"CRA", "Audio encryption"},
		[2]string{"CRM", "Encrypted meta frame"},
		[2]string{"ETC", "Event timing codes"},
		[2]string{"EQU", "Equalization"},
		[2]string{"GEO", "General encapsulated object"},
		[2]string{"IPL", "Involved people list"},
		[2]string{"LNK", "Linked information"},
		[2]string{"MCI", "Music CD Identifier"},
		[2]string{"MLL", "MPEG location lookup table"},
		[2]string{"PIC", "Attached picture"},
		[2]string{"POP", "Popularimeter"},
		[2]string{"REV", "Reverb"},
		[2]string{"RVA", "Relative volume adjustment"},
		[2]string{"SLT", "Synchronized lyric/text"},
		[2]string{"STC", "Synced tempo codes"},

		// Text frames.
		[2]string{"TAL", "Album/Movie/Show title"},
		[2]string{"TBP", "BPM (Beats Per Minute)"},
		[2]string{"TCM", "Composer"},
		[2]string{"TCO", "Content type"},
		[2]string{"TCR", "Copyright message"},
		[2]string{"TDA", "Date"},
		[2]string{"TDY", "Playlist delay"},
		[2]string{"TEN", "Encoded by"},
		[2]string{"TFT", "File type"},
		[2]string{"TIM", "Time"},
		[2]string{"TKE", "Initial key"},
		[2]string{"TLA", "Language(s)"},
		[2]string{"TLE", "Length"},
		[2]string{"TMT", "Media type"},
		[2]string{"TOA", "Original artist(s)/performer(s)"},
		[2]string{"TOF", "Original filename"},
		[2]string{"TOL", "Original Lyricist(s)/text writer(s)"},
		[2]string{"TOR", "Original release year"},
		[2]string{"TOT", "Original album/Movie/Show title"},
		[2]string{"TP1", "Lead artist(s)/Lead performer(s)/Soloist(s)/Performing group"},
		[2]string{"TP2", "Band/Orchestra/Accompaniment"},
		[2]string{"TP3", "Conductor/Performer refinement"},
		[2]string{"TP4", "Interpreted, remixed, or otherwise modified by"},
		[2]string{"TPA", "Part of a set"},
		[2]string{"TPB", "Publisher"},
		[2]string{"TRC", "ISRC (International Standard Recording Code)"},
		[2]string{"TRD", "Recording dates"},
		[2]string{"TRK", "Track number/Position in set"},
		[2]string{"TSI", "Size"},
		[2]string{"TSS", "Software/hardware and settings used for encoding"},
		[2]string{"TT1", "Content group description"},
		[2]string{"TT2", "Title/Songname/Content description"},
		[2]string{"TT3", "Subtitle/Description refinement"},
		[2]string{"TXT", "Lyricist/text writer"},
		[2]string{"TXX", "User defined text information frame"},  // special
		[2]string{"TYE", "Year"},

		[2]string{"UFI", "Unique file identifier"},
		[2]string{"ULT", "Unsychronized lyric/text transcription"},

		// URLs.
		[2]string{"WAF", "Official audio file webpage"},
		[2]string{"WAR", "Official artist/performer webpage"},
		[2]string{"WAS", "Official audio source webpage"},
		[2]string{"WCM", "Commercial information"},
		[2]string{"WCP", "Copyright/Legal information"},
		[2]string{"WPB", "Publishers official webpage"},
		[2]string{"WXX", "User defined URL link frame"},
	}

	return makeMap(parts[:], pull)
}
