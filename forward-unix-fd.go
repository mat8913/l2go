// forward-unix-fd
// Usage: ./forward-unix-fd <SOCKET>
// For use as ssh ProxyCommand with ProxyUseFdpass=yes
// See: https://www.gabriel.urdhr.fr/2016/08/07/openssh-proxyusefdpass/

package main

import (
	"net"
	"os"
	"reflect"
	"syscall"
)

// Source: https://github.com/higebu/netfd/blob/ed17b5f1ac32df732afbeeab4acd7a911e9eeacb/netfd.go
func get_conn_fd(c net.Conn) int {
	v := reflect.Indirect(reflect.ValueOf(c))
	conn := v.FieldByName("conn")
	netFD := reflect.Indirect(conn.FieldByName("fd"))
	pfd := netFD.FieldByName("pfd")
	fd := int(pfd.FieldByName("Sysfd").Int())
	return fd
}

func main() {
	if len(os.Args) != 2 {
		panic("Invalid usage")
	}
	sock_path := os.Args[1]

	conn, err := net.Dial("unix", sock_path)
	if err != nil {
		panic(err)
	}

	fd := get_conn_fd(conn)

	rights := syscall.UnixRights(fd)
	err = syscall.Sendmsg(1, nil, rights, nil, 0)
	if err != nil {
		panic(err)
	}
}
