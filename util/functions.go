package util

func PointerNotNil[T any](ptr *T) bool {
	return ptr != nil
}

func PointerNotNilIdx[T any](ptr *T, _ int) bool {
	return ptr != nil
}
