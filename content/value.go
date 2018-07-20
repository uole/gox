package content

import (
	"encoding/json"
	"encoding/xml"
	"strconv"
)

type ByteValue []byte

func (value ByteValue) Json(v interface{}) error {
	return json.Unmarshal(value, v)
}

func (value ByteValue) Xml(v interface{}) error {
	return xml.Unmarshal(value, v)
}

func (value ByteValue) Int() (int, error) {
	return strconv.Atoi(value.String())
}

func (value ByteValue) MustInt() int {
	v, _ := value.Int()
	return v
}

func (value ByteValue) Int64() (int64, error) {
	return strconv.ParseInt(value.String(), 10, 64)
}

func (value ByteValue) MustInt64() int64 {
	v, _ := value.Int64()
	return v
}

func (value ByteValue) Float64() (float64, error) {
	return strconv.ParseFloat(value.String(), 64)
}

func (value ByteValue) MustFloat64() float64 {
	v, _ := value.Float64()
	return v
}

func (value ByteValue) Bool() (bool, error) {
	return strconv.ParseBool(value.String())
}

func (value ByteValue) MustBool() bool {
	v, _ := value.Bool()
	return v
}

func (value ByteValue) String() string {
	return string(value)
}
