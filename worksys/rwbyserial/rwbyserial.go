package rwbyserial

import (
	"bufio"
	"fmt"
	"time"
	"worksys/common"
	"worksys/field"

	"github.com/tarm/serial"
)

// RWBySerial 通过串口进行读写数据的实体类
type RWBySerial struct {
	*common.Modbus // Modbus  传输协议结构体
	*field.RWField // RWField 参数字段结构体
}

// Receive 接收数据,死循环且阻塞
func (rwbs *RWBySerial) Receive(s interface{}) {
	def := rwbs.decodeFunc() // 首先确定处理数据包的函数
	buf := make([]byte, 0)
	br := bufio.NewReader(s.(*serial.Port))
	for {
		// 阻塞地读取数据
		b, err := br.ReadByte()
		if err != nil {
			fmt.Println(err.Error())
		}

		// 至少读到5个字节后才继续往下执行
		buf = append(buf, b)
		if len(buf) < 4 {
			continue
		}

		//  判断功能码，计算并取得本次数据包的字节长度，以应对以对应读返回或者写返回数据包
		code := int(buf[1])
		length := rwbs.PacketLen(code, buf)
		if len(buf) != length {
			continue
		}

		result := rwbs.Unpacket(length, buf, def)
		switch result.(type) {
		case []interface{}:
			fmt.Println("取到数据", result.([]interface{}))
		case bool:
			fmt.Println("校验码错误", result.(bool))
		case []string:
			fmt.Println("从设备发生错误", result.([]string))
		case nil:
			fmt.Println("写入数据成功")
		}

		buf = make([]byte, 0)
	}
}

// Send 发送数据,死循环且阻塞
func (rwbs *RWBySerial) Send(s interface{}, cmd *common.Command) {
	var err error
	timer := time.Tick(1 * time.Second) // 每秒执行一次循环
	for {
		select {
		case <-timer:
			// 设备地址要写正确
			// PLC的寄存器地址从1开始的，纯modbus的寄存器地址从0开始的,寄存器地址也要写正确
			// 比如说取浮点数，在PLC中数据是从第一个寄存器开始的,那么起始寄存器地址只能是1,5,9,13...,不能是2,4,6...这些
			// 这是浮点数所占字节长度与PLC寄存器地址综合的结果，硬要从2,4,6开始，读到的数据是错误的
			// 取纯modbus的寄存器中的浮点数据也同理，起始地址只能是0,4,8,12...
			// 浮点数占两个寄存器,共4个字节，整数占1个寄存器，共2个字节
			t := time.Now().Second()
			if t%cmd.CmdInterval == 0 { // 每隔interval(毫秒)时间执行一次发送操作
				// ---测试开始---
				if cmd.CmdInterval == 10 {
					cmd.SetCMD(1, 3, 0, 4, 10)
				} else {
					sendVal := rwbs.RandDataGenerator(10, 1) //1 表示浮点数
					cmd.SetCMD(1, 16, 0, len(sendVal)/2, sendVal, 11)
				}
				// ---测试结束---
				d := rwbs.Packet(cmd)
				_, err = s.(*serial.Port).Write(d)
				if err != nil {
					fmt.Println("发送数据请求指令出错:" + err.Error())
				}
			}
		}
	}
}

// Packet 打包数据
func (rwbs *RWBySerial) Packet(cmd *common.Command) []byte {
	pkb := cmd.GetCMD()
	pkb = append(pkb, rwbs.ModbusCRC16ForByte(pkb)...)
	fmt.Println("指令执行时间:" + time.Now().Format("2006-01-02 15:04:05"))
	return pkb
}

