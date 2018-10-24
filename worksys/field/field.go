package field

import (
	"github.com/akkuman/parseConfig"
)

// RWField 结构体参数
type RWField struct {
	IsActive           byte    //主动或被动模式
	Interval           byte    // 指令接收或发送间隔
	DataType           byte    // 电流或电压或实测值
	FloatType          byte    // 浮点数的类型
	IntType            byte    // 整数的类型
	ByteOrder          byte    // 接收回来的数据的大小端对齐标记
	LowerLimit         float64 // 量程最小值
	UpperLimit         float64 // 量程最大值
	LinearCoefficient  float64 // 线性系数
	CorrectedIntercept float64 // 修正截距
	PollutionType      byte    // 污染源类型
}

// NewRWField 生成读写结构体字段并返回该结构体
func NewRWField(conf *parseConfig.Config) *RWField {
	rwf := new(RWField)
	isActive := conf.Get("isActive").(float64)
	interval := conf.Get("interval").(float64)
	dataType := conf.Get("dataType").(float64)
	floatType := conf.Get("floatType").(float64)
	intType := conf.Get("intType").(float64)
	byteOrder := conf.Get("byteOrder").(float64)
	lowerLimit := conf.Get("lowerLimit").(float64)
	upperLimit := conf.Get("upperLimit").(float64)
	lc := conf.Get("linearCoefficient").(float64)
	ci := conf.Get("correctedIntercept").(float64)
	pt := conf.Get("pollutionType").(float64)

	rwf.IsActive = byte(isActive)
	rwf.Interval = byte(interval)
	rwf.DataType = byte(dataType)
	rwf.FloatType = byte(floatType)
	rwf.IntType = byte(intType)
	rwf.ByteOrder = byte(byteOrder)
	rwf.LowerLimit = lowerLimit
	rwf.UpperLimit = upperLimit
	rwf.LinearCoefficient = lc
	rwf.CorrectedIntercept = ci
	rwf.PollutionType = byte(pt)

	return rwf
}
