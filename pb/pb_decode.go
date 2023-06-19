package pb

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	"google.golang.org/protobuf/encoding/protowire"
)

// 特殊字符的定义
const (
	// HorizontalTab 水平制表符
	HorizontalTab = 9
	// NewLineChar 换行符
	NewLineChar = 10
	// CarriageReturn 回车符
	CarriageReturn = 13
	// MaxCtrlChar 最大的控制字符
	MaxCtrlChar = 31
	// DeleteChar 删除符
	DeleteChar = 127
)

var (
	// errPBTagTooBig pb的tag值太大
	errPBTagTooBig = errors.New("pb's tag too big")
	// errUnknownType 未知的PB类型
	errUnknownType = errors.New("unknown type")
)

// FieldMeta 保存Protobuf字段序列化或者反序列化的元数据
type FieldMeta struct {
	// Tag 字段的tag值
	Tag uint64
	// Type 字段的type值
	Type Type
}

// readTagType 从序列化后的二进制数据中读取tag和type，并且返回剩余的数据
func readTagType(raw []byte) (tagType *FieldMeta, rest []byte, err error) {
	tag, typ, length := protowire.ConsumeTag(raw)
	if length < 0 {
		return nil, nil, protowire.ParseError(length)
	}
	if tag > MaxTagValue {
		return nil, nil, errPBTagTooBig
	}

	tagType = &FieldMeta{
		Tag:  uint64(tag),
		Type: Type(typ),
	}
	return tagType, raw[length:], nil
}

// isString 判断raw中的二进制数据是否是字符串. 根据其中是否有控制字符来判断
// 有控制字符则代表是不是字符串
func isString(raw []byte) bool {

	for _, c := range raw {
		if c == HorizontalTab || c == NewLineChar || c == CarriageReturn {
			// 水平制表符、换行符和回车符认为是合法字符串字符
			return true
		} else if c == DeleteChar {
			// 删除符认为是非法字符串字符
			return false
		} else if c <= MaxCtrlChar {
			return false
		}
	}
	return true
}

// DecodeInterface 将PB二进制数据反序列化为map[string]interface{}数据
// raw: 要进行反序列化的PB数据
// opts: 用户针对每个字段的干预选择
func DecodeInterface(raw []byte, opts Options) (map[string]interface{}, error) {
	res, err := decode(raw, opts)
	if err != nil {
		return nil, err
	}
	res.FixTagTypeNames()
	return map[string]interface{}(res), nil
}

