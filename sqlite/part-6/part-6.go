package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

/*
*
* const
*
**/
const (
	COLUMN_USERNAME_SIZE = 32  // USERNAME字段大小
	COLUMN_EMAIL_SIZE    = 255 // EMAIL字段大小

	ID_SIZE       = 4 // int32 字节是4
	USERNAME_SIZE = COLUMN_USERNAME_SIZE
	EMAIL_SIZE    = COLUMN_EMAIL_SIZE

	ID_OFFSET       = 0
	USERNAME_OFFSET = ID_OFFSET + ID_SIZE
	EMAIL_OFFSET    = USERNAME_OFFSET + USERNAME_SIZE

	ROW_SIZE = ID_SIZE + USERNAME_SIZE + EMAIL_SIZE

	PAGE_SIZE       = 4096
	TABLE_MAX_PAGES = 100
	ROWS_PER_PAGE   = PAGE_SIZE / ROW_SIZE
	TABLE_MAX_ROWS  = ROWS_PER_PAGE * TABLE_MAX_PAGES

	TABLE_SIZE = PAGE_SIZE * TABLE_MAX_PAGES
)

/*
*
* enum
*
**/
type ExecuteResult int32

const (
	EXECUTE_SUCCESS ExecuteResult = iota
	EXECUTE_TABLE_FULL
)

type MetaCommandResult int32

const (
	META_COMMAND_SUCCESS MetaCommandResult = iota
	META_COMMAND_UNRECOGNIZED_COMMAND
)

type PrepareResult int32

const (
	PREPARE_SUCCESS PrepareResult = iota
	PREPARE_STRING_TOO_LONG
	PREPARE_NEGATIVE_ID
	PREPARE_SYNTAX_ERROR
	PREPARE_UNRECOGNIZED_STATEMENT
)

type StatementType int32

const (
	STATEMENT_INSERT StatementType = iota
	STATEMENT_SELECT
)

/*
*
* Statement
*
**/
type Statement struct {
	statementType StatementType
	rowToInsert   Row //only used by insert statement
}

/*
*
* InputBuffer
*
**/
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

func (r *InputBuffer) doMetaCommand(table *Table) MetaCommandResult {
	if r.buffer == ".exit" {
		r.closeInputBuffer()
		table.freeTable()
		os.Exit(0)
	} else {
		return META_COMMAND_UNRECOGNIZED_COMMAND
	}
	return META_COMMAND_SUCCESS
}

func (r *InputBuffer) prepareStatement(statement *Statement) PrepareResult {
	if strings.HasPrefix(r.buffer, "insert") {
		return r.prepareInsert(statement)
	}
	if strings.HasPrefix(r.buffer, "select") {
		statement.statementType = STATEMENT_SELECT
		return PREPARE_SUCCESS
	}
	return PREPARE_UNRECOGNIZED_STATEMENT
}

