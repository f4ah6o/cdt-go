package extract

import "strings"

type Meta map[string]string

func ParseMeta(s string) Meta {
	m := Meta{}
	for _, field := range strings.Fields(s) {
		if strings.Contains(field, "=") {
			parts := strings.SplitN(field, "=", 2)
			m[parts[0]] = strings.Trim(parts[1], `"'`)
		} else {
			m[field] = "true"
		}
	}
	return m
}

func (m Meta) Bool(key string) bool {
	v, ok := m[key]
	return ok && (v == "true" || v == "1" || v == "yes")
}

func (m Meta) First(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(m[k]); v != "" {
			return v
		}
	}
	return ""
}
