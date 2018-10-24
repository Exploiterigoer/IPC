package common

import (
	"fmt"
	"strings"

	"github.com/biu"
)

// Modbus1 协议实体结构
type Modbus1 struct{}

// zeroGenerator 字符'0'对应的ASCII码生成器,补'0'的时候会用到
func (m *Modbus1) zeroGenerator(number int) []byte {
	zero := make([]byte, 0)
	for i := 0; i < number; i++ {
		zero = append(zero, byte(0x30))
	}

	return zero
}

// bigEndianConvert 将小端对齐的字节转为大端对齐
func (m *Modbus1) bigEndianConvert(b []byte) []byte {
	bigEndianData := make([]byte, 0)
	for i := len(b); i > 0; i-- {
		if i > 0 && i%2 == 0 {
			bigEndianData = append(bigEndianData, b[i-2:i]...)
		}
	}
	return bigEndianData
}

// CountLength 计算本次数据包的长度
func (m *Modbus1) CountLength(code int, buf []byte) int {
	var length int
	if code == ReadKReg {
		// 地址位(1) + 功能码(2) +数据长度(x) + 校验码(2)
		length = 3 + int(buf[2]) + 2
	} else if code == ReadWReg {

	} else if code == WriteKReg {

	} else if code == WriteMKReg {
		// 写入正确,格式为:
		// 设备地址(1) + 功能码(1) + 寄存器起始地址(2) + 寄存器个数(2) + 校验码(2)
		length = 8
	} else {
		length = 5
		// 各种错误,格式为:
		// 设备地址(1) + 功能码(1) + 错误码(1) + 校验码(2)
		// 其中功能码已经在原来的功能码上加上了0x80
	}

	return length
}

// AFloat32Littlendian 电流值单精度浮点数小端对齐的处理
// 单精度浮点数占4个字节,每两个寄存器才能得到一个单精度浮点数
// 因而从读取回来的字节数组需要每4个字节分为一组,即得到一个
// 浮点数的16进制形式.IEEE规则下的浮点数格式: 1.xxx * 2^e
// 这个16进制还要转为二进制字符串,根据/二进制字符串的第1位是
// '1'表示该浮点数是负数,是'0'表示该浮点数// 正数,再根据二进制
// 的2~10共8位二进制位计算出10进制值,就得到a×b^m// 中的m的值,
// 剩下23位就是浮点数的尾数.因为a恒等于1,在计算机中是被省略
// 了的,在计算时要补回来,即'1+尾数'.
// 如果m<0,则在'1+尾数'前补上m个'0',前面补上了m个'0'后的'1+尾数'的
// 第1个'0'表示整数部分二进制,剩下的比特位// 全是小数部分的二进制位,
// 根据小数部分的二进制为，利用递归算法还原为10进制的小数,此种情况下,整数为0.
// 如果m>0,则整数部分二进制位为'1+尾数'中从左往右数的m位比特位,小数部分就是
// 剩下的二进制位，但是要注意,往往m恰好等于'1+尾数'的长度,这意味着没有小数对应
// 的二进制位,此时需要把小数部分的二进制位补'0'，补一个即可
func (m *Modbus1) AFloat32Littlendian(b []byte) {
	fmt.Println("电流值单精度浮点数小端对齐的处理")
	// 输入字节先转为大端对齐，再处理成浮点数
	data := m.convertFloatBinStrToFloat(4, m.bigEndianConvert(b))
	fmt.Println(data)
}

// AFloat32Bigendian 电流值单精度浮点数大端对齐的处理
func (m *Modbus1) AFloat32Bigendian(b []byte) {
	fmt.Println("电流值单精度浮点数大端对齐的处理")
	data := m.convertFloatBinStrToFloat(4, b)
	fmt.Println(data)
}

// AFloat64Littlendian 电流值双精度浮点数小端对齐的处理
func (m *Modbus1) AFloat64Littlendian(b []byte) {
	fmt.Println("电流值双精度浮点数小端对齐的处理")
	data := m.convertFloatBinStrToFloat(64, m.bigEndianConvert(b))
	fmt.Println(data)
}

// AFloat64Bigendian 电流值双精度浮点数大端对齐的处理
func (m *Modbus1) AFloat64Bigendian(b []byte) {
	fmt.Println("电流值双精度浮点数大端对齐的处理")
	data := m.convertFloatBinStrToFloat(64, b)
	fmt.Println(data)
}

