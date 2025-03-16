package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var config = make(map[string]string)
var expiryTimes = make(map[string]time.Time)

var flag = make(map[string][]string)

func Init() {
	os.Args = os.Args[1:]

	flag["port"] = []string{"6379"}

	for len(os.Args) > 0 {
		switch os.Args[0] {
		case "--dir":
			flag["dir"] = []string{os.Args[1]}
			os.Args = os.Args[2:]
		case "--dbfilename":
			flag["dbfilename"] = []string{os.Args[1]}
			os.Args = os.Args[2:]
		case "--port":
			flag["port"] = []string{os.Args[1]}
			os.Args = os.Args[2:]
		case "--replicaof":
			flag["replicaof"] = []string{os.Args[1], os.Args[2]}
			os.Args = os.Args[3:]
		default:
			os.Args = os.Args[1:]
		}

	}

	if dir, ok := flag["dir"]; ok {
		config["dir"] = dir[0]
	}
	if dbfilename, ok := flag["dbfilename"]; ok {
		config["dbfilename"] = dbfilename[0]
	}

	readFromRDB(config["dir"] + "/" + config["dbfilename"])
}

func main() {
	Init()

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", flag["port"][0]))
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleClient(conn)
		defer conn.Close()
	}
}

func handleClient(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		// read the incoming connection into the buffer
		n, err := conn.Read(buf)
		log.Println("Received:", string(buf[:n]))
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			fmt.Println("Error reading:", err.Error())
			os.Exit(1)
		}
		cmdParser := NewParser(buf[:n])
		cmd, err := cmdParser.Parse()
		if err != nil {
			log.Println("Command error:", err)
			conn.Write([]byte(fmt.Sprintf("-ERR unknown command '%s'\r\n", err.Error())))
			continue
		}
		_, err = cmd.Execute(conn)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

