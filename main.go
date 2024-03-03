package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

const (
	httpVer = "HTTP/1.1"
	rateLim = 10
	port    = 8080
)

func openDoc(document string) (string, string) {
	file, err := os.Open("www/" + document)

	fileContents := make([]byte, 150)

	if err != nil {
		log.Panicln("Can not opne file!", err)

	}
	n, _ := file.Read(fileContents)

	file.Close()

	return string(fileContents[0:n]), "200 OK"

}

func handleResponse(connCh chan net.Conn) {
	conn := <-connCh
	currTime := time.Now()
	file, respCode := openDoc("index.html")
	conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %s\r\nDate:%+v\r\nServer: Matts Srvr\r\nContent-Type:text/html\r\nContent-Length:%d\r\n\r\n%s", respCode, currTime, len(file), file)))
	conn.Close()

}

func handleRequest(conn net.Conn, connCh chan net.Conn) {

	requestStream := make([]byte, 1042)

	_, err := conn.Read(requestStream)

	if err != nil {
		log.Panicln(err, "error in reading request")
	}
	connCh <- conn
}

func main() {

	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))

	log.Println("lisenting on", ln.Addr().String())

	if err != nil {
		log.Fatal("can not listen...closing connection", err)
	}

	connCh := make(chan net.Conn)

	for {

		conn, err := ln.Accept()

		if err != nil {
			log.Panicln("Error!", err)
			return
		}

		go handleRequest(conn, connCh)
		go handleResponse(connCh)

	}
}
