package jce

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"pb_json/pb"
)

var (
	// errPBTagTooBig pb的tag值太大
	errPBTagTooBig = errors.New("pb's tag too big")
	// errUnknownType 未知的PB类型
	errUnknownType = errors.New("unknown type")
)

type jceImpl struct{}

func (j *jceImpl) Do(raw []byte, opts ...pb.Options) ([]byte, error) {
	result := pb.JSONResult{}
	raw, err := jceDecode(raw, result)
	if err != nil {
		return nil, err
	}
	if len(raw) != 0 {
		return nil, errInvalidData()
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

const (
	// Char char类型
	Char pb.Type = 0
	// Short short类型
	Short pb.Type = 1
	// Int int类型
	Int pb.Type = 2
	// Int64 int64类型
	Int64 pb.Type = 3
	// Float float类型
	Float pb.Type = 4
	// Double double类型
	Double pb.Type = 5
	// String1 string1类型
	String1 pb.Type = 6
	// String4 string4类型
	String4 pb.Type = 7
	// Map map类型
	Map pb.Type = 8
	// List list类型
	List pb.Type = 9
	// StructBegin struct_begin类型
	StructBegin pb.Type = 10
	// StructEnd struct_end类型
	StructEnd pb.Type = 11
	// Zero zero类型
	Zero pb.Type = 12
	// SimpleList simplelist类型
	SimpleList pb.Type = 13
	// EmptyMap 空map类型
	EmptyMap pb.Type = 17
	// EmptyList 空list类型
	EmptyList pb.Type = 18
	// EmptySimpleList 空simplelist类型
	EmptySimpleList pb.Type = 19

	// MaxFieldNum 一个结构体中字段的最大数量
	MaxFieldNum = 10000
)

var (
	// jceTypeNamesFormat 类型对应的名称
	jceTypeNamesFormat = map[pb.Type]string{
		Zero:            "%04d_zero",
		Char:            "%04d_char",
		Short:           "%04d_short",
		Int:             "%04d_int",
		Int64:           "%04d_int64",
		Float:           "%04d_float",
		Double:          "%04d_double",
		String1:         "%04d_string",
		String4:         "%04d_string",
		Map:             "%04d_map",
		List:            "%04d_list",
		SimpleList:      "%04d_simplelist",
		StructBegin:     "%04d_struct",
		EmptyMap:        "%04d_emptymap",
		EmptyList:       "%04d_emptylist",
		EmptySimpleList: "%04d_emptysimplelist",
	}

	// errInvalidData 数据为异常的jce数据
	errInvalidData = func() error { return fmt.Errorf("jce data invalid") }
)

// JCEFieldMeta 保存JCE字段序列化或者反序列化的元数据
type JCEFieldMeta struct {
	Tag  uint64  // 字段的tag值
	Type pb.Type // 字段的type值
}

// jceDecode 将JCE二进制数据反序列化为json数据格式的JSONResult
func jceDecode(raw []byte, result pb.JSONResult) ([]byte, error) {
	var (
		err error
		end bool
	)
	for len(raw) > 0 && !end {
		end, raw, err = readOneValue(raw, result)
		if err != nil {
			return nil, err
		}
	}
	return raw, nil
}

// jceReadTagType 从序列化后的二进制数据中读取tag和type，并且返回剩余的数据
func jceReadTagType(raw []byte) (tagType *JCEFieldMeta, rest []byte, err error) {
	len := len(raw)
	if len < 1 {
		return nil, nil, errInvalidData()
	}
	tagType = &JCEFieldMeta{
		Type: pb.Type(raw[0] & 0xF),
		Tag:  uint64(raw[0] >> 4),
	}
	if tagType.Tag < 15 {
		return tagType, raw[1:], nil
	}
	// 还需要一位用作tag
	if len < 2 {
		return nil, nil, errInvalidData()
	}
	tagType.Tag = uint64(raw[1])
	return tagType, raw[2:], nil
}

// readZero 读取zero类型
func readZero(tag uint64, result pb.JSONResult) {
	key := fmt.Sprintf(jceTypeNamesFormat[Zero], tag)
	result.Append(key, 0)
}

// readChar 读取char类型
func readChar(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	if len(raw) < 1 {
		return nil, errInvalidData()
	}
	key := fmt.Sprintf(jceTypeNamesFormat[Char], tag)
	result.Append(key, int(raw[0]))
	return raw[1:], nil
}

// readShort 读取short类型数据
func readShort(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	if len(raw) < 2 {
		return nil, errInvalidData()
	}
	key := fmt.Sprintf(jceTypeNamesFormat[Short], tag)
	result.Append(key, int(binary.BigEndian.Uint16(raw)))
	return raw[2:], nil
}

// readInt 读取int类型数据
func readInt(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	if len(raw) < 4 {
		return nil, errInvalidData()
	}
	key := fmt.Sprintf(jceTypeNamesFormat[Int], tag)
	result.Append(key, int(binary.BigEndian.Uint32(raw)))
	return raw[4:], nil
}

// readInt64 读取int64类型数据
func readInt64(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	if len(raw) < 8 {
		return nil, errInvalidData()
	}
	key := fmt.Sprintf(jceTypeNamesFormat[Int64], tag)
	result.Append(key, int64(binary.BigEndian.Uint64(raw)))
	return raw[8:], nil
}

// readFloat 读取float类型数据
func readFloat(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	if len(raw) < 4 {
		return nil, errInvalidData()
	}
	key := fmt.Sprintf(jceTypeNamesFormat[Float], tag)
	result.Append(key, math.Float32frombits(binary.BigEndian.Uint32(raw)))
	return raw[4:], nil
}

// readDouble 读取double类型数据
func readDouble(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	if len(raw) < 8 {
		return nil, errInvalidData()
	}
	key := fmt.Sprintf(jceTypeNamesFormat[Double], tag)
	result.Append(key, math.Float64frombits(binary.BigEndian.Uint64(raw)))
	return raw[8:], nil
}

// readString1 读取string1类型数据
func readString1(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	if len(raw) < 1 {
		return nil, errInvalidData()
	}
	length := int(raw[0])
	if len(raw) < length+1 {
		return nil, errInvalidData()
	}
	key := fmt.Sprintf(jceTypeNamesFormat[String1], tag)
	result.Append(key, string(raw[1:length+1]))
	return raw[length+1:], nil
}

// readString4 读取string4类型数据
func readString4(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	if len(raw) < 4 {
		return nil, errInvalidData()
	}
	length := int(binary.BigEndian.Uint32(raw))
	if len(raw) < length+4 {
		return nil, errInvalidData()
	}
	key := fmt.Sprintf(jceTypeNamesFormat[String4], tag)
	result.Append(key, string(raw[4:length+4]))
	return raw[length+4:], nil
}

// readStruct 读取结构体数据
func readStruct(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	newResult := pb.JSONResult{}
	raw, err := jceDecode(raw, newResult)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf(jceTypeNamesFormat[StructBegin], tag)
	result.Append(key, newResult)
	return raw, nil
}

// readLength 读取长度值
func readLength(raw []byte) (length int, rest []byte, err error) {
	// 读取tag和type
	var tagType *JCEFieldMeta
	tagType, raw, err = jceReadTagType(raw)
	if err != nil {
		return 0, nil, errInvalidData()
	}
	switch tagType.Type {
	case Zero:
	case Char:
		if len(raw) < 1 {
			err = errInvalidData()
			break
		}
		length = int(raw[0])
		raw = raw[1:]
	case Short:
		if len(raw) < 2 {
			err = errInvalidData()
			break
		}
		length = int(binary.BigEndian.Uint16(raw))
		raw = raw[2:]
	case Int:
		if len(raw) < 4 {
			err = errInvalidData()
			break
		}
		length = int(binary.BigEndian.Uint32(raw))
		raw = raw[4:]
	default:
		return 0, nil, errUnknownType
	}
	if err != nil {
		return 0, nil, err
	}
	return length, raw, nil
}

// readMap 读取map类型数据
func readMap(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	var length int
	var err error
	length, raw, err = readLength(raw)
	if err != nil {
		return nil, err
	}
	if length == 0 {
		key := fmt.Sprintf(jceTypeNamesFormat[EmptyMap], tag)
		result.Append(key, nil)
		return raw, nil
	}
	key := fmt.Sprintf(jceTypeNamesFormat[Map], tag)
	for i := 0; i < length; i++ {
		mapItem := pb.JSONResult{}
		// 读取map key
		raw, err = readMapKey(raw, mapItem)
		if err != nil {
			return nil, err
		}
		// 读取map value
		_, raw, err = readOneValue(raw, mapItem)
		if err != nil {
			return nil, err
		}
		result.AppendArrayItem(key, mapItem)
	}
	return raw, nil
}

// readMapKey 读取map的key值
func readMapKey(raw []byte, result pb.JSONResult) ([]byte, error) {
	tagType, raw, err := jceReadTagType(raw)
	if err != nil {
		return nil, err
	}
	switch tagType.Type {
	case Char:
		raw, err = readChar(raw, tagType.Tag, result)
	case Short:
		raw, err = readShort(raw, tagType.Tag, result)
	case Int:
		raw, err = readInt(raw, tagType.Tag, result)
	case Int64:
		raw, err = readInt64(raw, tagType.Tag, result)
	case Float:
		raw, err = readFloat(raw, tagType.Tag, result)
	case Double:
		raw, err = readDouble(raw, tagType.Tag, result)
	case String1:
		raw, err = readString1(raw, tagType.Tag, result)
	case String4:
		raw, err = readString4(raw, tagType.Tag, result)
	case StructBegin:
		raw, err = readStruct(raw, tagType.Tag, result)
	case StructEnd:
		return raw, nil
	default:
		return nil, errUnknownType
	}
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// readOneValue 读取map的value值
// raw: 要被处理的数据
// result: 结果
// return:
// end: 当前struct是否已经结束
// rest: 剩余为处理的数据
// err: 出错信息
func readOneValue(raw []byte, result pb.JSONResult) (end bool, rest []byte, err error) {
	// 读取tag和type
	tagType, raw, err := jceReadTagType(raw)
	if err != nil {
		return false, nil, err
	}
	switch tagType.Type {
	case Char:
		raw, err = readChar(raw, tagType.Tag, result)
	case Short:
		raw, err = readShort(raw, tagType.Tag, result)
	case Int:
		raw, err = readInt(raw, tagType.Tag, result)
	case Int64:
		raw, err = readInt64(raw, tagType.Tag, result)
	case Float:
		raw, err = readFloat(raw, tagType.Tag, result)
	case Double:
		raw, err = readDouble(raw, tagType.Tag, result)
	case String1:
		raw, err = readString1(raw, tagType.Tag, result)
	case String4:
		raw, err = readString4(raw, tagType.Tag, result)
	case Map:
		raw, err = readMap(raw, tagType.Tag, result)
	case List:
		raw, err = readList(raw, tagType.Tag, result)
	case StructBegin:
		raw, err = readStruct(raw, tagType.Tag, result)
	case StructEnd:
		return true, raw, nil
	case Zero:
		readZero(tagType.Tag, result)
	case SimpleList:
		raw, err = readSimpleList(raw, tagType.Tag, result)
	default:
		return false, nil, errUnknownType
	}
	if err != nil {
		return false, nil, err
	}
	return false, raw, nil
}

// readSimpleList 读取simplelist类型数据([]byte类型)
func readSimpleList(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	var err error
	// jce的simplelist当前仅支持[]byte类型
	_, raw, err = jceReadTagType(raw)
	if err != nil {
		return nil, err
	}
	var length int
	length, raw, err = readLength(raw)
	if err != nil {
		return nil, err
	}
	if length == 0 {
		key := fmt.Sprintf(jceTypeNamesFormat[EmptySimpleList], tag)
		result.Append(key, nil)
		return raw, nil
	}
	simpleList := make([]int, 0, length)
	for _, b := range raw[:length] {
		simpleList = append(simpleList, int(b))
	}
	key := fmt.Sprintf(jceTypeNamesFormat[SimpleList], tag)
	result.Append(key, simpleList)
	return raw[length:], nil
}

// readList 读取lsit类型数据
func readList(raw []byte, tag uint64, result pb.JSONResult) ([]byte, error) {
	length, raw, err := readLength(raw)
	if err != nil {
		return nil, err
	}
	if length == 0 {
		key := fmt.Sprintf(jceTypeNamesFormat[EmptyList], tag)
		result.Append(key, nil)
		return raw, nil
	}
	key := fmt.Sprintf(jceTypeNamesFormat[List], tag)
	for i := 0; i < length; i++ {
		listItem := pb.JSONResult{}
		_, raw, err = readOneValue(raw, listItem)
		if err != nil {
			return nil, err
		}
		result.AppendArrayItem(key, listItem)
	}
	return raw, nil
}
