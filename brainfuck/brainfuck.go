package main

import (
	"bytes"
	"fmt"
	"github.com/swd543/interpret"
	"os"
	"runtime/debug"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Printf("Usage: %s filename\n", args[0])
		return
	}
	filename := args[1]
	reader, _ := os.Open(filename)
	var b bytes.Buffer
	state := interpret.NewState[int](reader, &b, nil)
	defer func() {
		if r := recover(); r != nil {
			_ = fmt.Errorf("something broke, state: %v", state)
			debug.PrintStack()
		}
	}()
	for {
		if symbol, err := state.GetNextSymbol(); err == nil {
			state.GetCommand(symbol)(state)
		} else {
			fmt.Println(b.String(), err)
			break
		}
	}
}
