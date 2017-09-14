package tools

func contains(list []interface{}, value interface{}) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
