package blockchain

// Simulate a network by using events to enable simpler testing
import (
	"encoding/json"

	"github.com/chuckpreslar/emission"
)

type NetClient interface {
	GetAddress() string
	GetEmitter() *emission.Emitter
}

type FakeNet struct {
	Clients map[string]*NetClient
}

// Registers clients to the network.
// Clients and Miners are registered by public key.
func (f *FakeNet) Register(clientList ...*NetClient) {
	for _, client := range clientList {
		f.Clients[(*client).GetAddress()] = client
	}
}

// Broadcasts to all clients within this.clients.
func (f *FakeNet) Broadcast(msg string, o interface{}) {
	for address, _ := range f.Clients {
		f.SendMessage(address, msg, o)
	}
}

// Tests whether a client is registered with the network.
func (f *FakeNet) Recognizes(client *NetClient) bool {
	if _, ok := f.Clients[(*client).GetAddress()]; ok {
		return true
	} else {
		return false
	}
}

func (f *FakeNet) SendMessage(addr string, msg string, o interface{}) {
	jsonByte, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	o2 := o
	err = json.Unmarshal(jsonByte, &o2)
	if err != nil {
		panic(err)
	}

	client := f.Clients[addr]
	(*client).GetEmitter().Emit(msg, o2)
}

func NewFakeNet() *FakeNet {
	var f FakeNet
	f.Clients = make(map[string]*NetClient)

	return &f
}
