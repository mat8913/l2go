// echo server
// test with: socat STDIO UNIX-CONNECT:echo.sock

package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func handle_connection(conn net.Conn) {
	defer conn.Close()
	log.Println("got connection", conn)
	io.Copy(conn, conn)
	log.Println("reached EOF", conn)
}

func listener_to_channel(listener net.Listener, connections chan<- net.Conn) {
	defer close(connections)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		connections <- conn
	}
}

func main() {
	listener, err := net.Listen("unix", "echo.sock")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	log.Println("listening", listener)

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	connections := make(chan net.Conn)
	go listener_to_channel(listener, connections)

	for {
		select {
		case conn, more := <-connections:
			if !more {
				log.Println("no more connections - exiting")
				return
			}
			go handle_connection(conn)
		case sig := <-sigs:
			log.Println("got signal", sig, "- exiting")
			return
		}
	}
}
