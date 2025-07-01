package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	initCommands()
	cfg := newConfig()

	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		if !sc.Scan() {
			break
		}
		words := cleanInput(sc.Text())
		if len(words) == 0 {
			continue
		}
		if cmd, ok := commands[words[0]]; ok {
			if err := cmd.callback(cfg); err != nil {
				fmt.Println("Error:", err)
			}
		} else {
			fmt.Println("Unknown command:", words[0])
		}
	}
}
