package main

import (
	"bytes"
	"io"
	"log"
	"testing"
	"time"
)

func TestSetCommand_Execute_No_Expiry(t *testing.T) {
	testString := "*4\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"

	cmdParser := NewParser([]byte(testString))
	command, err := cmdParser.Parse()
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)

	_, err = command.Execute(io.Writer(buf))
	if err != nil {
		t.Errorf("Error executing command: %s", err)
	}
	if buf.String() != "+OK\r\n" {
		t.Errorf("Expected +OK, got %s", buf.String())
	}

	command = GetCommand{key: "key"}
	buf.Reset()
	_, err = command.Execute(io.Writer(buf))
	if buf.String() != "+value\r\n" {
		t.Errorf("Expected value, got %s", buf.String())
	}

}
func TestSetCommand_Execute_With_Expiry(t *testing.T) {
	testString := "*6\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n$2\r\nPX\r\n$3\r\n100\r\n"

	cmdParser := NewParser([]byte(testString))
	command, err := cmdParser.Parse()
	if err != nil {
		return
	}
	// new string builder
	buf := new(bytes.Buffer)
	n, err := command.Execute(io.Writer(buf))
	if err != nil {
		t.Errorf("Error executing command: %s", err)
	}
	if n != 5 {
		t.Errorf("Expected 5, got %d", n)
	}
	log.Println(buf.String())

	buf.Reset()
	time.Sleep(200 * time.Millisecond)

	command = GetCommand{key: "key"}
	_, err = command.Execute(io.Writer(buf))
	log.Println(buf.String())
	if buf.String() != "$-1\r\n" {
		t.Errorf("Expected nil, got %s", buf.String())
	}
}

func TestConfigGetCommand_Execute(t *testing.T) {
	config["dir"] = "/tmp/redis-files"

	testString := "*3\r\n$6\r\nCONFIG\r\n$3\r\nGET\r\n$3\r\ndir\r\n"

	cmdParser := NewParser([]byte(testString))
	command, err := cmdParser.Parse()
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)

	_, err = command.Execute(io.Writer(buf))
	if err != nil {
		t.Errorf("Error executing command: %s", err)
	}
	if buf.String() != "*2\r\n$3\r\ndir\r\n$16\r\n/tmp/redis-files\r\n" {
		t.Errorf("Expected *2\n$3\ndir\n$16\n/tmp/redis-files\n, got %s", buf.String())
	}
}
