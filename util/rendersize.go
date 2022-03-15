package util

import (
	"errors"
	"fmt"
	"sort"
)

// ErrNoSolution is returned by RenderSize when it can't find a solution.
var ErrNoSolution = fmt.Errorf("no solution")

// RenderSize will return the rendered result as close as possible to the max size.
func RenderSize(maxSize int, input string, renderFunc func(string) (string, error)) (string, error) {
	result, err := renderFunc(input)
	if err != nil {
		return "", err
	}
	if len(result) <= maxSize {
		return result, nil
	}

	newSize := sort.Search(len(input), func(n int) bool {
		if err != nil {
			return true
		}

		result, err = renderFunc(input[:n])
		return len(result) > maxSize
	})
	if err != nil {
		return "", err
	}
	if newSize == 0 {
		return "", fmt.Errorf("failed to render string to size %d: %w", maxSize, ErrNoSolution)
	}

	// newSize is the first size that is too large
	return renderFunc(input[:newSize-1])
}

// RenderSizeN works like RenderSize but accepts multiple inputs.
//
// Inputs are trimmed one by one starting with the fist element.
func RenderSizeN(maxSize int, inputs []string, renderFunc func([]string) (string, error)) (result string, err error) {
	if len(inputs) == 0 {
		return RenderSize(maxSize, "", func(string) (string, error) {
			return renderFunc(nil)
		})
	}

	for i := range inputs {
		result, err = RenderSize(maxSize, inputs[i], func(inputN string) (string, error) {
			inputs[i] = inputN
			return renderFunc(inputs)
		})
		if errors.Is(err, ErrNoSolution) {
			inputs[i] = ""
			continue
		}
		if err != nil {
			return "", err
		}
		return result, err
	}

	return "", ErrNoSolution
}
