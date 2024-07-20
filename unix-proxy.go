// unix-proxy
// Create a unix socket which proxies to an existing unix socket
// Usage: ./unix-proxy <EXISTING SOCKET> <PROXY SOCKET>

package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func handle_connection(conn net.Conn, connect_sock string) {
	log.Println("got connection", conn)
	conn2, err := net.Dial("unix", connect_sock)
	if err != nil {
		log.Println("error on", conn, err)
		conn.Close()
		return
	}

	uc1 := conn.(*net.UnixConn)
	uc2 := conn2.(*net.UnixConn)
	proxy(uc1, uc2)
}

func proxy(conn1, conn2 *net.UnixConn) {
	defer conn1.Close()
	defer conn2.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer conn1.CloseWrite()
		io.Copy(conn1, conn2)
	}()

	go func() {
		defer wg.Done()
		defer conn2.CloseWrite()
		io.Copy(conn2, conn1)
	}()

	wg.Wait()
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
	if len(os.Args) != 3 {
		panic("Invalid usage")
	}

	connect_sock := os.Args[1]
	listen_sock := os.Args[2]

	listener, err := net.Listen("unix", listen_sock)
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
			go handle_connection(conn, connect_sock)
		case sig := <-sigs:
			log.Println("got signal", sig, "- exiting")
			return
		}
	}
}
