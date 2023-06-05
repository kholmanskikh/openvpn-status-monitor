package openvpn_status

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Line struct {
	CommonName         string
	RealAddress        string
	VirtualAddress     string
	VirtualIPv6Address string
	BytesReceived      uint64
	BytesSent          uint64
	ConnectedSince     time.Time
	Username           string
	ClientId           uint32
	PeerId             uint32
	DataChannelCipher  string
}

type Status struct {
	Lines []Line
	Time  time.Time
}

const ClientCommonName = "Common Name"
const ClientRealAddress = "Real Address"
const ClientVirtualAddress = "Virtual Address"
const ClientVirtualIpv6Address = "Virtual IPv6 Address"
const ClientBytesReceived = "Bytes Received"
const ClientBytesSent = "Bytes Sent"
const ClientConnectedSince = "Connected Since (time_t)"
const ClientUsername = "Username"
const ClientId = "Client ID"
const ClientPeerId = "Peer ID"
const ClientDataChannelCipher = "Data Channel Cipher"

func ReadFromFile(path string) (*Status, error) {
	status := Status{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	timeLineRegexp := regexp.MustCompile(`^TIME,[^,]*,(\d+)`)
	headerClientListRegexp := regexp.MustCompile(`^HEADER,CLIENT_LIST,(.*)`)
	clientListRegexp := regexp.MustCompile(`^CLIENT_LIST,(.*)`)

	clientListLines := make([]string, 0)
	nameToPos := make(map[string]int)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := timeLineRegexp.FindStringSubmatch(line); matches != nil {
			val, err := strconv.ParseInt(matches[1], 10, 64)
			if err != nil {
				return nil, err
			}
			status.Time = time.Unix(val, 0)
		} else if matches := headerClientListRegexp.FindStringSubmatch(line); matches != nil {
			names := strings.Split(matches[1], ",")
			if len(names) == 0 {
				return nil, errors.New("Invalid format of the HEADER,CLIENT_LIST line")
			}

			for idx, name := range names {
				nameToPos[name] = idx
			}

			for _, name := range []string{ClientCommonName, ClientRealAddress, ClientVirtualAddress,
				ClientVirtualIpv6Address, ClientBytesReceived, ClientBytesSent,
				ClientConnectedSince, ClientUsername, ClientId, ClientPeerId,
				ClientDataChannelCipher} {
				if _, ok := nameToPos[name]; !ok {
					return nil, errors.New(fmt.Sprintf("Name '%s' is not present in line: %s ",
						name, line))
				}
			}
		} else if matches := clientListRegexp.FindStringSubmatch(line); matches != nil {
			clientListLines = append(clientListLines, matches[1])
		}
	}

	for _, clientLine := range clientListLines {
		tokens := strings.Split(clientLine, ",")
		if len(tokens) != len(nameToPos) {
			return nil, errors.New("Mismatch in field CLIENT_LIST names and values")
		}

		statusLine := Line{
			CommonName:         tokens[nameToPos[ClientCommonName]],
			RealAddress:        tokens[nameToPos[ClientRealAddress]],
			VirtualAddress:     tokens[nameToPos[ClientVirtualAddress]],
			VirtualIPv6Address: tokens[nameToPos[ClientVirtualIpv6Address]],
			Username:           tokens[nameToPos[ClientUsername]],
			DataChannelCipher:  tokens[nameToPos[ClientDataChannelCipher]],
		}

		statusLine.BytesReceived, err = strconv.ParseUint(tokens[nameToPos[ClientBytesReceived]], 10, 64)
		if err != nil {
			return nil, err
		}

		statusLine.BytesSent, err = strconv.ParseUint(tokens[nameToPos[ClientBytesSent]], 10, 64)
		if err != nil {
			return nil, err
		}

		clientId, err := strconv.ParseUint(tokens[nameToPos[ClientId]], 10, 32)
		if err != nil {
			return nil, err
		}
		statusLine.ClientId = uint32(clientId)

		peerId, err := strconv.ParseUint(tokens[nameToPos[ClientPeerId]], 10, 32)
		if err != nil {
			return nil, err
		}
		statusLine.PeerId = uint32(peerId)

		connectedSince, err := strconv.ParseInt(tokens[nameToPos[ClientConnectedSince]], 10, 64)
		if err != nil {
			return nil, err
		}
		statusLine.ConnectedSince = time.Unix(connectedSince, 0)

		status.Lines = append(status.Lines, statusLine)
	}

	if status.Time.IsZero() {
		return nil, errors.New("TIME is not set")
	}

	return &status, nil
}
