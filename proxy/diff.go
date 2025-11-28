package main

import (
	"bytes"
	"encoding/json"
	"strings"
)

// Keys whose values we ignore entirely (removed from the structure)
var ignoreKeys = map[string]struct{}{
	"timestamp":  {},
	"trace_id":   {},
	"traceId":    {},
	"request_id": {},
	"requestId":  {},
}

// Keys we treat as PII and mask.
var piiKeys = map[string]struct{}{
	"email":    {},
	"phone":    {},
	"name":     {},
	"cc_last4": {},
}

const piiMask = "***"

// equalJSONBodies parses, normalizes (ignoring volatile keys + masking PII),
// and compares the two JSON bodies.
func equalJSONBodies(a, b []byte) (bool, error) {
	var va any
	var vb any

	// If either fails to parse as JSON, fall back to raw byte compare.
	if err := json.Unmarshal(a, &va); err != nil {
		return bytes.Equal(normalizeRaw(a), normalizeRaw(b)), nil
	}
	if err := json.Unmarshal(b, &vb); err != nil {
		return bytes.Equal(normalizeRaw(a), normalizeRaw(b)), nil
	}

	na := sanitizeJSON(va)
	nb := sanitizeJSON(vb)

	ba, err := json.Marshal(na)
	if err != nil {
		return false, err
	}
	bb, err := json.Marshal(nb)
	if err != nil {
		return false, err
	}
	return bytes.Equal(ba, bb), nil
}

// sanitizeJSON recursively removes ignored keys and masks PII.
func sanitizeJSON(v any) any {
	switch vv := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(vv))
		for k, val := range vv {
			lk := strings.ToLower(k)
			if _, ok := ignoreKeys[lk]; ok {
				continue
			}
			if _, ok := piiKeys[lk]; ok {
				out[k] = piiMask
				continue
			}
			out[k] = sanitizeJSON(val)
		}
		return out
	case []any:
		for i, el := range vv {
			vv[i] = sanitizeJSON(el)
		}
		return vv
	default:
		return vv
	}
}

// normalizeRaw trims whitespace for non-JSON or parse-fail cases.
func normalizeRaw(b []byte) []byte {
	return bytes.TrimSpace(b)
}
