package validators

import (
	"fmt"
	"os"
	"regexp"

	"github.com/ninibe/bigduration"
)

func NonEmptyString(n string, value interface{}) error {
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("%q expected string", n)
	}

	if v == "" {
		return fmt.Errorf("%q is a required parameter", n)
	}
	return nil
}

func StringIsValidPath(n string, value interface{}) error {
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("%q expected string", n)
	}

	if _, err := os.Stat(v); err != nil {
		return fmt.Errorf("%q (%q) is not a valid path: %s", n, v, err.Error())
	}
	return nil
}

func Float64Percent(n string, value interface{}) error {
	v, ok := value.(float64)
	if !ok {
		return fmt.Errorf("%q expected float", n)
	}

	if v < 0 || 1.0 < v {
		return fmt.Errorf("%q represents a percentage and must be between => 0 and <= 1.0, value: %f", n, v)
	}
	return nil
}

func IsEnvironmentNameValid(n string, value interface{}) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9_]{1,4}$`)
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("%q expected string", n)
	}

	if len(v) < 1 {
		return fmt.Errorf("%q (environment name) is a required parameter.", n)
	}

	if len(v) > 4 {
		return fmt.Errorf("%q (%q) cannot be longer than 4 characters.", n, v)
	}

	if matches := re.MatchString(v); !matches {
		return fmt.Errorf("%q (%q) may only contain alphanumeric characters and underscores.", n, v)
	}

	return nil
}

func IsApplicationNameValid(n string, value interface{}) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9_-]{1,26}$`)
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("%q expected string", n)
	}

	if v == "shared-services" {
		return fmt.Errorf("%q cannot be called shared-services (for historical reasons), sorry!", v)
	}

	if len(v) < 1 {
		return fmt.Errorf("%q (application name) is a required parameter.", n)
	}

	if len(v) > 26 {
		// This is because we use this along with env name in ELB names
		// and those are limited to 32 characters
		return fmt.Errorf("%q (%q) cannot be longer than 26 characters.", n, v)
	}

	if matches := re.MatchString(v); !matches {
		return fmt.Errorf("%q (%q) may only contain alphanumeric characters, hyphens and underscores.", n, v)
	}

	return nil
}

func IsVersionValid(n string, value interface{}) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9_]{1,25}$`)
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("%q expected string", n)
	}

	if len(v) < 1 {
		return fmt.Errorf("%q (application version) is a required parameter.", n)
	}

	if len(v) > 25 {
		return fmt.Errorf("%q (%q) cannot be longer than 25 characters.", n, v)
	}

	if matches := re.MatchString(v); !matches {
		return fmt.Errorf("%q (%q) may only contain alphanumeric characters and underscores.", n, v)
	}

	return nil
}

func IsSlotIDValid(n string, value interface{}) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9\.\-_]{0,25}$`)
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("%q expected string", n)
	}

	if len(v) > 25 {
		return fmt.Errorf("%q (%q) cannot be longer than 25 characters.", n, v)
	}

	if len(v) > 1 {
		if matches := re.MatchString(v); !matches {
			return fmt.Errorf("%q (%q) may only contain alphanumeric characters and underscores.", n, v)
		}
	}

	return nil
}

func IsBigDurationValid(n string, value interface{}) error {
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("%q expected string", n)
	}

	if v == "" {
		return fmt.Errorf("%q is a required parameter", n)
	}

	_, err := bigduration.ParseBigDuration(v)
	if err != nil {
		return fmt.Errorf("Expected time duration: %s", err)
	}

	return nil
}

func IsNamespaceValid(n string, value interface{}) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9\/\-_\.*\'()]{1,255}$`)
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("%q expected string", n)
	}

	if len(v) < 1 {
		return fmt.Errorf("%q is a required parameter", n)
	}

	if len(v) > 255 {
		return fmt.Errorf("%q (%q) cannot be longer than 255 characters.", n, v)
	}

	if matches := re.MatchString(v); !matches {
		return fmt.Errorf("%q (%q) may only contain alphanumeric characters, '/', '-', '_', '.', ''', '(', ')' and '*'.", n, v)
	}

	return nil
}
