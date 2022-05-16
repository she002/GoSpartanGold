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
	"path/filepath"
	"strconv"
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

func LoadMinerConfig(fileName string) *SaveJsonType {
	dat, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("LoadMinerConfig() Read file fail:", err)
		return nil
	}

	var jsonData SaveJsonType
	err = json.Unmarshal(dat, &jsonData)
	if err != nil {
		fmt.Println("LoadMinerConfig() Unmarshal fail:", err)
		return nil
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
		return nil
	}

	unsignErr := rsa.VerifyPKCS1v15(&jsonData.KeyPair.PublicKey, crypto.SHA256, hashed[:], signature)
	if unsignErr != nil {
		fmt.Println("Error from signing: ", unsignErr)
		return nil
	}

	return &jsonData

}

func LoadStartingBalances(filePath1 string) map[string]uint32 {
	file, err := os.Open(filePath1)
	Balances := make(map[string]uint32)
	if err != nil {
		fmt.Println("LoadStartingBalances() fails to open file", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		res := strings.Split(scanner.Text(), ",")
		if len(res) != 2 {
			return nil
		}

		num, err := strconv.Atoi(res[1])
		if err != nil {
			fmt.Println("LoadStartingBalances() Atoi fails", err)
		}
		Balances[res[0]] = uint32(num)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("LoadStartingBalances() fails to read file", err)
	}
	return Balances
}

func readUserInput(m *TcpMiner) {
	for {
		reader := bufio.NewReader(os.Stdin)
		var menu string = ""
		menu += fmt.Sprintf("Funds: %d\n", m.AvailableGold())
		menu += fmt.Sprintf("Address: %s\n", (*m).Address)
		menu += fmt.Sprintf("Pending transactions: %s\n", m.ShowPendingOut())
		menu += "What would you like to do?\n"
		menu += "*(c)onnect to miner?\n"
		menu += "*(t)ransfer funds?\n"
		menu += "*(r)esend pending transactions?\n"
		menu += "*show (b)alances?\n"
		menu += "*show blocks for (d)ebugging and exit?\n"
		menu += "*(s)ave your state?\n"
		menu += "*e(x)it without saving?\n"

		fmt.Println(menu)
		fmt.Print("Your choice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSuffix(choice, "\n")
		switch choice {
		case "x":
			fmt.Println(`Shutting down.  Have a nice day.`)
			os.Exit(0)
		case "b":
			fmt.Println("  Balances: ")
			m.ShowAllBalances()
		case "c":
			fmt.Print("  port: ")
			port, _ := reader.ReadString('\n')
			port = strings.TrimSuffix(port, "\n")
			m.RegisterWith(port)
			fmt.Printf("Registering with miner at port %s\n", port)
		case "t":
			fmt.Print("  amount: ")
			amt, _ := reader.ReadString('\n')
			amt = strings.TrimSuffix(amt, "\n")
			amtInt, err := strconv.Atoi(amt)
			if err != nil {
				fmt.Println("Wrong input")
			} else if amtInt <= 0 {
				fmt.Println("Wrong input")
			} else {
				amtUint := uint32(amtInt)
				if amtUint > m.AvailableGold() {
					fmt.Printf("***Insufficient gold.  You only have %d\n", m.AvailableGold())
				} else {
					fmt.Print("  address: ")
					addr, _ := reader.ReadString('\n')
					addr = strings.TrimSuffix(addr, "\n")
					var outputs []Output
					var output1 Output
					output1.Address = addr
					output1.Amount = amtUint
					outputs = append(outputs, output1)
					m.PostTransaction(outputs, m.Config.defaultTxFee)
				}
			}
		case "r":
			m.ResendPendingTransactions()
		case "s":
			fmt.Print("  file name: ")
			savePath, _ := reader.ReadString('\n')
			savePath = strings.TrimSuffix(savePath, "\n")
			m.SaveJson(savePath)
		case "d":
			for _, val := range m.Blocks {
				var txStr string = ""
				for _, tx := range val.Transactions {
					txStr += fmt.Sprintf("%s ", tx.Id)
				}
				if txStr != "" {
					fmt.Printf("%s transactions: %s\n", val.GetHashStr(), txStr)
				}
			}
			fmt.Println()
			m.ShowBlockchain()
		default:
			fmt.Printf("Unrecognized choice: %s\n", choice)

		}

		fmt.Print("  Press enter to back to menu: ")
		reader.ReadString('\n')
	}
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
	configfilepath := arguments[2]

	if option == "-c" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please enter your name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSuffix(name, "\n")
		fmt.Print("Please enter your port: ")
		port, _ := reader.ReadString('\n')
		port = strings.TrimSuffix(port, "\n")
		NewMinerSaveJson(configfilepath, name, port)
		fmt.Print("End program.\n")
	} else if option == "-g" {
		minerConfig := LoadMinerConfig(configfilepath)
		if minerConfig == nil {
			fmt.Print("Failed to load config file...End program.\n")
			return
		}
		fmt.Print("Load successful.\n")
		filepath2, err := filepath.Abs("./config/starting_balances.txt")
		if err == nil {
			fmt.Println("Absolute:", filepath2)
		}
		startingBalances := LoadStartingBalances(filepath2)
		//fmt.Printf("starting balances:\n%v\n", *startingBalances)
		genesis, config, _ := MakeGenesis(20, COINBASE_AMT_ALLOWED, DEFAULT_TX_FEE, CONFIRMED_DEPTH, startingBalances)
		net := NewRealNet()
		miner1 := NewTcpMiner(minerConfig.Name, net, NUM_ROUNDS_MINING, genesis, &minerConfig.KeyPair, minerConfig.Connection, config)
		miner1.Initialize(minerConfig.KnownTcpConnections)
		readUserInput(miner1)
		fmt.Print("End program.\n")
	} else {
		fmt.Print("Invalid option\n")
		fmt.Print("End program.\n")
	}
}
