package protocol

type GatewayProtocol struct {
	PackLen      uint32 `json:"pack_len"`
	Cmd          uint8  `json:"cmd"`
	LocalIP      uint32 `json:"local_ip"`
	LocalPort    uint16 `json:"local_port"`
	ClientIP     uint32 `json:"client_ip"`
	ClientPort   uint16 `json:"client_port"`
	ConnectionID uint32 `json:"connection_id"`
	Flag         uint8  `json:"flag"`
	GatewayPort  uint16 `json:"gateway_port"`
	ExtLen       uint32 `json:"ext_len"`
	ExtData      []byte `json:"ext_data"`
	Body         []byte `json:"body"`
}
