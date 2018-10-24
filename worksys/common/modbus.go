package common

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/biu"
)

// 32位字符'0'对应的ASCII码,补'0'的时候会用到
var zeroBytes = []byte{
	0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
	0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
	0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
	0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
	0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
	0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
	0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
	0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
}

// Modbus 协议实体结构
type Modbus struct{}

// PacketLen 计算本次数据包的长度
func (m *Modbus) PacketLen(code int, buf []byte) int {
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

// CombineToIntSlice 合成整数,每两个字节合成
func (m *Modbus) CombineToIntSlice(b []byte) int {
	var itd int
	intStr := biu.BytesToBinaryString(b)
	intStr = strings.Replace(intStr, " ", "", -1)
	intStr = strings.Replace(intStr, "[", "", -1)
	intStr = strings.Replace(intStr, "]", "", -1)

	if intStr[0]-'0' == 1 {
		// 负数的处理方式
		// 根据负整数二进制计算公式转换为10进制负数
		n := len(intStr)
		itd = int((-1)*math.Pow(2, float64(n-1))) + m.BinStrToInt(intStr[1:])
	} else {
		// 整数的处理方式
		// 直接将两个字节合成10进制正数
		itd = int(ByteToInt(b, 1, 16))
	}

	return itd
}

// CombineToFloatSlice 合成浮点数,每4个字节合成
func (m *Modbus) CombineToFloatSlice(flag, exp int, tail string) float64 {
	// 恰好为0的浮点数
	if flag == 0 && exp == -127 {
		return 0.00
	}

	intBinStr, floatBinStr := m.FixedBinStr(exp, tail) // 二进制的整、小数字符序列
	intd := m.BinStrToInt(intBinStr)                   //整数部分
	floatd := m.BinStrToFloat(floatBinStr)             //小数部分
	ifdata := float64(intd) + floatd
	if flag == 1 {
		ifdata *= -1
	}
	return ifdata
}

// CombineToFloatBinStr 合成浮点数的二进制字符序列,只支持单精度浮点数的合成
func (m *Modbus) CombineToFloatBinStr(f float64) string {
	// 如果是负数,要去掉前面的'-'
	fstr := strconv.FormatFloat(f, 'f', 8, 64)
	if f < 0 {
		fstr = strings.Replace(fstr, "-", "", 1)
	}

	// 找到小数点的位置
	dotIndex := strings.Index(fstr, ".")
	if dotIndex < 0 {
		fstr += ".0"
		dotIndex = strings.Index(fstr, ".")
	}
	// 计算整数、小数的二进制串
	it, _ := strconv.ParseInt(fstr[:dotIndex], 10, 64)
	ft, _ := strconv.ParseFloat(fstr[dotIndex:], 64)
	intBinStr := m.IntToBinStr(int(it)) // 整数部分的二进制串
	floatBinStr := m.FloatToBinStr(ft)  // 小数部分的二进制串

	// 计算浮点数二进制的阶码
	exp := 0
	if it == 0 {
		// 如果整数部分为0,则阶码就是小数部分的二进制串第一个'1'出现的下标
		// 找出浮点数部分二进制串第一个1出现的位置
		idx := strings.Index(floatBinStr, "1")
		exp = -(idx + 1) + 127
	} else {
		exp = len(intBinStr) - 1 + 127
	}

	// 阶码的二进制序列串,要补足8位
	expBinStr := m.IntToBinStr(exp)
	if expl := len(expBinStr); expl < 8 {
		expBinStr = string(zeroBytes[:8-expl]) + expBinStr
	}

	// 如果是整数部分为0,则从左边的第一个'1'开始,
	// 计算浮点数二进制串的尾数是否足够23位,不足则补全
	if it == 0 {
		idx := strings.Index(floatBinStr, "1")
		floatBinStr += string(zeroBytes[:idx])
	}

	// 如果整数部分为0,需要从小数部分找出第一个'1'的位置,尾数就是该位置后面的二进制串
	// 如果整数不为0,需要合并整数和小数二进制串,再找出第一个'1'的位置,尾数就是该位置后面的二进制串
	// fbidx+1和fbidx+1表示省略掉浮点数科学计数法1.AAA * 2^x中的 '1'
	// fbidx+23+1 表示取足够的23位尾数
	combineStr := ""
	if len(intBinStr) == 0 {
		fbidx := strings.Index(floatBinStr, "1")
		combineStr = floatBinStr[fbidx+1 : fbidx+23+1]
	} else {
		ifcb := intBinStr + floatBinStr
		fbidx := strings.Index(ifcb, "1")
		combineStr = ifcb[fbidx+1 : fbidx+23+1]
	}

	if f < 0 {
		combineStr = "1" + expBinStr + combineStr
	} else {
		combineStr = "0" + expBinStr + combineStr
	}

	if len(combineStr) > 32 {
		combineStr = combineStr[:32]
	}

	return combineStr
}

// BinStrToInt 递归算出整数部分二进制字符序列代表的10进制整数
func (m *Modbus) BinStrToInt(s string) int {
	slen := len(s)
	if slen == 1 {
		return int(s[0] - '0')
	}

	return m.BinStrToInt(s[0:slen-1])*2 + int(s[slen-1]-'0')
}

// IntToBinStr 递归算出浮点数整数部分的二进制字符序列
func (m *Modbus) IntToBinStr(i int) string {
	if i == 0 {
		return ""
	}

	if i < 0 {
		i *= -1
	}

	a := strconv.Itoa(i % 2)
	b := i / 2
	return m.IntToBinStr(b) + a
}

// BinStrToFloat 递归算出小数部分二进制字符序列代表的10进制浮点数,只支持单精度浮点数
func (m *Modbus) BinStrToFloat(s string) float64 {
	slen := len(s)
	if slen == 1 {
		return float64((s[0] - '0')) / 2.0
	}

	return float64((s[0]-'0'))/2.0 + m.BinStrToFloat(s[1:])/2.0
}

// FloatToBinStr 递归算出浮点数小数部分的二进制字符序列,只支持单精度浮点数
func (m *Modbus) FloatToBinStr(f float64) string {
	// 固定浮点数的精度为8为小数
	precise := strconv.FormatFloat(f-float64(int(f)), 'f', 8, 64)
	p, _ := strconv.ParseFloat(precise, 64)
	if p < 1e-64 { // 判断的精度为64位小数
		return "" + strconv.Itoa(int(f))
	}

	a := int(f * 2)
	b := f*2 - float64(a)

	return "" + strconv.Itoa(a) + m.FloatToBinStr(b)
}

// FixedBinStr "修复"二进制字符串且拆分整数、小数两部分
// "修复"的意思是负数的阶码时,在得到的小数尾数前面补0,且补上阶码绝对值个数的0
// 此时,整数部分是0,小数部分就是从下标为1起的补0后的尾数
// 正数阶码时,从左往右取阶码绝对值个数的尾数位作为整数部分,剩下的作为小数部分,
// 只支持单精度浮点数
func (m *Modbus) FixedBinStr(exp int, bin string) (string, string) {
	var intPart, floatPart string
	zeroLen := exp

	if exp < 0 {
		// 阶码小于0需要左移动小数点，移动就是相当于前边补'0'
		// '0'的个数就是阶码-1
		zeroLen *= -1
		intPart = "0"
		floatPart = (string(zeroBytes[:zeroLen]) + bin)[1:]
	} else {
		// binLen-1表示去除开头第一个'1'后剩下的字符长度
		// zeroLen-binLen+2 = zeroLen-(binLen-1)+1
		// zeroLen-(binLen-1)+1表示补充zeroLen-(binLen-1)个'0'字符
		// 因为数组不包含最后的下标对应的元素,所以再加上1
		if binLen := len(bin); binLen-1 <= zeroLen {
			bin += string(zeroBytes[:zeroLen-binLen+2])
		}

		// 整数部分就是bin字符串0到下标zeroLen之间所包含的字符串(注意数组下标取值'左闭右开')
		// 小数部分就是bin字符串下标zeroLen之后所包含的字符串
		intPart = bin[:zeroLen+1]
		floatPart = bin[zeroLen+1:]
	}
	return intPart, floatPart
}

// IntLittleEndian 小端对齐的整数字节转大端对齐
// 整数在寄存器中允许的最大值为65535,对应无符号的整数最大值
// 最小值为-32768,对应有符号整数最小值
// 整数只占一个寄存器,同样分大小端对齐
// 整数且小端对齐
func (m *Modbus) IntLittleEndian(data []byte) []byte {
	dataLength := len(data)
	for i := 0; i < dataLength; i++ {
		if (i+1)%2 == 0 {
			data[i-1], data[i] = data[i], data[i-1]
		}
	}
	return data
}

// FloatLittleEndian 小端对齐的整数字节转大端对齐
// 浮点数在寄存器中以4个字节保存,即2个寄存器
// 而这两个寄存器地址总是从低到高的
// 低地址位的寄存器存浮点数的第3,4个字节(低位字节)
// 高地址位的寄存器存浮点数的第1,2个字节(高位字节)
// 按标准的modbus协议传输方式,每个字节
// 要按照先高8位后低8位存储
// modbus传输时总是先传输低位字节,后传输高位字节(小端对齐)
// 在合成浮点数时,顺序一定是先高位字节后低位字节(大端对齐)
// 所以需要根据实际情况,在上位机统一成大端对齐的数据。
func (m *Modbus) FloatLittleEndian(data []byte) []byte {
	validData := make([]byte, 0)
	dataLength := len(data)
	for i := dataLength; i > 0; i-- {
		if i > 0 && i%2 == 0 {
			validData = append(validData, data[i-2:i]...)
		}
	}
	return validData
}

// LittleBigEndian 字节大小端对齐转化，统一转为大端对齐
func (m *Modbus) LittleBigEndian(isFloat, byteOrder byte, data []byte) []byte {
	validData := make([]byte, 0)
	if isFloat == 1 && byteOrder == 0 {
		validData = m.FloatLittleEndian(data) // 浮点数且小端对齐的处理
	} else if isFloat == 0 && byteOrder == 0 {
		validData = m.IntLittleEndian(data) // 整数且小端对齐的处理
	} else {
		validData = data // 浮点数或整数大端对齐的处理
	}
	return validData
}

// BinStrTrim 清除二进制字符串中多余的字符
func (m *Modbus) BinStrTrim(b []byte) string {
	binStr := biu.ToBinaryString(b)
	binStr = strings.Replace(binStr, " ", "", -1)
	binStr = strings.Replace(binStr, "[", "", -1)
	binStr = strings.Replace(binStr, "]", "", -1)
	return binStr
}

// GetDataResult 获取最终处理结果,不管接收到的是浮点数还是整数
// flag 表示整数还是负数的标记,1表示负数,0表示正数
// 减去'0'可实现字符与数字之间的转换
// 0~9的ASCII码为48~57,是连续的,'x'-'0'相当于
// x-48,如果x=48,则'x'-'0'=0
// 浮点数的阶码,即2的幂次,可正可负
// 浮点数尾数,并且补足24位,如果exp>0,整数部分的二进制为
// tail[:exp+1],小数部分的二进制为 tail[exp+1:]
func (m *Modbus) GetDataResult(isFloat byte, validData []byte) interface{} {
	returnVal := make([]interface{}, 0)

	vdl := len(validData)
	for i := 0; i < vdl; i++ {
		if (i+1)%4 == 0 && isFloat == 1 { // 浮点数每4个字节处理一次
			binStr := m.BinStrTrim(validData[i-3 : i+1])
			flag := binStr[0] - '0'
			exp := m.BinStrToInt(binStr[1:9]) - 127
			tail := "1" + binStr[9:]
			returnVal = append(returnVal, m.CombineToFloatSlice(int(flag), exp, tail))
		} else if (i+1)%2 == 0 && isFloat != 1 { // 整数每2个字节处理一次
			returnVal = append(returnVal, m.CombineToIntSlice(validData[i-1:i+1]))
		}
	}

	return returnVal
}

// ModbusCRC16ForString modbus CRC16 of checkout algorithm.
func (m *Modbus) ModbusCRC16ForString(dataString string) string {
	crc := 0xFFFF
	length := len(dataString)
	for i := 0; i < length; i++ {
		// gets the low 8 bits for calculating.
		crc = crc ^ int(dataString[i])
		for j := 0; j < 8; j++ {
			flag := crc & 0x0001
			crc >>= 1
			if flag == 1 {
				crc ^= 0xA001
			}
		}
	}

	//formats the check code with littleEndian.
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, int16(crc))
	hex := TohexString(bytesBuffer.Bytes())
	return hex[0] + hex[1]
}

