package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)


func readInputTokens(reader *bufio.Reader) ([]Token, error) {
	var tokens []Token

	for {
		err := parserDiscardWhitespace(reader)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.New(fmt.Sprintf("Error while reading whitespace: %s", err))
		}

		byte, err := reader.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.New(fmt.Sprintf("Error while reading next byte: %s", err))
		}

		char := string(byte)
		if char == "[" {
			// A filepath looks like [/path/to/file].
			// The closing bracket is optional. 
			token, err := parserReadFilePath(reader)
			if token.Value != "" {
				tokens = append(tokens, token)
			}
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
		} else if ((char == "#") || (char == "]")) {
			err := parserDiscardToEOL(reader)
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
		} else if char == ":" {
			// read field value
			err := parserDiscardWhitespace(reader)
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, errors.New(fmt.Sprintf("Error while reading whitespace: %s", err))
			}
			token, err := parserReadFieldValue(reader)
			if token.Value != "" {
				tokens = append(tokens, token)
			}
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
		} else {
			// Make the read byte readable by the reader's next read.
			err := reader.UnreadByte()
			if err != nil {
				panic("Unexpected error rewinding reader.")
			}
			token, err := parserReadFieldKey(reader)
			if token.Value != "" {
				tokens = append(tokens, token)
			}
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
		}
	}

	return tokens, nil
}

func parserReadFilePath(reader *bufio.Reader) (Token, error) {
	var token Token
	check := func (char string) bool {
		return !((char == "]") || (char == "\n"))
	}
	path, err := parserReadWhile(reader, check)
	if ((err != nil) && (err != io.EOF)) {
		return token, err
	}
	return parserMakeToken(TokenFilePath, path), err
}

func parserReadFieldKey(reader *bufio.Reader) (Token, error) {
	var token Token
	check := func (char string) bool {
		return !((char == ":") || (char == "\n"))
	}
	key, err := parserReadWhile(reader, check)
	if ((err != nil) && (err != io.EOF)) {
		return token, err
	}
	return parserMakeToken(TokenFieldKey, key), err
}

func parserReadFieldValue(reader *bufio.Reader) (Token, error) {
	byte, err := reader.Peek(1)
	if err != nil {
		return Token{ }, err
	}

	char := string(byte)
	if ((char == "\"") || (char == "'")) {
		bytes, err := reader.ReadBytes(byte[0])
		return parserMakeToken(TokenFieldValue, string(bytes[0: (len(bytes) - 1)])), err
	} else {
		bytes, err := reader.ReadBytes('\n')
		return parserMakeToken(TokenFieldValue, string(bytes[0: (len(bytes) - 1)])), err
	}
}

func parserDiscardWhitespace(reader *bufio.Reader) error {
	check := func (char string) bool {
		return ((char == " ") || (char == "\t") || (char == "\n"))
	}
	_, err := parserReadWhile(reader, check)
	return err
}

func parserReadWhile(reader *bufio.Reader, wantChar func(string) bool) (string, error) {
	var str strings.Builder
	for {
		byte, err := reader.Peek(1)
		if err == io.EOF {
			return str.String(), err
		} else if err != nil {
			return "", errors.New(fmt.Sprintf("Error while reading byte: %s", err))
		}

		char := string(byte)
		if (wantChar(char)) {
			fmt.Fprintf(&str, char)
			_, err := reader.ReadByte()
			if err != nil {
				panic("Unexpected error advancing reader.")
			}
		} else {
			break
		}
	}
	return str.String(), nil
}

func parserDiscardToEOL(reader *bufio.Reader) error {
	_, err := reader.ReadBytes('\n')
	return err
}

func parserMakeToken(kind TokenType, value string) Token {
	token := Token{
		Type: kind,
		Value: value,
	}
	return token
}