func (r *InputBuffer) prepareInsert(statement *Statement) PrepareResult {
	statement.statementType = STATEMENT_INSERT

	insertList := strings.Fields(r.buffer)
	//insertList := strings.Split(r.buffer, " ")

	if len(insertList) < 4 {
		return PREPARE_SYNTAX_ERROR
	}

	idString := strings.TrimSpace(insertList[1])
	username := strings.TrimSpace(insertList[2])
	email := strings.TrimSpace(insertList[3])

	if idString == "" || username == "" || email == "" {
		return PREPARE_SYNTAX_ERROR
	}
	id, err := strconv.ParseInt(idString, 10, 32)
	if err != nil || id < 0 {
		return PREPARE_NEGATIVE_ID
	}
	if len(username) > COLUMN_USERNAME_SIZE {
		return PREPARE_STRING_TOO_LONG
	}
	if len(email) > COLUMN_EMAIL_SIZE {
		return PREPARE_STRING_TOO_LONG
	}

	statement.rowToInsert.id = int32(id)

	statement.rowToInsert.username = username

	statement.rowToInsert.email = email

	return PREPARE_SUCCESS
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

/*
*
* Row
*
**/
type Row struct {
	id       int32
	username string
	email    string
}

func (r *Row) printRow() {
	fmt.Printf("(%d, %s, %s)\n", r.id, r.username, r.email)
}

func (r *Row) serializeRow(destination []byte) []byte {
	// TODO 序列化
	destination = memcpy(destination, Int32ToBytes(r.id, ID_SIZE), ID_OFFSET, 0, ID_SIZE)
	destination = memcpy(destination, StringToBytes(r.username, USERNAME_SIZE), USERNAME_OFFSET, 0, USERNAME_SIZE)
	destination = memcpy(destination, StringToBytes(r.email, EMAIL_SIZE), EMAIL_OFFSET, 0, EMAIL_SIZE)
	return destination
}

func (r *Row) deserializeRow(src []byte) {
	// TODO 反序列化
	r.id = BytesToInt32(memcpy(make([]byte, ID_SIZE), src, 0, ID_OFFSET, ID_SIZE))
	r.username = BytesToString(memcpy(make([]byte, USERNAME_SIZE), src, 0, USERNAME_OFFSET, USERNAME_SIZE))
	r.email = BytesToString(memcpy(make([]byte, EMAIL_SIZE), src, 0, EMAIL_OFFSET, EMAIL_SIZE))
}

/*
*
* Pager
*
**/
type Pager struct {
	// TODO
	file       *os.File
	fileFD     uintptr
	fileName   string
	fileLength int64
	pages      [][]byte
}

func NewPager(fileName string) *Pager {
	pager := &Pager{}

	_, b := IsFile(fileName)
	var file *os.File
	var err error
	if b {
		file, err = os.OpenFile(fileName, os.O_RDWR, 0666)
	} else {
		file, err = os.Create(fileName)
	}
	if err != nil {
		fmt.Printf("Unable to open file\n")
		os.Exit(0)
	}

	// defer pager.closeFile(file)

	fileLength, _ := file.Seek(0, io.SeekEnd)
	pager.fileLength = fileLength
	pager.file = file
	pager.fileFD = file.Fd()
	pager.fileName = file.Name()
	pages := make([][]byte, TABLE_MAX_PAGES)
	pager.pages = pages

	return pager
}

func (r *Pager) getPage(pageNum int) []byte {
	if pageNum > TABLE_MAX_PAGES {
		fmt.Printf("Tried to fetch page number out of bounds. %d > %d\n", pageNum, TABLE_MAX_PAGES)
		os.Exit(0)
	}

	page := r.pages[pageNum]
	if page == nil {
		r.pages[pageNum] = make([]byte, PAGE_SIZE)

		// Cache miss. Allocate memory and load from file.
		numPages := r.fileLength / PAGE_SIZE

		// We might save a partial page at the end of the file
		if r.fileLength%PAGE_SIZE != 0 {
			numPages += 1
		}

		if pageNum <= int(numPages) {
			// TODO
			//file := os.NewFile(r.fileFD, r.fileName)
			//defer r.closeFile(file)

			// 偏移到头部
			_, _ = r.file.Seek(int64(pageNum)*PAGE_SIZE, io.SeekStart)
			// 文件读到内存
			buf := make([]byte, PAGE_SIZE)
			re := bufio.NewReader(r.file)
			n, err := re.Read(buf)
			if (err != nil && err != io.EOF) || n < 0 {
				fmt.Printf("Error reading file: %v \n", err)
				os.Exit(0)
			}
			copy(r.pages[pageNum], buf)
		}
	}
	return r.pages[pageNum]
}

func (r *Pager) pagerFlush(pageNum, size int) {
	page := r.pages[pageNum]
	if page == nil {
		fmt.Printf("Tried to flush null page\n")
		os.Exit(0)
	}

	// TODO
	//file := os.NewFile(r.fileFD, r.fileName)
	//defer r.closeFile(file)

	offset, err := r.file.Seek(int64(pageNum*PAGE_SIZE), io.SeekStart)
	if err != nil || offset < 0 {
		fmt.Printf("Tried to flush null page\n")
		os.Exit(0)
	}

	// 内存写入文件
	n, err2 := r.file.Write(page[:size])
	if (err2 != nil && err2 != io.EOF) || n < 0 {
		fmt.Printf("Tried to flush null page  %v \n", err2)
		os.Exit(0)
	}
}

func (r *Pager) closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		fmt.Printf("Error closing db file.\n")
		os.Exit(0)
	}
}

