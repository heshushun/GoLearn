package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
	//_, err := fmt.Scanln(&r.buffer)
	reader := bufio.NewReader(os.Stdin)
	res, _, err := reader.ReadLine()
	r.buffer = strings.TrimSpace(string(res))
	if err != nil {
		fmt.Printf("Error reading input %v \n", err)
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

type MetaCommandResult int32

const (
	META_COMMAND_SUCCESS MetaCommandResult = iota
	META_COMMAND_UNRECOGNIZED_COMMAND

	//META_COMMAND_SUCCESS MetaCommandResult = 0
	//META_COMMAND_UNRECOGNIZED_COMMAND MetaCommandResult = 1
)

type PrepareResult int32

const (
	PREPARE_SUCCESS PrepareResult = iota
	PREPARE_UNRECOGNIZED_STATEMENT
)

type StatementType int32

const (
	STATEMENT_INSERT StatementType = iota
	STATEMENT_SELECT
)

type Statement struct {
	statementType StatementType
}

func (r *InputBuffer) doMetaCommand() MetaCommandResult {
	if r.buffer == ".exit" {
		r.closeInputBuffer()
		os.Exit(0)
	} else {
		return META_COMMAND_UNRECOGNIZED_COMMAND
	}
	return META_COMMAND_SUCCESS
}

func (r *InputBuffer) prepareStatement(statement *Statement) PrepareResult {
	if strings.HasPrefix(r.buffer, "insert") {
		statement.statementType = STATEMENT_INSERT
		return PREPARE_SUCCESS
	}
	if strings.HasPrefix(r.buffer, "select") {
		statement.statementType = STATEMENT_SELECT
		return PREPARE_SUCCESS
	}
	return PREPARE_UNRECOGNIZED_STATEMENT
}

func (r *InputBuffer) executeStatement(statement *Statement) {
	switch statement.statementType {
	case STATEMENT_INSERT:
		fmt.Printf("This is where we would do an insert.\n")
	case STATEMENT_SELECT:
		fmt.Printf("This is where we would do a select.\n")
	default:
		fmt.Printf("no this statementType.\n")
	}
}

func main() {

	for {
		inputBuffer := NewInputBuffer()
		inputBuffer.printPrompt()
		inputBuffer.readInput()

		if len(inputBuffer.buffer) == 0 {
			continue
		}

		if inputBuffer.buffer[0] == '.' {
			switch inputBuffer.doMetaCommand() {
			case META_COMMAND_SUCCESS:
			case META_COMMAND_UNRECOGNIZED_COMMAND:
				fmt.Printf("Unrecognized command '%s'\n", inputBuffer.buffer)
				continue
			}
		}

		var statement Statement
		switch inputBuffer.prepareStatement(&statement) {
		case PREPARE_SUCCESS:
		case PREPARE_UNRECOGNIZED_STATEMENT:
			fmt.Printf("Unrecognized keyword at start of '%s'.\n", inputBuffer.buffer)
			continue
		}

		inputBuffer.executeStatement(&statement)
		fmt.Printf("Executed.\n")
	}
}
