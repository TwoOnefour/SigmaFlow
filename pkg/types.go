package pkg

import (
	"strconv"
	"strings"
)

type TextFloat64 float64

func (tf *TextFloat64) UnmarshalJSON(data []byte) error {

	str := string(data)

	str = strings.Trim(str, "\"")

	if str == "" || str == "null" {
		*tf = 0
		return nil
	}

	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}

	*tf = TextFloat64(f)
	return nil
}

func (tf TextFloat64) Float64() float64 {
	return float64(tf)
}
