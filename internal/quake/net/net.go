package net

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	OutOfBandHeader  = "\xff\xff\xff\xff"
	GetInfoCommand   = "getinfo"
	GetStatusCommand = "getstatus"
)

func SendCommand(addr, cmd string) ([]byte, error) {
	raddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	buffer := make([]byte, 1024*1024)
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, err
	}
	n, err := conn.WriteTo([]byte(fmt.Sprintf("%s%s", OutOfBandHeader, cmd)), raddr)
	if err != nil {
		return nil, err
	}
	n, _, err = conn.ReadFrom(buffer)
	if err != nil {
		return nil, err
	}
	return buffer[:n], nil
}

func parseMap(data []byte) map[string]string {
	if i := bytes.Index(data, []byte("\n")); i >= 0 {
		data = data[i+1:]
	}
	data = bytes.TrimPrefix(data, []byte("\\"))
	data = bytes.TrimSuffix(data, []byte("\n"))
	parts := bytes.Split(data, []byte("\\"))
	m := make(map[string]string)
	for i := 0; i < len(parts)-1; i += 2 {
		m[string(parts[i])] = string(parts[i+1])
	}
	return m
}

type Player struct {
	Name  string
	Ping  int
	Score int
}

func parsePlayers(data []byte) ([]Player, error) {
	players := make([]Player, 0)
	for _, player := range bytes.Split(data, []byte("\n")) {
		parts := bytes.SplitN(player, []byte(" "), 3)
		if len(parts) != 3 {
			continue
		}
		name, err := strconv.Unquote(string(parts[2]))
		if err != nil {
			return nil, err
		}
		ping, err := strconv.Atoi(string(parts[1]))
		if err != nil {
			return nil, err
		}
		score, err := strconv.Atoi(string(parts[0]))
		if err != nil {
			return nil, err
		}
		players = append(players, Player{
			Name:  name,
			Ping:  ping,
			Score: score,
		})
	}
	return players, nil
}

func GetInfo(addr string) (map[string]string, error) {
	resp, err := SendCommand(addr, GetInfoCommand)
	if err != nil {
		return nil, err
	}
	return parseMap(resp), nil
}

type StatusResponse struct {
	Configuration map[string]string
	Players       []Player
}

func GetStatus(addr string) (*StatusResponse, error) {
	resp, err := SendCommand(addr, GetStatusCommand)
	if err != nil {
		return nil, err
	}
	data := bytes.TrimSuffix(resp, []byte("\n"))
	parts := bytes.SplitN(data, []byte("\n"), 3)
	switch len(parts) {
	case 2:
		status := &StatusResponse{
			Configuration: parseMap(parts[1]),
			Players:       make([]Player, 0),
		}
		return status, nil
	case 3:
		status := &StatusResponse{
			Configuration: parseMap(parts[1]),
		}
		status.Players, _ = parsePlayers(parts[2])
		return status, nil
	default:
		return nil, errors.Errorf("cannot parse response: %q", resp)
	}
}