// Decode 将PB二进制数据反序列化为json数据
// raw: 要进行反序列化的PB数据
// opts: 用户针对每个字段的干预选择
func Decode(raw []byte, opts Options) (string, error) {
	res, err := decode(raw, opts)
	if err != nil {
		return "", err
	}

	res.FixTagTypeNames()

	data, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// decode 将PB二进制数据反序列化为json数据格式的JSONResult
// raw: 要进行反序列化的PB数据
// opts: 用户针对每个字段的干预选择
func decode(raw []byte, opts Options) (JSONResult, error) {

	result := JSONResult{}
	var err error
	for len(raw) > 0 {
		// 读取tag和type
		var tagType *FieldMeta
		tagType, raw, err = readTagType(raw)
		if err != nil {
			return nil, err
		}

		switch tagType.Type {
		case Varint:
			raw, err = readVarint(raw, tagType.Tag, opts, result)
		case Bytes:
			data, length := protowire.ConsumeBytes(raw)
			if length < 0 {
				return nil, protowire.ParseError(length)
			}
			raw = raw[length:]
			err = readBytes(data, tagType.Tag, opts, result)
		case Fixed32:
			raw, err = readFixed32(raw, tagType.Tag, opts, result)
		case Fixed64:
			raw, err = readFixed64(raw, tagType.Tag, opts, result)
		default:
			return nil, errUnknownType
		}

		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// readVarint 解析varint类型
// raw: 要反序列化的PB数据
// tag: 要反序列化的字段的tag
// opts: 用户干预反序列化的选择
// result: 反序列化的结果
func readVarint(raw []byte, tag uint64, opts Options,
	result JSONResult) ([]byte, error) {
	value, length := protowire.ConsumeVarint(raw)
	if length < 0 {
		return raw, protowire.ParseError(length)
	}
	raw = raw[length:]

	// 根据用户选择进行类型转换，默认Varint类型
	typ := opts.GetTypeByTag(strconv.FormatUint(tag, 10))
	typeName := fmt.Sprintf(typeNamesFormat[typ], tag)
	switch typ {
	case Int32:
		result.Append(typeName, int32(value))
	case Int64:
		result.Append(typeName, int64(value))
	case UInt:
		result.Append(typeName, uint64(value))
	case SInt:
		result.Append(typeName, protowire.DecodeZigZag(value))
	case Bool:
		if value == 0 {
			result.Append(typeName, false)
			break
		}
		result.Append(typeName, true)
	default:
		typeName = fmt.Sprintf(typeNamesFormat[Varint], tag)
		result.Append(typeName, value)
	}
	return raw, nil
}

// readBytes 解析bytes类型
// data: 要反序列化的PB数据
// tag: 要反序列化的字段的tag
// opts: 用户干预反序列化的选择
// result: 反序列化的结果
func readBytes(data []byte, tag uint64, opts Options,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readBytes] %w", err)
		}
	}()

	// 根据用户选择进行类型转换，默认进行推测
	sTag := strconv.FormatUint(tag, 10)
	typ := opts.GetTypeByTag(sTag)
	typeName := fmt.Sprintf(typeNamesFormat[typ], tag)
	switch {
	case typ == Bytes:
		result.Append(typeName, hex.EncodeToString(data))
	case typ == String:
		result.Append(typeName, string(data))
	case typ == Message:
		// 递归解析
		res, nerr := decode(data, opts.GetOptionsByTag(sTag))
		if nerr != nil {
			return nerr
		}
		result.Append(typeName, res)
	case typ >= Packed:
		// packed=true的repeated类型数据
		return readPacked(data, tag, typ, result)
	default:
		// 先推测为嵌套类型
		res, nerr := decode(data, opts)
		if nerr == nil {
			typeName := fmt.Sprintf(typeNamesFormat[Message], tag)
			result.Append(typeName, res)
			return nil
		}
		// 在判断是否有控制字符，有控制字符，则认为是bytes
		if !isString(data) {
			typeName := fmt.Sprintf(typeNamesFormat[Bytes], tag)
			result.Append(typeName, hex.EncodeToString(data))
			return nil
		}
		// 字符串类型，直接赋值
		typeName := fmt.Sprintf(typeNamesFormat[String], tag)
		result.Append(typeName, string(data))
	}
	return nil
}

// readPacked 解析packed类型
// raw: 要反序列化的PB数据
// tag: 要反序列化的字段的tag
// typ: 用户干预反序列化的选择
// result: 反序列化的结果
func readPacked(data []byte, tag uint64, typ Type,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readPacked] %w", err)
		}
	}()

	// 根据类型进行解析
	switch typ {
	case Packed + Int32:
		err = readInt32Packed(data, tag, result)
	case Packed + Int64:
		err = readInt64Packed(data, tag, result)
	case Packed + UInt:
		err = readUIntPacked(data, tag, result)
	case Packed + SInt:
		err = readSIntPacked(data, tag, result)
	case Packed + Bool:
		err = readBoolPacked(data, tag, result)
	case Packed + Fixed32:
		err = readFixed32Packed(data, tag, result)
	case Packed + Float:
		err = readFloatPacked(data, tag, result)
	case Packed + SFixed32:
		err = readSFixed32Packed(data, tag, result)
	case Packed + Fixed64:
		err = readFixed64Packed(data, tag, result)
	case Packed + Double:
		err = readDoublePacked(data, tag, result)
	case Packed + SFixed64:
		err = readSFixed64Packed(data, tag, result)
	default:
		return errUnknownType
	}
	return err
}

// readSFixed64Packed 解析Packed SFixed64类型
func readSFixed64Packed(data []byte, tag uint64,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readSFixed64Packed] %w", err)
		}
	}()

	typeName := fmt.Sprintf(typeNamesFormat[Packed+SFixed64], tag)
	for len(data) > 0 {
		value, length := protowire.ConsumeFixed64(data)
		if length < 0 {
			return protowire.ParseError(length)
		}
		data = data[length:]
		// 采用字符串，防止溢出
		result.Append(typeName, strconv.FormatInt(int64(value), 10))
	}
	return nil
}

