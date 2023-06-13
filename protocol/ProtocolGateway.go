package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
)

const (
	a = iota
)

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

func Input(buffer []byte) uint32 {
	if len(buffer) < HEAD_LEN {
		return 0
	}
	packLen := binary.BigEndian.Uint32(buffer[:HEAD_LEN])
	return packLen
}

func Encode(data map[string]interface{}) []byte {
	flag := 0
	body := data["body"]
	if _, ok := body.(string); !ok {
		bodyBytes, _ := json.Marshal(body)
		body = string(bodyBytes)
		flag = 1
	}
	extData := data["ext_data"].(string)
	extLen := len(extData)
	packageLen := HEAD_LEN + extLen + len(body.(string))

	buffer := make([]byte, packageLen)
	binary.BigEndian.PutUint32(buffer[:4], uint32(packageLen))
	binary.BigEndian.PutUint32(buffer[4:8], uint32(data["cmd"].(int)))
	binary.BigEndian.PutUint32(buffer[8:12], uint32(data["local_ip"].(int)))
	binary.BigEndian.PutUint16(buffer[12:14], uint16(data["local_port"].(int)))
	binary.BigEndian.PutUint32(buffer[14:18], uint32(data["client_ip"].(int)))
	binary.BigEndian.PutUint16(buffer[18:20], uint16(data["client_port"].(int)))
	binary.BigEndian.PutUint32(buffer[20:24], uint32(data["connection_id"].(int)))
	binary.BigEndian.PutUint32(buffer[24:28], uint32(flag))
	binary.BigEndian.PutUint16(buffer[28:30], uint16(data["gateway_port"].(int)))
	binary.BigEndian.PutUint32(buffer[30:34], uint32(extLen))

	copy(buffer[34:34+extLen], []byte(extData))
	copy(buffer[34+extLen:], []byte(body.(string)))

	return buffer
}

// 从二进制数据转换为数组
func decode(buffer []byte) map[string]interface{} {
	data := make(map[string]interface{})

	packLen := binary.BigEndian.Uint32(buffer[0:4])
	cmd := buffer[4]
	localIP := binary.BigEndian.Uint32(buffer[5:9])
	localPort := binary.BigEndian.Uint16(buffer[9:11])
	clientIP := binary.BigEndian.Uint32(buffer[11:15])
	clientPort := binary.BigEndian.Uint16(buffer[15:17])
	connectionID := binary.BigEndian.Uint32(buffer[17:21])
	flag := buffer[21]
	gatewayPort := binary.BigEndian.Uint16(buffer[22:24])
	extLen := binary.BigEndian.Uint16(buffer[24:26])

	data["pack_len"] = packLen
	data["cmd"] = cmd
	data["local_ip"] = localIP
	data["local_port"] = localPort
	data["client_ip"] = clientIP
	data["client_port"] = clientPort
	data["connection_id"] = connectionID
	data["flag"] = flag
	data["gateway_port"] = gatewayPort
	data["ext_len"] = extLen

	if extLen > 0 {
		extData := buffer[26 : 26+extLen]
		data["ext_data"] = extData

		if flag&FLAG_BODY_IS_SCALAR > 0 {
			body := buffer[26+extLen:]
			data["body"] = body
		} else {
			body := buffer[26+extLen:]
			data["body"] = unserialize(body)
		}
	} else {
		data["ext_data"] = ""

		if flag&FLAG_BODY_IS_SCALAR > 0 {
			body := buffer[26:]
			data["body"] = body
		} else {
			body := buffer[26:]
			data["body"] = unserialize(body)
		}
	}

	return data
}

// 反序列化
func unserialize(data []byte) interface{} {
	var result interface{}
	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&result)
	if err != nil {
		return nil
	}
	return result
}

// 定义一个枚举类型 Command，底层类型为 uint8
type Command uint8

// 发给worker，gateway有一个新的连接
const CMD_ON_CONNECT = 1

// 发给worker的，客户端有消息
const CMD_ON_MESSAGE = 3

// 发给worker上的关闭链接事件
const CMD_ON_CLOSE = 4

// 发给gateway的向单个用户发送数据
const CMD_SEND_TO_ONE = 5

// 发给gateway的向所有用户发送数据
const CMD_SEND_TO_ALL = 6

// 发给gateway的踢出用户
// 1、如果有待发消息，将在发送完后立即销毁用户连接
// 2、如果无待发消息，将立即销毁用户连接
const CMD_KICK = 7

// 发给gateway的立即销毁用户连接
const CMD_DESTROY = 8

// 发给gateway，通知用户session更新
const CMD_UPDATE_SESSION = 9

// 获取在线状态
const CMD_GET_ALL_CLIENT_SESSIONS = 10

// 判断是否在线
const CMD_IS_ONLINE = 11

// client_id绑定到uid
const CMD_BIND_UID = 12

// 解绑
const CMD_UNBIND_UID = 13

// 向uid发送数据
const CMD_SEND_TO_UID = 14

// 根据uid获取绑定的clientid
const CMD_GET_CLIENT_ID_BY_UID = 15

// 加入组
const CMD_JOIN_GROUP = 20

// 离开组
const CMD_LEAVE_GROUP = 21

// 向组成员发消息
const CMD_SEND_TO_GROUP = 22

// 获取组成员
const CMD_GET_CLIENT_SESSIONS_BY_GROUP = 23

// 获取组在线连接数
const CMD_GET_CLIENT_COUNT_BY_GROUP = 24

// 按照条件查找
const CMD_SELECT = 25

// 获取在线的群组ID
const CMD_GET_GROUP_ID_LIST = 26

// 取消分组
const CMD_UNGROUP = 27

// worker连接gateway事件
const CMD_WORKER_CONNECT = 200

// 心跳
const CMD_PING = 201

// GatewayClient连接gateway事件
const CMD_GATEWAY_CLIENT_CONNECT = 202

// 根据client_id获取session
const CMD_GET_SESSION_BY_CLIENT_ID = 203

// 发给gateway，覆盖session
const CMD_SET_SESSION = 204

// 当websocket握手时触发，只有websocket协议支持此命令字
const CMD_ON_WEBSOCKET_CONNECT = 205

// 包体是标量
const FLAG_BODY_IS_SCALAR = 0x01

// 通知gateway在send时不调用协议encode方法，在广播组播时提升性能
const FLAG_NOT_CALL_ENCODE = 0x02

// 包头长度
const HEAD_LEN = 28
