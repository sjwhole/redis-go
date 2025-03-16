package main

import (
	"bufio"
	"encoding/binary"
	"log"
	"os"
	"strconv"
	"time"
)

func readFromRDB(fileName string) {
	// read from rdb file
	file, err := os.Open(fileName)
	if err != nil {
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	err = parseRDB(file)
	if err != nil {
		return
	}
}

func parseRDB(file *os.File) error {
	reader := bufio.NewReader(file)
	for opcode, err := reader.ReadByte(); opcode != 0xFE; opcode, err = reader.ReadByte() {
		if err != nil {
			return err
		}
	}
	_, err := reader.Discard(2)
	if err != nil {
		return err
	}

	lenMask := 0xc0
	for i := 0; i < 2; i++ {
		b, err := reader.ReadByte()
		if err != nil {
			return err
		}
		switch lenMask & int(b) {
		case 0b00000000:
			_, err = reader.Discard(0)
		case 0b01000000:
			_, err = reader.Discard(1)
		case 0b10000000:
			_, err = reader.Discard(4)
		}
	}

	for b, _ := reader.Peek(1); b[0] != 0xFF; b, _ = reader.Peek(1) {
		var expiryTime time.Time
		log.Println(expiryTime)
		switch b, _ := reader.Peek(1); int(b[0]) {
		//expiry time in seconds,
		case 0xFD:
			_, err = reader.Discard(1)
			expiryTimeBytes := make([]byte, 4)
			_, err = reader.Read(expiryTimeBytes)
			i, err := strconv.ParseInt(strconv.FormatUint(binary.LittleEndian.Uint64(expiryTimeBytes), 10), 10, 64)
			if err != nil {
				return err
			}
			expiryTime = time.Unix(i, 0)
			log.Println(expiryTime)

		//expiry time in ms
		case 0xFC:
			_, err = reader.Discard(1)
			expiryTimeBytes := make([]byte, 8)
			_, err = reader.Read(expiryTimeBytes)
			i, err := strconv.ParseInt(strconv.FormatUint(binary.LittleEndian.Uint64(expiryTimeBytes), 10), 10, 64)
			if err != nil {
				return err
			}
			expiryTime = time.UnixMilli(i)
			log.Println(expiryTime)
		//no expiry time
		default:

		}

		// Discard format
		_, err = reader.Discard(1)
		if err != nil {
			return err
		}
		// Read key
		b, err := reader.ReadByte()
		if err != nil {
			return err
		}
		keyLen := int(b)
		key := make([]byte, keyLen)
		_, err = reader.Read(key)
		if err != nil {
			return err
		}

		// Read value
		b, err = reader.ReadByte()
		if err != nil {
			return err
		}
		valueLen := int(b)
		value := make([]byte, valueLen)
		_, err = reader.Read(value)
		if err != nil {
			return err
		}
		db[string(key)] = string(value)
		// check if expiry time is set
		if !expiryTime.IsZero() {
			log.Println("Set")
			expiryTimes[string(key)] = expiryTime
		}
	}

	return nil
}
