package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)


func newLexer(reader *bufio.Reader) Lexer {
	lexer := Lexer{Reader: reader}
	return lexer
}

func (lexer *Lexer) More() bool {
	err := lexer.DiscardWhitespace()
	if err != nil {
		panic(err)
	}
	return !lexer.EOF()
}

func (lexer *Lexer) EOF() bool {
	_, err := lexer.Reader.Peek(1)
	if ((err != nil) && (err != io.EOF)) {
		panic(fmt.Sprintf("Error checking for lexer EOF: %s", err))
	}
	return err == io.EOF
}

func (lexer *Lexer) Next() (Token, error) {
	if !lexer.More() {
		return lexer.MakeToken(TokenEOF, ""), nil
	}

	byte, err := lexer.Reader.ReadByte()
	// The error won't be io.EOF, which is checked by calling `More`.
	if err != nil {
		return lexer.UnknownToken(), errors.New(fmt.Sprintf("Error while reading next byte: %s", err))
	}

	char := string(byte)
	if char == "[" {
		// A filepath looks like [/path/to/file].
		// The closing bracket is optional. 
		return lexer.ReadFilePath()
	} else if ((char == "#") || (char == "]")) {
		err := lexer.IgnoreToEOL()
		if err == nil {
			return lexer.Next()
		} else {
			return lexer.UnknownToken(), err
		}
	} else if char == ":" {
		err := lexer.DiscardWhitespace()
		if err == io.EOF {
			return lexer.UnknownToken(), nil
		} else if err != nil {
			return lexer.UnknownToken(), errors.New(fmt.Sprintf("Error while reading whitespace: %s", err))
		}
		return lexer.ReadFieldValue()
	} else {
		// Make the read byte readable by the reader's next read.
		err := lexer.Reader.UnreadByte()
		if err != nil {
			panic("Unexpected error rewinding reader.")
		}
		return lexer.ReadFieldKey()
	}
}

func (lexer *Lexer) ReadFilePath() (Token, error) {
	var token Token
	check := func (char string) bool {
		return !((char == "]") || (char == "\n"))
	}
	path, err := lexer.ReadWhile(check)
	if ((err != nil) && (err != io.EOF)) {
		return token, err
	}
	return lexer.MakeToken(TokenFilePath, path), nil
}

func (lexer *Lexer) ReadFieldKey() (Token, error) {
	var token Token
	check := func (char string) bool {
		return !((char == ":") || (char == "\n"))
	}
	key, err := lexer.ReadWhile(check)
	if ((err != nil) && (err != io.EOF)) {
		return token, err
	}
	byte, err := lexer.Reader.Peek(1)
	if ((err != nil) && (err != io.EOF)) {
		return token, errors.New(fmt.Sprintf("Error peeking byte following field key: %s", err))
	}
	if string(byte) == ":" {
		return lexer.MakeToken(TokenFieldKey, key), nil
	} else {
		return lexer.MakeToken(TokenUnknown, key), nil
	}
}

func (lexer *Lexer) ReadFieldValue() (Token, error) {
	byte, err := lexer.Reader.Peek(1)
	if err != nil {
		return Token{ }, err
	}

	char := string(byte)
	if ((char == "\"") || (char == "'")) {
		bytes, err := lexer.Reader.ReadBytes(byte[0])
		return lexer.MakeToken(TokenFieldValue, string(bytes[0: (len(bytes) - 1)])), err
	} else {
		bytes, err := lexer.Reader.ReadBytes('\n')
		return lexer.MakeToken(TokenFieldValue, string(bytes[0: (len(bytes) - 1)])), err
	}
}

func (lexer *Lexer) DiscardWhitespace() error {
	check := func (char string) bool {
		return ((char == " ") || (char == "\t") || (char == "\n"))
	}
	_, err := lexer.ReadWhile(check)
	return err
}

func (lexer *Lexer) ReadWhile(wantChar func(string) bool) (string, error) {
	var str strings.Builder
	for {
		byte, err := lexer.Reader.Peek(1)
		if err == io.EOF {
			break
		} else if err != nil {
			return "", errors.New(fmt.Sprintf("Error while reading byte: %s", err))
		}

		char := string(byte)
		if (wantChar(char)) {
			fmt.Fprintf(&str, char)
			_, err := lexer.Reader.ReadByte()
			if err != nil {
				panic("Unexpected error advancing reader.")
			}
		} else {
			break
		}
	}
	return str.String(), nil
}

func (lexer *Lexer) IgnoreToEOL() error {
	_, err := lexer.Reader.ReadBytes('\n')
	return err
}

func (lexer *Lexer) UnknownToken() Token {
	return lexer.MakeToken(TokenUnknown, "")
}

func (lexer *Lexer) MakeToken(kind TokenType, value string) Token {
	token := Token{
		Type: kind,
		Value: value,
	}
	return token
}
