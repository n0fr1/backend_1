package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	var nickname string

	ch := make(chan string)
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	ch <- "You are " + who
	messages <- who + " has arrived"
	entering <- ch

	log.Println(who + " has arrived11")

	ch <- "Please, input your nickname:"
	input := bufio.NewScanner(conn)
	for input.Scan() {

		if len(nickname) == 0 {
			nickname = input.Text()
			ch <- "your nickname is: " + nickname
			log.Println(who + " changed to: " + nickname)
		} else {
			messages <- nickname + ": " + input.Text()
		}

	}

	leaving <- ch
	if len(nickname) == 0 {
		messages <- who + " has left"
	} else {
		messages <- nickname + " has left"
	}

	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		log.Println(msg)
		fmt.Fprintln(conn, msg)
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {

		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}
