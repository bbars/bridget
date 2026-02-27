package handler

import (
	"sort"
	"strconv"
	"strings"
)

type acceptType struct {
	mime string
	q    float64
}

func matchContentType(acceptHeader string, supported ...string) string {
	if acceptHeader == "" {
		if len(supported) > 0 {
			return supported[0]
		} else {
			return ""
		}
	}

	var accepted []acceptType
	for _, part := range strings.Split(acceptHeader, ",") {
		if part == "" {
			continue
		}

		pair := strings.Split(strings.TrimSpace(part), ";")
		t := acceptType{mime: pair[0], q: 1.0}

		for _, param := range pair[1:] {
			if s, found := strings.CutPrefix(strings.TrimSpace(param), "q="); found {
				t.q, _ = strconv.ParseFloat(s, 64)
			}
		}
		accepted = append(accepted, t)
	}

	sort.SliceStable(accepted, func(i, j int) bool {
		return accepted[i].q > accepted[j].q
	})

	if len(supported) == 0 {
		if len(accepted) > 0 {
			return accepted[0].mime
		} else {
			return ""
		}
	}

	for _, p := range accepted {
		for _, a := range supported {
			if p.mime == a || p.mime == "*/*" {
				return a
			}
		}
	}

	return ""
}
