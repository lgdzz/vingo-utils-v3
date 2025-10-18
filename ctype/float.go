// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/26
// 描述：
// *****************************************************************************

package ctype

import "strconv"

type Float float64

// Float Float -> float64
func (f Float) Float() float64 {
	return float64(f)
}

// String Float -> string
func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

// Int Float -> int (取整)
func (f Float) Int() int {
	return int(f)
}
