package main

import (
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func NewMinerSaveJson(fileName string, name string, port string) {

	privKey, _, _ := GenerateKeypair()

	var jsonData SaveJsonType
	jsonData.Name = name
	jsonData.KeyPair = *privKey
	jsonData.Connection = port
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Println("SaveJson() Marshal fail:", err)
		return
	}

	err = os.WriteFile(fileName, jsonBytes, 0755)
	if err != nil {
		fmt.Println("SaveJson() Write file fail:", err)
		return
	}
}

func LoadMinerConfig(fileName string) bool {
	dat, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("LoadMinerConfig() Read file fail:", err)
		return false
	}

	var jsonData SaveJsonType
	err = json.Unmarshal(dat, &jsonData)
	if err != nil {
		fmt.Println("LoadMinerConfig() Unmarshal fail:", err)
		return false
	}

	address := GenerateAddress(&jsonData.KeyPair.PublicKey)

	fmt.Printf("Name: %s, Port: %s Address: %s\n", jsonData.Name, jsonData.Connection, address)
	fmt.Printf("Known Miners: %v\n", jsonData.KnownTcpConnections)
	rng := rand.Reader
	bytes := []byte(jsonData.Name)
	hashed := sha256.Sum256(bytes)
	signature, signErr := rsa.SignPKCS1v15(rng, &jsonData.KeyPair, crypto.SHA256, hashed[:])
	if signErr != nil {
		fmt.Println("Error from signing: ", signErr)
		return false
	}

	unsignErr := rsa.VerifyPKCS1v15(&jsonData.KeyPair.PublicKey, crypto.SHA256, hashed[:], signature)
	if unsignErr != nil {
		fmt.Println("Error from signing: ", unsignErr)
		return false
	}

	return true

}

func main() {
	arguments := os.Args
	if len(arguments) != 3 {
		fmt.Println("Instruction: ./app [-option] <filepath>")
		fmt.Println("option:")
		fmt.Println("    -c : create a new miner account. <filepath> should be the filepath to save miner config")
		fmt.Println("    -g : load miner config file. <filepath> should be the filepath to load miner config file")
		return
	}

	option := arguments[1]
	filepath := arguments[2]

	if option == "-c" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please enter your name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSuffix(name, "\n")
		fmt.Print("Please enter your port: ")
		port, _ := reader.ReadString('\n')
		port = strings.TrimSuffix(port, "\n")
		NewMinerSaveJson(filepath, name, port)
		fmt.Print("End program.\n")
	} else if option == "-g" {
		if LoadMinerConfig(filepath) {
			fmt.Print("Load successful.\n")
		} else {
			fmt.Print("Load failed.\n")
		}
		fmt.Print("End program.\n")
	} else {
		fmt.Print("Invalid option\n")
		fmt.Print("End program.\n")
	}
}
