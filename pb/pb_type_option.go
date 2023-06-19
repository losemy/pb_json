package pb

import (
	"encoding/json"
	"fmt"
)

// Type Proto序列化后的数据类型
type Type int8

const (
	// Varint varint，变长1-10字节(int32, int64, uint32, uint64, bool, enum, sint32, sint64)
	Varint Type = 0
	// Fixed32 fixed32，固定4字节(fixed32, sfixed32, float)
	Fixed32 Type = 5
	// Fixed64 fixed64，固定8字节(fixed64, sfixed64, double)
	Fixed64 Type = 1
	// Bytes bytes，变长(string, bytes, embedded messages, packed, repeated)
	Bytes Type = 2
	// StartGroup 弃用
	StartGroup Type = 3
	// EndGroup 弃用
	EndGroup Type = 4

	// Unkown 未知类型
	Unkown Type = 9
	// String pb中的string类型
	String Type = 10
	// Message pb中的message类型
	Message Type = 11
	// Int32 pb中的int32类型
	Int32 Type = 12
	// Int64 pb中的int64类型
	Int64 Type = 13
	// Uint pb中的uint32、uint64类型
	UInt Type = 14
	// SInt pb中的sint32、sint64类型
	SInt Type = 15
	// Bool pb中的bool类型
	Bool Type = 16
	// Double pb中的double类型
	Double Type = 17
	// Float pb中的float类型
	Float Type = 18
	// SFixed32 pb中的sfixed32类型
	SFixed32 Type = 19
	// SFixed64 pb中的sfixed64类型
	SFixed64 Type = 20
	// Packed 字段设置了[packed=true]
	Packed Type = 21

	// MaxTagValue 支持的tag最大值
	MaxTagValue = 9999
)

