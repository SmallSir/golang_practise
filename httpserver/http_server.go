package httpServer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"

	"golang.org/x/sync/errgroup"
)

func startHttpServer() error {
	httpServer := http.NewServeMux()
	httpServer.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "hello")
	})
	if err := http.ListenAndServe("0.0.0.0:8080", httpServer); err != nil {
		return errors.New("http server error")
	}
	return nil
}

func startRpcServer() error {
	if err := http.ListenAndServe("0.0.0.0:8080", http.DefaultServeMux); err != nil {
		return errors.New("rpc error")
	}
	return nil
}

func singleError() error {
	return errors.New("single error")
}

type HttpServer struct {
	g           *errgroup.Group
	serverChan  chan struct{}
	errChan     chan error
	serverError error
}

func (h *HttpServer) StartServer() {
	h.g.Go(startRpcServer)
	h.g.Go(startHttpServer)

	err := h.g.Wait()
	if err != nil {
		h.serverError = err
		h.serverChan <- struct{}{}
	}

	select {
	case <-h.serverChan: // 等待服务结束信号
		h.g.Go(singleError)
		_ = h.g.Wait()
		h.serverChan <- struct{}{}
	}
}

func NewHttpServer(context context.Context, serverChan chan struct{}, errChan chan error) *HttpServer {
	g, _ := errgroup.WithContext(context)
	return &HttpServer{
		g:          g,
		serverChan: serverChan,
		errChan:    errChan,
	}
}

func main() {
	serverChan := make(chan struct{}, 3)   // 用来将linux信号的结果发送到服务内
	singleChan := make(chan os.Signal, 10) //用来接收linux注册的信号
	errChan := make(chan error)            // 用来传递服务的报错信号
	defer close(singleChan)
	defer close(serverChan)
	defer close(errChan)

	ctx := context.Background()
	server := NewHttpServer(ctx, serverChan, errChan)
	go server.StartServer() // 启动服务

	signal.Notify(singleChan, syscall.SIGTERM) //注册信号
	select {
	case <-singleChan: //收到信号结果后会提示服务暂停, 同时打日志等待服务结束
		serverChan <- struct{}{}
		log.Fatalf("single error")
		<-errChan
		break
	case <-errChan: //收到服务报错, 打日志记录报错信息
		log.Fatalf("err is %+v", server.serverError)
		break
	}
}
