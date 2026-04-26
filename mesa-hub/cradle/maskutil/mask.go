package maskutil

// MaskedValue 对敏感值做脱敏处理
func MaskedValue(val string) string {
	if len(val) <= 8 {
		return "****"
	}
	return val[:4] + "****" + val[len(val)-4:]
}
