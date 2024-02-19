package main

import (
	"errors"
	"log"
	"net"
	"os"
	"regexp"
)

const (
	httpProtVer  = "HTTP/1.1"
	srvr         = "matts server"
	rootFilePath = "/www/http/src"
)

type HTTPResponse struct {
	respCode    string
	contentType string
	contentLen  string
	body        string
}

type request struct {
	identity net.Conn
	method   string
	resource string
	httpVer  string
}

func fileOrPathExists(resource string) bool {
	fullPath := rootFilePath + resource
	_, err := os.Open(fullPath)

	if err != nil {
		log.Panicln("error!", err)
		return false
	}

	return true

}

func buildResponse(responseObj *HTTPResponse) []byte {
	respString := ""

	return []byte(respString)
}

func handleRequest(reqChan chan *request) {
	req := <-reqChan
	conn := req.identity
	response := &HTTPResponse{}

	if !fileOrPathExists(req.resource) {
		response.respCode = "404 Not Found"
		conn.Write(buildResponse(response))
		conn.Close()
	} else {
		response.respCode = "200 OK"
	}
}

func parseRequest(conn net.Conn, inputStream *[]byte) (*request, error) {

	req := &request{identity: conn, method: "", resource: "", httpVer: ""}
	expressions := []string{`^(GET|POST|PUT|DELETE|OPTIONS)`, `\s\/[^\s]+`, `HTTP\/[^\s]+`}

	for i := range expressions {
		re := regexp.MustCompile(expressions[i])
		switch i {
		case 0:

			method := string(re.Find(*inputStream))
			if method != "GET" {
				return nil, errors.New("server does not support this request type")
			}
			req.method = string(re.Find(*inputStream))

		case 1:

			resource := string(re.Find(*inputStream))
			req.resource = resource

		case 2:

			version := string(re.Find(*inputStream))

			if version != httpProtVer {
				return nil, errors.New("can not accept requests from this version of http")
			}
			req.httpVer = version

		}

	}

	return req, nil
}

func handleConnection(conn net.Conn, reqChan chan *request) {
	inputStream := make([]byte, 1042)

	_, err := conn.Read(inputStream)

	if err != nil {
		log.Panicln(err, "error in reading request")
	}

	parsed, err := parseRequest(conn, &inputStream)

	if err != nil {
		log.Fatalln("FATAL", err)
	}

	reqChan <- parsed

}

func main() {

	ln, err := net.Listen("tcp", "localhost:8080")

	log.Println("lisenting on", ln.Addr().String())

	if err != nil {
		log.Fatal("can not listen...closing connection", err)
	}

	reqChan := make(chan *request) //request channel

	for {

		conn, err := ln.Accept()

		if err != nil {
			log.Panicln("Error!", err)
			return
		}

		go handleConnection(conn, reqChan)
		go handleRequest(reqChan)

	}
}