var (

	// typeNamesFormat 类型对应的名称
	typeNamesFormat = map[Type]string{
		Varint:            "%04d_varint",
		Fixed32:           "%04d_fixed32",
		Fixed64:           "%04d_fixed64",
		Bytes:             "%04d_bytes",
		String:            "%04d_string",
		Message:           "%04d_message",
		Int32:             "%04d_int32",
		Int64:             "%04d_int64",
		UInt:              "%04d_uint",
		SInt:              "%04d_sint",
		Bool:              "%04d_bool",
		Double:            "%04d_double",
		Float:             "%04d_float",
		SFixed32:          "%04d_sfixed32",
		SFixed64:          "%04d_sfixed64",
		Packed + Fixed32:  "%04d_packed.fixed32",
		Packed + Fixed64:  "%04d_packed.fixed64",
		Packed + Int32:    "%04d_packed.int32",
		Packed + Int64:    "%04d_packed.int64",
		Packed + UInt:     "%04d_packed.uint",
		Packed + SInt:     "%04d_packed.sint",
		Packed + Bool:     "%04d_packed.bool",
		Packed + Double:   "%04d_packed.double",
		Packed + Float:    "%04d_packed.float",
		Packed + SFixed32: "%04d_packed.sfixed32",
		Packed + SFixed64: "%04d_packed.sfixed64",
	}

	// namesToType 名称和对应类型的映射
	namesToType = map[string]Type{
		"varint":           Varint,
		"fixed32":          Fixed32,
		"fixed64":          Fixed64,
		"bytes":            Bytes,
		"string":           String,
		"message":          Message,
		"int32":            Int32,
		"int64":            Int64,
		"uint":             UInt,
		"sint":             SInt,
		"bool":             Bool,
		"double":           Double,
		"float":            Float,
		"sfixed32":         SFixed32,
		"sfixed64":         SFixed64,
		"packed.fixed32s":  Packed + Fixed32,
		"packed.fixed64s":  Packed + Fixed64,
		"packed.int32s":    Packed + Int32,
		"packed.int64s":    Packed + Int64,
		"packed.uints":     Packed + UInt,
		"packed.sints":     Packed + SInt,
		"packed.bools":     Packed + Bool,
		"packed.doubles":   Packed + Double,
		"packed.floats":    Packed + Float,
		"packed.sfixed32s": Packed + SFixed32,
		"packed.sfixed64s": Packed + SFixed64,
		"strings":          String,
		"messages":         Message,
		"varints":          Varint,
		"fixed32s":         Fixed32,
		"fixed64s":         Fixed64,
		"int32s":           Int32,
		"int64s":           Int64,
		"uints":            UInt,
		"sints":            SInt,
		"bools":            Bool,
		"doubles":          Double,
		"floats":           Float,
		"sfixed32s":        SFixed32,
		"sfixed64s":        SFixed64,
	}

	// varintNamesToType varint类型数据
	varintNamesToType = map[string]Type{
		"varint": Varint,
		"int32":  Int32,
		"int64":  Int64,
		"uint":   UInt,
		"sint":   SInt,
		"bool":   Bool,
	}

	// fixed32NamesToType fixed32类型数据
	fixed32NamesToType = map[string]Type{
		"fixed32":  Fixed32,
		"float":    Float,
		"sfixed32": SFixed32,
	}

	// fixed64NamesToType fixed64类型数据
	fixed64NamesToType = map[string]Type{
		"fixed64":  Fixed64,
		"double":   Double,
		"sfixed64": SFixed64,
	}

	// simpleBytesNamesToType 简单bytes类型数据
	simpleBytesNamesToType = map[string]Type{
		"bytes":   Bytes,
		"string":  String,
		"message": Message,
	}

	// listNamesToType unpacked repeated类型
	listNamesToType = map[string]Type{
		"strings":   String,
		"messages":  Message,
		"varints":   Varint,
		"fixed32s":  Fixed32,
		"fixed64s":  Fixed64,
		"int32s":    Int32,
		"int64s":    Int64,
		"uints":     UInt,
		"sints":     SInt,
		"bools":     Bool,
		"doubles":   Double,
		"floats":    Float,
		"sfixed32s": SFixed32,
		"sfixed64s": SFixed64,
	}

	// packedNamesToType packed repeated类型数据
	packedNamesToType = map[string]Type{
		"packed.fixed32s":  Packed + Fixed32,
		"packed.fixed64s":  Packed + Fixed64,
		"packed.int32s":    Packed + Int32,
		"packed.int64s":    Packed + Int64,
		"packed.uints":     Packed + UInt,
		"packed.sints":     Packed + SInt,
		"packed.bools":     Packed + Bool,
		"packed.doubles":   Packed + Double,
		"packed.floats":    Packed + Float,
		"packed.sfixed32s": Packed + SFixed32,
		"packed.sfixed64s": Packed + SFixed64,
	}
)

// Options 用户对PB数据解析的干预选择
type Options map[string]interface{}

// NewOptions 创建一个Options实例，失败则返回nil
func NewOptions(data []byte) Options {
	opts := Options{}
	err := json.Unmarshal(data, &opts)
	if err != nil {
		return nil
	}
	return opts
}

// GetOptionsByTag 通过tag获取对应的Options实例，如果失败则返回nil
func (o Options) GetOptionsByTag(tag string) Options {
	if o == nil {
		return nil
	}

	if opts, ok := o[GetOptionsKey(tag)].(map[string]interface{}); ok {
		return Options(opts)
	}
	if opts, ok := o[GetOptionsKey(tag)].(Options); ok {
		return opts
	}
	return nil
}

// GetOptionsKey 根据tag生成对应的Message使用的key
func GetOptionsKey(tag string) string {
	return fmt.Sprintf("%voptions", tag)
}

// GetTypeByTag 通过tag获取对应的type类型，失败返回
func (o Options) GetTypeByTag(tag string) Type {
	if o == nil {
		return Unkown
	}

	// 先判断是否是结构体类型
	if _, ok := o[tag].(map[string]interface{}); ok {
		return Message
	}

	// 再判断其它类型
	if name, ok := o[tag].(string); ok {
		if typ, ok := namesToType[name]; ok {
			return typ
		}
	}
	return Unkown
}