// Unpacket 解包数据
// 解包需要根据配置参数中指定的数据类型、是什么浮点数类型、是什么整数类型、
// 是大端还是小端传输来进行解包数据包
func (rwbs *RWBySerial) Unpacket(length int, buffer []byte, function func([]byte)) interface{} {
	// 取到完整的数据包后，进行校验码比对，确认发送过来的数据包是否已经被改变
	// 被改变则丢弃本次数据包
	oldCRCBytes := buffer[length-2:]
	newCRCBytes := rwbs.ModbusCRC16ForByte(buffer[:length-2])
	if oldCRCBytes[0] != newCRCBytes[0] || oldCRCBytes[1] != newCRCBytes[1] {
		buffer = make([]byte, 0)
		return false
	}

	// 从设备响应异常,返回的功能码有0x81,0x82,0x83,0x84,0x85,0x86,0x8F,0x90
	if int(buffer[1]) > 16 {
		return common.TohexString(buffer)
	} else if int(buffer[1]) == 16 {
		return nil
	}

	function(buffer[3 : length-2]) // 传入浮点数对应的字节
	/*
		// 调整字节大小端顺序
		validData := rwbs.LittleBigEndian(rwbs.IsFloat, rwbs.ByteOrder, buf[3:length-2])

		// 取得最终结果
		result := rwbs.GetDataResult(rwbs.IsFloat, validData)
	*/
	return nil
}

