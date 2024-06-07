package problem

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/IssueMatcher.cs#L230
type Configs struct {
	ProblemMatcher []Config `json:"problemMatcher,omitempty"`
}

func (c *Configs) Validate() error {
	distinctOwners := sets.New[string]()
	for _, config := range c.ProblemMatcher {
		if err := config.Validate(); err != nil {
			return err
		}
		if distinctOwners.Has(config.Owner) {
			return fmt.Errorf("duplicate owner name %s", config.Owner)
		}
		distinctOwners.Insert(config.Owner)
	}
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/IssueMatcher.cs#L274
type Config struct {
	Owner    string    `json:"owner,omitempty"`
	Severity string    `json:"severity,omitempty"`
	Pattern  []Pattern `json:"pattern,omitempty"`
}

var validSeverities = []string{"", "ERROR", "WARNING", "NOTICE"}

func (c *Config) Validate() error {
	if c.Owner == "" {
		return fmt.Errorf("owner must not be empty")
	}

	c.Severity = strings.ToUpper(c.Severity)
	if slices.Contains(validSeverities, c.Severity) {
		return fmt.Errorf("matcher %s contains unexpected default severity: %s", c.Owner, c.Severity)
	}

	patternCount := len(c.Pattern)
	if patternCount == 0 {
		return fmt.Errorf("matcher %s pattern must not be empty", c.Owner)
	}

	var (
		file     = -1
		line     = -1
		column   = -1
		severity = -1
		code     = -1
		message  = -1
		fromPath = -1
	)
	for i, p := range c.Pattern {
		re, err := regexp.Compile(p.Regexp)
		if err != nil {
			return err
		}

		if p.Loop {
			if patternCount == 1 || i != patternCount-1 {
				return fmt.Errorf("only the last pattern in a multiline matcher may set 'loop'")
			}
			if p.Message == nil {
				return fmt.Errorf("the loop pattern must set 'message'")
			}
		}

		groupCount := re.NumSubexp()
		if err = c.validatePatternField("file", p.File, groupCount, &file); err != nil {
			return err
		}
		if err = c.validatePatternField("line", p.Line, groupCount, &line); err != nil {
			return err
		}
		if err = c.validatePatternField("column", p.Column, groupCount, &column); err != nil {
			return err
		}
		if err = c.validatePatternField("severity", p.Severity, groupCount, &severity); err != nil {
			return err
		}
		if err = c.validatePatternField("code", p.Code, groupCount, &code); err != nil {
			return err
		}
		if err = c.validatePatternField("message", p.Message, groupCount, &message); err != nil {
			return err
		}
		if err = c.validatePatternField("fromPath", p.FromPath, groupCount, &fromPath); err != nil {
			return err
		}
	}

	if message < 0 {
		return fmt.Errorf("at least one pattern must set 'message'")
	}

	return nil
}

func (c *Config) validatePatternField(fieldName string, fieldValue *int, groupCount int, value *int) error {
	if fieldValue == nil {
		return nil
	}

	v := *fieldValue
	if groupCount < v {
		return fmt.Errorf("the value %d of property '%s' is out of range", v, fieldName)
	}

	if value != nil && *value >= 0 {
		return fmt.Errorf("the property %s is set twice", fieldName)
	}

	*value = v
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/IssueMatcher.cs#L398
type Pattern struct {
	File     *int   `json:"file,omitempty"`
	Line     *int   `json:"line,omitempty"`
	Column   *int   `json:"column,omitempty"`
	Severity *int   `json:"severity,omitempty"`
	Code     *int   `json:"code,omitempty"`
	Message  *int   `json:"message,omitempty"`
	FromPath *int   `json:"fromPath,omitempty"`
	Loop     bool   `json:"loop,omitempty"`
	Regexp   string `json:"regexp,omitempty"`
}
