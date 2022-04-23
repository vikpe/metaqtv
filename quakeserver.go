package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

func ReadServerQtv(serverAddress SocketAddress, retries int, timeout int) (QtvServer, error) {
	conn, err := net.Dial("udp4", serverAddress.toString())
	if err != nil {
		return QtvServer{}, err
	}
	defer conn.Close()

	qtvStatusSequence := []byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', ' ', '3', '2', 0x0a}
	buffer := make([]byte, bufferMaxSize)
	bufferLength := 0

	for i := 0; i < retries; i++ {
		conn.SetDeadline(timeInFuture(timeout))

		_, err = conn.Write(qtvStatusSequence)
		if err != nil {
			return QtvServer{}, err
		}

		conn.SetDeadline(timeInFuture(timeout))
		bufferLength, err = conn.Read(buffer)
		if err != nil {
			continue
		}

		break
	}

	if err != nil {
		// no logging here. it seems that servers may not reply if they do not support
		// this specific "32" status request.
		return QtvServer{}, err
	}

	validResponseSequence := []byte{0xff, 0xff, 0xff, 0xff, 'n', 'q', 't', 'v'}
	responseSequence := buffer[:len(validResponseSequence)]
	isValidResponse := bytes.Equal(responseSequence, validResponseSequence)
	if !isValidResponse {
		// some servers react to the specific "32" status message but will send the regular
		// status message because they misunderstood our command.
		return QtvServer{}, err
	}

	reader := csv.NewReader(strings.NewReader(string(buffer[5:bufferLength])))
	reader.Comma = ' '

	qtvRecord, err := reader.Read()
	if err != nil {
		return QtvServer{}, err
	}

	IndexTitle := 2
	IndexAddress := 3

	if qtvRecord[IndexAddress] == "" {
		// these are the servers that are not configured correctly,
		// that means they are not reporting their qtv ip as they should.
		return QtvServer{}, err
	}

	return QtvServer{
		Title:      qtvRecord[IndexTitle],
		Address:    qtvRecord[IndexAddress],
		Spectators: make([]string, 0),
	}, nil
}

func ReadServers(serverAddresses []SocketAddress, retries int, timeout int) []QuakeServer {
	quakeServers := make([]QuakeServer, 0)
	processedAddresses := make(map[SocketAddress]bool, 0)

	var (
		wg    sync.WaitGroup
		mutex sync.Mutex
	)

	for _, address := range serverAddresses {
		if _, ok := processedAddresses[address]; ok {
			continue
		}

		wg.Add(1)

		go func(address SocketAddress) {
			defer wg.Done()

			qserver, err := ReadServer(address, retries, timeout)

			if err != nil {
				return
			}

			mutex.Lock()
			processedAddresses[address] = true
			quakeServers = append(quakeServers, qserver)
			mutex.Unlock()
		}(address)
	}

	wg.Wait()

	return quakeServers
}

