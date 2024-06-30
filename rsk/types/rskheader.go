package types

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

type BlockNonce [8]byte

//go:generate gencodec -type Header -field-override headerMarshaling -out gen_header_json.go
//go:generate go run ../rlp/rlpgen -type Header -out gen_header_rlp.go

type UncleCount uint64

// RskHeader represents a block header in the RSK blockchain.
type RskHeader struct {
	ParentHash                common.Hash    `json:"parentHash"                 gencodec:"required"`
	UncleHash                 common.Hash    `json:"sha3Uncles"                 gencodec:"required"`
	Coinbase                  common.Address `json:"miner"`
	Root                      common.Hash    `json:"stateRoot"                  gencodec:"required"`
	TxHash                    common.Hash    `json:"transactionsRoot"           gencodec:"required"`
	ReceiptHash               common.Hash    `json:"receiptsRoot"               gencodec:"required"`
	Bloom                     Bloom          `json:"logsBloom"                  gencodec:"required"`
	Difficulty                *big.Int       `json:"difficulty"                 gencodec:"required"`
	Number                    *big.Int       `json:"number"                     gencodec:"required"`
	GasLimit                  uint64         `json:"gasLimit"                   gencodec:"required"`
	GasUsed                   uint64         `json:"gasUsed"                    gencodec:"required"`
	Time                      uint64         `json:"timestamp"                  gencodec:"required"`
	Extra                     []byte         `json:"extraData"                  gencodec:"required"`
	PaidFees                  *big.Int       `json:"paidFees"                   gencodec:"required"`
	MinimumGasPrice           uint64         `json:"minimumGasPrice"            gencodec:"required"`
	Uncles                    []string       `json:"uncles"                                             rlp:"-"`
	UncleCount                uint64         `json:"uncleCount"`
	UmmRoot                   []byte         `json:"ummRoot"`
	BitcoinMergedMiningHeader []byte         `json:"bitcoinMergedMiningHeader"  gencodec:"required"`
	OriginalHash              common.Hash    `json:"hash"                       gencodec:"required"     rlp:"-"`
}

// field type overrides for gencodec
type headerMarshaling struct {
	Difficulty                *hexutil.Big
	Number                    *hexutil.Big
	GasLimit                  hexutil.Uint64
	GasUsed                   hexutil.Uint64
	Time                      hexutil.Uint64
	Extra                     hexutil.Bytes
	PaidFees                  *hexutil.Big
	MinimumGasPrice           hexutil.Uint64
	BitcoinMergedMiningHeader hexutil.Bytes
	Hash                      common.Hash `json:"hash"` // adds call to Hash() in Marsh
}

// Hash returns the block hash of the header, which is simply the keccak256 hash of its
// RLP encoding.
func (h *RskHeader) Hash() common.Hash {
  h.UncleCount = uint64(len(h.Uncles)) // I don't find a way to custom this in Unmarshaling
	return rlpHash(h)
}

// rlpHash encodes x and hashes the encoded bytes.
func rlpHash(x interface{}) (h common.Hash) {
	sha := hasherPool.Get().(crypto.KeccakState)
	defer hasherPool.Put(sha)
	sha.Reset()
	rlp.Encode(sha, x)
	sha.Read(h[:])
	return h
}

// hasherPool holds LegacyKeccak256 hashers for rlpHash.
var hasherPool = sync.Pool{
	New: func() interface{} { return sha3.NewLegacyKeccak256() },
}
