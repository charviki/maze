package maskutil

import "testing"

func TestMaskedValue_Short(t *testing.T) {
	// 长度 <= 8 的字符串应返回 "****"
	result := MaskedValue("short")
	if result != "****" {
		t.Errorf("期望 ****, 实际=%s", result)
	}
}

func TestMaskedValue_Long(t *testing.T) {
	// 长度 > 8 的字符串应返回 前4位 + "****" + 后4位
	result := MaskedValue("abcdefghijklmnop")
	if result != "abcd****mnop" {
		t.Errorf("期望 abcd****mnop, 实际=%s", result)
	}
}

func TestMaskedValue_Empty(t *testing.T) {
	// 空字符串应返回 "****"
	result := MaskedValue("")
	if result != "****" {
		t.Errorf("期望 ****, 实际=%s", result)
	}
}
