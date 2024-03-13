package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"time"
)

const (
	port           = 8080
	allowedHTTPVer = "HTTP/1.1"
	strikeLimit    = 3
)

type request struct {
	conn     net.Conn
	method   string
	resource string
	httpVer  string
}

type profile struct {
	strikes     int //instances messages that are within 5 seconds of
	lastmessage time.Time
	isBanned    bool
	bannedTime  time.Time
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

	r := regexp.MustCompile(`(?m)^(GET) (\S+) (HTTP\/1\.1)$`)
	matches := r.FindStringSubmatch(req)
	log.Println(matches)
	return matches
}

func handleRequest(conn net.Conn, connCh chan request) {

	requestStream := make([]byte, 1042)
	requestorProfile := make(map[string]profile)

	if _, exists := requestorProfile[conn.RemoteAddr().String()]; !exists {

		requestorProfile[conn.RemoteAddr().String()] = profile{strikes: 0, lastmessage: time.Now(), isBanned: false, bannedTime: time.Time{}}

	} else {
		if banCheck := requestorProfile[conn.RemoteAddr().String()].isBanned; banCheck {
			conn.Close()
		} else {
			currentTime := time.Now()
			profile := requestorProfile[conn.RemoteAddr().String()]
			if currentTime.Sub(profile.lastmessage) <= (5 * time.Second) {
				profile.strikes += 1
				if profile.strikes >= 3 {
					profile.bannedTime = time.Now()
					profile.isBanned = true
					conn.Close()
				}
			}
		}

	}
	_, err := conn.Read(requestStream)

	if err != nil {
		log.Panicln(err, "error in reading request")
	}

	parsed := parseHTTPreq(string(requestStream))
	log.Println(parsed)
	connCh <- request{conn: conn, method: "GET", resource: "/", httpVer: allowedHTTPVer}

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
		log.Printf("%s connected!", conn.RemoteAddr().String())
		if err != nil {
			log.Panicln("Error!", err)
			return
		}

		go handleRequest(conn, connCh)
		go handleResponse(connCh)

	}
}