// ModbusCRC16ForByte modbus CRC16 of checkout algorithm.
func (m *Modbus) ModbusCRC16ForByte(b []byte) []byte {
	crc := 0xFFFF
	length := len(b)
	for i := 0; i < length; i++ {
		// gets the low 8 bits for calculating.
		crc = crc ^ int(b[i])
		for j := 0; j < 8; j++ {
			flag := crc & 0x0001
			crc >>= 1
			if flag == 1 {
				crc ^= 0xA001
			}
		}
	}

	//formats the check code with littleEndian.
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, int16(crc))
	//hex := TohexString(bytesBuffer.Bytes())
	//return hex[0] + hex[1]
	return bytesBuffer.Bytes()
}

//HJt212CRC16ForString CRC16 of checkout algorithm.
func (m *Modbus) HJt212CRC16ForString(dataString string) string {
	crc := 0xFFFF
	length := len(dataString)
	for i := 0; i < length; i++ {
		// gets the heigh 8 bits for calculating.
		crc = (crc >> 8) ^ int(dataString[i])
		for j := 0; j < 8; j++ {
			flag := crc & 0x0001
			crc >>= 1
			if flag == 1 {
				crc ^= 0xA001
			}
		}
	}

	//formats the check code with littleEndian.
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, int16(crc))
	hex := TohexString(bytesBuffer.Bytes())
	return hex[1] + hex[0]
}

