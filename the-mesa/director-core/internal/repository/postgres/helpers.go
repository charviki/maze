package postgres

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
