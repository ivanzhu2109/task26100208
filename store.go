package companydedup

// Company 表示库中的一家企业，保留原始名称与规范化名称。
type Company struct {
	Raw        string
	Normalized string
}

// ImportResult 表示批量导入中单行的判定结果。
type ImportResult struct {
	Raw      string
	Accepted bool
	Matches  []Company
}

// Store 是进程内存中的企业名称查重库。
type Store struct {
	companies []Company
}

// NewStore 创建空的企业库。
func NewStore() *Store {
	return &Store{}
}

// Query 查询库内所有与给定名称疑似重复的企业，不修改库。
func (s *Store) Query(name string) []Company {
	n := normalize(name)
	var matches []Company
	for _, c := range s.companies {
		if isDuplicate(c.Normalized, n) {
			matches = append(matches, c)
		}
	}
	return matches
}

// Import 严格按输入顺序逐行判定：无命中的行被接纳并加入库，
// 有命中的行被拒绝并返回当时全部命中的企业。
// 同批已接纳的行立即参与后续行的查重；被拒绝的行不入库，
// 也不影响后续判定。
func (s *Store) Import(names []string) []ImportResult {
	results := make([]ImportResult, 0, len(names))
	for _, raw := range names {
		n := normalize(raw)
		var matches []Company
		for _, c := range s.companies {
			if isDuplicate(c.Normalized, n) {
				matches = append(matches, c)
			}
		}
		if len(matches) == 0 {
			s.companies = append(s.companies, Company{Raw: raw, Normalized: n})
			results = append(results, ImportResult{Raw: raw, Accepted: true})
		} else {
			results = append(results, ImportResult{Raw: raw, Accepted: false, Matches: matches})
		}
	}
	return results
}
