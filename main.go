package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"time"
)

const port = 8080

type request struct {
	conn     net.Conn
	method   string
	resource string
	httpVer  string
}

func openDoc(document string) (string, string) {
	file, err := os.Open("www/" + document)

	fileContents := make([]byte, 150)

	if err != nil {
		log.Panicln("Can not open file!", err)

	}
	n, _ := file.Read(fileContents)

	file.Close()

	return string(fileContents[0:n]), "200 OK"

}

func handleResponse(connCh chan request) {
	req := <-connCh

	conn := req.conn
	currTime := time.Now()

	if req.method == "GET" {

		file, respCode := openDoc("index.html")
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %s\r\nDate:%+v\r\nServer: Matts Srvr\r\nContent-Type:text/html\r\nContent-Length:%d\r\n\r\n%s", respCode, currTime, len(file), file)))
		conn.Close()

	} else {
		conn.Close()
	}

}

func parseHTTPreq(req string) []string {

	r, _ := regexp.Compile(`^(GET) (\S+) (HTTP\/1\.1)$`)
	matches := r.FindStringSubmatch(req)

	return matches[1:]
}

func handleRequest(conn net.Conn, connCh chan request) {

	requestStream := make([]byte, 1042)

	_, err := conn.Read(requestStream)

	if err != nil {
		log.Panicln(err, "error in reading request")
	}

	parsed := parseHTTPreq(string(requestStream))

	if len(parsed) == 0 || parsed[2] != "HTTP/1.1" {
		connCh <- request{conn: conn, method: parsed[0], resource: "error!"}
	} else {
		connCh <- request{conn: conn, method: parsed[0], resource: parsed[1], httpVer: parsed[2]}
	}

}

func main() {

	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))

	log.Println("lisenting on", ln.Addr().String())

	if err != nil {
		log.Fatal("can not listen...closing connection", err)
	}

	connCh := make(chan request)

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
