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

	/*
	 * Common Node Header Layout
	 */
	NODE_TYPE_SIZE          = 1
	NODE_TYPE_OFFSET        = 0
	IS_ROOT_SIZE            = 1
	IS_ROOT_OFFSET          = NODE_TYPE_SIZE
	PARENT_POINTER_SIZE     = 4
	PARENT_POINTER_OFFSET   = IS_ROOT_OFFSET + IS_ROOT_SIZE
	COMMON_NODE_HEADER_SIZE = NODE_TYPE_SIZE + IS_ROOT_SIZE + PARENT_POINTER_SIZE
	/*
	 * Leaf Node Header Layout
	 */
	LEAF_NODE_NUM_CELLS_SIZE   = 4
	LEAF_NODE_NUM_CELLS_OFFSET = COMMON_NODE_HEADER_SIZE
	LEAF_NODE_HEADER_SIZE      = COMMON_NODE_HEADER_SIZE + LEAF_NODE_NUM_CELLS_SIZE
	/*
	 * Leaf Node Body Layout
	 */
	LEAF_NODE_KEY_SIZE        = 4
	LEAF_NODE_KEY_OFFSET      = 0
	LEAF_NODE_VALUE_SIZE      = ROW_SIZE
	LEAF_NODE_VALUE_OFFSET    = LEAF_NODE_KEY_OFFSET + LEAF_NODE_KEY_SIZE
	LEAF_NODE_CELL_SIZE       = LEAF_NODE_KEY_SIZE + LEAF_NODE_VALUE_SIZE
	LEAF_NODE_SPACE_FOR_CELLS = PAGE_SIZE - LEAF_NODE_HEADER_SIZE
	LEAF_NODE_MAX_CELLS       = LEAF_NODE_SPACE_FOR_CELLS / LEAF_NODE_CELL_SIZE
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

type NodeType int32

const (
	NODE_INTERNAL NodeType = iota
	NODE_LEAF
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
		table.dbClosed()
		os.Exit(0)
	} else if r.buffer == ".btree" {
		fmt.Printf("Tree:\n")
		page := table.pager.getPage(table.rootPageNum)
		node := NewNode(&page)
		node.printLeafNode()
		return META_COMMAND_SUCCESS
	} else if r.buffer == ".constants" {
		fmt.Printf("Constants:\n")
		page := table.pager.getPage(table.rootPageNum)
		node := NewNode(&page)
		node.printConstants()
		return META_COMMAND_SUCCESS
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
	// 序列化
	destination = memcpy(destination, Int32ToBytes(r.id, ID_SIZE), ID_OFFSET, 0, ID_SIZE)
	destination = memcpy(destination, StringToBytes(r.username, USERNAME_SIZE), USERNAME_OFFSET, 0, USERNAME_SIZE)
	destination = memcpy(destination, StringToBytes(r.email, EMAIL_SIZE), EMAIL_OFFSET, 0, EMAIL_SIZE)
	return destination
}

func (r *Row) deserializeRow(src []byte) {
	// 反序列化
	r.id = BytesToInt32(memcpy(make([]byte, ID_SIZE), src, 0, ID_OFFSET, ID_SIZE))
	r.username = BytesToString(memcpy(make([]byte, USERNAME_SIZE), src, 0, USERNAME_OFFSET, USERNAME_SIZE))
	r.email = BytesToString(memcpy(make([]byte, EMAIL_SIZE), src, 0, EMAIL_OFFSET, EMAIL_SIZE))
}

/*
*
* Node
*
**/
type Node struct {
	data []byte
}

func NewNode(page *[]byte) *Node {
	return &Node{data: *page}
}

func (r *Node) leafNodeNumCells() int {
	offset := LEAF_NODE_NUM_CELLS_OFFSET
	numCells := BytesToInt32(r.data[offset : offset+LEAF_NODE_NUM_CELLS_SIZE])
	return int(numCells)
}

