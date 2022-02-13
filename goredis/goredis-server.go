package main

import (
	"GoLearn/goredis/core"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	DefaultAofFile = "./goredis.aof"
)

// 服务端实例
var goredis = core.NewServer()

func main() {
	/*---- 命令行参数处理 ----*/
	argv := os.Args
	argc := len(os.Args)
	if argc >= 2 {
		if argv[1] == "-v" || argv[1] == "--version" {
			version()
		}
		if argv[1] == "--help" || argv[1] == "-h" {
			usage()
		}
	}

	/*---- 监听信号 平滑退出 ----*/
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go sigHandler(c)

	/*---- 初始化服务端实例 ----*/
	initServer()

	/*---- 网络处理 ----*/
	netListen, err := net.Listen("tcp", "127.0.0.1:9736")
	if err != nil {
		log.Print("listen err ")
	}
	defer netListen.Close()
	for {
		conn, err := netListen.Accept()
		if err != nil {
			continue
		}
		go handle(conn)
	}
}

// 处理请求
func handle(conn net.Conn) {
	c := goredis.CreateClient()
	for {
		if c.Flags&core.CLIENT_PUBSUB > 0 {
			if c.Buf != "" {
				responseConn(conn, c)
				c.Buf = ""
			}
			time.Sleep(1)
		} else {
			err := c.ReadQueryFromClient(conn)
			if err != nil {
				log.Println("ReadQueryFromClient err", err)
				return
			}
			err = c.ProcessInputBuffer()
			if err != nil {
				log.Println("ProcessInputBuffer err", err)
				return
			}
			goredis.ProcessCommand(c)
			responseConn(conn, c)
		}
	}
}

// 响应返回给客户端
func responseConn(conn net.Conn, c *core.Client) {
	_, _ = conn.Write([]byte(c.Buf))
}

// 初始化服务端实例
func initServer() {
	goredis.Pid = os.Getpid()
	goredis.DbNum = 16
	initDb()
	goredis.Start = time.Now().UnixNano() / 1000000
	//var getf server.CmdFun
	goredis.AofFilename = DefaultAofFile

	getCommand := &core.GoredisCommand{Name: "get", Proc: core.GetCommand}
	setCommand := &core.GoredisCommand{Name: "set", Proc: core.SetCommand}
	subscribeCommand := &core.GoredisCommand{Name: "subscribe", Proc: core.SubscribeCommand}
	publishCommand := &core.GoredisCommand{Name: "publish", Proc: core.PublishCommand}
	geoaddCommand := &core.GoredisCommand{Name: "geoadd", Proc: core.GeoAddCommand}
	geohashCommand := &core.GoredisCommand{Name: "geohash", Proc: core.GeoHashCommand}
	geoposCommand := &core.GoredisCommand{Name: "geopos", Proc: core.GeoPosCommand}
	geodistCommand := &core.GoredisCommand{Name: "geodist", Proc: core.GeoDistCommand}
	georadiusCommand := &core.GoredisCommand{Name: "georadius", Proc: core.GeoRadiusCommand}
	georadiusbymemberCommand := &core.GoredisCommand{Name: "georadiusbymember", Proc: core.GeoRadiusByMemberCommand}

	goredis.Commands = map[string]*core.GoredisCommand{
		"get":               getCommand,
		"set":               setCommand,
		"geoadd":            geoaddCommand,
		"geohash":           geohashCommand,
		"geopos":            geoposCommand,
		"geodist":           geodistCommand,
		"georadius":         georadiusCommand,
		"georadiusbymember": georadiusbymemberCommand,
		"subscribe":         subscribeCommand,
		"publish":           publishCommand,
	}
	tmp := make(map[string]*core.List)
	goredis.PubSubChannels = &tmp
	LoadData()
}

// 初始化db
func initDb() {
	goredis.Dbs = make([]*core.GoredisDb, goredis.DbNum)
	for i := 0; i < goredis.DbNum; i++ {
		goredis.Dbs[i] = new(core.GoredisDb)
		goredis.Dbs[i].Dict = make(map[string]*core.GoredisObject, 100)
	}
}

// 持久化load dada
func LoadData() {
	c := goredis.CreateClient()
	c.FakeFlag = true
	pros := core.ReadAof(goredis.AofFilename)
	for _, v := range pros {
		c.QueryBuf = string(v)
		err := c.ProcessInputBuffer()
		if err != nil {
			log.Println("ProcessInputBuffer err", err)
		}
		goredis.ProcessCommand(c)
	}
}

// 监听信号处理
func sigHandler(c chan os.Signal) {
	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			exitHandler()
		default:
			fmt.Println("signal ", s)
		}
	}
}

func exitHandler() {
	fmt.Println("exiting smoothly ...")
	fmt.Println("bye ")
	os.Exit(0)
}

func version() {
	println("Goredis server v=0.0.1 sha=xxxxxxx:001 malloc=libc-go bits=64 ")
	os.Exit(0)
}

func usage() {
	println("Usage: ./goredis-server [/path/to/redis.conf] [options]")
	println("       ./goredis-server - (read config from stdin)")
	println("       ./goredis-server -v or --version")
	println("       ./goredis-server -h or --help")
	println("Examples:")
	println("       ./goredis-server (run the server with default conf)")
	println("       ./goredis-server /etc/redis/6379.conf")
	println("       ./goredis-server --port 7777")
	println("       ./goredis-server --port 7777 --slaveof 127.0.0.1 8888")
	println("       ./goredis-server /etc/myredis.conf --loglevel verbose")
	println("Sentinel mode:")
	println("       ./goredis-server /etc/sentinel.conf --sentinel")
	os.Exit(0)
}
