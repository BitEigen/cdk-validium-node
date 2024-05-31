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
	batchesData = [][]byte{[]byte(mess)}
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

	daMessage = "8f442a10d8bcc87541d6d4fa4719c0fef2ab6af662c92daf26c937f199b2f5246030344aed6204a92f1b01b0ab793b97ab9664bc433673c6b48bb180eaac676f1c00000000001da272e7277f5ce35b9a2f9c11331979c3f935d7650efb50e12bb836d1c76787c4e998"
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
