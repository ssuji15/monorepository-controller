package controller

import (
	"strings"
)

func FilterLabelsOrAnnotations(in map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		if !strings.HasPrefix(k, "kapp.k14s.io/") {
			out[k] = v
		}
	}
	return out
}
