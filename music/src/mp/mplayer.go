package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

)

func handleLibCommands(tokens []string) {
	switch tokens[1] {
	case "list":
		for i := 0; i < lib.Len(); i++ {

		}
	}
}

func handlePlayCommand(tokens []string) {
	if len(tokens) != 2 {
		fmt.Println("USAGE: play <name>")
		return
	}

	e := lib.Find(tokens[i])
	if e == nil {
		fmt.Println("The Music", tokens[1], "does not exist")
		return
	}

	mp.Play(e.source, e.Type, ctrl, signal)
}

func main() {
	lib = library.NewMusicManager()

	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter command->")
		rawLine,_, _ := r.ReadLine()
		line := string(rawLine)

		if line == "q" || line == "e" {
			break
		}

		tokens := strings.Split(line, " ")
		if tokens[0] == "lib" {
			handleLibCommands(tokens)
		} else if tokens[0] == "play" {
			handleLibCommands(tokens)
		} else {
			fmt.Println("Unrecognized command: ", tokens[0])
		}
	}
}