// convertFloatBinStrToFloat 将浮点数16字节转为10进制的数,包括float32和float64
func (m *Modbus1) convertFloatBinStrToFloat(floatType int, b []byte) []float64 {
	returnData := make([]float64, 0)
	var fIndex int // 一个浮点数的起始字节对应的二进制字符串起始位
	var eIndex int // 一个浮点数的阶码的起始位
	var base int   // 一个浮点数底数的指数
	bitSize := floatType / 8
	if bitSize == 4 {
		fIndex = 3
		eIndex = 9
		base = 127
	} else {
		fIndex = 7
		eIndex = 12
		base = 1023
	}

	zeroBytes1 := m.zeroGenerator(base)

	blen := len(b)
	for i := 0; i < blen; i++ {
		if (i+1)%bitSize == 0 {
			// 算出二进制字符串、正负标记、阶码、尾数
			binStr := m.byteToBinStr(b[i-fIndex : i+1])   // IEEE浮点数格式浮点数的二进制字符串
			flag := binStr[0] - '0'                       // 正数负数标记
			exp := m.binStrToInt(binStr[1:eIndex]) - base // 浮点数阶码,即m的值
			tail := "1" + binStr[eIndex:]                 // 补'1'后的尾数
			intBinStr, floatBinStr := "", ""              // 小数点位置=exp+1

			// 分割整数部分、小数部分的二进制字符串
			if exp > 0 && exp < len(tail)-1 { // 尾数足够长,第1个'1'不算尾数,xxx才是
				intBinStr = tail[:exp+1]
				floatBinStr = tail[exp+1:]
			}

			if exp > 0 && exp > len(tail) { // 尾数不足,补上exp-len(tail)个'0'
				tail += string(zeroBytes1[:exp-len(tail)+1])
				fmt.Println(tail, len(tail))
				intBinStr = tail[:exp+1]
				floatBinStr = tail[exp+1:]
			}

			if exp < 0 {
				zeroLen := (-1) * exp // 需要补的'0'的个数
				intBinStr = "0"
				floatBinStr = string(zeroBytes1[:zeroLen-1]) + tail //zeroLen-1表示不包含整数部分的'0'
			}

			if len(floatBinStr) == 0 { // 避免整数部分的二进制字符串恰好是整个二进制字符串
				floatBinStr = "0" // 导致小数部分没有二进制字符串,需要补一个'0'作为小数二进制字符串
			}

			// 计算整数部分、小数部分各自对应的十进制数值
			intPart := m.binStrToInt(intBinStr)
			floatPart := m.binStrToFloat(floatBinStr)
			floatData := float64(intPart) + floatPart

			// 判断正负数
			if flag == 1 {
				floatData *= -1
			}
			returnData = append(returnData, floatData)
		}
	}

	return returnData
}

// AInt16Littlendian 电流值int16型整数小端对齐的处理
func (m *Modbus1) AInt16Littlendian(b []byte) {
	fmt.Println("电流值int16型整数小端对齐的处理")
}

// AInt16Bigendian 电流值int16型整数大端对齐的处理
func (m *Modbus1) AInt16Bigendian(b []byte) {
	fmt.Println("电流值int16型整数大端对齐的处理")
}

// AInt32Littlendian 电流值int32型整数小端对齐的处理
func (m *Modbus1) AInt32Littlendian(b []byte) {
	fmt.Println("电流值int32型整数小端对齐的处理")
}

// AInt32Bigendian 电流值int32型整数大端对齐的处理
func (m *Modbus1) AInt32Bigendian(b []byte) {
	fmt.Println("电流值int32型整数大端对齐的处理")
}

// AInt64Littlendian 电流值int64型整数小端对齐的处理
func (m *Modbus1) AInt64Littlendian(b []byte) {
	fmt.Println("电流值int64型整数小端对齐的处理")
}

// AInt64Bigendian 电流值int64型整数大端对齐的处理
func (m *Modbus1) AInt64Bigendian(b []byte) {
	fmt.Println("电流值int64型整数大端对齐的处理")
}

// VFloat32Littlendian 电压值单精度浮点数小端对齐的处理
func (m *Modbus1) VFloat32Littlendian(b []byte) {

}

