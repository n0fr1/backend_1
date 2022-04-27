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

//Для отработки практических навыков
//Решил реализовать через интерфейс и структуру
type TimeInfo interface {
	start()
	catchMessage()
	handleConn(context.Context, net.Conn)
}

type Server struct {
	messages    chan string
	connections map[net.Conn]bool
	wg          sync.WaitGroup
}

func main() {

	message := make(chan string)
	connection := make(map[net.Conn]bool)

	var srv TimeInfo = &Server{
		messages:    message,
		connections: connection,
	}

	srv.start()

}

func (s *Server) start() {

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

				s.wg.Add(1)
				s.connections[conn] = true
				go s.catchMessage()
				go s.handleConn(ctx, conn)
			}
		}
	}()

	<-ctx.Done()

	log.Println("done")
	l.Close()
	s.wg.Wait()
	log.Println("exit")
}

func (s *Server) catchMessage() { //ловим сообщения от сервера

	var msg string

	for {
		fmt.Fscanln(os.Stdin, &msg)
		s.messages <- msg
	}

}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {

	defer s.wg.Done()
	defer conn.Close()

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
