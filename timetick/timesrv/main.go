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
	wg          sync.WaitGroup
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

	log.Println("im started!")

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println(err)
				return
			} else {
				srv.wg.Add(1)
				srv.connections[conn] = true
				go srv.handleConn(ctx, conn)
			}
		}
	}()

	<-ctx.Done()

	log.Println("done")
	l.Close()
	srv.wg.Wait()
	log.Println("exit")
}

func (s *Server) catchMessage() { //ловим сообщения от сервера

	var msg string

	for {
		fmt.Fscan(os.Stdin, &msg)
		s.messages <- msg
	}

}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {

	defer s.wg.Done()
	defer conn.Close()

	go s.catchMessage()

	// каждую 1 секунду отправлять клиентам текущее время сервера
	tck := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			fmt.Fprintf(conn, "%s\n", "Bye-bye!")
			delete(s.connections, conn) //очищаем соединения
			return
		case t := <-tck.C:
			fmt.Fprintf(conn, "now: %s\n", t)
		case msg := <-s.messages: //выводим сообщение всем клиентам
			for connect := range s.connections {
				fmt.Fprintf(connect, "!!! Message: %s\n", msg)
			}
		}

	}
}
