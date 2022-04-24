package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"sync"
)

type RawServerSocketAddress struct {
	IpParts [4]byte
	Port    uint16
}

func (addr RawServerSocketAddress) toSocketAddress() SocketAddress {
	ip := net.IPv4(addr.IpParts[0], addr.IpParts[1], addr.IpParts[2], addr.IpParts[3]).String()

	return SocketAddress{
		Host: ip,
		Port: int(addr.Port),
	}
}

func ReadMasterServer(socketAddress string, retries int, timeout int) ([]SocketAddress, error) {
	addresses := make([]SocketAddress, 0)

	conn, err := net.Dial("udp4", socketAddress)
	if err != nil {
		return addresses, err
	}

	defer conn.Close()

	statusPacket := []byte{0x63, 0x0a, 0x00}
	buffer := make([]byte, 8192)
	bufferLength := 0

	for i := 0; i < retries; i++ {
		conn.SetDeadline(timeInFuture(timeout))

		_, err = conn.Write(statusPacket)
		if err != nil {
			return addresses, err
		}

		conn.SetDeadline(timeInFuture(timeout))
		bufferLength, err = conn.Read(buffer)
		if err != nil {
			continue
		}

		break
	}

	if err != nil {
		return addresses, err
	}

	validHeader := []byte{0xff, 0xff, 0xff, 0xff, 0x64, 0x0a}
	responseHeader := buffer[:len(validHeader)]
	isValidHeader := bytes.Equal(responseHeader, validHeader)

	if !isValidHeader {
		err = errors.New(socketAddress + ": Response error")
		return addresses, err
	}

	reader := bytes.NewReader(buffer[6:bufferLength])

	for {
		var rawAddress RawServerSocketAddress

		err = binary.Read(reader, binary.BigEndian, &rawAddress)
		if err != nil {
			break
		}

		addresses = append(addresses, rawAddress.toSocketAddress())
	}

	return addresses, nil
}

func ReadMasterServers(masterAddresses []SocketAddress, retries int, timeout int) []SocketAddress {
	var (
		wg           sync.WaitGroup
		mutex        sync.Mutex
		allAddresses = make([]SocketAddress, 0)
	)

	for _, masterAddress := range masterAddresses {
		wg.Add(1)

		go func(masterAddress SocketAddress) {
			defer wg.Done()

			addresses, err := ReadMasterServer(masterAddress.toString(), retries, timeout)

			if err != nil {
				log.Println(err)
				return
			}

			mutex.Lock()
			allAddresses = append(allAddresses, addresses...)
			mutex.Unlock()
		}(masterAddress)
	}

	wg.Wait()

	addressMap := make(map[SocketAddress]bool, 0)
	uniqueAddresses := make([]SocketAddress, 0)

	for _, address := range allAddresses {
		if !addressMap[address] {
			uniqueAddresses = append(uniqueAddresses, address)
			addressMap[address] = true
		}
	}

	return uniqueAddresses
}
