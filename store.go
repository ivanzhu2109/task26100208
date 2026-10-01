package entdup

// Enterprise 是库中保存的一家企业。
type Enterprise struct {
	// Raw 是导入时的原始名称，用于结果识别。
	Raw string
	// Norm 是 Normalize(Raw) 的结果，所有匹配均使用它。
	Norm string
}

// ImportResult 是批量导入中一行的判定结果。
type ImportResult struct {
	// Raw 是该行的原始名称。
	Raw string
	// Accepted 为 true 表示该行被接纳并加入企业库。
	Accepted bool
	// Hits 是被拒绝时命中的全部疑似重复企业（按入库顺序）。
	Hits []Enterprise
}

// Store 是进程内存中的企业名称查重库，供 CSV 导入流程直接调用。
type Store struct {
	ents []Enterprise
}

// NewStore 创建一个空的企业库。
func NewStore() *Store {
	return &Store{}
}

// Len 返回库中企业数量。
func (s *Store) Len() int {
	return len(s.ents)
}

// hits 返回库内所有与 norm 疑似重复的企业，按入库顺序。
func (s *Store) hits(norm string) []Enterprise {
	var out []Enterprise
	for _, e := range s.ents {
		l := len([]rune(e.Norm))
		if n := len([]rune(norm)); n > l {
			l = n
		}
		if IsDuplicate(Distance(e.Norm, norm), l) {
			out = append(out, e)
		}
	}
	return out
}

// Query 返回库内所有与 raw 疑似重复的企业，不修改企业库。
func (s *Store) Query(raw string) []Enterprise {
	return s.hits(Normalize(raw))
}

// Import 严格按输入顺序逐行判定：无命中的行被接纳并立即加入
// 企业库参与后续行的查重；有命中的行被拒绝，不进入企业库，
// 也不影响后续判定。疑似重复关系不做传递合并。
func (s *Store) Import(raws []string) []ImportResult {
	results := make([]ImportResult, len(raws))
	for i, raw := range raws {
		norm := Normalize(raw)
		hits := s.hits(norm)
		if len(hits) > 0 {
			results[i] = ImportResult{Raw: raw, Accepted: false, Hits: hits}
			continue
		}
		s.ents = append(s.ents, Enterprise{Raw: raw, Norm: norm})
		results[i] = ImportResult{Raw: raw, Accepted: true}
	}
	return results
}
