package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client chan<- string

type clients struct {
	entering chan client
	leaving  chan client
	messages chan string
}

func NewClients() clients {

	entering := make(chan client)
	leaving := make(chan client)
	messages := make(chan string)

	return clients{
		entering: entering,
		leaving:  leaving,
		messages: messages,
	}
}

func main() {

	nclients := NewClients()

	listener, err := net.Listen("tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	go nclients.broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go nclients.handleConn(conn)
	}
}

func (c *clients) handleConn(conn net.Conn) {
	var nickname string

	ch := make(chan string)
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	ch <- "You are " + who
	c.messages <- who + " has arrived"
	c.entering <- ch

	log.Println(who + " has arrived")

	ch <- "Please, input your nickname:" //добавляем возможность указывать nickname
	input := bufio.NewScanner(conn)
	for input.Scan() {

		if len(nickname) == 0 {
			nickname = input.Text()
			ch <- "your nickname is: " + nickname
			log.Println(who + " changed to: " + nickname)
		} else {
			c.messages <- nickname + ": " + input.Text()
		}

	}

	c.leaving <- ch

	if len(nickname) == 0 {
		c.messages <- who + " has left"
	} else {
		c.messages <- nickname + " has left"
	}

	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func (c *clients) broadcaster() {

	clients := make(map[client]bool)

	for {
		select {

		case msg := <-c.messages:
			for cli := range clients {
				cli <- msg
			}
		case cli := <-c.entering:
			clients[cli] = true

		case cli := <-c.leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}
