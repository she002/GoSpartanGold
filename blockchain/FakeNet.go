package main

// Simulate a network by using events to enable simpler testing
import (
	"sync"

	"github.com/chuckpreslar/emission"
)

type NetClient interface {
	GetAddress() string
	GetEmitter() *emission.Emitter
}

type FakeNet struct {
	Clients map[string]NetClient
	mu      sync.Mutex
}

// Registers clients to the network.
// Clients and Miners are registered by public key.
func (f *FakeNet) Register(clientList ...NetClient) {
	(*f).mu.Lock()
	defer (*f).mu.Unlock()
	for _, client := range clientList {
		f.Clients[(client).GetAddress()] = client
	}
}

// Broadcasts to all clients within this.clients.
func (f *FakeNet) Broadcast(msg string, data []byte) {
	(*f).mu.Lock()
	defer (*f).mu.Unlock()
	for address := range f.Clients {
		client := f.Clients[address]
		go (client).GetEmitter().Emit(msg, data)
	}
}

// Tests whether a client is registered with the network.
func (f *FakeNet) Recognizes(client *NetClient) bool {
	(*f).mu.Lock()
	defer (*f).mu.Unlock()
	if _, ok := f.Clients[(*client).GetAddress()]; ok {
		return true
	} else {
		return false
	}
}

/*
func (f *FakeNet) SendMessage(addr string, msg string, o interface{}) {

	jsonByte, err := json.Marshal(o)
	if err != nil {
		fmt.Println("SendMessage() Marshal Panic:")
		panic(err)
	}
	o2 := o
	err = json.Unmarshal(jsonByte, &o2)
	if err != nil {
		fmt.Println("SendMessage() unmarshal Panic:")
		panic(err)
	}

	client := f.Clients[addr]
	(client).GetEmitter().Emit(msg, o2)
}*/

func (f *FakeNet) SendMessage(addr string, msg string, jsonByte []byte) {

	/*
		jsonByte, err := json.Marshal(o)
		if err != nil {
			fmt.Println("SendMessage() Marshal Panic:")
			panic(err)
		}*/
	(*f).mu.Lock()
	defer (*f).mu.Unlock()
	client := f.Clients[addr]
	go (client).GetEmitter().Emit(msg, jsonByte)
}

func NewFakeNet() *FakeNet {
	var f FakeNet
	f.Clients = make(map[string]NetClient)

	return &f
}