//HJt212CRC16ForByte CRC16 of checkout algorithm.
func (m *Modbus) HJt212CRC16ForByte(b []byte) []byte {
	crc := 0xFFFF
	length := len(b)
	for i := 0; i < length; i++ {
		// gets the heigh 8 bits for calculating.
		crc = (crc >> 8) ^ int(b[i])
		for j := 0; j < 8; j++ {
			flag := crc & 0x0001
			crc >>= 1
			if flag == 1 {
				crc ^= 0xA001
			}
		}
	}

	//formats the check code with littleEndian.
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, int16(crc))
	//hex := TohexString(bytesBuffer.Bytes())
	//return hex[1] + hex[0]
	return bytesBuffer.Bytes()
}

// RandDataGenerator 随机数生成器,生成的是随机数对应的字节
// number 表示生成随机数的个数
// isFloat表示生成随机数的类型,1--整型,0--浮点型
// 本函数为临时测试函数
func (m *Modbus) RandDataGenerator(number int, isFloat int) []byte {
	data := make([]byte, 0)
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for i := 0; i < number; i++ {
		if isFloat == 1 {
			d := r.Float64()*float64((i+1)*(i+1)) + float64(i+1)
			d = d * (-1) / 1e8 // 指数最高为8
			dd := m.CombineToFloatBinStr(float64(d))
			ddd := m.BinStrToInt(dd)
			dddd := IntToByte(ddd, 1, 4)
			data = append(data, dddd...)
		} else {
			data = append(data, IntToByte(r.Intn(32768)*(-1), 1, 2)...)
		}
	}

	return data
}