func (rwbs *RWBySerial) decodeFunc() func([]byte) {
	m1 := new(common.Modbus1)
	// 电流值、单精度浮点数、小端模式
	flag1 := rwbs.DataType == 0 && rwbs.FloatType == 0 && rwbs.ByteOrder == 0
	if flag1 {
		return m1.AFloat32Littlendian
	}
	// 电流值、单精度浮点数、大端模式
	flag2 := rwbs.DataType == 0 && rwbs.FloatType == 0 && rwbs.ByteOrder == 1
	if flag2 {
		return m1.AFloat32Bigendian
	}
	// 电流值、双精度浮点数、小端模式
	flag3 := rwbs.DataType == 0 && rwbs.FloatType == 1 && rwbs.ByteOrder == 0
	if flag3 {
		return m1.AFloat64Littlendian
	}
	// 电流值、双精度浮点数、大端模式
	flag4 := rwbs.DataType == 0 && rwbs.FloatType == 1 && rwbs.ByteOrder == 1
	if flag4 {
		return m1.AFloat64Bigendian
	}
	// 电流值、int16整型、小端模式
	flag5 := rwbs.DataType == 0 && rwbs.IntType == 0 && rwbs.ByteOrder == 0
	if flag5 {
		return m1.AInt16Littlendian
	}
	// 电流值、int16整型、大端模式
	flag6 := rwbs.DataType == 0 && rwbs.IntType == 0 && rwbs.ByteOrder == 1
	if flag6 {
		return m1.AInt16Bigendian
	}
	// 电流值、int32整型、小端模式
	flag7 := rwbs.DataType == 0 && rwbs.IntType == 1 && rwbs.ByteOrder == 0
	if flag7 {
		return m1.AInt32Littlendian
	}
	// 电流值、int32整型、大端模式
	flag8 := rwbs.DataType == 0 && rwbs.IntType == 1 && rwbs.ByteOrder == 1
	if flag8 {
		return m1.AInt32Bigendian
	}
	// 电流值、int64整型、小端模式
	flag9 := rwbs.DataType == 0 && rwbs.IntType == 2 && rwbs.ByteOrder == 0
	if flag9 {
		return m1.AInt64Littlendian
	}
	// 电流值、int64整型、大端模式
	flag10 := rwbs.DataType == 0 && rwbs.IntType == 2 && rwbs.ByteOrder == 1
	if flag10 {
		return m1.AInt64Bigendian
	}
	// 电压值、单精度浮点数、小端模式
	flag11 := rwbs.DataType == 1 && rwbs.FloatType == 0 && rwbs.ByteOrder == 0
	if flag11 {
		return m1.VFloat32Littlendian
	}
	// 电压值、单精度浮点数、大端模式
	flag12 := rwbs.DataType == 1 && rwbs.FloatType == 0 && rwbs.ByteOrder == 1
	if flag12 {
		return m1.VInt32Bigendian
	}
	// 电压值、双精度浮点数、小端模式
	flag13 := rwbs.DataType == 1 && rwbs.FloatType == 1 && rwbs.ByteOrder == 0
	if flag13 {
		return m1.VInt64Littlendian
	}
	// 电压值、双精度浮点数、大端模式
	flag14 := rwbs.DataType == 1 && rwbs.FloatType == 1 && rwbs.ByteOrder == 1
	if flag14 {
		return m1.VInt64Bigendian
	}
	// 电压值、int16整型、小端模式
	flag15 := rwbs.DataType == 1 && rwbs.IntType == 0 && rwbs.ByteOrder == 0
	if flag15 {
		return m1.VInt16Littlendian
	}
	// 电压值、int16整型、大端模式
	flag16 := rwbs.DataType == 1 && rwbs.IntType == 0 && rwbs.ByteOrder == 1
	if flag16 {
		return m1.VInt16Bigendian
	}
	// 电压值、int32整型、小端模式
	flag17 := rwbs.DataType == 1 && rwbs.IntType == 1 && rwbs.ByteOrder == 0
	if flag17 {
		return m1.VInt32Littlendian
	}
	// 电压值、int32整型、大端模式
	flag18 := rwbs.DataType == 1 && rwbs.IntType == 1 && rwbs.ByteOrder == 1
	if flag18 {
		return m1.VInt32Bigendian
	}
	// 电压值、int64整型、小端模式
	flag19 := rwbs.DataType == 1 && rwbs.IntType == 2 && rwbs.ByteOrder == 0
	if flag19 {
		return m1.VInt64Littlendian
	}
	// 电压值、int64整型、大端模式
	flag20 := rwbs.DataType == 1 && rwbs.IntType == 2 && rwbs.ByteOrder == 1
	if flag20 {
		return m1.VInt64Bigendian
	}
	// 实测值、单精度浮点数、小端模式
	flag21 := rwbs.DataType == 2 && rwbs.FloatType == 0 && rwbs.ByteOrder == 0
	if flag21 {
		return m1.RFloat32Littlendian
	}
	// 实测值、单精度浮点数、大端模式
	flag22 := rwbs.DataType == 2 && rwbs.FloatType == 0 && rwbs.ByteOrder == 1
	if flag22 {
		return m1.RFloat32Bigendian
	}
	// 实测值、双精度浮点数、小端模式
	flag23 := rwbs.DataType == 2 && rwbs.FloatType == 1 && rwbs.ByteOrder == 0
	if flag23 {
		return m1.RFloat64Littlendian
	}
	// 实测值、双精度浮点数、大端模式
	flag24 := rwbs.DataType == 2 && rwbs.FloatType == 1 && rwbs.ByteOrder == 1
	if flag24 {
		return m1.RFloat64Bigendian
	}
	// 实测值、int16整型、小端模式
	flag25 := rwbs.DataType == 2 && rwbs.IntType == 0 && rwbs.ByteOrder == 0
	if flag25 {
		return m1.RInt16Littlendian
	}
	// 实测值、int16整型、大端模式
	flag26 := rwbs.DataType == 2 && rwbs.IntType == 0 && rwbs.ByteOrder == 1
	if flag26 {
		return m1.RInt16Bigendian
	}
	// 实测值、int32整型、小端模式
	flag27 := rwbs.DataType == 2 && rwbs.IntType == 1 && rwbs.ByteOrder == 0
	if flag27 {
		return m1.RInt32Littlendian
	}
	// 实测值、int32整型、大端模式
	flag28 := rwbs.DataType == 2 && rwbs.IntType == 1 && rwbs.ByteOrder == 1
	if flag28 {
		return m1.RInt32Bigendian
	}
	// 实测值、int64整型、小端模式
	flag29 := rwbs.DataType == 2 && rwbs.IntType == 2 && rwbs.ByteOrder == 0
	if flag29 {
		return m1.RInt64Littlendian
	}
	// 实测值、int64整型、大端模式
	flag30 := rwbs.DataType == 2 && rwbs.IntType == 2 && rwbs.ByteOrder == 1
	if flag30 {
		return m1.RInt64Bigendian
	}
	return nil
}
