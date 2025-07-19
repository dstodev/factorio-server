package internal

func SliceTo[T any](values ...any) []T {
	result := make([]T, 0, len(values))
	for _, v := range values {
		if casted, ok := v.(T); ok {
			result = append(result, casted)
		}
	}
	return result
}
