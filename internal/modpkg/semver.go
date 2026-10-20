package modpkg

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var semverPattern = regexp.MustCompile("^(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\\.[0-9A-Za-z-]+)*))?(?:\\+[0-9A-Za-z-]+(?:\\.[0-9A-Za-z-]+)*)?$")

type Version struct {
	Major uint64
	Minor uint64
	Patch uint64
	Pre   []string
	Raw   string
}

func ParseVersion(s string) (Version, error) {
	m := semverPattern.FindStringSubmatch(s)
	if m == nil {
		return Version{}, fmt.Errorf("invalid SemVer %q", s)
	}
	major, _ := strconv.ParseUint(m[1], 10, 64)
	minor, _ := strconv.ParseUint(m[2], 10, 64)
	patch, _ := strconv.ParseUint(m[3], 10, 64)
	v := Version{Major: major, Minor: minor, Patch: patch, Raw: s}
	if m[4] != "" {
		v.Pre = strings.Split(m[4], ".")
		for _, part := range v.Pre {
			if isNumericIdentifier(part) && len(part) > 1 && part[0] == '0' {
				return Version{}, fmt.Errorf("invalid SemVer prerelease %q: numeric identifiers cannot have leading zeroes", s)
			}
		}
	}
	return v, nil
}

func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major { return -1 }
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor { return -1 }
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch { return -1 }
		return 1
	}
	if len(v.Pre) == 0 && len(other.Pre) == 0 { return 0 }
	if len(v.Pre) == 0 { return 1 }
	if len(other.Pre) == 0 { return -1 }
	n := len(v.Pre)
	if len(other.Pre) < n { n = len(other.Pre) }
	for i := 0; i < n; i++ {
		a, b := v.Pre[i], other.Pre[i]
		if a == b { continue }
		an, bn := isNumericIdentifier(a), isNumericIdentifier(b)
		switch {
		case an && bn:
			ai, _ := strconv.ParseUint(a, 10, 64)
			bi, _ := strconv.ParseUint(b, 10, 64)
			if ai < bi { return -1 }
			if ai > bi { return 1 }
		case an:
			return -1
		case bn:
			return 1
		default:
			if a < b { return -1 }
			if a > b { return 1 }
		}
	}
	if len(v.Pre) < len(other.Pre) { return -1 }
	if len(v.Pre) > len(other.Pre) { return 1 }
	return 0
}

func isNumericIdentifier(s string) bool {
	if s == "" { return false }
	for _, r := range s {
		if r < '0' || r > '9' { return false }
	}
	return true
}

type constraint struct {
	op string
	v  Version
}

func MatchVersion(version, expression string) (bool, error) {
	v, err := ParseVersion(version)
	if err != nil { return false, err }
	expression = strings.TrimSpace(expression)
	if expression == "" || expression == "*" {
		return true, nil
	}
	parts := strings.Fields(expression)
	if len(parts) == 1 && !strings.HasPrefix(parts[0], ">") && !strings.HasPrefix(parts[0], "<") && !strings.HasPrefix(parts[0], "=") {
		want, err := ParseVersion(parts[0])
		if err != nil { return false, err }
		return v.Compare(want) == 0, nil
	}
	for _, part := range parts {
		c, err := parseConstraint(part)
		if err != nil { return false, err }
		cmp := v.Compare(c.v)
		ok := false
		switch c.op {
		case "=":
			ok = cmp == 0
		case ">":
			ok = cmp > 0
		case ">=":
			ok = cmp >= 0
		case "<":
			ok = cmp < 0
		case "<=":
			ok = cmp <= 0
		}
		if !ok { return false, nil }
	}
	return true, nil
}

func parseConstraint(s string) (constraint, error) {
	for _, op := range []string{">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(s, op) {
			raw := strings.TrimPrefix(s, op)
			if raw == "" {
				return constraint{}, fmt.Errorf("missing version in constraint %q", s)
			}
			v, err := ParseVersion(raw)
			if err != nil { return constraint{}, err }
			return constraint{op: op, v: v}, nil
		}
	}
	return constraint{}, fmt.Errorf("invalid version constraint %q", s)
}
