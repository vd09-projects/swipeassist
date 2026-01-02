package utils

import (
	"cmp"
	"sort"
)

// NormalizeAndDedupe normalizes items, removes empties, deduplicates, and sorts.
func NormalizeAndDedupe[T any](
	in []T,
	normalize func(T) (T, bool), // returns (value, keep?)
	less func(a, b T) bool, // sorting comparator
	key func(T) string, // dedupe key
) []T {
	seen := make(map[string]struct{}, len(in))
	out := make([]T, 0, len(in))

	for _, v := range in {
		nv, ok := normalize(v)
		if !ok {
			continue
		}

		k := key(nv)
		if _, exists := seen[k]; exists {
			continue
		}

		seen[k] = struct{}{}
		out = append(out, nv)
	}

	sort.Slice(out, func(i, j int) bool {
		return less(out[i], out[j])
	})

	return out
}

// Identity returns the value itself.
func Identity[T any](v T) T { return v }

// Less returns true if a < b for ordered types.
func Less[T cmp.Ordered](a, b T) bool { return a < b }

func MergeUniqueAndSort[T any, K comparable](
	dst []T,
	src []T,
	key func(T) K,
	less func(a, b T) bool,
) []T {
	if len(src) == 0 {
		sort.Slice(dst, func(i, j int) bool { return less(dst[i], dst[j]) })
		return dst
	}

	seen := make(map[K]struct{}, len(dst)+len(src))
	for _, d := range dst {
		seen[key(d)] = struct{}{}
	}

	for _, s := range src {
		k := key(s)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		dst = append(dst, s)
	}

	sort.Slice(dst, func(i, j int) bool { return less(dst[i], dst[j]) })
	return dst
}
