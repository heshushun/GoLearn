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
	NODE_TYPE_SIZE          = 4
	NODE_TYPE_OFFSET        = 0
	IS_ROOT_SIZE            = 4
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
	LEAF_NODE_KEY_SIZE          = 4
	LEAF_NODE_KEY_OFFSET        = 0
	LEAF_NODE_VALUE_SIZE        = ROW_SIZE
	LEAF_NODE_VALUE_OFFSET      = LEAF_NODE_KEY_OFFSET + LEAF_NODE_KEY_SIZE
	LEAF_NODE_CELL_SIZE         = LEAF_NODE_KEY_SIZE + LEAF_NODE_VALUE_SIZE
	LEAF_NODE_SPACE_FOR_CELLS   = PAGE_SIZE - LEAF_NODE_HEADER_SIZE
	LEAF_NODE_MAX_CELLS         = LEAF_NODE_SPACE_FOR_CELLS / LEAF_NODE_CELL_SIZE
	LEAF_NODE_RIGHT_SPLIT_COUNT = (LEAF_NODE_MAX_CELLS + 1) / 2
	LEAF_NODE_LEFT_SPLIT_COUNT  = (LEAF_NODE_MAX_CELLS + 1) - LEAF_NODE_RIGHT_SPLIT_COUNT
	/*
	 * Internal Node Header Layout
	 */
	INTERNAL_NODE_NUM_KEYS_SIZE      = 4
	INTERNAL_NODE_NUM_KEYS_OFFSET    = COMMON_NODE_HEADER_SIZE
	INTERNAL_NODE_RIGHT_CHILD_SIZE   = 4
	INTERNAL_NODE_RIGHT_CHILD_OFFSET = INTERNAL_NODE_NUM_KEYS_OFFSET + INTERNAL_NODE_NUM_KEYS_SIZE
	INTERNAL_NODE_HEADER_SIZE        = COMMON_NODE_HEADER_SIZE + INTERNAL_NODE_NUM_KEYS_SIZE + INTERNAL_NODE_RIGHT_CHILD_SIZE
	/*
	 * Internal Node Body Layout
	 */
	INTERNAL_NODE_KEY_SIZE   = 4
	INTERNAL_NODE_CHILD_SIZE = 4
	INTERNAL_NODE_CELL_SIZE  = INTERNAL_NODE_CHILD_SIZE + INTERNAL_NODE_KEY_SIZE
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
	EXECUTE_DUPLICATE_KEY
	EXECUTE_NOT_FIND
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
		table.pager.printTree(0, 0)
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

/*
*
* leafNode
*
**/
func (r *Node) leafNodeNumCells() int {
	offset := LEAF_NODE_NUM_CELLS_OFFSET
	numCells := BytesToInt32(r.data[offset : offset+LEAF_NODE_NUM_CELLS_SIZE])
	return int(numCells)
}

func (r *Node) setLeafNodeNumCells(num int) {
	offset := LEAF_NODE_NUM_CELLS_OFFSET
	copy(r.data[offset:offset+LEAF_NODE_NUM_CELLS_SIZE], Int32ToBytes(int32(num), LEAF_NODE_NUM_CELLS_SIZE))
}

func (r *Node) leafNodeKeyOffset(cellNum int) int {
	offset := LEAF_NODE_HEADER_SIZE + cellNum*LEAF_NODE_CELL_SIZE
	return offset
}

func (r *Node) leafNodeKey(cellNum int) int32 {
	offset := r.leafNodeKeyOffset(cellNum)
	return BytesToInt32(r.data[offset : offset+LEAF_NODE_KEY_SIZE])
}

func (r *Node) setLeafNodeKey(cellNum int, key int32) {
	offset := r.leafNodeKeyOffset(cellNum)
	copy(r.data[offset:offset+LEAF_NODE_KEY_SIZE], Int32ToBytes(key, LEAF_NODE_KEY_SIZE))
}

