package celestia

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	daTypes "github.com/0xPolygon/cdk-data-availability/types"
	"github.com/0xPolygonHermez/zkevm-node/log"
	openrpc "github.com/celestiaorg/celestia-openrpc"
	"github.com/celestiaorg/celestia-openrpc/types/blob"
	"github.com/celestiaorg/celestia-openrpc/types/share"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const blobSize = 40

// DAConfig is config for Celestia DA
type DAConfig struct {
	GasPrice            float64
	Rpc                 string
	NamespaceId         string
	AuthToken           string
	SequencerPrivateKey *ecdsa.PrivateKey
}

// CelestiaBackend implements the Celestia integration
type CelestiaBackend struct {
	Cfg    DAConfig
	Client *openrpc.Client
	// Trpc        *http.HTTP
	Namespace share.Namespace
	// BlobstreamX *blobstreamx.BlobstreamX
}

// New creates new CelestiaBackend
func New(cfg DAConfig) (*CelestiaBackend, error) {
	if cfg.NamespaceId == "" {
		return nil, errors.New("namespace id cannot be blank")
	}
	nsBytes, err := hex.DecodeString(cfg.NamespaceId)
	if err != nil {
		return nil, err
	}

	namespace, err := share.NewBlobNamespaceV0(nsBytes)
	if err != nil {
		return nil, err
	}

	return &CelestiaBackend{
		Cfg:       cfg,
		Client:    nil,
		Namespace: namespace,
	}, nil
}

// Init inits Celestia Backend from config
func (c *CelestiaBackend) Init() error {
	daClient, err := openrpc.NewClient(context.Background(), c.Cfg.Rpc, c.Cfg.AuthToken)
	if err != nil {
		return err
	}
	c.Client = daClient
	return nil
}

// post batchdata to celestia and return BlobPointer in []byte
func (c *CelestiaBackend) postBatchData(ctx context.Context, batchData []byte) ([]byte, error) {
	dataBlob, err := blob.NewBlobV0(c.Namespace, batchData)
	if err != nil {
		log.Warn("Error creating blob", "err", err)
		return nil, err
	}

	commitment, err := blob.CreateCommitment(dataBlob)
	if err != nil {
		log.Warn("Error creating commitment", "err", err)
		return nil, err
	}

	height, err := c.Client.Blob.Submit(ctx, []*blob.Blob{dataBlob}, openrpc.GasPrice(c.Cfg.GasPrice))
	if err != nil {
		log.Warn("Blob Submission error", "err", err)
		return nil, err
	}

	if height == 0 {
		log.Warn("Unexpected height from blob response", "height", height)
		return nil, errors.New("unexpected response code")
	}

	proofs, err := c.Client.Blob.GetProof(ctx, height, c.Namespace, commitment)
	if err != nil {
		log.Warn("Error retrieving proof", "err", err)
		return nil, err
	}

	included, err := c.Client.Blob.Included(ctx, height, c.Namespace, proofs, commitment)
	if err != nil || !included {
		log.Warn("Error checking for inclusion", "err", err, "proof", proofs)
		return nil, err
	}
	log.Info("Successfully posted blob height: ", height, " commitment: ", hex.EncodeToString(commitment))

	txCommitment := [32]byte{}
	copy(txCommitment[:], commitment)

	blobPointer := BlobPointer{
		BlockHeight:  height,
		TxCommitment: txCommitment,
	}

	blobPointerData, err := blobPointer.MarshalBinary()

	if err != nil {
		log.Warn("BlobPointer MashalBinary error", "err", err)
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, blobPointerData)
	if err != nil {
		log.Warn("blob pointer data serialization failed", "err", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

// PostSequence posts batches data to Celestia and returns DA message
func (c *CelestiaBackend) PostSequence(ctx context.Context, batchesData [][]byte) ([]byte, error) {
	var sequence daTypes.Sequence
	for _, batchData := range batchesData {
		sequence = append(sequence, batchData)
	}

	// submit sequence to celestia
	// encode sequence to []byte
	blobData, err := c.encodeSequence(sequence)
	if err != nil {
		return nil, err
	}
	// celestia does not support key value data
	// so we need blobPointer to query blob data
	blobPointer, err := c.postBatchData(ctx, blobData)
	if err != nil {
		log.Error("Cannot post batchdata to celestia")
		return nil, err
	}

	signedSequence, err := sequence.Sign(c.Cfg.SequencerPrivateKey)
	if err != nil {
		log.Error("Cannot sign sequence")
		return nil, err
	}

	daMessage := append(signedSequence.Signature, blobPointer...)
	return daMessage, nil
}

// GetSequence gets batches from celestia
func (c *CelestiaBackend) GetSequence(ctx context.Context, batchHashes []common.Hash, dataAvailabilityMessage []byte) ([][]byte, error) {
	msgLen := len(dataAvailabilityMessage)

	if msgLen != blobSize+crypto.SignatureLength {
		return nil, fmt.Errorf("wrong da message length: %d", msgLen)
	}

	start := crypto.SignatureLength
	blobMessage := dataAvailabilityMessage[start:msgLen]

	blobPointer := BlobPointer{}
	err := blobPointer.UnmarshalBinary(blobMessage[:])
	if err != nil {
		log.Errorf("Cannot unmarshal BlobMessage %s\n", hex.EncodeToString(blobMessage))
		return nil, err
	}
	blob, err := c.Client.Blob.Get(ctx, blobPointer.BlockHeight, c.Namespace, blobPointer.TxCommitment[:])
	if err != nil {
		return nil, err
	}
	res, err := c.decodeBlobData(blob.Data)
	if err != nil {
		log.Errorf("Cannot decode blob data\n")
		return nil, err
	}
	return res, nil
}

func (c *CelestiaBackend) encodeSequence(sequence daTypes.Sequence) ([]byte, error) {
	sequenceOfText := []string{}

	for _, batchData := range ([]daTypes.ArgBytes)(sequence) {
		batchText, err := batchData.MarshalText()
		if err != nil {
			log.Errorf("Cannot marshal batchData %s\n", hex.EncodeToString(batchData))
			return nil, err
		}
		sequenceOfText = append(sequenceOfText, string(batchText))
	}

	res, err := json.Marshal(sequenceOfText)
	if err != nil {
		log.Errorf("Cannot marshal sequence\n")
		return nil, err
	}
	return res, nil
}

func (c *CelestiaBackend) decodeBlobData(blobData []byte) ([][]byte, error) {
	encodedBatchData := []daTypes.ArgBytes{}
	err := json.Unmarshal(blobData, &encodedBatchData)
	if err != nil {
		log.Errorf("Cannot unmarshal blobData: %v", encodedBatchData)
		return nil, err
	}
	res := [][]byte{}
	for _, batchData := range encodedBatchData {
		err := batchData.UnmarshalText(batchData)
		if err != nil {
			log.Errorf("Cannot unmarshal batchData %s\n", hex.EncodeToString(batchData))
			return nil, err
		}
		res = append(res, batchData)
	}
	return res, nil
}
