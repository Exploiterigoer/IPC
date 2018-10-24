package common

import (
	"fmt"
)

// ReadCoil 等一系列常量为寄存器对应的功能码
const (
	ReadCoil   = 0x01 // 读线圈状态(暂不支持)
	ReadStat   = 0x02 // 读离散输入状态(暂不支持)
	ReadKReg   = 0x03 // 读保持寄存器
	ReadWReg   = 0x04 // 读输入寄存器
	WriteCoil  = 0x05 // 写单个线圈(暂不支持)
	WriteKReg  = 0x06 // 写单个保持寄存器
	WriteMCoil = 0x0F // 写多个线圈(暂不支持)
	WriteMKReg = 0x10 // 写多个保持寄存器
)

// Command 指令实体结构
type Command struct {
	deviceAddr   byte
	functionCode byte
	regStartAddr []byte
	regNumber    []byte
	byteValue    []byte
	CmdInterval  int // 指令发送的时间间隔
}

// ReadCMD 获取读数据的指令
func (c *Command) readCMD() []byte {
	rb := make([]byte, 0)
	rb = append(rb, c.deviceAddr)      // 设备地址
	rb = append(rb, c.functionCode)    // 功能码
	rb = append(rb, c.regStartAddr...) // 寄存器起始地址
	rb = append(rb, c.regNumber...)    // 寄存器个数
	return rb
}

// WriteCMD 获取写数据的指令
func (c *Command) writeCMD() []byte {
	wb := make([]byte, 0)
	wb = append(wb, c.deviceAddr)           // 设备地址
	wb = append(wb, c.functionCode)         // 功能码
	wb = append(wb, c.regStartAddr...)      // 寄存器起始地址
	wb = append(wb, c.regNumber...)         // 寄存器个数
	wb = append(wb, byte(len(c.byteValue))) // 写入的字节个数
	wb = append(wb, c.byteValue...)         // 写入的字节值
	return wb
}

// GetCMD 获取指令
func (c *Command) GetCMD() []byte {
	fc := c.functionCode
	if fc == ReadKReg || fc == ReadWReg {
		return c.readCMD()
	}
	return c.writeCMD()
}

// SetCMD 设置指令的字段值
// 关于args可变参数,前4个依次表示设备地址、功能码、寄存器起始地址、寄存器个数
// 对于读取指令,args共5个参数,最后一个表示本指令发送的时间间隔,单位是秒
// 对于写入指令,args共6个参数,第5个参数表示将要写入的字节,
// 最后一个一个表示本指令发送的时间间隔,单位是秒
func (c *Command) SetCMD(args ...interface{}) {
	argsLen := len(args)
	if argsLen < 5 {
		fmt.Println("指令配置存在错误")
		return
	}

	c.deviceAddr = byte(args[0].(int))
	c.functionCode = byte(args[1].(int))
	c.regStartAddr = IntToByte(args[2].(int), 1, 2)
	c.regNumber = IntToByte(args[3].(int), 1, 2)

	if argsLen == 5 {
		c.CmdInterval = args[4].(int)
	}

	if argsLen >= 6 {
		c.byteValue = args[4].([]byte)
		c.CmdInterval = args[5].(int)
	}
}