func (r *Node) leafNodeValueOffset(cellNum int) int {
	offset := r.leafNodeKeyOffset(cellNum) + LEAF_NODE_KEY_SIZE
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

func (r *Node) leafNodeCell(cellNum int) []byte {
	offset := r.leafNodeKeyOffset(cellNum)
	return r.data[offset : offset+LEAF_NODE_CELL_SIZE]
}

func (r *Node) setLeafNodeCell(cellNum int, val []byte) {
	offset := r.leafNodeKeyOffset(cellNum)
	copy(r.data[offset:offset+LEAF_NODE_CELL_SIZE], val)
}

func (r *Node) initializeLeafNode() {
	r.setNodeType(NODE_LEAF)
	r.setNodeRoot(false)
	r.setLeafNodeNumCells(0)
}

func (r *Node) getNodeType() NodeType {
	offset := NODE_TYPE_OFFSET
	return NodeType(BytesToInt32(r.data[offset : offset+NODE_TYPE_SIZE]))
}

func (r *Node) setNodeType(nodeType NodeType) {
	offset := NODE_TYPE_OFFSET
	copy(r.data[offset:offset+NODE_TYPE_SIZE], Int32ToBytes(int32(nodeType), NODE_TYPE_SIZE))
}

func (r *Node) isNodeRoot() bool {
	offset := IS_ROOT_OFFSET
	value := r.data[offset : offset+IS_ROOT_SIZE]
	isRoot := false
	if BytesToInt32(value) == 0 {
		isRoot = false
	}
	if BytesToInt32(value) == 1 {
		isRoot = true
	}
	return isRoot
}

func (r *Node) setNodeRoot(isRoot bool) {
	offset := IS_ROOT_OFFSET
	if isRoot {
		copy(r.data[offset:offset+IS_ROOT_SIZE], Int32ToBytes(1, IS_ROOT_SIZE))
	} else {
		copy(r.data[offset:offset+IS_ROOT_SIZE], Int32ToBytes(0, IS_ROOT_SIZE))
	}
}

/*
*
* internalNode
*
**/
func (r *Node) internalNodeNumKeys() int {
	offset := INTERNAL_NODE_NUM_KEYS_OFFSET
	numKeys := BytesToInt32(r.data[offset : offset+INTERNAL_NODE_NUM_KEYS_SIZE])
	return int(numKeys)
}

func (r *Node) setInternalNodeNumKeys(numKeys int) {
	offset := INTERNAL_NODE_NUM_KEYS_OFFSET
	copy(r.data[offset:offset+INTERNAL_NODE_NUM_KEYS_SIZE], Int32ToBytes(int32(numKeys), INTERNAL_NODE_NUM_KEYS_SIZE))
}

func (r *Node) internalNodeRightChild() []byte {
	offset := INTERNAL_NODE_RIGHT_CHILD_OFFSET
	return r.data[offset : offset+INTERNAL_NODE_RIGHT_CHILD_SIZE]
}

func (r *Node) setInternalNodeRightChild(rightChild []byte) {
	offset := INTERNAL_NODE_RIGHT_CHILD_OFFSET
	copy(r.data[offset:offset+INTERNAL_NODE_RIGHT_CHILD_SIZE], rightChild)
}

func (r *Node) internalNodeCellOffset(cellNum int) int {
	offset := INTERNAL_NODE_HEADER_SIZE + cellNum*INTERNAL_NODE_CELL_SIZE
	return offset
}

func (r *Node) internalNodeCell(cellNum int) []byte {
	offset := r.internalNodeCellOffset(cellNum)
	return r.data[offset : offset+INTERNAL_NODE_CELL_SIZE]
}

func (r *Node) setInternalNodeCell(cellNum int, val []byte) {
	offset := r.internalNodeCellOffset(cellNum)
	copy(r.data[offset:offset+INTERNAL_NODE_CELL_SIZE], val)
}

func (r *Node) internalNodeChild(childNum int) []byte {
	numKeys := r.internalNodeNumKeys()
	if childNum > numKeys {
		fmt.Printf("Tried to access child_num %d > num_keys %d\n", childNum, numKeys)
		os.Exit(0)
	} else if childNum == numKeys {
		return r.internalNodeRightChild()
	} else {
		return r.internalNodeCell(childNum)
	}
	return []byte{}
}

func (r *Node) setInternalNodeChild(childNum int, val []byte) {
	numKeys := r.internalNodeNumKeys()
	if childNum > numKeys {
		fmt.Printf("Tried to access child_num %d > num_keys %d\n", childNum, numKeys)
		os.Exit(0)
	} else if childNum == numKeys {
		r.setInternalNodeRightChild(val)
	} else {
		r.setInternalNodeCell(childNum, val)
	}
}

func (r *Node) internalNodeKey(keyNum int) int32 {
	offset := r.internalNodeCellOffset(keyNum) + INTERNAL_NODE_CHILD_SIZE
	return BytesToInt32(r.data[offset : offset+INTERNAL_NODE_KEY_SIZE])
}

func (r *Node) setInternalNodeKey(keyNum int, nodeKey int32) {
	offset := r.internalNodeCellOffset(keyNum) + INTERNAL_NODE_CHILD_SIZE
	copy(r.data[offset:offset+INTERNAL_NODE_KEY_SIZE], Int32ToBytes(nodeKey, INTERNAL_NODE_KEY_SIZE))
}

func (r *Node) getNodeMaxKey() int32 {
	switch r.getNodeType() {
	case NODE_INTERNAL:
		return r.internalNodeKey(r.internalNodeNumKeys() - 1)
	case NODE_LEAF:
		return r.leafNodeKey(r.leafNodeNumCells() - 1)
	default:
		fmt.Printf("NodeType error %d\n", r.getNodeType())
		os.Exit(0)
	}
	return 0
}

func (r *Node) initializeInternalNode() {
	r.setNodeType(NODE_INTERNAL)
	r.setNodeRoot(false)
	r.setInternalNodeNumKeys(0)
}

func (r *Node) printConstants() {
	fmt.Printf("ROW_SIZE: %d\n", ROW_SIZE)
	fmt.Printf("COMMON_NODE_HEADER_SIZE: %d\n", COMMON_NODE_HEADER_SIZE)
	fmt.Printf("LEAF_NODE_HEADER_SIZE: %d\n", LEAF_NODE_HEADER_SIZE)
	fmt.Printf("LEAF_NODE_CELL_SIZE: %d\n", LEAF_NODE_CELL_SIZE)
	fmt.Printf("LEAF_NODE_SPACE_FOR_CELLS: %d\n", LEAF_NODE_SPACE_FOR_CELLS)
	fmt.Printf("LEAF_NODE_MAX_CELLS: %d\n", LEAF_NODE_MAX_CELLS)
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

func (r *Pager) getUnusedPageNum() int {
	return r.numPages
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

func (r *Pager) indent(level int) {
	for i := 0; i < level; i++ {
		fmt.Print("  ")
	}
}

func (r *Pager) printTree(pageNum, indentationLevel int) {
	page := r.getPage(pageNum)
	node := NewNode(&page)
	numKeys := 0

	switch node.getNodeType() {
	case NODE_LEAF:
		numKeys = node.leafNodeNumCells()
		r.indent(indentationLevel)
		fmt.Printf("- leaf (size %d)\n", numKeys)
		for i := 0; i < numKeys; i++ {
			r.indent(indentationLevel + 1)
			fmt.Printf("- %d\n", node.leafNodeKey(i))
		}
		break
	case NODE_INTERNAL:
		var child []byte
		numKeys = node.internalNodeNumKeys()
		r.indent(indentationLevel)
		fmt.Printf("- internal (size %d)\n", numKeys)
		for i := 0; i < numKeys; i++ {
			child = node.internalNodeChild(i)
			r.printTree(BytesToInt(child), indentationLevel+1)

			r.indent(indentationLevel + 1)
			fmt.Printf("- key %d\n", node.internalNodeKey(i))
		}
		child = node.internalNodeRightChild()
		r.printTree(BytesToInt(child), indentationLevel+1)
		break
	}
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
		rootNode.setNodeRoot(true)
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
	numCells := node.leafNodeNumCells()
	insertRow := &statement.rowToInsert

	ketToInsert := insertRow.id
	cursor := TableFind(r, ketToInsert)
	if cursor == nil {
		return EXECUTE_NOT_FIND
	}

	if cursor.cellNum < numCells {
		ketAtIndex := node.leafNodeKey(cursor.cellNum)
		if ketAtIndex == ketToInsert {
			return EXECUTE_DUPLICATE_KEY
		}
	}
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

func (r *Table) createNewRoot(rightChildPageNum int) {
	/*
	 Handle splitting the root.
	 Old root copied to new page, becomes left child.
	 Address of right child passed in.
	 Re-initialize root page to contain the new root node.
	 New root node points to two children.
	*/
	// root
	root := r.pager.getPage(r.rootPageNum)
	rootNode := NewNode(&root)

	// right
	//rightChild := r.pager.getPage(rightChildPageNum)
	//rightNode := NewNode(&rightChild)

	// left
	leftChildPageNum := r.pager.getUnusedPageNum()
	leftChild := r.pager.getPage(leftChildPageNum)
	leftNode := NewNode(&leftChild)

	/* Left child has data copied from old root */
	copy(leftChild, root)
	leftNode.setNodeRoot(false)

	/* Root node is a new internal node with one key and two children */
	rootNode.initializeInternalNode()
	rootNode.setNodeRoot(true)
	rootNode.setInternalNodeNumKeys(1)
	rootNode.setInternalNodeChild(0, IntToBytes(leftChildPageNum))
	rootNode.setInternalNodeKey(0, leftNode.getNodeMaxKey())
	rootNode.setInternalNodeRightChild(IntToBytes(rightChildPageNum))
}

/*r
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

func TableFind(table *Table, key int32) *Cursor {

	rootPageNum := table.rootPageNum

	page := table.pager.getPage(table.rootPageNum)
	rootNode := NewNode(&page)

	if rootNode.getNodeType() == NODE_LEAF {
		return LeafNodeFind(table, rootPageNum, key)
	} else {
		return InternalNodeFind(table, rootPageNum, key)
	}
}

func LeafNodeFind(table *Table, pageNum int, key int32) *Cursor {
	page := table.pager.getPage(pageNum)
	node := NewNode(&page)
	numCells := node.leafNodeNumCells()

	cursor := &Cursor{table: table}
	cursor.pageNum = pageNum

	// Binary search
	minIndex := 0
	onePastMaxIndex := numCells
	for {
		if onePastMaxIndex == minIndex {
			break
		}
		index := (minIndex + onePastMaxIndex) / 2
		keyAtIndex := node.leafNodeKey(index)
		if key == keyAtIndex {
			cursor.cellNum = index
			return cursor
		}
		if key < keyAtIndex {
			onePastMaxIndex = index
		} else {
			minIndex = index + 1
		}
	}
	cursor.cellNum = minIndex

	return cursor
}

func InternalNodeFind(table *Table, pageNum int, key int32) *Cursor {
	page := table.pager.getPage(pageNum)
	node := NewNode(&page)
	numKeys := node.internalNodeNumKeys()

	/* Binary search to find index of child to search */
	minIndex := 0
	/* there is one more child than key */
	maxIndex := numKeys

	for {
		if minIndex == maxIndex {
			break
		}
		index := (minIndex + maxIndex) / 2
		keyToRight := node.internalNodeKey(index)
		if keyToRight >= key {
			maxIndex = index
		} else {
			minIndex = index + 1
		}
	}

	childNum := node.internalNodeChild(minIndex)
	childPage := table.pager.getPage(BytesToInt(childNum))
	child := NewNode(&childPage)
	switch child.getNodeType() {
	case NODE_LEAF:
		return LeafNodeFind(table, BytesToInt(childNum), key)
	case NODE_INTERNAL:
		return InternalNodeFind(table, BytesToInt(childNum), key)
	default:
		fmt.Printf("NodeType error %v", child.getNodeType())
		os.Exit(0)
	}
	return nil
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
		r.leafNodeSplitAndInsert(key, row)
		return
	}

	if r.cellNum < numCells {
		// Make room for new cell
		for i := numCells; i > r.cellNum; i-- {
			ds := node.leafNodeKeyOffset(i)
			ss := node.leafNodeKeyOffset(i - 1)
			pageData := r.table.pager.pages[pageNum]
			copy(pageData[ds:ds+LEAF_NODE_CELL_SIZE], pageData[ss:ss+LEAF_NODE_CELL_SIZE])
		}
	}

	// numCells
	node.setLeafNodeNumCells(node.leafNodeNumCells() + 1)

	// key
	node.setLeafNodeKey(r.cellNum, key)

	// value
	dest := make([]byte, LEAF_NODE_VALUE_SIZE)
	dest = row.serializeRow(dest)
	node.setLeafNodeValue(r.cellNum, dest)

	r.table.pager.pages[pageNum] = node.data
}

func (r *Cursor) leafNodeSplitAndInsert(key int32, row *Row) {
	/*
	 Create a new node and move half the cells over.
	 Insert the new value in one of the two nodes.
	 Update parent or create a new parent.
	*/
	oldPage := r.table.pager.getPage(r.pageNum)
	oldNode := NewNode(&oldPage)

	newPageNum := r.table.pager.getUnusedPageNum()
	newPage := r.table.pager.getPage(newPageNum)
	newNode := NewNode(&newPage)
	newNode.initializeLeafNode()

	/*
	 All existing keys plus new key should be divided
	 evenly between old (left) and new (right) nodes.
	 Starting from the right, move each key to correct position.
	*/
	for i := LEAF_NODE_MAX_CELLS; i >= 0; i-- {
		var destinationNode *Node
		if i >= LEAF_NODE_LEFT_SPLIT_COUNT {
			// new (right)
			destinationNode = newNode
		} else {
			// old (left)
			destinationNode = oldNode
		}

		// 插入位置
		indexWithinNode := i % LEAF_NODE_LEFT_SPLIT_COUNT
		if i == r.cellNum {
			// key
			destinationNode.setLeafNodeKey(indexWithinNode, key)
			// value
			destination := make([]byte, LEAF_NODE_VALUE_SIZE)
			destination = row.serializeRow(destination)
			destinationNode.setLeafNodeValue(indexWithinNode, destination)
		} else if i > r.cellNum {
			destinationNode.setLeafNodeCell(indexWithinNode, oldNode.leafNodeCell(i-1))
		} else {
			destinationNode.setLeafNodeCell(indexWithinNode, oldNode.leafNodeCell(i))
		}
	}

	/* Update cell count on both leaf nodes */
	oldNode.setLeafNodeNumCells(LEAF_NODE_LEFT_SPLIT_COUNT)
	newNode.setLeafNodeNumCells(LEAF_NODE_RIGHT_SPLIT_COUNT)

	r.table.pager.pages[r.pageNum] = oldNode.data
	r.table.pager.pages[newPageNum] = newNode.data

	if oldNode.isNodeRoot() {
		// 创建rootNode
		r.table.createNewRoot(newPageNum)
	} else {
		fmt.Printf("Need to implement updating parent after split\n")
		os.Exit(0)
	}

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
		case EXECUTE_DUPLICATE_KEY:
			fmt.Printf("Error: Duplicate key.\n")
		case EXECUTE_NOT_FIND:
			fmt.Printf("Error: Not find.\n")
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

func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	_ = binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
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
