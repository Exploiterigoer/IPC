package rwbytcpip

import (
	"worksys/common"
	"worksys/field"
)

// RWByTCPIP 通过TCP/IP进行读写数据的实体类
type RWByTCPIP struct {
	*common.Modbus // modbus结构体
	*field.RWField // 参数字段
}

// Receive 接收
func (rbt *RWByTCPIP) Receive(s interface{}) {

}

// Send 发送
func (rbt *RWByTCPIP) Send(inter interface{}, cmd *common.Command) {

}

// Packet 打包
func (rbt *RWByTCPIP) Packet(cmd *common.Command) []byte {
	return nil
}

// Unpacket 解包
func (rbt *RWByTCPIP) Unpacket(length int, buffer []byte, function func([]byte)) interface{} {
	return nil
}
