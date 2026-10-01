package entdup

import "testing"

func TestDistance(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "abcd", 1},
		{"abcd", "abc", 1},
		{"kitten", "sitting", 3},
		// 按码点而非 UTF-8 字节：每个汉字/全角字符算 1
		{"北京", "北京", 0},
		{"北京", "上海", 2},
		{"北京", "北京市", 1},
		{"科技", "科技股份", 2},
	}
	for _, c := range cases {
		if got := Distance(c.a, c.b); got != c.want {
			t.Errorf("Distance(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestIsDuplicateBoundary(t *testing.T) {
	// 阈值精确等价于 20*d <= 3*L，不做四舍五入。
	cases := []struct {
		d, l int
		want bool
	}{
		{0, 1, true},
		{3, 20, true},  // 恰好 85%：20*3 == 3*20
		{4, 20, false}, // 略低于 85%
		{1, 7, true},   // 20 <= 21
		{2, 7, false},  // 40 > 21
		{3, 19, false}, // 60 > 57，若四舍五入 84.2%->84% 仍否；确保是精确比较
		{1, 6, false},  // 20 > 18，精确比较不四舍五入
	}
	for _, c := range cases {
		if got := IsDuplicate(c.d, c.l); got != c.want {
			t.Errorf("IsDuplicate(d=%d, l=%d) = %v, want %v", c.d, c.l, got, c.want)
		}
	}
}
