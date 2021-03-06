package sim2

import (
	"log"
	"fmt"
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/rand"
	"time"
)

type EthAPI struct {
	conn *ethclient.Client
	mrm *MoovRideManager
	auth *bind.TransactOpts
	newRideEvent chan *MoovRideManagerNewRideRequest
	lastNewRequestIndex uint
	lastCheck time.Time
}

type BlockchainInterface interface {
	GetRideAddressIfAvailable() (available bool, address string)
	AcceptRequest(address string) (status bool)
	GetLocations(address string) (from string, to string)
}


func NewEthApi(mrmAddress string, privateKeyString string) (*EthAPI)  {
	var err error
	var ethApi EthAPI
	//c.ethApi.conn, err = ethclient.Dial("http://127.0.0.1:7545")
	ethApi.conn, err = ethclient.Dial("ws://127.0.0.1:8546")
	if err != nil {
		log.Fatalf("could not create ipc client: %v", err)
	}
	ethApi.mrm, err = NewMoovRideManager(common.HexToAddress(mrmAddress), ethApi.conn)
	if err != nil {
		log.Fatalf("could not connect to mrm: %v", err)
	}

	privateKey, err := crypto.HexToECDSA(privateKeyString)
	if err != nil {
		log.Fatalf("could not convert private key to hex: %v", err)
	}
	ethApi.auth = bind.NewKeyedTransactor(privateKey)
	ethApi.newRideEvent = make(chan *MoovRideManagerNewRideRequest)
	_, err = ethApi.mrm.WatchNewRideRequest(nil, ethApi.newRideEvent);
	if err != nil {
		log.Fatalf("could not watch for New Ride event: %v", err)
	}
	ethApi.lastCheck = time.Now()
	return &ethApi
}

func (ethApi *EthAPI) GetRideAddressIfAvailable() (available bool, address string) {
	ethApi.checkConnection()
	select {
	case msg := <-ethApi.newRideEvent:
		if msg.Raw.Index > ethApi.lastNewRequestIndex {
			msg.Raw.Index = ethApi.lastNewRequestIndex
			addresses, err := ethApi.mrm.GetAvailableRides(nil)
			if err != nil {
				log.Println("could not get addresses from car: ", err)
			}
			if len(addresses) > 0 {
				//ethApi.lastNewRequestIndex = msg.Raw.Index
				s1 := rand.NewSource(time.Now().UnixNano())
				r1 := rand.New(s1)
				address = addresses[uint(r1.Int()%len(addresses))].String()
				available = true
			} else {
				available = false
			}
		} else {
			available = false
		}
	default:
		if time.Now().Sub(ethApi.lastCheck) > time.Second * 3 {
			addresses, err := ethApi.mrm.GetAvailableRides(nil)
			if err != nil {
				log.Println("could not get addresses from car: ", err)
			}
			if len(addresses) > 0 {
				//ethApi.lastNewRequestIndex = msg.Raw.Index
				s1 := rand.NewSource(time.Now().UnixNano())
				r1 := rand.New(s1)
				address = addresses[uint(r1.Int()%len(addresses))].String()
				available = true
			} else {
				available = false
			}
			ethApi.lastCheck = time.Now()
		}
	}

	return
}

func (ethApi *EthAPI) AcceptRequest(address string) (status bool) {
	ethApi.checkConnection()
	transaction, err := ethApi.mrm.AcceptRideRequest(&bind.TransactOpts{
		From:     ethApi.auth.From,
		Signer:   ethApi.auth.Signer,
		GasLimit: 2381623,
	}, common.HexToAddress(address))
	if err != nil {
		log.Println("Could not accept ride request from car: ", err)
		status = false
	} else {
		fmt.Println("Transaction initiated")
		receipt, err := bind.WaitMined(context.Background(), ethApi.conn, transaction)
		if err != nil {
			log.Fatalf("Wait for mining error: %v \n", err)
		} else if receipt.Status == types.ReceiptStatusFailed {
			status = false
		} else {
			status = true
		}
	}
	return
}

func (ethApi *EthAPI) GetLocations(address string) (from string, to string) {
	ethApi.checkConnection()
	ride, err := ethApi.mrm.Rides(nil, common.HexToAddress(address))
	if err != nil {
		log.Println("get locations error: ", err)
	} else {
		from = ride.From
		to = ride.To
	}
	return
}
func (ethApi *EthAPI) checkConnection() {
	_, err := ethApi.conn.NetworkID(context.TODO())
	if err!= nil {
		log.Fatalln("Geth Connection Problem")
	}
}

