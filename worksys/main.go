package main

import (
	"fmt"
	"worksys/abstract"
	"worksys/common"
	"worksys/field"
	"worksys/rwbyserial"
	"worksys/rwbytcpip"

	"github.com/akkuman/parseConfig"
	"github.com/tarm/serial"
)

func main() {
	//m := new(common.Modbus)
	//x := m.CombineToFloatBinStr(-0.0015118864014539925)
	//x := m.BinStrToFloat("10010110000011100110011001100110011001100110011001100 ")
	//x := m.BinStrToInt("10100000110100010101011000")
	//fmt.Println(x)
	//return

	config := parseConfig.New("conf.json")
	model := config.Get("model").(float64) // 通讯模式

	// 利用接口实现不同通讯方式但使用相同的方法读写数据
	var performer abstract.ReadWriteWorker
	var mdl interface{}
	switch model {
	case 1:
		//ip := config.Get("tcp > ip").(string)
		//port := config.Get("tcp > iport").(float64)
		performer = &rwbytcpip.RWByTCPIP{
			RWField: field.NewRWField(&config),
		}
	case 0:
		baud := config.Get("serial > baudRate").(float64)
		com := config.Get("serial > com").(string)
		performer = &rwbyserial.RWBySerial{
			RWField: field.NewRWField(&config),
		}

		sc := &serial.Config{Name: com, Baud: int(baud)}
		s, err := serial.OpenPort(sc)

		defer s.Close()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		mdl = s
	}

	//cmd1.SetCMD(1, 3, 0, 8, 10) //取8个寄存器的数据
	cmd1 := &common.Command{}
	cmd1.CmdInterval = 10

	//wd := RandDataGenerator(performer, 4, 0) // 被设置到寄存器的字节数组
	//cmd2.SetCMD(1, 16, 0, len(wd)/2, wd, 7) // 写寄存器个数 = 写入字节数 / 2
	//cmd2 := &common.Command{}
	//cmd2.CmdInterval = 7

	go performer.Send(mdl, cmd1) // 发送读取指令
	//go performer.Send(mdl, cmd2) // 发送写入指令
	performer.Receive(mdl) // 执行接收任务
}

/*
	判断奇偶：x & 1
	除以2的n次幂（乘法左移）：x >> n
	对2的n次幂取余：x & ~(~0<<n)
	从低位开始,将x的第n位置1：x | (1<<n)
	从低位开始,将x的第n位置0：x & ~(1<<n)
	测试x的第n位是否为1：x & (1<<n)

	如果是浮点数,那一次是读四个数,然后类型转换后得到浮点
	如果是整数那就要用量程转换了:
	最小量程+返回值*(最大量程-最小量程)/(最大值-最小值)
*/
