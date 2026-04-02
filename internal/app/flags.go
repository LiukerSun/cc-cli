package app

import (
	"fmt"
	"strings"
)

type kvFlag struct {
	values map[string]string
}

type stringListFlag struct {
	values []string
}

type boolFlag struct {
	value bool
	set   bool
}

func (f *kvFlag) String() string {
	return ""
}

func (f *kvFlag) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid env value %q: expected KEY=VALUE", value)
	}
	key := strings.TrimSpace(parts[0])
	if key == "" {
		return fmt.Errorf("invalid env value %q: key is empty", value)
	}
	if f.values == nil {
		f.values = map[string]string{}
	}
	f.values[key] = parts[1]
	return nil
}

func (f *stringListFlag) String() string {
	return strings.Join(f.values, ",")
}

func (f *stringListFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	f.values = append(f.values, value)
	return nil
}

func (f *boolFlag) String() string {
	if !f.set {
		return ""
	}
	if f.value {
		return "true"
	}
	return "false"
}

func (f *boolFlag) Set(value string) error {
	f.set = true
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "true", "1", "yes":
		f.value = true
	case "false", "0", "no":
		f.value = false
	default:
		return fmt.Errorf("invalid boolean value %q", value)
	}
	return nil
}

func (f *boolFlag) IsBoolFlag() bool {
	return true
}

func removeString(values []string, target string) []string {
	target = strings.TrimSpace(target)
	if target == "" {
		return values
	}
	out := values[:0]
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			continue
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
