package tracing

import (
	"GoLearn/etcd"
	"GoLearn/etcd/expample/pb"
	"context"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func Test_Tracer(t *testing.T) {
	tracer, closer := NewJaegerTracer("jaeger-test-demo")
	defer closer.Close()
	// 创建 root span
	span := tracer.StartSpan("span_root")
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	r1 := spanDemo("Hello spanDemo", ctx)
	r2 := spanDemo2("Hello spanDemo2", ctx)
	fmt.Printf("Resp demo: %v \n", r1)
	fmt.Printf("Resp demo2: %v\n", r2)
	span.Finish()
}

var (
	serv = flag.String("service", "hello_service", "service name")
	port = flag.Int("port", 50001, "listening port")
	reg  = flag.String("reg", "http://127.0.0.1:2379", "register etcd address")
)

const (
	clientName = "HelloClient"
	serverName  = "HelloServer"
)

func Test_TraceClient(t *testing.T){
	flag.Parse()
	fmt.Printf("service: %v:%v", *serv, *port)

	// 使用拦截器加入tracer，并创建连接
	resolver := etcd.NewResolver(*serv)
	robin := grpc.RoundRobin(resolver)
	tracer, _ := NewJaegerTracer(clientName)
	chainInter := grpc_middleware.ChainUnaryClient(
		OpenTracingClientInterceptor(tracer),
	)
	dialOpts := []grpc.DialOption{
		grpc.WithBalancer(robin),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(chainInter),
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := grpc.DialContext(ctx, *reg, dialOpts...)
	if err != nil {
		panic(err)
	}
	fmt.Println("conn...")

	// 生成client 并发送请求
	ticker := time.NewTicker(1 * time.Second)
	for t := range ticker.C {
		client := pb.NewGreeterClient(conn)
		resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "world " + strconv.Itoa(t.Second())})
		if err == nil {
			fmt.Printf("%v: Reply is %s\n", t, resp.Message)
		} else {
			fmt.Println(err)
		}
	}
}

func Test_TraceServer(t *testing.T){
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		panic(err)
	}

	err = etcd.Register(*serv, "127.0.0.1", *port, *reg, time.Second*10, 15)
	if err != nil {
		panic(err)
	}

	// 使用拦截器加入tracer
	tracer, closer := NewJaegerTracer(serverName)
	defer closer.Close()
	chainInter := grpc_middleware.ChainUnaryServer(
		OpentracingServerInterceptor(tracer),
	)
	servOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(chainInter),
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		log.Printf("receive signal '%v'", s)
		_ = etcd.UnRegister()
		os.Exit(1)
	}()

	log.Printf("starting hello service at %d", *port)
	s := grpc.NewServer(servOpts...)
	pb.RegisterGreeterServer(s, &server{})
	s.Serve(lis)
}

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("%v: Receive is %s\n", time.Now(), in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}