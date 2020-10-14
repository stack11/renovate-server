package util

func GetOrDefault(m map[string]string, key, def string) string {
	if len(m) == 0 {
		return def
	}

	if v, ok := m[key]; ok {
		return v
	}

	return def
}
