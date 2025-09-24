package collector

import (
	"crypto/rand"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// WencaiTokenGenerator 问财token生成器
type WencaiTokenGenerator struct {
	tokenServerTime int64
	first           []interface{}
	second          []interface{}
	vn              int
	un              int
	s               *QnBuffer
}

// QnBuffer 缓冲区结构
type QnBuffer struct {
	baseFileds []int
	data       []int
}

// NewWencaiTokenGenerator 创建新的token生成器
func NewWencaiTokenGenerator() *WencaiTokenGenerator {
	gen := &WencaiTokenGenerator{
		tokenServerTime: time.Now().Unix(),
		vn:              0,
		un:              63,
	}

	// 初始化first和second数组（简化版本）
	gen.initArrays()
	gen.init()

	return gen
}

// initArrays 初始化数组数据
func (w *WencaiTokenGenerator) initArrays() {
	// 这里是简化版本，实际应该包含完整的数组数据
	w.first = []interface{}{
		"", 9527, "String", "Boolean", "eh", "ad", "Bu", "ileds", "1", "\b",
		"Array", "7", "base", "64De", "\u2543\u252b", "etatS", "pa", "e",
		"FromUrl", "getOrigi", "nFromUrl", "\u255b\u253e", "b?\x18q)", "ic",
		"k", "sted", "he", "wser", "oNo", "ckw", "ent", "hst", "^And", "RM",
		"systemL", 5, "\u255f\u0978\u095b\u09f5", "TR8", "!'", "gth", "er", "TP",
		83, "r", true, "v", "v-nixeh", "RegExp", "thsi.cn", "K\x19\"]K^xVV",
		"KXxAPD?\x1b[Y", "document", 0, "allow", 1, "; ", "length", "Init", "=",
	}

	w.second = []interface{}{
		1, "", 0, "he", "ad", 29, "\x180G\x1f", "?>=<;:\\\\/,+", "ng", "to",
		"ff", "Number", "Error", "11", "6", "er", "ro", "code", "co", "_?L",
		"ed", "@S\x15D*", "Object", "len", "gth", "on", "lo", "RegExp", "ySta",
		13, "eel", "ee", "ouse", "ll", "\u2544\u2530\u2555\u2531", "FCm-",
		"isTru", "getC", "Pos", "ve", "or", "ae", "^", "On", "Sho", "can",
		"ont", "roid", "anguage", "\u2502", "ta", "tna", "Date", "3", "am",
	}
}

// init 初始化缓冲区
func (w *WencaiTokenGenerator) init() {
	// 创建QnBuffer实例
	w.s = &QnBuffer{
		baseFileds: []int{63, 63, 63, 63, 0, 0, 0, 11, 1, 1, 1, 1, 1, 1, 1, 63, 1, 0},
		data:       make([]int, 18),
	}

	// 设置初始值
	w.s.data[0] = w.serverTimeNow()
	w.updateRandom()
	w.s.data[15] = w.vn
	w.s.data[14] = w.un
	w.s.data[13] = 0
	w.s.data[6] = w.strhash()
	w.s.data[10] = w.getBrowserFeature()
	w.s.data[12] = w.getPlatform()
	w.s.data[11] = w.getBrowserIndex()
	w.s.data[9] = w.getPluginNum()
}

// serverTimeNow 获取服务器时间
func (w *WencaiTokenGenerator) serverTimeNow() int {
	return int(w.tokenServerTime)
}

// updateRandom 更新随机数
func (w *WencaiTokenGenerator) updateRandom() {
	// 生成随机数
	b := make([]byte, 4)
	rand.Read(b)
	randomValue := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if randomValue < 0 {
		randomValue = -randomValue
	}
	w.s.data[1] = randomValue % 10000
}

// strhash 计算字符串哈希
func (w *WencaiTokenGenerator) strhash() int {
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"
	hash := 0
	for _, char := range userAgent {
		hash = (hash << 5) - hash + int(char)
		hash = int(uint32(hash)) // 模拟JavaScript的无符号右移
	}
	return hash
}

// getBrowserFeature 获取浏览器特征
func (w *WencaiTokenGenerator) getBrowserFeature() int {
	return 3812
}

// getPlatform 获取平台信息
func (w *WencaiTokenGenerator) getPlatform() int {
	return 7
}

// getBrowserIndex 获取浏览器索引
func (w *WencaiTokenGenerator) getBrowserIndex() int {
	return 10
}

// getPluginNum 获取插件数量
func (w *WencaiTokenGenerator) getPluginNum() int {
	return 5
}

// timeNow 获取当前时间戳
func (w *WencaiTokenGenerator) timeNow() int {
	return int(time.Now().UnixMilli() / 1000)
}

// toBuffer 转换为缓冲区
func (q *QnBuffer) toBuffer() []int {
	result := make([]int, 0)
	bitPos := -1

	for i := 0; i < len(q.baseFileds); i++ {
		value := q.data[i]
		bits := q.baseFileds[i]
		bitPos += bits

		for bits > 0 {
			result = append(result, value&255)
			bits--
			if bits > 0 {
				bitPos--
				value >>= 8
			}
		}
	}

	return result
}

// Base64编码相关函数
var base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// base64Encode Base64编码
func base64Encode(data []int) string {
	var result strings.Builder

	for i := 0; i < len(data); i += 3 {
		var chunk int

		// 第一个字节
		if i < len(data) {
			chunk |= data[i] << 16
		}
		// 第二个字节
		if i+1 < len(data) {
			chunk |= data[i+1] << 8
		}
		// 第三个字节
		if i+2 < len(data) {
			chunk |= data[i+2]
		}

		// 生成4个Base64字符
		result.WriteByte(base64Chars[(chunk>>18)&63])
		result.WriteByte(base64Chars[(chunk>>12)&63])
		result.WriteByte(base64Chars[(chunk>>6)&63])
		result.WriteByte(base64Chars[chunk&63])
	}

	return result.String()
}

// encode 编码函数
func encode(data []int) string {
	// 计算校验和
	checksum := calculateChecksum(data)

	// 创建编码数据
	encoded := []int{86, checksum} // 86是固定值

	// 加密数据
	encryptData(data, encoded, checksum)

	return base64Encode(encoded)
}

// calculateChecksum 计算校验和
func calculateChecksum(data []int) int {
	checksum := 0
	for _, value := range data {
		checksum = (checksum << 5) - checksum + value
	}
	return checksum & 0xFFFFFFFF
}

// encryptData 加密数据
func encryptData(source []int, dest []int, key int) {
	for i := 0; i < len(source); i++ {
		encrypted := source[i] ^ (key & 255)
		dest = append(dest, encrypted)
		key = ^(key * 131) // 简化的密钥更新
	}
}

// Update 更新并生成token
func (w *WencaiTokenGenerator) Update() string {
	// 更新计数器
	w.s.data[13]++

	// 更新时间戳
	w.s.data[0] = w.serverTimeNow()
	w.s.data[2] = w.timeNow()

	// 重置其他值
	w.s.data[15] = w.vn
	w.s.data[7] = 0
	w.s.data[8] = 0
	w.s.data[3] = 0
	w.s.data[4] = 0
	w.s.data[5] = 0
	w.s.data[16] = 0

	// 转换为缓冲区并编码
	buffer := w.s.toBuffer()
	return encode(buffer)
}

// GenerateToken 生成问财token的便捷方法
func GenerateWencaiToken() string {
	generator := NewWencaiTokenGenerator()
	return generator.Update()
}

// 辅助函数：安全的类型转换
func safeInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		return 0
	case float64:
		return int(val)
	default:
		return 0
	}
}

// 辅助函数：安全的字符串转换
func safeString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case float64:
		return fmt.Sprintf("%.0f", val)
	default:
		return ""
	}
}

// 数学辅助函数
func mathRandom() float64 {
	b := make([]byte, 8)
	rand.Read(b)
	// 转换为0-1之间的浮点数
	return float64(b[0]) / 255.0
}

// 时间戳辅助函数
func getTimestamp() int64 {
	return time.Now().UnixMilli()
}

// 模拟JavaScript的parseInt函数
func parseInt(s string, base int) int {
	if base == 0 {
		base = 10
	}
	if i, err := strconv.ParseInt(s, base, 64); err == nil {
		return int(i)
	}
	return 0
}

// 模拟JavaScript的Math.floor函数
func mathFloor(x float64) int {
	return int(math.Floor(x))
}
