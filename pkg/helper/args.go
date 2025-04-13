package helper

func Pointer[T any](v T) *T {
	return &v
}

func ParseMCPArgs[T any](key string, mcpArgs map[string]any) (*T, error) {
	value, ok := mcpArgs[key].(T)
	if !ok {
		return nil, nil
	}
	return Pointer(value), nil
}
