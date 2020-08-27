package tools

func Unique(ids []string, AllowEmpty bool) (res []string) {
	res = make([]string, 0, len(ids))
	mm := map[string]bool{}
	for _, id := range ids {
		_, exist := mm[id]
		if exist || (!AllowEmpty && id == "") {
			continue
		}
		res = append(res, id)
		mm[id] = true
	}
	return
}