func (r *Node) setLeafNodeNumCells(num int) {
	offset := LEAF_NODE_NUM_CELLS_OFFSET
	copy(r.data[offset:offset+LEAF_NODE_NUM_CELLS_SIZE], Int32ToBytes(int32(num), LEAF_NODE_NUM_CELLS_SIZE))
}

func (r *Node) leafNodeCell(cellNum int) int {
	offset := LEAF_NODE_HEADER_SIZE + cellNum*LEAF_NODE_CELL_SIZE
	return offset
}

func (r *Node) leafNodeKey(cellNum int) int32 {
	offset := r.leafNodeCell(cellNum)
	return BytesToInt32(r.data[offset : offset+LEAF_NODE_KEY_SIZE])
}

func (r *Node) setLeafNodeKey(cellNum int, key int32) {
	offset := r.leafNodeCell(cellNum)
	copy(r.data[offset:offset+LEAF_NODE_KEY_SIZE], Int32ToBytes(key, LEAF_NODE_KEY_SIZE))
}

func (r *Node) leafNodeValueOffset(cellNum int) int {
	offset := r.leafNodeCell(cellNum) + LEAF_NODE_KEY_SIZE
	return offset
}

func (r *Node) leafNodeValue(cellNum int) []byte {
	offset := r.leafNodeValueOffset(cellNum)
	return r.data[offset : offset+LEAF_NODE_VALUE_SIZE]
}

func (r *Node) setLeafNodeValue(cellNum int, val []byte) {
	offset := r.leafNodeValueOffset(cellNum)
	copy(r.data[offset:offset+LEAF_NODE_VALUE_SIZE], val)
}

func (r *Node) initializeLeafNode() {
	offset := LEAF_NODE_NUM_CELLS_OFFSET
	copy(r.data[offset:offset+LEAF_NODE_NUM_CELLS_SIZE], Int32ToBytes(0, LEAF_NODE_NUM_CELLS_SIZE))
}

func (r *Node) printConstants() {
	fmt.Printf("ROW_SIZE: %d\n", ROW_SIZE)
	fmt.Printf("COMMON_NODE_HEADER_SIZE: %d\n", COMMON_NODE_HEADER_SIZE)
	fmt.Printf("LEAF_NODE_HEADER_SIZE: %d\n", LEAF_NODE_HEADER_SIZE)
	fmt.Printf("LEAF_NODE_CELL_SIZE: %d\n", LEAF_NODE_CELL_SIZE)
	fmt.Printf("LEAF_NODE_SPACE_FOR_CELLS: %d\n", LEAF_NODE_SPACE_FOR_CELLS)
	fmt.Printf("LEAF_NODE_MAX_CELLS: %d\n", LEAF_NODE_MAX_CELLS)
}

func (r *Node) printLeafNode() {
	numCells := r.leafNodeNumCells()
	fmt.Printf("leaf (size %d)\n", numCells)
	for i := 0; i < numCells; i++ {
		key := r.leafNodeKey(i)
		fmt.Printf("  - %d : %d\n", i, key)
	}
}

/*
*
* Pager
*
**/
type Pager struct {
	file       *os.File
	fileFD     uintptr
	fileName   string
	fileLength int64
	numPages   int // 页数
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
	pager.numPages = int(fileLength / PAGE_SIZE)
	pages := make([][]byte, TABLE_MAX_PAGES)
	pager.pages = pages

	if fileLength%PAGE_SIZE != 0 {
		fmt.Printf("Db file is not a whole number of pages. Corrupt file.\n")
		os.Exit(0)
	}

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

		if pageNum >= r.numPages {
			r.numPages = pageNum + 1
		}
	}
	return r.pages[pageNum]
}

func (r *Pager) pagerFlush(pageNum int) {
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
	n, err2 := r.file.Write(page[:PAGE_SIZE])
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
	// 目前只有退出才关闭文件
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
	rootPageNum int // root页码
	pager       *Pager
}

