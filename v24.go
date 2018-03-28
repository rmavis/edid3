package main

import (
	"bufio"
	"fmt"
)


func v24GetFrames(reader *bufio.Reader) []ID3v2Frame {
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
		header.Size = calcSynchsafe(readBytes(reader, 4))
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
	for _, frame := range frames {
		if frame.Header.Id[0:1] == "T" {
			v24PrintTextFrame(frame)
		} else {
			fmt.Printf("Frame is not text frame (%v)\n", frame.Header.Id)
		}
	}
}


func v24PrintTextFrame(frame ID3v2Frame) {
	var key string
	switch frame.Header.Id {
	case "TALB":  //  Album/Movie/Show title
		key = "Album"
	case "TBPM":  //  BPM (beats per minute)
		key = "BPM"
	case "TCOM":  //  Composer
		key = "Composer"
	case "TCON":  //  Content type
		key = "Content type"
	case "TCOP":  //  Copyright message
		key = "Copyright message"
	case "TDEN":  //  Encoding time
		key = "Encoding time"
	case "TDLY":  //  Playlist delay
		key = "Playlist delay"
	case "TDOR":  //  Original release time
		key = "Original release time"
	case "TDRC":  //  Recording time
		key = "Recording time"
	case "TDRL":  //  Release time
		key = "Release time"
	case "TDTG":  //  Tagging time
		key = "Tagging time"
	case "TENC":  //  Encoded by
		key = "Encoded by"
	case "TEXT":  //  Lyricist/Text writer
		key = "Lyricist/Text writer"
	case "TFLT":  //  File type
		key = "File type"
	case "TIPL":  //  Involved people list
		key = "Involved people list"
	case "TIT1":  //  Content group description
		key = "Content group description"
	case "TIT2":  //  Title/songname/content description
		key = "Title/songname/content description"
	case "TIT3":  //  Subtitle/Description refinement
		key = "Subtitle/Description refinement"
	case "TKEY":  //  Initial key
		key = "Initial key"
	case "TLAN":  //  Language(s)
		key = "Language(s)"
	case "TLEN":  //  Length
		key = "Length"
	case "TMCL":  //  Musician credits list
		key = "Musician credits list"
	case "TMED":  //  Media type
		key = "Media type"
	case "TMOO":  //  Mood
		key = "Mood"
	case "TOAL":  //  Original album/movie/show title
		key = "Original album/movie/show title"
	case "TOFN":  //  Original filename
		key = "Original filename"
	case "TOLY":  //  Original lyricist(s)/text writer(s)
		key = "Original lyricist(s)/text writer(s)"
	case "TOPE":  //  Original artist(s)/performer(s)
		key = "Original artist(s)/performer(s)"
	case "TOWN":  //  File owner/licensee
		key = "File owner/licensee"
	case "TPE1":  //  Lead performer(s)/Soloist(s)
		key = "Lead performer(s)/Soloist(s)"
	case "TPE2":  //  Band/orchestra/accompaniment
		key = "Band/orchestra/accompaniment"
	case "TPE3":  //  Conductor/performer refinement
		key = "Conductor/performer refinement"
	case "TPE4":  //  Interpreted, remixed, or otherwise modified by
		key = "Interpreted, remixed, or otherwise modified by"
	case "TPOS":  //  Part of a set
		key = "Part of a set"
	case "TPRO":  //  Produced notice
		key = "Produced notice"
	case "TPUB":  //  Publisher
		key = "Publisher"
	case "TRCK":  //  Track number/Position in set
		key = "Track number/Position in set"
	case "TRSN":  //  Internet radio station name
		key = "Internet radio station name"
	case "TRSO":  //  Internet radio station owner
		key = "Internet radio station owner"
	case "TSOA":  //  Album sort order
		key = "Album sort order"
	case "TSOP":  //  Performer sort order
		key = "Performer sort order"
	case "TSOT":  //  Title sort order
		key = "Title sort order"
	case "TSRC":  //  ISRC (international standard recording code)
		key = "ISRC (international standard recording code)"
	case "TSSE":  //  Software/Hardware and settings used for encoding
		key = "Software/Hardware and settings used for encoding"
	case "TSST":  //  Set subtitle
		key = "Set subtitle"
	case "TXXX":  //  User defined text information frame
		key = "User defined text information frame"
	default:
		key = frame.Header.Id
	}
	fmt.Printf("%v: %v\n", key, parseString(frame.Body))
}


/*

//  Audio encryption
"AENC" = null,
//  Attached picture
"APIC" = null,
//  Audio seek point index
"ASPI" = null,

//  Comments
"COMM" = null,
//  Commercial frame
"COMR" = null,

//  Encryption method registration
"ENCR" = null,
//  Equalisation (2)
"EQU2" = null,
//  Event timing codes
"ETCO" = null,

//  General encapsulated object
"GEOB" = null,
//  Group identification registration
"GRID" = null,

//  Linked information
"LINK" = null,

//  Music CD identifier
"MCDI" = null,
//  MPEG location lookup table
"MLLT" = null,

//  Ownership frame
"OWNE" = null,

//  Private frame
"PRIV" = null,
//  Play counter
"PCNT" = null,
//  Popularimeter
"POPM" = null,
//  Position synchronisation frame
"POSS" = null,

//  Recommended buffer size
"RBUF" = null,
//  Relative volume adjustment (2)
"RVA2" = null,
//  Reverb
"RVRB" = null,

//  Seek frame
"SEEK" = null,
//  Signature frame
"SIGN" = null,
//  Synchronised lyric/text
"SYLT" = null,
//  Synchronised tempo codes
"SYTC" = null,

//  Unique file identifier
"UFID" = null,
//  Terms of use
"USER" = null,
//  Unsynchronised lyric/text transcription
"USLT" = null,

//  Commercial information
"WCOM" = null,
//  Copyright/Legal information
"WCOP" = null,
//  Official audio file webpage
"WOAF" = null,
//  Official artist/performer webpage
"WOAR" = null,
//  Official audio source webpage
"WOAS" = null,
//  Official Internet radio station homepage
"WORS" = null,
//  Payment
"WPAY" = null,
//  Publishers official webpage
"WPUB" = null,
//  User defined URL link frame
"WXXX" = null,

*/
