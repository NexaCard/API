package upstream

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

const (
	// HeaderApiKey API Key header
	HeaderApiKey = "NexaCard-Api-Key"
	// HeaderTimestamp 时间戳 header
	HeaderTimestamp = "NexaCard-Timestamp"
	// HeaderSignature 签名 header
	HeaderSignature = "NexaCard-Signature"

	LegacyHeaderApiKey    = "Dujiao-Next-Api-Key"
	LegacyHeaderTimestamp = "Dujiao-Next-Timestamp"
	LegacyHeaderSignature = "Dujiao-Next-Signature"

	ChannelHeaderKey       = "NexaCard-Channel-Key"
	ChannelHeaderTimestamp = "NexaCard-Channel-Timestamp"
	ChannelHeaderSignature = "NexaCard-Channel-Signature"

	LegacyChannelHeaderKey       = "Dujiao-Next-Channel-Key"
	LegacyChannelHeaderTimestamp = "Dujiao-Next-Channel-Timestamp"
	LegacyChannelHeaderSignature = "Dujiao-Next-Channel-Signature"

	// MaxTimestampSkew 最大时间戳偏差（秒）
	MaxTimestampSkew = 60
)

func AuthHeaders(header http.Header) (apiKey, timestamp, signature string) {
	return headerValue(header, HeaderApiKey, LegacyHeaderApiKey),
		headerValue(header, HeaderTimestamp, LegacyHeaderTimestamp),
		headerValue(header, HeaderSignature, LegacyHeaderSignature)
}

func ChannelAuthHeaders(header http.Header) (channelKey, timestamp, signature string) {
	return headerValue(header, ChannelHeaderKey, LegacyChannelHeaderKey),
		headerValue(header, ChannelHeaderTimestamp, LegacyChannelHeaderTimestamp),
		headerValue(header, ChannelHeaderSignature, LegacyChannelHeaderSignature)
}

func headerValue(header http.Header, names ...string) string {
	for _, name := range names {
		if value := header.Get(name); value != "" {
			return value
		}
	}
	return ""
}

// Sign 生成 HMAC-SHA256 签名
// signString = "{method}\n{path}\n{timestamp}\n{body_md5}"
func Sign(secret, method, path string, timestamp int64, body []byte) string {
	bodyMD5 := md5Hex(body)
	signString := fmt.Sprintf("%s\n%s\n%d\n%s", method, path, timestamp, bodyMD5)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signString))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify 验证签名
func Verify(secret, method, path, signature string, timestamp int64, body []byte) bool {
	expected := Sign(secret, method, path, timestamp, body)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// IsTimestampValid 检查时间戳是否在有效范围内
func IsTimestampValid(timestamp int64) bool {
	now := time.Now().Unix()
	return math.Abs(float64(now-timestamp)) <= MaxTimestampSkew
}

// ParseTimestamp 解析时间戳字符串
func ParseTimestamp(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func md5Hex(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