// VFloat32Bigendian 电压值单精度浮点数大端对齐的处理
func (m *Modbus1) VFloat32Bigendian(b []byte) {

}

// VFloat64Littlendian 电压值双精度浮点数小端对齐的处理
func (m *Modbus1) VFloat64Littlendian(b []byte) {

}

// VFloat64Bigendian 电压值双精度浮点数大端对齐的处理
func (m *Modbus1) VFloat64Bigendian(b []byte) {

}

// VInt16Littlendian 电压值int16型整数小端对齐的处理
func (m *Modbus1) VInt16Littlendian(b []byte) {

}

// VInt16Bigendian 电压值int16型整数大端对齐的处理
func (m *Modbus1) VInt16Bigendian(b []byte) {

}

// VInt32Littlendian 电压值int32型整数小端对齐的处理
func (m *Modbus1) VInt32Littlendian(b []byte) {

}

// VInt32Bigendian 电压值int32型整数大端对齐的处理
func (m *Modbus1) VInt32Bigendian(b []byte) {

}

// VInt64Littlendian 电压值int64型整数小端对齐的处理
func (m *Modbus1) VInt64Littlendian(b []byte) {

}

// VInt64Bigendian 电压值int64型整数大端对齐的处理
func (m *Modbus1) VInt64Bigendian(b []byte) {

}

// RFloat32Littlendian 实测值单精度浮点数小端对齐的处理
func (m *Modbus1) RFloat32Littlendian(b []byte) {

}

// RFloat32Bigendian 实测值单精度浮点数大端对齐的处理
func (m *Modbus1) RFloat32Bigendian(b []byte) {

}

// RFloat64Littlendian 实测值双精度浮点数小端对齐的处理
func (m *Modbus1) RFloat64Littlendian(b []byte) {

}

// RFloat64Bigendian 实测值双精度浮点数大端对齐的处理
func (m *Modbus1) RFloat64Bigendian(b []byte) {

}

// RInt16Littlendian 实测值int16型整数小端对齐的处理
func (m *Modbus1) RInt16Littlendian(b []byte) {

}

// RInt16Bigendian 实测值int16型整数大端对齐的处理
func (m *Modbus1) RInt16Bigendian(b []byte) {

}

// RInt32Littlendian 实测值int32型整数小端对齐的处理
func (m *Modbus1) RInt32Littlendian(b []byte) {

}

// RInt32Bigendian 实测值int32型整数大端对齐的处理
func (m *Modbus1) RInt32Bigendian(b []byte) {

}

// RInt64Littlendian 实测值int64型整数小端对齐的处理
func (m *Modbus1) RInt64Littlendian(b []byte) {

}

// RInt64Bigendian 实测值int64型整数大端对齐的处理
func (m *Modbus1) RInt64Bigendian(b []byte) {

}

// byteToBinStr 将字节数据转化为二进制字符串
func (m *Modbus1) byteToBinStr(b []byte) string {
	binStr := biu.ToBinaryString(b)
	binStr = strings.Replace(binStr, " ", "", -1)
	binStr = strings.Replace(binStr, "[", "", -1)
	binStr = strings.Replace(binStr, "]", "", -1)
	return binStr
}

// binStrToInt 递归算出整数对应的二进制字符序列所代表的10进制整数
func (m *Modbus1) binStrToInt(binStr string) int {
	binStrLen := len(binStr)
	if binStrLen == 1 {
		return int(binStr[0] - '0')
	}

	return m.binStrToInt(binStr[0:binStrLen-1])*2 + int(binStr[binStrLen-1]-'0')
}

//  binStrToInt 整数二进制字符串转对应的十进制数(尾递归)
func (m *Modbus) binStrToInt(binStr string) int {
	binStrLen := len(binStr)
	if binStrLen == 1 {
		return int(binStr[0] - '0')
	}

	return m.binStrToInt(binStr[0:binStrLen-1])*2 + int(binStr[binStrLen-1]-'0')
}

// binStrToFloat 浮点数二进制字符串转对应的十进制数(尾递归)
func (m *Modbus1) binStrToFloat(binStr string) float64 {
	binStrLen := len(binStr)
	if binStrLen == 1 {
		return float64((binStr[0] - '0')) / 2.0
	}

	return float64((binStr[0]-'0'))/2.0 + m.binStrToFloat(binStr[1:])/2.0
}
