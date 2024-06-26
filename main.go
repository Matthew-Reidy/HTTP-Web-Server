package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"time"
)

const (
	port           = 8080
	allowedHTTPVer = "HTTP/1.1"
	strikeLimit    = 10
	NotFound       = "404 Not Found"
	Ok             = "200 OK"
)

type request struct {
	conn     net.Conn
	method   string
	resource string
	httpVer  string
}

type profile struct {
	strikes     int64     //instances messages that are within x seconds of eachother
	lastmessage time.Time //last message or request made by a requestor
	isBanned    bool      //boolean is user banned
	bannedTime  time.Time //time of ban, used in unban logic
}

func openDoc(document string) (string, string) {

	pathway := "www" + document

	if _, err := os.Stat(pathway); !os.IsNotExist(err) {

		file, err := os.Open(pathway)

		f, _ := file.Stat()

		if err != nil {
			log.Println("error! ", err)
		}
		defer file.Close()

		reader := bufio.NewReader(file)

		fileContents := make([]byte, f.Size())

		n, readErr := reader.Read(fileContents)

		if readErr != nil {
			log.Println("error occured reading document! ", pathway, readErr)
		}

		return string(fileContents[:n]), Ok

	} else {

		file, err := os.Open("www/404.html")

		if err != nil {
			log.Println(err)
		}
		defer file.Close()

		fileContent := make([]byte, 150)

		n, readErr := file.Read(fileContent)

		if readErr != nil {
			log.Println("error occured reading document! ", pathway, readErr)
		}

		return string(fileContent[:n]), NotFound
	}

}

func handleResponse(connCh chan *request) {
	req := <-connCh
	//todo: add logic to handle 403 responses, 404 responses
	//todo: add some sort of functionality to identify js and css files used by the HTML and send with the
	conn := req.conn
	currTime := time.Now()
	defer conn.Close()

	if req.httpVer == allowedHTTPVer {

		if req.method == "GET" {
			file, respCode := openDoc(req.resource)
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %s\r\nDate:%+v\r\nServer: Matts Srvr\r\nContent-Type:text/html\r\nContent-Length:%d\r\n\r\n%s", respCode, currTime, len(file), file)))

		} else {
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 403 Forbidden\r\nDate:%+v\r\nServer: Matts Srvr\r\nContent-Type:text/html\r\nContent-Length:25\r\n\r\n<h1>forbidden!</h1>", currTime)))

		}
	}

}

func parseHTTPreq(req []byte) []string {

	r := regexp.MustCompile(`(GET|POST|PUT|DELETE) (.+?) (HTTP\/1\.1)`)
	matches := r.FindStringSubmatch(string(req))

	return matches
}

func handleRequest(conn net.Conn, connCh chan *request, requestorProfile map[string]*profile) {

	requestStream := make([]byte, 1042)
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)

	if _, exists := requestorProfile[remoteAddr.IP.String()]; !exists {

		log.Printf("%s added to profile list", remoteAddr.IP.String())
		requestorProfile[remoteAddr.IP.String()] = &profile{
			strikes:     0,
			lastmessage: time.Now(),
			isBanned:    false,
			bannedTime:  time.Time{},
		}

	} else {

		log.Printf("%s is visiting again from the profile list", remoteAddr.IP.String())
		profile := requestorProfile[remoteAddr.IP.String()]

		if banCheck := requestorProfile[remoteAddr.IP.String()].isBanned; banCheck {
			//todo: add some sort of unban logic
			log.Printf("%s is banned!", remoteAddr.IP.String())
			conn.Close()
			return

		} else {

			currentTime := time.Now()

			if currentTime.Sub(profile.lastmessage) <= 5*time.Second {

				log.Printf("%s strike! num of strikes! %d", remoteAddr.IP.String(), requestorProfile[remoteAddr.IP.String()].strikes)
				profile.strikes++

				if profile.strikes >= strikeLimit {

					profile.bannedTime = time.Now()
					profile.isBanned = true
					conn.Close()
					return

				}
			} else {

				profile.strikes = 0

			}

		}

	}
	n, err := conn.Read(requestStream)

	if err != nil && err != io.EOF {
		log.Panicln(err, "error in reading request")
	}

	parsed := parseHTTPreq(requestStream[:n])

	connCh <- &request{
		conn:     conn,
		method:   parsed[1],
		resource: parsed[2],
		httpVer:  parsed[3],
	}

}

func main() {

	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))

	log.Println("lisenting on", ln.Addr().String())

	if err != nil {
		log.Fatal("can not listen...closing connection", err)
	}

	connCh := make(chan *request)
	requestorProfile := make(map[string]*profile)

	for {

		conn, err := ln.Accept()
		log.Printf("%s connected!", conn.RemoteAddr().String())
		if err != nil {
			log.Panicln("Error!", err)
			return
		}

		go handleRequest(conn, connCh, requestorProfile)
		go handleResponse(connCh)

	}
}
