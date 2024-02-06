package main

import (
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", "localhost:8080")
	log.Println("lisenting", ln.Addr().String())
	if err != nil {
		log.Fatal("can not listen...closing connection", err)

	}

	defer ln.Close()

	for {

		conn, err := ln.Accept()

		if err != nil {
			log.Panicln("Error!", err)
			return
		}

		go handleConnection(conn)

	}
}

func handleConnection(conn net.Conn) {

}
