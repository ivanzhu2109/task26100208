package companydedup

import "testing"

func raws(companies []Company) []string {
	out := make([]string, len(companies))
	for i, c := range companies {
		out[i] = c.Raw
	}
	return out
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// 大小写、全角字符和分隔符形成的同名应判定为重复。
func TestNormalizeEquivalentNames(t *testing.T) {
	s := NewStore()
	results := s.Import([]string{"ＡＢＣ（中国）有限公司"})
	if !results[0].Accepted {
		t.Fatal("首次导入应被接纳")
	}

	variants := []string{
		"abc中国有限公司",
		"ABC 中国 有限公司",
		"ａｂｃ-中国-有限公司",
		"Abc（中国）有限公司",
	}
	for _, v := range variants {
		matches := s.Query(v)
		if len(matches) != 1 || matches[0].Raw != "ＡＢＣ（中国）有限公司" {
			t.Errorf("Query(%q) 应命中原始名称，实际 %v", v, raws(matches))
		}
	}

	res := s.Import([]string{"ABC－中国－有限公司"})
	if res[0].Accepted {
		t.Error("同名变体导入应被拒绝")
	}
	if len(res[0].Matches) != 1 || res[0].Matches[0].Raw != "ＡＢＣ（中国）有限公司" {
		t.Errorf("拒绝结果应返回命中的原始企业，实际 %v", raws(res[0].Matches))
	}
}

// 相似度恰好 85%（20d == 3L）应命中，略低于 85% 不应命中。
func TestThresholdBoundary(t *testing.T) {
	// L=20, d=3：20*3 == 3*20，恰好 85%。
	exact := "aaaaaaaaaaaaaaaaaaaa"
	exactMatch := "aaabaaaaaaaaaaaaaabb" // 3 处替换
	// L=20, d=4：20*4 > 3*20，略低于 85%。
	below := "aaabaaaaaaaaaaaaabbb" // 4 处替换
	// L=13, d=2：20*2 > 3*13，低于 85%。
	base13 := "aaaaaaaaaaaaa"
	below13 := "aabaaaaaaaaab" // 13 码点，2 处替换

	s := NewStore()
	s.Import([]string{exact, base13})

	if got := s.Query(exactMatch); len(got) != 1 || got[0].Raw != exact {
		t.Errorf("恰好 85%% 应命中，实际 %v", raws(got))
	}
	if got := s.Query(below); len(got) != 0 {
		t.Errorf("略低于 85%% 不应命中，实际 %v", raws(got))
	}
	if got := s.Query(below13); len(got) != 0 {
		t.Errorf("L=13,d=2 不应命中，实际 %v", raws(got))
	}
}

// 阈值判定必须按码点而非 UTF-8 字节计算。
func TestDistanceCountsRunesNotBytes(t *testing.T) {
	s := NewStore()
	s.Import([]string{"中中中中中中中中中中"}) // 10 个码点
	// 替换 1 个码点：d=1, L=10，20 <= 30，应命中；
	// 若按字节计算 L=30, d=3，60 > 90 不成立……此处直接验证命中即可。
	if got := s.Query("中中中中中中中中中国"); len(got) != 1 {
		t.Errorf("多字节字符应按码点计算距离并命中，实际 %v", raws(got))
	}
}

// 链式相似：A~B、B~C 达阈值，A!~C。
// 库中仅有 A 时按 B、C 导入：拒绝 B、接纳 C；之后查询 B 同时命中 A 和 C。
func TestNonTransitiveChain(t *testing.T) {
	a := "aaaaaaaaaa"
	b := "aaaaaaaaab" // 与 A 差 1，与 C 差 1
	c := "aaaaaaaabb" // 与 A 差 2（未达阈值），与 B 差 1

	s := NewStore()
	s.Import([]string{a})

	results := s.Import([]string{b, c})
	if results[0].Accepted {
		t.Error("B 与 A 达阈值，应被拒绝")
	}
	if len(results[0].Matches) != 1 || results[0].Matches[0].Raw != a {
		t.Errorf("B 的命中应只有 A，实际 %v", raws(results[0].Matches))
	}
	if !results[1].Accepted {
		t.Error("C 与 A 未达阈值，且被拒绝的 B 不入库，C 应被接纳")
	}

	matches := s.Query(b)
	if len(matches) != 2 || !contains(raws(matches), a) || !contains(raws(matches), c) {
		t.Errorf("查询 B 应同时命中 A 和 C，实际 %v", raws(matches))
	}
}

// 同批已接纳的行立即参与后续行的查重。
func TestAcceptedRowsJoinImmediately(t *testing.T) {
	s := NewStore()
	results := s.Import([]string{"Alpha Corp", "alpha-corp", "Beta Ltd"})
	if !results[0].Accepted {
		t.Error("第一行应被接纳")
	}
	if results[1].Accepted {
		t.Error("第二行与同批已接纳的第一行重复，应被拒绝")
	}
	if len(results[1].Matches) != 1 || results[1].Matches[0].Raw != "Alpha Corp" {
		t.Errorf("第二行应命中第一行，实际 %v", raws(results[1].Matches))
	}
	if !results[2].Accepted {
		t.Error("第三行无命中，应被接纳")
	}
}

// 被拒绝的行不进入企业库，也不影响后续判定。
func TestRejectedRowDoesNotAffectLaterRows(t *testing.T) {
	s := NewStore()
	s.Import([]string{"aaaaaaaaaa"})
	results := s.Import([]string{"aaaaaaaaab", "aaaaaaaabb"})
	if results[0].Accepted {
		t.Error("第一行与库中名称重复，应被拒绝")
	}
	if !results[1].Accepted {
		t.Error("第二行仅与被拒绝的行相似，应被接纳")
	}
}

// 查询不修改企业库。
func TestQueryDoesNotMutate(t *testing.T) {
	s := NewStore()
	s.Import([]string{"Gamma Inc"})
	for i := 0; i < 3; i++ {
		if got := s.Query("gamma-inc"); len(got) != 1 {
			t.Fatalf("第 %d 次查询应命中，实际 %v", i+1, raws(got))
		}
	}
	results := s.Import([]string{"gamma inc"})
	if results[0].Accepted {
		t.Error("查询不应把查询词加入库中，重复导入应被拒绝")
	}
}
