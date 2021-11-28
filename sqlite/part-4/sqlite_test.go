package main

import (
	"testing"
)

func Test_Insert(t *testing.T) {
	table := NewTable()
	for {
		inputBuffer := NewInputBuffer()
		inputBuffer.printPrompt()
		inputBuffer.readInput()

		if len(inputBuffer.buffer) == 0 {
			continue
		}

		// input
		if inputBuffer.buffer[0] == '.' {
			switch inputBuffer.doMetaCommand(table) {
			case META_COMMAND_SUCCESS:
			case META_COMMAND_UNRECOGNIZED_COMMAND:
				t.Error("Test_Insert error")
				continue
			}
		}

		// prepare
		var statement Statement
		switch inputBuffer.prepareStatement(&statement) {
		case PREPARE_SUCCESS:
		case PREPARE_NEGATIVE_ID:
			t.Error("ID must be positive.\n")
			continue
		case PREPARE_STRING_TOO_LONG:
			t.Error("String is too long.\n")
			continue
		case PREPARE_SYNTAX_ERROR:
			t.Error("Syntax error. Could not parse statement.\n")
			continue
		case PREPARE_UNRECOGNIZED_STATEMENT:
			t.Errorf("Unrecognized keyword at start of '%s'.\n", inputBuffer.buffer)
			continue
		}

		// execute
		switch table.executeStatement(&statement) {
		case EXECUTE_SUCCESS:
		case EXECUTE_TABLE_FULL:
			t.Error("Error: Table full.\n")
		}

	}
}
