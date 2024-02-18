package main

import (
	"log"
	"net"
	"regexp"
	"errors"
)

const (
	httpProtVer = "HTTP/1.1"
	srvr    = "matts server"
)

type HTTPResponse struct {
	respCode    string
	contentType string
	contentLen  string
	body        string
}



type request struct {
	identity net.Conn
	method string
	resource string
	httpVer string

}

func fileOrPathExists(resource string) bool{
	return false
}

func handleRequest(reqChan chan *request){

	/*TODO
	
	*/
}

func parseRequest(conn net.Conn, inputStream *[]byte) (*request, error) {
	
	req := &request{identity: conn}
	expressions := []string{`^(GET|POST|PUT|DELETE|OPTIONS)`, `\/[^\s]+`, `HTTP/(\d\.\d)$`}
	
	for i := range expressions {
		re := regexp.MustCompile(expressions[i])
		switch i {
		case 0:
			req.method = string(re.Find(*inputStream))
		case 1:
			req.resource = string(re.Find(*inputStream))
		case 2:
			version := string(re.Find(*inputStream))
			if version != httpProtVer{
				return nil , errors.New("can not accept requests from this version of http")
			}
			req.httpVer = string(re.Find(*inputStream))
		
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

	if err != nil{
		log.Fatalln("FATAL" , err)
	}

	reqChan <- parsed

}

func main() {

	ln, err := net.Listen("tcp", "localhost:8080")

	log.Println("lisenting on", ln.Addr().String())

	if err != nil {
		log.Fatal("can not listen...closing connection", err)
	}

	reqChan := make(chan *request)

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
