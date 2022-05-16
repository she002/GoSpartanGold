package main

// Simulate a network by using events to enable simpler testing
import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type RealNet struct {
	Clients map[string]TcpConnectionInfo
	mu      sync.Mutex
}

// Registers clients to the network.
// Clients and Miners are registered by public key.
func (f *RealNet) Register(clientInfo TcpConnectionInfo) {
	(*f).mu.Lock()
	defer (*f).mu.Unlock()
	(*f).Clients[clientInfo.Address] = clientInfo
}

// Broadcasts to all clients within this.clients.
func (f *RealNet) Broadcast(msg string, data []byte) {
	(*f).mu.Lock()
	defer (*f).mu.Unlock()
	for _, client := range f.Clients {

		var conn TcpData
		conn.Msg = msg
		conn.Data = make([]byte, len(data))
		copy(conn.Data, data)

		connBytes, err := json.Marshal(conn)
		if err != nil {
			fmt.Println("Broadcast() TcpData unmarshal Panic: ", err)
			return
		}

		c, err := net.Dial("tcp", client.Connection)
		if err != nil {
			fmt.Println(err)
			return
		}

		c.Write(connBytes)
		c.Close()
	}
}

// Tests whether a client is registered with the network.
func (f *RealNet) Recognizes(client *TcpConnectionInfo) bool {
	(*f).mu.Lock()
	defer (*f).mu.Unlock()
	if _, ok := f.Clients[client.Address]; ok {
		return true
	} else {
		return false
	}
}

func (f *RealNet) SendMessage(addr string, msg string, jsonByte []byte) {

	(*f).mu.Lock()
	defer (*f).mu.Unlock()

	client, ok := f.Clients[addr]
	if !ok {
		fmt.Printf("SendMessage(): address[%s] not found\n", addr)
		return
	}

	var conn TcpData
	conn.Msg = msg
	conn.Data = make([]byte, len(jsonByte))
	copy(conn.Data, jsonByte)

	connBytes, err := json.Marshal(conn)
	if err != nil {
		fmt.Println("SendMessage() TcpData unmarshal Panic: ", err)
		return
	}

	c, err := net.Dial("tcp", client.Connection)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.Write(connBytes)
	c.Close()
}

func NewRealNet() *RealNet {
	var f RealNet
	f.Clients = make(map[string]TcpConnectionInfo)

	return &f
}
