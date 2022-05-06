package blockchain

// Simulate a network by using events to enable simpler testing
import (
	"encoding/json"
)

type FakeNet struct {
	Clients map[string]*Client
	o2      interface{}
}

// Registers clients to the network.
// Clients and Miners are registered by public key.
func (f *FakeNet) register(clientList ...*Client) {
	for _, client := range clientList {
		f.Clients[client.Address] = client
	}
}

// Broadcasts to all clients within this.clients.
func (f *FakeNet) broadcast(msg string, o interface{}) {
	for address, _ := range f.Clients {
		f.sendMessage(address, msg, o)
	}
}

// Tests whether a client is registered with the network.
func (f *FakeNet) recognizes(client Client) bool {
	if _, ok := f.Clients[client.Address]; ok {
		return true
	} else {
		return false
	}
}

func (f *FakeNet) sendMessage(addr string, msg string, o interface{}) {
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
	client.Emitter.Emit(msg, o2)
}

func NewFakeNet() *FakeNet {
	var f FakeNet
	f.Clients = make(map[string]*Client)

	return &f
}