func ReadServer(serverAddress SocketAddress, retries int, timeout int) (QuakeServer, error) {
	conn, err := net.Dial("udp4", serverAddress.toString())
	if err != nil {
		return QuakeServer{}, err
	}
	defer conn.Close()

	statusSequence := []byte{0xff, 0xff, 0xff, 0xff, 's', 't', 'a', 't', 'u', 's', ' ', '2', '3', 0x0a}
	buffer := make([]byte, bufferMaxSize)
	bufferLength := 0

	for i := 0; i < retries; i++ {
		conn.SetDeadline(timeInFuture(timeout))

		_, err = conn.Write(statusSequence)
		if err != nil {
			return QuakeServer{}, err
		}

		conn.SetDeadline(timeInFuture(timeout))
		bufferLength, err = conn.Read(buffer)
		if err != nil {
			continue
		}

		break
	}

	if err != nil {
		return QuakeServer{}, err
	}

	validResponseSequence := []byte{0xff, 0xff, 0xff, 0xff, 'n', '\\'}
	responseSequence := buffer[:len(validResponseSequence)]
	isValidResponse := bytes.Equal(responseSequence, validResponseSequence)
	if !isValidResponse {
		log.Println(serverAddress.toString() + ": Response error")
		return QuakeServer{}, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(buffer[len(validResponseSequence):bufferLength])))
	scanner.Scan()

	settings := strings.FieldsFunc(scanner.Text(), func(r rune) bool {
		if r == '\\' {
			return true
		}
		return false
	})

	qserver := newQuakeServer()
	qserver.Address = serverAddress.toString()

	for i := 0; i < len(settings)-1; i += 2 {
		qserver.Settings[settings[i]] = settings[i+1]
	}

	if val, ok := qserver.Settings["hostname"]; ok {
		qserver.Settings["hostname"] = quakeTextToPlainText(val)
		qserver.Title = qserver.Settings["hostname"]
	}
	if val, ok := qserver.Settings["map"]; ok {
		qserver.Map = val
	}
	if val, ok := qserver.Settings["maxclients"]; ok {
		value, _ := strconv.Atoi(val)
		qserver.MaxPlayers = value
	}
	if val, ok := qserver.Settings["maxspectators"]; ok {
		value, _ := strconv.Atoi(val)
		qserver.MaxSpectators = value
	}

	for scanner.Scan() {
		reader := csv.NewReader(strings.NewReader(scanner.Text()))
		reader.Comma = ' '

		clientRecord, err := reader.Read()
		if err != nil {
			continue
		}

		client, err := parseClientRecord(clientRecord)
		if err != nil {
			continue
		}

		if client.IsSpec {
			qserver.Spectators = append(qserver.Spectators, Spectator{
				Name:    client.Name,
				NameInt: client.NameInt,
				IsBot:   client.IsBot,
			})
		} else {
			qserver.Players = append(qserver.Players, client.Player)
		}
	}

	qserver.NumPlayers = len(qserver.Players)
	qserver.NumSpectators = len(qserver.Spectators)

	qtvServer, _ := ReadServerQtv(serverAddress, retries, timeout)
	qserver.QtvAddress = qtvServer.Address

	return qserver, nil
}

func isBotName(name string) bool {
	switch name {
	case
		"[ServeMe]",
		"twitch.tv/vikpe":
		return true
	}
	return false
}

func isBotPing(ping int) bool {
	switch ping {
	case
		10:
		return true
	}
	return false
}

func parseClientRecord(clientRecord []string) (Client, error) {
	expectedColumnCount := 9
	columnCount := len(clientRecord)

	if columnCount != expectedColumnCount {
		err := errors.New(fmt.Sprintf("invalid player column count %d.", columnCount))
		return Client{}, err
	}

	const (
		IndexFrags              = 1
		IndexTime               = 2
		IndexPing               = 3
		IndexName               = 4
		IndexSkin               = 5
		IndexColorTop           = 6
		IndexColorBottom        = 7
		IndexTeam               = 8
		SpectatorPrefix  string = "\\s\\"
	)

	nameRawStr := clientRecord[IndexName]

	isSpec := strings.HasPrefix(nameRawStr, SpectatorPrefix)
	if isSpec {
		nameRawStr = strings.TrimPrefix(nameRawStr, SpectatorPrefix)
	}

	name := quakeTextToPlainText(nameRawStr)
	nameInt := stringToIntArray(name)
	team := quakeTextToPlainText(clientRecord[IndexTeam])
	teamInt := stringToIntArray(team)
	colorTop := fieldAsInt(clientRecord[IndexColorTop])
	colorBottom := fieldAsInt(clientRecord[IndexColorBottom])
	ping := fieldAsInt(clientRecord[IndexPing])

	return Client{
		Player: Player{
			Name:    name,
			NameInt: nameInt,
			Team:    team,
			TeamInt: teamInt,
			Skin:    clientRecord[IndexSkin],
			Colors:  [2]int{colorTop, colorBottom},
			Frags:   fieldAsInt(clientRecord[IndexFrags]),
			Ping:    ping,
			Time:    fieldAsInt(clientRecord[IndexTime]),
			IsBot:   isBotName(name) || isBotPing(ping),
		},
		IsSpec: isSpec,
	}, nil

}

func fieldAsInt(value string) int {
	valueAsInt, _ := strconv.Atoi(value)
	return valueAsInt
}