// readDoublePacked 解析Packed Double类型
func readDoublePacked(data []byte, tag uint64,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readDoublePacked] %w", err)
		}
	}()

	typeName := fmt.Sprintf(typeNamesFormat[Packed+Double], tag)
	for len(data) > 0 {
		value, length := protowire.ConsumeFixed64(data)
		if length < 0 {
			return protowire.ParseError(length)
		}
		data = data[length:]
		result.Append(typeName, math.Float64frombits(value))
	}
	return nil
}

// readFixed64Packed 解析Packed Fixed64类型
func readFixed64Packed(data []byte, tag uint64,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readFixed64Packed] %w", err)
		}
	}()

	typeName := fmt.Sprintf(typeNamesFormat[Packed+Fixed64], tag)
	for len(data) > 0 {
		value, length := protowire.ConsumeFixed64(data)
		if length < 0 {
			return protowire.ParseError(length)
		}
		data = data[length:]
		// 采用字符串，防止溢出
		result.Append(typeName, strconv.FormatUint(value, 10))
	}
	return nil
}

// readSFixed32Packed 解析Packed SFixed32类型
func readSFixed32Packed(data []byte, tag uint64,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readSFixed32Packed] %w", err)
		}
	}()

	typeName := fmt.Sprintf(typeNamesFormat[Packed+SFixed32], tag)
	for len(data) > 0 {
		value, length := protowire.ConsumeFixed32(data)
		if length < 0 {
			return protowire.ParseError(length)
		}
		data = data[length:]
		result.Append(typeName, int32(value))
	}
	return nil
}

// readFloatPacked 解析Packed Float类型
func readFloatPacked(data []byte, tag uint64,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readFloatPacked] %w", err)
		}
	}()

	typeName := fmt.Sprintf(typeNamesFormat[Packed+Float], tag)
	for len(data) > 0 {
		value, length := protowire.ConsumeFixed32(data)
		if length < 0 {
			return protowire.ParseError(length)
		}
		data = data[length:]
		result.Append(typeName, math.Float32frombits(value))
	}
	return nil
}

// readFixed32Packed 解析Packed Fixed32类型
func readFixed32Packed(data []byte, tag uint64,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readFixed32Packed] %w", err)
		}
	}()

	typeName := fmt.Sprintf(typeNamesFormat[Packed+Fixed32], tag)
	for len(data) > 0 {
		value, length := protowire.ConsumeFixed32(data)
		if length < 0 {
			return protowire.ParseError(length)
		}
		data = data[length:]
		result.Append(typeName, uint32(value))
	}
	return nil
}

// readBoolPacked 解析Packed Bool类型
func readBoolPacked(data []byte, tag uint64,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readBoolPacked] %w", err)
		}
	}()

	typeName := fmt.Sprintf(typeNamesFormat[Packed+Bool], tag)
	for len(data) > 0 {
		value, length := protowire.ConsumeVarint(data)
		if length < 0 {
			return protowire.ParseError(length)
		}
		data = data[length:]

		if value == 0 {
			result.AppendArrayItem(typeName, false)
			continue
		}
		result.AppendArrayItem(typeName, true)
	}
	return nil
}

// readSIntPacked 解析Packed SInt类型
func readSIntPacked(data []byte, tag uint64,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readSIntPacked] %w", err)
		}
	}()

	typeName := fmt.Sprintf(typeNamesFormat[Packed+SInt], tag)
	for len(data) > 0 {
		value, length := protowire.ConsumeVarint(data)
		if length < 0 {
			return protowire.ParseError(length)
		}
		data = data[length:]
		result.AppendArrayItem(typeName, protowire.DecodeZigZag(value))
	}
	return nil
}

// readUIntPacked 解析Packed UInt类型
func readUIntPacked(data []byte, tag uint64,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readUIntPacked] %w", err)
		}
	}()

	typeName := fmt.Sprintf(typeNamesFormat[Packed+UInt], tag)
	for len(data) > 0 {
		value, length := protowire.ConsumeVarint(data)
		if length < 0 {
			return protowire.ParseError(length)
		}
		data = data[length:]
		result.AppendArrayItem(typeName, uint64(value))
	}
	return nil
}

// readInt64Packed 解析Packed Int64类型
func readInt64Packed(data []byte, tag uint64,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readInt64Packed] %w", err)
		}
	}()

	typeName := fmt.Sprintf(typeNamesFormat[Packed+Int64], tag)
	for len(data) > 0 {
		value, length := protowire.ConsumeVarint(data)
		if length < 0 {
			return protowire.ParseError(length)
		}
		data = data[length:]
		result.AppendArrayItem(typeName, int64(value))
	}
	return nil
}

