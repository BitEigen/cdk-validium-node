package avail_test

import (
	"testing"

	"github.com/0xPolygonHermez/zkevm-node/dataavailability/avail"
	"github.com/ethereum/go-ethereum/common"
)

var daMessage string
var da *avail.AvailBackend
var batchesData [][]byte
var mess string

func init() {
	mess = "testData"
	batchesData = [][]byte{[]byte(mess), []byte("testData2"), []byte("testData3")}

	cfg := avail.DAConfig{
		Seed:         "nothing play clerk horse attack kick jelly joy rug banner magic position",
		WsApiUrl:     "ws://192.168.1.49:7000",
		// WsApiUrl:     "wss://avail-turing.public.blastapi.io",
		HttpApiUrl:   "http://192.168.1.49:7000",
		BridgeApiUrl: "https://bridgeUrl",
		AppID:        0,
		Timeout:      10000,
	}

	da, err := avail.New("https://public-node.testnet.rsk.co", common.HexToAddress("0x123"), cfg)
	if err != nil {
		panic("cannot create avail da")
	}
	err = da.Init()
	if err != nil {
		panic("cannot init da")
	}
}

func TestPostSequence(t *testing.T) {

}

func TestGetSequence(t *testing.T) {

}

func TestVerifyMessage(t *testing.T) {

}
