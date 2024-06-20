package dataavailability

import (
	"github.com/0xPolygonHermez/zkevm-node/dataavailability/avail"
	"github.com/0xPolygonHermez/zkevm-node/dataavailability/celestia"
)

// DABackendType is the data availability protocol for the CDK
type DABackendType string

const (
	// DataAvailabilityCommittee is the DAC protocol backend
	DataAvailabilityCommittee DABackendType = "DataAvailabilityCommittee"

	// Celestia protocol
	Celestia DABackendType = "Celestia"

	// Avail protocol
	Avail DABackendType = "Avail"
)

// Config for dataavailability
type Config struct {
	// config for DA protocol Celestia
	Celestia celestia.Config `mapstructure:"Celestia"`
  // config for DA protocol Avail
	Avail    avail.Config    `mapstructure:"Avail"`
}
