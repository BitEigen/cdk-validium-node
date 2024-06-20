package avail

// Config for avail-node
type Config struct {
	Seed         string `mapstructure:"Seed"`
	WsApiUrl     string `mapstructure:"WsApiUrl"`
	HttpApiUrl   string `mapstructure:"HttpApiUrl"`
	BridgeApiUrl string `mapstructure:"BridageApiUrl"`
	AppID        int    `mapstructure:"AppId"`
	Timeout      int    `mapstructure:"Timeout"`
}
