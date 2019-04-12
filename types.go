package main

import (
)


type ID3v2Tag struct {
	Header ID3v2TagHeader
	Frames []ID3v2Frame
}

type ID3v2TagHeader struct {
	Version           int
	MinorVersion      int
	Compression       bool  // In v2.2
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
	// This could be a struct of booleans or just add booleans to
	// this struct like with the TagHeader.  @TODO
	Flags []byte
}

type Item struct {
	Path           string
	Tag            ID3v2Tag
	FillTagHeader  func(*ID3v2TagHeader, []byte)
	ReadFrames     func() []ID3v2Frame
	PrintFrames    func([]ID3v2Frame)
}
