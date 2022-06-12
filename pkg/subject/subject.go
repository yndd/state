package subject

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
)

const (
	dotReplChar   = "^"
	spaceReplChar = "~"
)

var errMalformedXPath = errors.New("malformed xpath")
var errMalformedXPathKey = errors.New("malformed xpath key")
var escapedBracketsReplacer = strings.NewReplacer(`\]`, `]`, `\[`, `[`)

var regDot = regexp.MustCompile(`\.`)
var regSpace = regexp.MustCompile(`\s`)

func GNMIPathToSubject(p *gnmi.Path) string {
	if p == nil {
		return ""
	}
	sb := new(strings.Builder)
	if p.GetOrigin() != "" {
		fmt.Fprintf(sb, "%s.", p.GetOrigin())
	}
	for i, e := range p.GetElem() {
		if i > 0 {
			sb.WriteString(".")
		}
		sb.WriteString(e.Name)
		if len(e.Key) > 0 {
			// sort keys by name
			kNames := make([]string, 0, len(e.Key))
			for k := range e.Key {
				kNames = append(kNames, k)
			}
			sort.Strings(kNames)
			for _, k := range kNames {
				sk := sanitizeKey(e.GetKey()[k])
				fmt.Fprintf(sb, ".{%s=%s}", k, sk)
			}
		}
	}
	return sb.String()
}

func XPathToSubject(p string) (string, error) {
	lp := len(p)
	if lp == 0 {
		return "", nil
	}

	sb := new(strings.Builder)
	idx := strings.Index(p, ":")
	if idx >= 0 && p[0] != '/' && !strings.Contains(p[:idx], "/") &&
		((idx+1 < lp && p[idx+1] == '/') || (lp == idx+1)) {
		sb.WriteString(p[:idx])
		p = p[idx+1:]
		// if path is only origin
		if lp == idx+1 {
			sb.WriteString(".>")
			return sb.String(), nil
		}
		p = strings.TrimPrefix(p, "/")
		if len(p) == 0 {
			sb.WriteString(".>")
			return sb.String(), nil
		}
		sb.WriteString(".")
	}
	p = strings.TrimPrefix(p, "/")
	buffer := make([]rune, 0)
	null := rune(0)
	prevC := rune(0)
	// track if the loop is traversing a key
	inKey := false

	for _, r := range p {
		switch r {
		case '[':
			if inKey && prevC != '\\' {
				return "", errMalformedXPath
			}
			if prevC != '\\' {
				inKey = true
			}
		case ']':
			if !inKey && prevC != '\\' {
				return "", errMalformedXPath
			}
			if prevC != '\\' {
				inKey = false
			}
		case '/':
			if !inKey {
				buffer = append(buffer, null)
				prevC = r
				continue
			}
		}
		buffer = append(buffer, r)
		prevC = r
	}
	if inKey {
		return "", errMalformedXPath
	}

	stringElems := strings.Split(string(buffer), string(null))
	numElem := len(stringElems)

	for i, s := range stringElems {
		idx := -1
		prevC := rune(0)
		for j, r := range s {
			if r == '[' && prevC != '\\' {
				idx = j
				break
			}
			prevC = r
		}
		var kvs map[string]string
		var keys []string
		if idx > 0 {
			var err error
			keys, kvs, err = parseXPathKeys(s[idx:])
			if err != nil {
				return "", err
			}
			s = s[:idx]
		}
		sb.WriteString(s)

		numKeys := len(keys)
		if numKeys > 0 {
			sb.WriteString(".")
			for j, k := range keys {
				v := kvs[k]
				if v != "*" {
					sb.WriteString("{")
					sb.WriteString(k)
					sb.WriteString("=")
					sb.WriteString(v)
					sb.WriteString("}")
				} else {
					sb.WriteString("*")
				}
				if j+1 != numKeys {
					sb.WriteString(".")
				}
			}
		}
		if i+1 != numElem {
			sb.WriteString(".")
		}
	}
	subject := sb.String()
	if strings.HasSuffix(subject, ".*") {
		subject = subject[:len(subject)-1] + ">"
		return subject, nil
	}
	sb.WriteString(".>")
	return sb.String(), nil
}

// parseXPathKeys takes keys definition from an xpath, e.g [k1=v1][k2=v2] and return the keys and values as a map[string]string
func parseXPathKeys(s string) ([]string, map[string]string, error) {
	if len(s) == 0 {
		return nil, nil, nil
	}
	kvs := make(map[string]string)
	keys := make([]string, 0)
	inKey := false
	start := 0
	prevRune := rune(0)
	for i, r := range s {
		switch r {
		case '[':
			if prevRune == '\\' {
				prevRune = r
				continue
			}
			if inKey {
				return nil, nil, errMalformedXPathKey
			}
			inKey = true
			start = i + 1
		case ']':
			if prevRune == '\\' {
				prevRune = r
				continue
			}
			if !inKey {
				return nil, nil, errMalformedXPathKey
			}
			eq := strings.Index(s[start:i], "=")
			if eq < 0 {
				return nil, nil, errMalformedXPathKey
			}
			k, v := s[start:i][:eq], s[start:i][eq+1:]
			if len(k) == 0 || len(v) == 0 {
				return nil, nil, errMalformedXPathKey
			}
			sk := escapedBracketsReplacer.Replace(k)
			kvs[sk] = escapedBracketsReplacer.Replace(v)
			keys = append(keys, sk)
			inKey = false
		}
		prevRune = r
	}
	if inKey {
		return nil, nil, errMalformedXPathKey
	}
	sort.Strings(keys)
	return keys, kvs, nil
}

func sanitizeKey(k string) string {
	s := regDot.ReplaceAllString(k, dotReplChar)
	return regSpace.ReplaceAllString(s, spaceReplChar)
}
