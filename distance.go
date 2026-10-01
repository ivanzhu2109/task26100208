package entdup

// Distance 返回两个已规范化名称之间的编辑距离：
// 将一个名称变成另一个名称所需的最少单码点插入、删除、替换次数，
// 每次操作代价为 1。按码点（rune）计算，而非 UTF-8 字节。
func Distance(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	prev := make([]int, len(rb)+1)
	curr := make([]int, len(rb)+1)
	for j := 0; j <= len(rb); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			curr[j] = min3(
				prev[j]+1,      // 删除
				curr[j-1]+1,    // 插入
				prev[j-1]+cost, // 替换（相同则免费）
			)
		}
		prev, curr = curr, prev
	}
	return prev[len(rb)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// IsDuplicate 判定距离 d 与较长者码点数 L 是否达到疑似重复阈值，
// 精确等价于 20*d <= 3*L（即相似度 1-d/L >= 85%），不做四舍五入。
func IsDuplicate(d, l int) bool {
	if l <= 0 {
		return false
	}
	return 20*d <= 3*l
}
