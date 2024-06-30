package types_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	rskTypes "github.com/0xPolygonHermez/zkevm-node/rsk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

var cl, err = ethclient.Dial("https://public-node.testnet.rsk.co")
var startBlock = 4472997

func init() {
	if err != nil {
		panic(err)
	}
}

func TestMultipleHash(t *testing.T) {
	numOfTest := 100
	for i := startBlock; i < startBlock+numOfTest; i++ { // start of the execution block
		testHash(t, big.NewInt(int64(i)))
		time.Sleep(1 * time.Second)
	}
}

func TestOneHash(t *testing.T) {
	testHash(t, big.NewInt(int64(startBlock)))
}

func testHash(t *testing.T, number *big.Int) {
	fmt.Printf("Checking block %d\n", number)
	var head *rskTypes.RskHeader
	err := cl.Client().CallContext(context.Background(), &head, "eth_getBlockByNumber", toBlockNumArg(number), false)
	require.NoError(t, err)
	require.Equal(t, head.Hash().Hex(), head.OriginalHash.Hex())
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	if number.Sign() >= 0 {
		return hexutil.EncodeBig(number)
	}
	// It's negative.
	if number.IsInt64() {
		return rpc.BlockNumber(number.Int64()).String()
	}
	// It's negative and large, which is invalid.
	return fmt.Sprintf("<invalid %d>", number)
}
