package entdup

import (
	"strings"
	"testing"
)

func raws(ents []Enterprise) []string {
	out := make([]string, len(ents))
	for i, e := range ents {
		out[i] = e.Raw
	}
	return out
}

func TestSameNameVariantsAreDuplicates(t *testing.T) {
	s := NewStore()
	res := s.Import([]string{"ABC科技（北京）有限公司"})
	if !res[0].Accepted {
		t.Fatal("首行应被接纳")
	}
	variants := []string{
		"ａｂｃ科技(北京)有限公司",   // 全角字母 + 半角括号
		"abc-科技·北京 有限公司",     // 小写 + 分隔符
		"ＡＢＣ 科技 （北京） 有限公司", // 全角 + 空格
	}
	for _, v := range variants {
		hits := s.Query(v)
		if len(hits) != 1 || hits[0].Raw != "ABC科技（北京）有限公司" {
			t.Errorf("Query(%q) 应命中且仅命中原始名称，实际 %v", v, raws(hits))
		}
	}
	if s.Len() != 1 {
		t.Errorf("查询不应修改企业库，Len = %d", s.Len())
	}
}

func TestThresholdExact85AndJustBelow(t *testing.T) {
	s := NewStore()
	a := strings.Repeat("a", 20)
	s.Import([]string{a})

	// 恰好 85%：d=3, L=20，20*3 == 3*20，必须命中
	exact := "bbb" + strings.Repeat("a", 17)
	if hits := s.Query(exact); len(hits) != 1 {
		t.Errorf("恰好 85%% 应命中，实际 %v", raws(hits))
	}
	// 略低于 85%：d=4, L=20，20*4 > 3*20，必须不命中
	below := "bbbb" + strings.Repeat("a", 16)
	if hits := s.Query(below); len(hits) != 0 {
		t.Errorf("略低于 85%% 不应命中，实际 %v", raws(hits))
	}
}

func TestNonTransitiveImportAndQuery(t *testing.T) {
	// A~B、B~C 均达 85%，A~C 未达 85%。
	a := strings.Repeat("a", 20)
	b := "bbb" + strings.Repeat("a", 17)
	c := "bbbccc" + strings.Repeat("a", 14)

	s := NewStore()
	s.Import([]string{a})

	res := s.Import([]string{b, c})
	if res[0].Accepted {
		t.Error("B 与 A 疑似重复，应被拒绝")
	}
	if got := raws(res[0].Hits); len(got) != 1 || got[0] != a {
		t.Errorf("B 的命中应为 [%q]，实际 %v", a, got)
	}
	if !res[1].Accepted {
		t.Error("C 与 A 未达阈值，且 B 已被拒绝不入库，C 应被接纳")
	}

	// 查询 B 必须同时命中 A 和 C（按入库顺序）。
	hits := s.Query(b)
	want := []string{a, c}
	got := raws(hits)
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("Query(B) 应命中 %v，实际 %v", want, got)
	}

	// D 仅与被拒绝的 B 相似，与 A、C 均未达阈值，必须被接纳：
	// 被拒绝的行不能通过自身影响后续判定。
	d := "bbbb" + strings.Repeat("a", 15) + "b"
	res2 := s.Import([]string{d})
	if !res2[0].Accepted {
		t.Errorf("D 仅与被拒绝的 B 相似，应被接纳，命中 %v", raws(res2[0].Hits))
	}
}

func TestImportAcceptedRowsJoinImmediately(t *testing.T) {
	s := NewStore()
	res := s.Import([]string{"阿尔法科技有限公司", "阿尔法科技有限公司"})
	if !res[0].Accepted {
		t.Fatal("首行应被接纳")
	}
	if res[1].Accepted {
		t.Error("同批后续相同名称应立即命中已接纳行并被拒绝")
	}
	if got := raws(res[1].Hits); len(got) != 1 || got[0] != "阿尔法科技有限公司" {
		t.Errorf("命中应保留原始名称，实际 %v", got)
	}
	if s.Len() != 1 {
		t.Errorf("被拒绝的行不应进入企业库，Len = %d", s.Len())
	}
}
