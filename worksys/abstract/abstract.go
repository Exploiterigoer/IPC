package abstract

import (
	"worksys/common"
)

// ReadWriteWorker 数据的收发接口
type ReadWriteWorker interface {
	// 收取的动作
	Receive(interface{})

	// 发送的动作
	Send(interface{}, *common.Command)

	// 打包的动作
	Packet(*common.Command) []byte

	// 解包的动作
	Unpacket(int, []byte, func([]byte)) interface{}
}
