// *****************************************************************************
// 作者: lgdz
// 创建时间: 2025/6/25
// 描述：身份证信息
// *****************************************************************************

package ctype

import (
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"strconv"
	"strings"
	"time"
)

const (
	GenderMale   = "男"
	GenderFemale = "女"
)

type IdCard string

// 身份证号系数
var idCardFactors = []uint{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}

// 身份证号效验码
var idCardCodes = map[uint]string{0: "1", 1: "0", 2: "X", 3: "9", 4: "8", 5: "7", 6: "6", 7: "5", 8: "4", 9: "3", 10: "2"}

// IdCardInfo 身份证信息
type IdCardInfo struct {
	IdCard     string // 身份证号码
	RegionCode string // 6位行政区域编码
	Birthday   string // 2006-01-02 格式日期
	Age        int    // 年龄：精确到月份
	UniformAge int    // 年龄：按年份计算
	Gender     string // 性别
	GenderInt  int    // 性别：1-男，2-女
}

// Analysis 解析身份证信息
func (s IdCard) Analysis() IdCardInfo {
	id := string(s)
	if !s.IsValid() {
		panic("身份证号不正确")
	}
	now := time.Now()
	thisMonth, _ := convertor.ToInt(now.Format("01"))
	info := IdCardInfo{IdCard: id, RegionCode: id[:6]}
	year, _ := convertor.ToInt(id[6:10])
	month, _ := convertor.ToInt(id[10:12])
	day := id[12:14]
	info.Birthday = fmt.Sprintf("%v-%v-%v", year, month, day)
	info.UniformAge = now.Year() - int(year)
	if int(thisMonth) < int(month) {
		info.Age = info.UniformAge - 1
	} else {
		info.Age = info.UniformAge
	}
	if i, _ := strconv.Atoi(string(id[16])); i%2 == 0 {
		info.GenderInt = 2
		info.Gender = GenderFemale
	} else {
		info.GenderInt = 1
		info.Gender = GenderMale
	}
	return info
}

// IsValid 验证身份证号是否正确（校验码验证）
// 身份证号码的最后一位校验码是根据前面的17位数字计算出来的。计算步骤如下：
// 1．将身份证号码的前17位数字分别乘以对应的系数。系数从左到右依次是：{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
// 2．将上述乘积求和
// 3. 将求和结果除以11，取余数
// 4. 根据余数，从对应表中找出校验码：{0: "1", 1: "0", 2: "X", 3: "9", 4: "8", 5: "7", 6: "6", 7: "5", 8: "4", 9: "3", 10: "2"}
func (s IdCard) IsValid() bool {
	id := string(s)
	if len(id) != 18 {
		return false
	}
	strSlice := strings.Split(id, "")
	var sum uint
	for i, factor := range idCardFactors {
		v, _ := convertor.ToInt(strSlice[i])
		sum += uint(v) * factor
	}
	// 余数
	code := sum % 11
	if last, ok := idCardCodes[code]; ok {
		// 余数取到的效验码等于身份证最后一位则为有效
		return last == strings.ToUpper(strSlice[17])
	}
	return false
}