func (r *Pager) freePager() {
	// TODO
	r.closeFile(r.file)

	r.fileFD = 0
	r.fileName = ""
	r.fileLength = 0
	r.pages = [][]byte{}
}

/*
*
* Table
*
**/
type Table struct {
	numRows int
	pager   *Pager
}

func NewTable(fileName string) *Table {
	pager := NewPager(fileName)
	numRows := int(pager.fileLength / ROW_SIZE)

	table := &Table{}
	table.pager = pager
	table.numRows = numRows
	return table
}

func (r *Table) rowSlot(rowNum int) (int, int) {
	pageNum := rowNum / ROWS_PER_PAGE // 第几页
	_ = r.pager.getPage(pageNum)

	rowOffset := rowNum % ROWS_PER_PAGE // row偏移
	byteOffset := rowOffset * ROW_SIZE  // byte偏移

	return pageNum, byteOffset
}

func (r *Table) freeTable() {
	pager := r.pager

	numFullPages := r.numRows / ROWS_PER_PAGE // 满页数量
	for i := 0; i < numFullPages; i++ {
		if pager.pages[i] == nil {
			continue
		}
		pager.pagerFlush(i, PAGE_SIZE)
	}

	numAdditionalRows := r.numRows % ROWS_PER_PAGE // 残页行数
	if numAdditionalRows > 0 {
		pageNum := numFullPages
		if pager.pages[pageNum] != nil {
			pager.pagerFlush(pageNum, numAdditionalRows*ROW_SIZE)
		}
	}

	r.numRows = 0
	pager.freePager()
}

func (r *Table) executeInsert(statement *Statement) ExecuteResult {
	// TODO 执行insert
	if r.numRows >= TABLE_MAX_ROWS {
		return EXECUTE_TABLE_FULL
	}
	insertRow := &statement.rowToInsert

	cursor := TableEnd(r)
	pageNum, byteOffset := cursor.cursorValue()
	if r.numRows >= TABLE_MAX_ROWS {
		return EXECUTE_TABLE_FULL
	}
	dest := make([]byte, ROW_SIZE)
	dest = insertRow.serializeRow(dest)
	copy(r.pager.pages[pageNum][byteOffset:byteOffset+ROW_SIZE], dest)

	r.numRows = r.numRows + 1
	return EXECUTE_SUCCESS
}

func (r *Table) executeSelect() ExecuteResult {
	cursor := TableStart(r)
	// TODO 执行select
	for !cursor.endOfTable {
		selectRow := &Row{}
		pageNum, byteOffset := cursor.cursorValue()
		src := r.pager.pages[pageNum][byteOffset : byteOffset+ROW_SIZE]
		selectRow.deserializeRow(src)

		selectRow.printRow()

		cursor.cursorAdvance()
	}
	return EXECUTE_SUCCESS
}

func (r *Table) executeStatement(statement *Statement) ExecuteResult {
	switch statement.statementType {
	case STATEMENT_INSERT:
		return r.executeInsert(statement)
	case STATEMENT_SELECT:
		return r.executeSelect()
	}
	return EXECUTE_SUCCESS
}

/*
*
* Cursor
*
**/
type Cursor struct {
	rowNum     int
	endOfTable bool
	table      *Table
}

func TableStart(table *Table) *Cursor {
	cursor := &Cursor{table: table}
	cursor.rowNum = 0
	cursor.endOfTable = table.numRows == 0
	return cursor
}

func TableEnd(table *Table) *Cursor {
	cursor := &Cursor{table: table}
	cursor.rowNum = table.numRows
	cursor.endOfTable = true
	return cursor
}

