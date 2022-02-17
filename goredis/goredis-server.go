package main

import (
	"GoLearn/goredis/core"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	DefaultAofFile = "./server.aof"
)

// 服务端实例
var server = core.NewServer()

func main() {
	IPPort := "127.0.0.1:9736"

	/*---- 监听信号 平滑退出 ----*/
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go sigHandler(c)

	/*---- 初始化服务端实例 ----*/
	initServer()

	/*---- 网络处理 ----*/
	netListen, err := net.Listen("tcp", IPPort)
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
	c := core.NewClient(server.Dbs[0])
	for {
		if c.Flags&core.CLIENT_PUBSUB > 0 {
			if c.RespBuf != "" {
				c.ResponseToClient(conn)
			}
			time.Sleep(1)
		} else {
			err := c.ReadQueryFromClient(conn)
			if err != nil {
				log.Println("ReadQueryFromClient err", err)
				return
			}
			err = c.ProcessQueryBuffer()
			if err != nil {
				log.Println("ProcessQueryBuffer err", err)
				return
			}
			server.ProcessCommand(c)
			c.ResponseToClient(conn)
		}
	}
}

// 初始化服务端实例
func initServer() {
	server.Pid = os.Getpid()
	server.Start = time.Now().UnixNano() / 1000000
	server.AofFilename = DefaultAofFile
	server.PubSubChannels = make(map[string]*core.List)

	initDb()
	initCommand()
	loadAof()
}

func initDb() {
	server.DbNum = 16
	server.Dbs = make([]*core.GoredisDb, server.DbNum)
	for i := 0; i < server.DbNum; i++ {
		server.Dbs[i] = new(core.GoredisDb)
		server.Dbs[i].Dict = make(map[string]*core.GoredisObject, 100)
	}
}

func initCommand() {
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

	server.Commands = map[string]*core.GoredisCommand{
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
}

func loadAof() {
	c := core.NewClient(server.Dbs[0])
	c.FakeFlag = true
	pros := core.ReadAof(server.AofFilename)
	for _, v := range pros {
		c.QueryBuf = v
		err := c.ProcessQueryBuffer()
		if err != nil {
			log.Println("ProcessQueryBuffer err", err)
		}
		server.ProcessCommand(c)
	}
}

// 监听信号处理
func sigHandler(c chan os.Signal) {
	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			exitHandler()
		default:
			println("signal ", s)
		}
	}
}

func exitHandler() {
	println("exiting success ...")
	println("bye bye")
	os.Exit(0)
}
