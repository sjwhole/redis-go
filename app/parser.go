package main

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

var db = make(map[string]string)

type Parser struct {
	message []byte
}

func NewParser(message []byte) *Parser {
	return &Parser{message: message}
}

func (p *Parser) Parse() (Command, error) {
	log.Println(string(p.message))

	parts := strings.Split(string(p.message), "\r\n")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid message")
	}

	cmd := strings.ToUpper(parts[2])
	switch cmd {
	case "PING":
		return PingCommand{}, nil
	case "ECHO":
		return EchoCommand{message: parts[4]}, nil
	case "SET":
		if len(parts) < 12 {
			return SetCommand{key: parts[4], value: parts[6], expiryMilSec: -1}, nil
		} else {
			parsedMilSec, err := strconv.Atoi(parts[10])
			if err != nil {
				return nil, fmt.Errorf("invalid expiry time")
			}
			return SetCommand{key: parts[4], value: parts[6], expiryMilSec: parsedMilSec}, nil
		}
	case "GET":
		return GetCommand{key: parts[4]}, nil
	case "CONFIG":
		if strings.ToUpper(parts[4]) == "GET" {
			return ConfigGetCommand{key: parts[6]}, nil
		} else {
			return nil, fmt.Errorf("unknown command %s", cmd)
		}
	case "KEYS":
		return KeysCommand{pattern: parts[4]}, nil

	case "INFO":
		if flag["replicaof"] != nil {
			return InfoCommand{role: "slave"}, nil
		}
		return InfoCommand{role: "master"}, nil

	default:
		return nil, fmt.Errorf("unknown command %s", cmd)
	}
}

type Command interface {
	Execute(io io.Writer) (int, error)
}

type PingCommand struct {
}

func (p PingCommand) Execute(io io.Writer) (int, error) {
	n, err := io.Write([]byte("+PONG\r\n"))
	if err != nil {
		return -1, err
	}
	return n, nil
}

type EchoCommand struct {
	message string
}

func (e EchoCommand) Execute(io io.Writer) (int, error) {
	n, err := io.Write([]byte(fmt.Sprintf("+%s\r\n", e.message)))
	if err != nil {
		return -1, err
	}
	return n, nil
}

type SetCommand struct {
	key          string
	value        string
	expiryMilSec int
}

func (s SetCommand) Execute(io io.Writer) (int, error) {
	db[s.key] = s.value
	if s.expiryMilSec > 0 {
		timer := time.After(time.Duration(s.expiryMilSec) * time.Millisecond)
		go func() {
			<-timer
			delete(db, s.key)
		}()
	}

	n, err := io.Write([]byte("+OK\r\n"))
	if err != nil {
		return -1, err
	}
	return n, nil
}

type GetCommand struct {
	key string
}

func (g GetCommand) Execute(io io.Writer) (int, error) {
	if _, ok := db[g.key]; !ok {
		n, err := io.Write([]byte("$-1\r\n"))
		if err != nil {
			return -1, err
		}
		return n, nil
	}

	if expiry, exists := expiryTimes[g.key]; exists {
		if time.Now().After(expiry) {
			n, err := io.Write([]byte("$-1\r\n"))
			if err != nil {
				return -1, err
			}
			return n, nil
		}
	}

	n, err := io.Write([]byte(fmt.Sprintf("+%s\r\n", db[g.key])))
	if err != nil {
		return -1, err
	}
	return n, nil
}

type ConfigGetCommand struct {
	key string
}

func (c ConfigGetCommand) Execute(io io.Writer) (int, error) {
	if _, ok := config[c.key]; !ok {
		n, err := io.Write([]byte("$-1\r\n"))
		if err != nil {
			return -1, err
		}
		return n, nil
	}

	n, err := io.Write([]byte(fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(c.key), c.key, len(config[c.key]), config[c.key])))
	if err != nil {
		return -1, err
	}
	return n, nil
}

type KeysCommand struct {
	pattern string
}

func (k KeysCommand) Execute(io io.Writer) (int, error) {
	if k.pattern == "*" {
		keys := make([]string, 0)
		for key := range db {
			if expiry, exists := expiryTimes[key]; exists {
				if time.Now().After(expiry) {
					continue
				}
			}
			keys = append(keys, key)
		}

		buf := new(strings.Builder)
		buf.Write([]byte("*" + strconv.Itoa(len(keys)) + "\r\n"))
		for _, key := range keys {
			buf.Write([]byte("$" + strconv.Itoa(len(key)) + "\r\n" + key + "\r\n"))
		}

		n, err := io.Write([]byte(buf.String()))
		if err != nil {
			return -1, err
		}
		return n, nil
	}
	return -1, fmt.Errorf("not implemented")
}

type InfoCommand struct {
	role string
}

func (i InfoCommand) Execute(io io.Writer) (int, error) {
	log.Println(i.role)
	response := fmt.Sprintf("role:%s", i.role)
	length := len(response)
	bulkResponse := fmt.Sprintf("$%d\r\n%s\r\n", length, response)

	n, err := io.Write([]byte(bulkResponse))
	if err != nil {
		return 0, err
	}
	return n, nil
}