func (r *Cursor) cursorValue() (int, int) {
	rowNum := r.rowNum
	pageNum := rowNum / ROWS_PER_PAGE // 第几页
	_ = r.table.pager.getPage(pageNum)

	rowOffset := rowNum % ROWS_PER_PAGE // row偏移
	byteOffset := rowOffset * ROW_SIZE  // byte偏移

	return pageNum, byteOffset
}

func (r *Cursor) cursorAdvance() {
	r.rowNum += 1
	if r.rowNum >= r.table.numRows {
		r.endOfTable = true
	}
}

func main() {

	fileName := "mydb.txt"
	table := NewTable(fileName)
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
				fmt.Printf("Unrecognized command '%s'\n", inputBuffer.buffer)
				continue
			}
		}

		// prepare
		var statement Statement
		switch inputBuffer.prepareStatement(&statement) {
		case PREPARE_SUCCESS:
		case PREPARE_NEGATIVE_ID:
			fmt.Printf("ID must be positive.\n")
			continue
		case PREPARE_STRING_TOO_LONG:
			fmt.Printf("String is too long.\n")
			continue
		case PREPARE_SYNTAX_ERROR:
			fmt.Printf("Syntax error. Could not parse statement.\n")
			continue
		case PREPARE_UNRECOGNIZED_STATEMENT:
			fmt.Printf("Unrecognized keyword at start of '%s'.\n", inputBuffer.buffer)
			continue
		}

		// execute
		switch table.executeStatement(&statement) {
		case EXECUTE_SUCCESS:
			fmt.Printf("Executed.\n")
		case EXECUTE_TABLE_FULL:
			fmt.Printf("Error: Table full.\n")
		}

		//inputBuffer.executeStatement(&statement)
	}
}

func memcpy(dest []byte, src []byte, ds, ss, n int) []byte {
	if len(dest) < ds+n {
		fmt.Printf("!!! dest %v, start %v, n %v \n", len(dest), ds, n)
		return []byte{}
	}
	if len(src) < ss+n {
		fmt.Printf("!!! src %v, start %v, n %v \n", len(src), ss, n)
		return []byte{}
	}
	copy(dest[ds:ds+n], src[ss:ss+n])
	return dest
}

func StringToBytes(str string, n int) []byte {
	buf := make([]byte, n)
	if n < len(str) {
		fmt.Printf("!!! str %v, n %v \n", len(str), n)
		return buf
	}
	copy(buf[:n], []byte(str))
	return buf
}

func BytesToString(buf []byte) string {
	return string(bytes.TrimRight(buf, "\x00"))
}

func Int32ToBytes(i int32, n int) []byte {
	buf := make([]byte, n)
	if n < 4 {
		fmt.Printf("!!! i %v, n %v \n", i, n)
		return buf
	}
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}

func BytesToInt32(buf []byte) int32 {
	return int32(binary.BigEndian.Uint32(buf))
}

func SizeOfAttribute(row Row, fieldName string) int {
	t := reflect.TypeOf(row)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return 0
	}
	fieldNum := t.NumField()
	for i := 0; i < fieldNum; i++ {
		if strings.ToUpper(t.Field(i).Name) == strings.ToUpper(fieldName) {
			v := reflect.ValueOf(row)
			fieldVal := v.FieldByName(t.Field(i).Name)
			return fieldVal.Len()
		}
	}
	return 0
}

func ToBytes(src interface{}, n int) []byte {
	ret := make([]byte, n)
	srcP := unsafe.Pointer(&src)
	ret = *(*[]byte)(srcP) // 类型转换
	return ret
}

func IsExists(path string) (os.FileInfo, bool) {
	f, err := os.Stat(path)
	return f, err == nil || os.IsExist(err)
}

func IsDir(path string) (os.FileInfo, bool) {
	f, flag := IsExists(path)
	return f, flag && f.IsDir()
}

func IsFile(path string) (os.FileInfo, bool) {
	f, flag := IsExists(path)
	return f, flag && !f.IsDir()
}
