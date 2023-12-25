package nanny

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

type RecordCondition struct {
	Name    string
	Enabled bool
	Path    string
	Value   string
	valueRe *regexp.Regexp
}

func (c RecordCondition) String() string {
	return c.Name
}

func NewCondition(name string, enabled bool, path string, value string) RecordCondition {
	rc := RecordCondition{
		Name:    name,
		Enabled: enabled,
		Path:    path,
		Value:   value,
	}
	rc = rc.withRegexp()
	return rc
}

// withRegexp returns a copy with a cached Regexp when the values specifies one.
func (r RecordCondition) withRegexp() RecordCondition {
	if strings.HasPrefix(r.Value, "/") && strings.HasSuffix(r.Value, "/") {
		expression := r.Value[1 : len(r.Value)-1]
		re, err := regexp.Compile(expression)
		if err != nil {
			slog.Warn(fmt.Sprintf("invalid regexp %q", r.Value))
			return r
		}
		r.valueRe = re
	}
	return r
}

// Matches returns true if the event matches the specification of this condition.
func (r RecordCondition) Matches(ev *Event) bool {
	if !r.Enabled {
		return true
	}
	tokens := strings.Split(r.Path, ".")
	if len(tokens) == 0 {
		return false
	}
	// first is field of event
	switch tokens[0] {
	case "level":
		return strings.ToLower(ev.Level.String()) == strings.ToLower(r.Value)
	case "message":
		if r.valueRe != nil {
			return r.valueRe.MatchString(ev.Message)
		}
		// exact
		return ev.Message == r.Value
	case "attrs":
		val := pathFindIn(1, tokens, ev.Attrs)
		return val != nil && fmt.Sprintf("%v", val) == r.Value
	default:
		return false
	}
}

func pathFindIn(index int, tokens []string, here interface{}) interface{} {
	//.Printf("%d %q %d, %v\n", index, tokens, len(tokens), here)
	if here == nil {
		return here
	}
	if index == len(tokens) {
		return here
	}
	token := tokens[index]
	if len(token) == 0 {
		return here
	}
	i, err := strconv.Atoi(token)
	if err == nil {
		// try index into array
		array, ok := here.([]interface{})
		if ok {
			if i >= len(array) {
				return nil
			}
			return pathFindIn(index+1, tokens, array[i])
		}
		return nil
	}
	// try key into hash
	hash, ok := here.(map[string]interface{})
	if ok {
		return pathFindIn(index+1, tokens, hash[token])
	}
	return nil
}