func dbOpen(fileName string) *Table {
	pager := NewPager(fileName)
	table := &Table{}
	table.pager = pager
	table.rootPageNum = 0

	if pager.numPages == 0 {
		// New database file. Initialize page 0 as leaf node.
		page := table.pager.getPage(0)
		rootNode := NewNode(&page)
		rootNode.initializeLeafNode()
		table.pager.pages[0] = rootNode.data
	}
	return table
}

func (r *Table) dbClosed() {
	pager := r.pager

	// 每页 内存写入回文件
	for i := 0; i < pager.numPages; i++ {
		if pager.pages[i] == nil {
			continue
		}
		pager.pagerFlush(i)
	}

	r.rootPageNum = 0
	pager.freePager()
}

func (r *Table) executeInsert(statement *Statement) ExecuteResult {
	// 执行insert
	page := r.pager.getPage(r.rootPageNum)
	node := NewNode(&page)
	if node.leafNodeNumCells() >= LEAF_NODE_MAX_CELLS {
		return EXECUTE_TABLE_FULL
	}
	insertRow := &statement.rowToInsert

	cursor := TableEnd(r)
	cursor.leafNodeInsert(insertRow.id, insertRow)

	return EXECUTE_SUCCESS
}

func (r *Table) executeSelect() ExecuteResult {
	cursor := TableStart(r)
	// 执行select
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
	pageNum    int // 页码
	cellNum    int // 单元格编号
	endOfTable bool
	table      *Table
}

func TableStart(table *Table) *Cursor {
	cursor := &Cursor{table: table}
	cursor.pageNum = table.rootPageNum
	cursor.cellNum = 0

	page := table.pager.getPage(table.rootPageNum)
	rootNode := NewNode(&page)
	numCells := rootNode.leafNodeNumCells()
	cursor.endOfTable = numCells == 0
	return cursor
}

func TableEnd(table *Table) *Cursor {
	cursor := &Cursor{table: table}
	cursor.pageNum = table.rootPageNum

	page := table.pager.getPage(table.rootPageNum)
	rootNode := NewNode(&page)
	numCells := rootNode.leafNodeNumCells()
	cursor.cellNum = numCells

	cursor.endOfTable = true
	return cursor
}

func (r *Cursor) cursorValue() (int, int) {
	pageNum := r.pageNum // 第几页
	page := r.table.pager.getPage(pageNum)
	node := NewNode(&page)
	offsetByte := node.leafNodeValueOffset(r.cellNum)
	return pageNum, offsetByte
}

func (r *Cursor) cursorAdvance() {
	pageNum := r.pageNum // 第几页
	page := r.table.pager.getPage(pageNum)
	node := NewNode(&page)
	r.cellNum += 1
	if r.cellNum >= node.leafNodeNumCells() {
		r.endOfTable = true
	}
}

func (r *Cursor) leafNodeInsert(key int32, row *Row) {
	// 用游标作为标识 来插入对应位置
	pageNum := r.pageNum // 第几页
	page := r.table.pager.getPage(pageNum)
	node := NewNode(&page)

	// 叶子节点单元格数量
	numCells := node.leafNodeNumCells()
	if numCells >= LEAF_NODE_MAX_CELLS {
		// Node full
		fmt.Printf("Need to implement splitting a leaf node.\n")
		os.Exit(0)
	}

	if r.cellNum < numCells {
		// Make room for new cell
		for i := numCells; i > r.cellNum; i-- {
			ds := node.leafNodeCell(i)
			ss := node.leafNodeCell(i - 1)
			pageData := r.table.pager.pages[pageNum]
			copy(pageData[ds:ds+LEAF_NODE_CELL_SIZE], pageData[ss:ss+LEAF_NODE_CELL_SIZE])
		}
	}

	node.setLeafNodeNumCells(node.leafNodeNumCells() + 1)
	node.setLeafNodeKey(r.cellNum, key)

	dest := make([]byte, LEAF_NODE_VALUE_SIZE)
	dest = row.serializeRow(dest)
	node.setLeafNodeValue(r.cellNum, dest)

	r.table.pager.pages[pageNum] = node.data
}

func main() {

	fileName := "mydb.txt"
	table := dbOpen(fileName)
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
				continue
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
