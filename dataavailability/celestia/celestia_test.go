package celestia_test

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"testing"

	daTypes "github.com/0xPolygon/cdk-data-availability/types"
	"github.com/0xPolygonHermez/zkevm-node/dataavailability/celestia"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

var daMessage string
var da *celestia.CelestiaBackend
var batchesData [][]byte
var mess string

func init() {
	mess = "testData"
	batchesData = [][]byte{[]byte(mess), []byte("testData2"), []byte("testData3")}
	pk, err := crypto.HexToECDSA("f26d6aca18e0c75ac948a262d4b9435a8173515f84c258d1f90d171143039024")
	if err != nil {
		panic("Cannot load private key")
	}

	cfg := celestia.DAConfig{
		GasPrice:            -1,
		Rpc:                 "ws://192.168.1.49:26658",
		NamespaceId:         "42690c204d39600fddd3",
		AuthToken:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJwdWJsaWMiLCJyZWFkIiwid3JpdGUiLCJhZG1pbiJdfQ.u4v_bgVxWKf-VNMmkjZPHbgHdTr80wBWCAPibdmXMcs",
		SequencerPrivateKey: pk,
	}

	daMessage = "a890635d027675de52aaa6dd0ffec3593ade0c4309139f3c7d61ec4c2c5e146654dbb150f00fe8dbad3e2115c63e8068527de99d254354d28f5212b4a78cf4161c00000000001dc982d397be9a0235e0816afb988f756c0782f77cc96a9c0898a22d2a822a1489137e"
	da, err = celestia.New(cfg)
	if err != nil {
		panic("cannot create celestia da")
	}
	err = da.Init()
	if err != nil {
		panic("cannot init da")
	}
}

func TestPostSequence(t *testing.T) {
	err := da.Init()
	require.NoError(t, err)

	msg, err := da.PostSequence(context.Background(), batchesData)
	require.NoError(t, err)
	fmt.Println("daMessage: ", common.Bytes2Hex(msg))
}

func TestGetSequence(t *testing.T) {
	err := da.Init()
	require.NoError(t, err)

	data, err := da.GetSequence(context.Background(), nil, common.Hex2Bytes(daMessage))
	require.NoError(t, err)
	require.Equal(t, mess, string(data[0]))
}

func TestVerifyMessage(t *testing.T) {
	daMess := common.Hex2Bytes(daMessage)
	sig := daMess[0:crypto.SignatureLength]

	publicKey := da.Cfg.SequencerPrivateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok, "pubkey is not ecdsa.PublicKey")
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Printf("address: %s\n", crypto.PubkeyToAddress(*publicKeyECDSA))

	sequence := daTypes.Sequence{}
	for _, batchData := range batchesData {
		sequence = append(sequence, batchData)
	}
	hash := sequence.HashToSign()
	fmt.Printf("signedHash: %s\n", hex.EncodeToString(hash))
	fmt.Printf("sig: %s\n", hex.EncodeToString(sig))

	// remove recover id
	sig = sig[:len(sig)-1]
	ok = crypto.VerifySignature(publicKeyBytes, hash, sig)
	require.True(t, ok, "Verify message failed")
}
