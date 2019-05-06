package main

import (
	"bufio"
	"fmt"
	"os"
)


func main() {
	// The first argument is the program name.
	has_args := len(os.Args) > 1
	has_data := !isFileEmpty(os.Stdin)

	if (!(has_args || has_data)) {
		printUsage(os.Args[0])
		return
	}

	if has_args {
		actOnArgs(os.Args[1:])
	}

	if has_data {
		actOnStdin()
	}
}

func actOnArgs(args []string) {
	x := len(args) - 1
	for _, arg := range args {
		if (arg[0] == '-') {
			fmt.Printf("FLAG ARGUMENT '%v'\n", arg)
		} else {
			element, err := elementFromPath(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
				continue
			}

			printElementData(element)
			if x > 0 {
				fmt.Println()
			}
		}
		x--
	}
}

func actOnStdin() {
	lexer := newLexer(bufio.NewReader(os.Stdin))
	for lexer.More() {
		token, err := lexer.Next()
		if err != nil {
			panic(fmt.Sprintf("Error getting lexer's next token: %s", err))
		}
		fmt.Printf("Got token: %v\n", token)
	}
}

func printUsage(program_name string) {
	fmt.Printf("Usage: %s [path(s) to mp3 file]\n", program_name)
}
