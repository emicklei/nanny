package nanny

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type RecordCondition struct {
	Name    string
	Enabled bool
	Path    string
	Value   string
}

func (c RecordCondition) String() string {
	return c.Name
}

// Matches returns true if the event matches the specification of this condition.
func (r RecordCondition) Matches(ev Event) bool {
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
		// is it a pattern?
		if strings.Contains(r.Value, "*") {
			// TODO cache this?
			re, err := regexp.Compile(strings.ReplaceAll(r.Value, "*", ".*"))
			if err != nil {
				return false
			}
			return re.MatchString(ev.Message)
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
