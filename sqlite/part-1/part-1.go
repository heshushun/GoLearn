package main

import (
	"fmt"
	"os"
)

type InputBuffer struct {
	buffer       string
	bufferLength int
	inputLength  int
}

func NewInputBuffer() *InputBuffer {
	inputBuffer := &InputBuffer{
		bufferLength: 0,
		inputLength:  0,
	}
	return inputBuffer
}

func (r *InputBuffer) printPrompt() {
	fmt.Printf("db > ")
}

func (r *InputBuffer) readInput() {
	_, err := fmt.Scanln(&r.buffer)
	if err != nil || len(r.buffer) <= 0 {
		fmt.Printf("Error reading input\n")
		os.Exit(0)
	}

	r.bufferLength = len(r.buffer)
	r.inputLength = len(r.buffer)
}

func (r *InputBuffer) closeInputBuffer() {
	r.buffer = ""
	r.bufferLength = 0
	r.inputLength = 0
}

func main() {
	for {
		inputBuffer := NewInputBuffer()
		inputBuffer.printPrompt()
		inputBuffer.readInput()

		if inputBuffer.buffer == ".exit" {
			inputBuffer.closeInputBuffer()
			os.Exit(0)
		} else {
			fmt.Printf("Unrecognized command '%s'.\n", inputBuffer.buffer)
		}
	}
}
