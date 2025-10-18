// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/26
// 描述：
// *****************************************************************************

package ctype

import "strconv"

type Int int

// Int Int -> int
func (i Int) Int() int {
	return int(i)
}

// String Int -> string
func (i Int) String() string {
	return strconv.Itoa(int(i))
}

// Float Int -> float64
func (i Int) Float() float64 {
	return float64(i)
}