// readInt32Packed 解析Packed Int32类型
func readInt32Packed(data []byte, tag uint64,
	result JSONResult) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[readInt32Packed] %w", err)
		}
	}()

	typeName := fmt.Sprintf(typeNamesFormat[Packed+Int32], tag)
	for len(data) > 0 {
		value, length := protowire.ConsumeVarint(data)
		if length < 0 {
			return protowire.ParseError(length)
		}
		data = data[length:]
		result.AppendArrayItem(typeName, int32(value))
	}
	return nil
}

// readFixed32 解析fixed32类型
// raw: 要反序列化的PB数据
// tag: 要反序列化的字段的tag
// opts: 用户干预反序列化的选择
// result: 反序列化的结果
func readFixed32(raw []byte, tag uint64, opts Options,
	result JSONResult) ([]byte, error) {
	value, length := protowire.ConsumeFixed32(raw)
	if length < 0 {
		return raw, protowire.ParseError(length)
	}
	raw = raw[length:]

	// 根据用户选择进行类型转换，默认Float类型
	typ := opts.GetTypeByTag(strconv.FormatUint(tag, 10))
	typeName := fmt.Sprintf(typeNamesFormat[typ], tag)
	switch typ {
	case Float:
		result.Append(typeName, math.Float32frombits(value))
	case SFixed32:
		result.Append(typeName, int32(value))
	case Fixed32:
		result.Append(typeName, uint32(value))
	default:
		typeName = fmt.Sprintf(typeNamesFormat[Float], tag)
		result.Append(typeName, math.Float32frombits(value))
	}
	return raw, nil
}

// readFixed64 解析fix32类型，默认认为是float64
func readFixed64(raw []byte, tag uint64, opts Options,
	result JSONResult) ([]byte, error) {
	value, length := protowire.ConsumeFixed64(raw)
	if length < 0 {
		return raw, protowire.ParseError(length)
	}
	raw = raw[length:]

	// 根据用户选择进行类型转换，默认Fixed64类型
	typ := opts.GetTypeByTag(strconv.FormatUint(tag, 10))
	typeName := fmt.Sprintf(typeNamesFormat[typ], tag)
	switch typ {
	case Double:
		result.Append(typeName, math.Float64frombits(value))
	case SFixed64:
		// 采用字符串，防止溢出
		result.Append(typeName, strconv.FormatInt(int64(value), 10))
	case Fixed64:
		// 采用字符串，防止溢出
		result.Append(typeName, strconv.FormatUint(value, 10))
	default:
		typeName := fmt.Sprintf(typeNamesFormat[Double], tag)
		result.Append(typeName, math.Float64frombits(value))
	}
	return raw, nil
}

// JSONResult Json结果
type JSONResult map[string]interface{}

// Append 往结果中添加数据，遇到相同的键则变为数组
func (j JSONResult) Append(key string, value interface{}) {
	if temp, ok := j[key]; ok {
		var nvalue []interface{}
		if nvalue, ok = temp.([]interface{}); ok {
			// 已经有数组值，添加
			nvalue = append(nvalue, value)
		} else {
			// 已经有非数组值，创建数组添加
			nvalue = []interface{}{temp, value}
		}
		j[key] = nvalue
		return
	}
	j[key] = value
}

// AppendArrayItem 往结果中对应键的数组中添加元素
func (j JSONResult) AppendArrayItem(key string, value interface{}) {
	if temp, ok := j[key]; ok {
		var nvalue []interface{}
		if nvalue, ok = temp.([]interface{}); ok {
			// 已经有数组值，添加
			nvalue = append(nvalue, value)
		} else {
			// 已经有非数组值，创建数组添加
			nvalue = []interface{}{temp, value}
		}
		j[key] = nvalue
		return
	}

	// 还没有值，添加数组值
	j[key] = []interface{}{value}
}

// FixTagTypeNames 修复解析结果中的TagType名称
func (j JSONResult) FixTagTypeNames() {
	// 数据类型结果后面加上s，如string数据的类型变为strings
	for k, v := range j {
		// 递归调用
		if nj, ok := v.(JSONResult); ok {
			nj.FixTagTypeNames()
		}
		if data, ok := v.([]interface{}); ok {
			delete(j, k)
			j[k+"s"] = data
		}
	}
}
