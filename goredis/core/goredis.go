package core

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

/*
*
* Client
*
**/
type Client struct {
	Cmd      *GoredisCommand  // 命令
	Argv     []*GoredisObject // 输入参数内容
	Argc     int              // 输入参数个数
	Db       *GoredisDb
	QueryBuf string
	Buf      string
}

func NewClient(conn net.Conn, db *GoredisDb) *Client {
	c := new(Client)
	c.Db = db
	c.Argv = make([]*GoredisObject, 5)
	c.QueryBuf = ""
	return c
}

func (c *Client) lookupKey(key string) *GoredisObject {
	if obj, ok := c.Db.Dict[key]; ok {
		return obj
	}
	return nil
}

func (c *Client) doReply(obj *GoredisObject) {
	c.Buf = obj.Ptr.(string)
}

func (c *Client) call(s *Server) {
	c.Cmd.Proc(c, s)
}

// 读取请求信息
func (c *Client) ReadQueryFromClient(conn net.Conn) (err error) {
	buff := make([]byte, 512)
	n, err := conn.Read(buff)

	if err != nil {
		log.Println("conn.Read err!=nil", err, "---len---", n, conn)
		conn.Close()
		return err
	}
	tmp := string(buff)
	parts := strings.Split(tmp, "\n")
	c.QueryBuf = parts[0]
	return nil
}

// 处理请求信息
func (c *Client) ProcessInputBuffer() {
	r := regexp.MustCompile("[^\\s]+")
	parts := r.FindAllString(c.QueryBuf, -1)
	argc, argv := len(parts), parts
	c.Argc = argc
	j := 0
	for _, v := range argv {
		c.Argv[j] = CreateObject(ObjectTypeString, v)
		j++
	}
}

/*
*
* Server
*
**/
type Server struct {
	Dbs              []*GoredisDb
	DbNum            int
	Start            int64
	Port             int32
	RdbFilename      string
	AofFilename      string
	NextClientID     int32
	SystemMemorySize int32
	Clients          int32
	Pid              int
	Commands         map[string]*GoredisCommand // 命令表
	Dirty            int64
	AofBuf           []string
}

func NewServer() *Server {
	s := new(Server)
	return s
}

func (s *Server) CreateClient(conn net.Conn) (c *Client) {
	c = new(Client)
	c.Db = s.Dbs[0]
	c.Argv = make([]*GoredisObject, 5)
	c.QueryBuf = ""
	return c
}

func (s *Server) ProcessCommand(c *Client) {
	name, ok := c.Argv[0].Ptr.(string)
	if !ok {
		log.Println("error cmd")
		os.Exit(0)
	}
	cmd := s.lookupCommand(name) // 查找命令
	if cmd != nil {
		c.Cmd = cmd
		c.call(s) // 执行命令
	} else {
		c.doReply(CreateObject(ObjectTypeString, fmt.Sprintf("(error) ERR unknown command '%s'", name)))
	}
}

func (s *Server) lookupCommand(name string) *GoredisCommand {
	cmd, ok := s.Commands[name]
	if ok {
		return cmd
	}
	return nil
}

func SetCommand(c *Client, s *Server) {
	objKey := c.Argv[1]
	objValue := c.Argv[2]
	if c.Argc != 3 {
		c.doReply(CreateObject(ObjectTypeString, "(error) ERR wrong number of arguments for 'set' command"))
		return
	}
	if stringKey, ok1 := objKey.Ptr.(string); ok1 {
		if stringValue, ok2 := objValue.Ptr.(string); ok2 {
			c.Db.Dict[stringKey] = CreateObject(ObjectTypeString, stringValue)
		}
	}
	c.doReply(CreateObject(ObjectTypeString, "OK"))
}

func GetCommand(c *Client, s *Server) {
	key := c.Argv[1].Ptr.(string)
	obj := c.lookupKey(key) // key查找obj
	if obj != nil {
		c.doReply(obj)
	} else {
		c.doReply(CreateObject(ObjectTypeString, "nil"))
	}
}
