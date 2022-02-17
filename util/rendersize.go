package util

import "sort"

// RenderSize will return the rendered result as close as possible to the max size.
func RenderSize(maxSize int, renderFunc func(n int) string) string {
	result := renderFunc(maxSize)
	if len(result) <= maxSize {
		return result
	}

	newSize := sort.Search(maxSize, func(max int) bool {
		return len(renderFunc(max)) > maxSize
	})
	if newSize == maxSize {
		return ""
	}

	// newSize is the first size that is too large
	return renderFunc(newSize - 1)
}
