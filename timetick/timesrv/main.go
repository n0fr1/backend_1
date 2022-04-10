package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Server struct {
	messages    chan string
	connections map[net.Conn]bool
}

func NewServer() Server {

	message := make(chan string)
	connection := make(map[net.Conn]bool)

	return Server{
		messages:    message,
		connections: connection,
	}
}

func main() {

	srv := NewServer()

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	cfg := net.ListenConfig{
		KeepAlive: time.Minute,
	}

	l, err := cfg.Listen(ctx, "tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	wg := &sync.WaitGroup{}
	log.Println("im started!")

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println(err)
				return
			} else {
				wg.Add(1)
				srv.connections[conn] = true
				go srv.handleConn(ctx, conn, wg)
			}
		}
	}()

	<-ctx.Done()

	log.Println("done")
	l.Close()
	wg.Wait()
	log.Println("exit")
}

func (s *Server) catchMessage() {

	var msg string

	for {
		fmt.Fscan(os.Stdin, &msg)
		s.messages <- msg
	}

}

func (s *Server) handleConn(ctx context.Context, conn net.Conn, wg *sync.WaitGroup) { //, connections map[net.Conn]bool) {
	defer wg.Done()
	defer conn.Close()

	go s.catchMessage()

	// каждую 1 секунду отправлять клиентам текущее время сервера
	tck := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			fmt.Fprintf(conn, "%s\n", "Bye!")
			delete(s.connections, conn)
			return
		case t := <-tck.C:
			fmt.Fprintf(conn, "now: %s\n", t)
		case msg := <-s.messages:
			for connect := range s.connections {
				fmt.Fprintf(connect, "!!! Message: %s\n", msg)
			}
		}

	}
}
