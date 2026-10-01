package companydedup

// levenshtein 计算两个码点序列之间的编辑距离
// （单码点插入、删除、替换，每次操作代价为 1）。
func levenshtein(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min3(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func min3(x, y, z int) int {
	m := x
	if y < m {
		m = y
	}
	if z < m {
		m = z
	}
	return m
}

// isDuplicate 判定两个已规范化名称是否达到疑似重复阈值：
// 相似度 = 1 - d/L >= 85%，精确等价于 20*d <= 3*L，
// 其中 d 为码点级编辑距离，L 为较长名称的码点数。
func isDuplicate(a, b string) bool {
	ra, rb := []rune(a), []rune(b)
	longer := len(ra)
	if len(rb) > longer {
		longer = len(rb)
	}
	if longer == 0 {
		return true
	}
	d := levenshtein(ra, rb)
	return 20*d <= 3*longer
}
