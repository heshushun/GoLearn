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
const CLIENT_PUBSUB = 1 << 18

/*
*
* Client
*
**/
type Client struct {
	Cmd            *GoredisCommand  // 命令
	Argv           []*GoredisObject // 输入参数内容
	Argc           int              // 输入参数个数
	Db             *GoredisDb       // 客户端缓存的DB数据
	QueryBuf       string           // 命令缓存
	RespBuf        string           // 响应缓存
	FakeFlag       bool             // 虚标志
	PubSubChannels map[string]*List // 订阅发布渠道
	PubSubPatterns *List            //
	Flags          int              // client flags
}

func NewClient(db *GoredisDb) *Client {
	c := new(Client)
	c.Db = db
	c.QueryBuf = ""
	c.PubSubChannels = make(map[string]*List, 0)
	c.Flags = 0
	return c
}

// key获取缓存
func (c *Client) lookupObject(key string) *GoredisObject {
	if obj, ok := c.Db.Dict[key]; ok {
		return obj
	}
	return nil
}

func (c *Client) addReplyStatus(s string) {
	r := proto.NewString([]byte(s))
	c.addReplyString(r)
}

func (c *Client) addReplyError(s string) {
	r := proto.NewError([]byte(s))
	c.addReplyString(r)
}

func (c *Client) addReplyString(r *proto.Resp) {
	if ret, err := proto.EncodeToBytes(r); err == nil {
		c.RespBuf = string(ret)
	}
}

// 读取请求
func (c *Client) ReadQueryFromClient(conn net.Conn) (err error) {
	buff := make([]byte, 512)
	n, err := conn.Read(buff)

	if err != nil {
		log.Println("conn.Read err!=nil", err, "len", n, conn)
		conn.Close()
		return err
	}
	c.QueryBuf = string(buff)
	return nil
}

// 转化请求
func (c *Client) ProcessQueryBuffer() error {
	//r := regexp.MustCompile("[^\\s]+")
	decoder := proto.NewDecoder(bytes.NewReader([]byte(c.QueryBuf)))
	if resp, err := decoder.DecodeMultiBulk(); err == nil {
		c.Argc = len(resp)
		c.Argv = make([]*GoredisObject, c.Argc)
		for k, s := range resp {
			c.Argv[k] = CreateObject(ObjectTypeString, string(s.Value))
		}
		return nil
	}
	return errors.New("ProcessQueryBuffer failed")
}

// 响应请求
func (c *Client) ResponseToClient(conn net.Conn) {
	_, _ = conn.Write([]byte(c.RespBuf))
	c.RespBuf = ""
}

/*
*
* Server
*
**/
type Server struct {
	Dbs              []*GoredisDb               //
	DbNum            int                        //
	Start            int64                      //
	Port             int32                      //
	RdbFilename      string                     //
	AofFilename      string                     // Aof存储文件名
	NextClientID     int32                      //
	SystemMemorySize int32                      //
	Clients          int32                      //
	Pid              int                        //
	Commands         map[string]*GoredisCommand // 命令表
	Dirty            int64                      //
	AofBuf           []string                   //
	PubSubChannels   map[string]*List           //
	PubSubPatterns   *List                      //
}

func NewServer() *Server {
	s := new(Server)
	return s
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
		s.call(c) // 执行命令
	} else {
		c.addReplyError(fmt.Sprintf("(error) ERR unknown command '%s'", name))
	}
}

func (s *Server) call(c *Client) {
	dirty := s.Dirty
	c.Cmd.Proc(c, s)
	dirty = s.Dirty - dirty
	if dirty > 0 && !c.FakeFlag {
		AppendToFile(s.AofFilename, c.QueryBuf)
	}
}

func (s *Server) lookupCommand(name string) *GoredisCommand {
	cmd, ok := s.Commands[name]
	if ok {
		return cmd
	}
	return nil
}
