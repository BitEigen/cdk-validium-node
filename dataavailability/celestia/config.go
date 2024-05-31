package celestia

import "github.com/0xPolygonHermez/zkevm-node/config/types"

// Config for celestia-node
type Config struct {
	GasPrice            float64                  `mapstructure:"GasPrice"`
	Rpc                 string                   `mapstructure:"Rpc"`
	NamespaceId         string                   `mapstructure:"NamespaceId"`
	AuthToken           string                   `mapstructure:"AuthToken"`
	SequencerPrivateKey types.KeystoreFileConfig `mapstructure:"SequencerPrivateKey"`
}
