package core

import (
	"GoLearn/goredis/core/proto"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
)

//flags 模式
const CLIENT_PUBSUB = (1 << 18)

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
	FakeFlag bool
	PubSubChannels *map[string]*List
	PubSubPatterns *List
	Flags          int //client flags
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

func (c *Client) addReply(obj *GoredisObject) {
	c.Buf = obj.Ptr.(string)
}

func (c *Client) addReplyStatus(s string) {
	r := proto.NewString([]byte(s))
	c.addReplyString(r)
}

func (c *Client) addReplyError(s string) {
	r := proto.NewError([]byte(s))
	c.addReplyString(r)
}

func (c *Client) addReplyString( r *proto.Resp) {
	if ret, err := proto.EncodeToBytes(r); err == nil {
		c.Buf = string(ret)
	}
}

func (c *Client) call(s *Server) {
	dirty := s.Dirty
	c.Cmd.Proc(c, s)
	dirty = s.Dirty - dirty
	if dirty > 0 && !c.FakeFlag {
		_ = AppendToFile(s.AofFilename, c.QueryBuf)
	}
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
	c.QueryBuf = string(buff)
	return nil
}

// 处理请求信息
func (c *Client) ProcessInputBuffer() error {
	//r := regexp.MustCompile("[^\\s]+")
	decoder := proto.NewDecoder(bytes.NewReader([]byte(c.QueryBuf)))
	//decoder := proto.NewDecoder(bytes.NewReader([]byte("*2\r\n$3\r\nget\r\n")))
	if resp, err := decoder.DecodeMultiBulk(); err == nil {
		c.Argc = len(resp)
		c.Argv = make([]*GoredisObject, c.Argc)
		for k, s := range resp {
			c.Argv[k] = CreateObject(ObjectTypeString, string(s.Value))
		}
		return nil
	}
	return errors.New("ProcessInputBuffer failed")
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
	PubSubChannels   *map[string]*List
	PubSubPatterns   *List
}

func NewServer() *Server {
	s := new(Server)
	return s
}

func (s *Server) CreateClient() (c *Client) {
	c = new(Client)
	c.Db = s.Dbs[0]
	c.QueryBuf = ""
	tmp := make(map[string]*List, 0)
	c.PubSubChannels = &tmp
	c.Flags = 0
	return c
}

func (s *Server) ProcessCommand(c *Client) {
	name, ok := c.Argv[0].Ptr.(string)
	if !ok {
		log.Println("error cmd")
		os.Exit(0)
	}
	cmd := s.lookupCommand(name) // 查找命令
	fmt.Println(cmd, name, s)
	if cmd != nil {
		c.Cmd = cmd
		c.call(s) // 执行命令
	} else {
		c.addReplyError(fmt.Sprintf("(error) ERR unknown command '%s'", name))
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
		c.addReplyError("(error) ERR wrong number of arguments for 'set' command")
	}
	if stringKey, ok1 := objKey.Ptr.(string); ok1 {
		if stringValue, ok2 := objValue.Ptr.(string); ok2 {
			c.Db.Dict[stringKey] = CreateObject(ObjectTypeString, stringValue)
		}
	}
	s.Dirty++
	c.addReplyStatus("OK")
}

func GetCommand(c *Client, s *Server) {
	key := c.Argv[1].Ptr.(string)
	obj := c.lookupKey(key) // key查找obj
	if obj != nil {
		c.addReplyStatus(obj.Ptr.(string))
	} else {
		c.addReplyStatus("nil")
	}
}
