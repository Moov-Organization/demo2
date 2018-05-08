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
)

type EthAPI struct {
	conn *ethclient.Client
	mrm *MoovRideManager
	auth *bind.TransactOpts
}

type EthApiInterface interface {
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
	return &ethApi
}

func (ethApi *EthAPI) GetRideAddressIfAvailable() (available bool, address string) {
	addresses, err := ethApi.mrm.GetAvailableRides(nil)
	if err != nil {
		log.Println("could not get addresses from car: %v", err)
	}
	if len(addresses) > 0 {
		available = true
		address = addresses[0].String()
	} else {
		available = false
	}
	return
}

func (ethApi *EthAPI) AcceptRequest(address string) (status bool) {
	transaction, err := ethApi.mrm.AcceptRideRequest(&bind.TransactOpts{
		From:     ethApi.auth.From,
		Signer:   ethApi.auth.Signer,
		GasLimit: 2381623,
	}, common.HexToAddress(address))
	if err != nil {
		log.Println("Could not accept ride request from car: %v", err)
	}
	fmt.Println("Transaction initiated")
	receipt, err := bind.WaitMined(context.Background(), ethApi.conn, transaction)
	if err != nil {
		log.Fatalf("Wait for mining error %s %v: ", err)
	} else if receipt.Status == types.ReceiptStatusFailed {
		status = true
	} else {
		status = false
	}
	return
}

func (ethApi *EthAPI) GetLocations(address string) (from string, to string) {
	ride, err := ethApi.mrm.Rides(nil, common.HexToAddress(address))
	if err != nil {
		log.Println("get locations error: %v", err)
	}
	from = ride.From
	to = ride.To
	return
}
