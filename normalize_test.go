package entdup

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"ASCII 原样", "ABC科技", "abc科技"},
		{"大小写折叠", "AbC科技", "abc科技"},
		{"全角字母数字", "ＡＢＣ１２３", "abc123"},
		{"全角括号等分隔符删除", "ＡＢＣ科技（北京）有限公司", "abc科技北京有限公司"},
		{"半角分隔符删除", "A-B C.D_科技", "abcd科技"},
		{"完整大小写折叠 ß", "Straße", "strasse"},
		{"数字保留", "公司2024号", "公司2024号"},
		{"空白与标点删除", " 公司 ，。 ", "公司"},
		{"空串", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Normalize(c.in); got != c.want {
				t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestNormalizeSameNameVariants(t *testing.T) {
	variants := []string{
		"ABC科技(北京)有限公司",
		"ａｂｃ科技（北京）有限公司",
		"abc-科技·北京 有限公司",
		"ＡＢＣ 科技 （北京） 有限公司",
	}
	want := Normalize(variants[0])
	for _, v := range variants[1:] {
		if got := Normalize(v); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", v, got, want)
		}
	}
}
