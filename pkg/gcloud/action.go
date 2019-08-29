package gcloud

import (
	"github.com/deislabs/porter/pkg/exec/builder"
	"github.com/pkg/errors"
)

type Action struct {
	Steps []Steps // using UnmarshalYAML so that we don't need a custom type per action
}

// UnmarshalYAML takes any yaml in this form
// ACTION:
// - gcloud: ...
// and puts the steps into the Action.Steps field
func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var steps []Steps
	results, err := builder.UnmarshalAction(unmarshal, &steps)
	if err != nil {
		return err
	}

	for _, result := range results {
		step := result.(*[]Steps)
		a.Steps = append(a.Steps, *step...)
	}
	return nil
}

func (a Action) GetSteps() []builder.ExecutableStep {
	steps := make([]builder.ExecutableStep, len(a.Steps))
	for i := range a.Steps {
		steps[i] = a.Steps[i]
	}

	return steps
}

type Steps struct {
	Step `yaml:"gcloud"`
}

type Step struct {
	Description string        `yaml:"description"`
	Groups      Groups        `yaml:"groups"`
	Command     string        `yaml:"command"`
	Arguments   []string      `yaml:"arguments,omitempty"`
	Flags       builder.Flags `yaml:"flags,omitempty"`
	Outputs     []Output      `yaml:"outputs,omitempty"`
}

func (s Step) GetCommand() string {
	return "gcloud"
}

func (s Step) GetArguments() []string {
	args := make([]string, 0, len(s.Arguments)+len(s.Groups)+2)

	// Always be in non-interactive mode, must be specified immediately after gcloud
	args = append(args, "--quiet")

	// Specify the gcloud group(s) and command
	args = append(args, s.Groups...)
	args = append(args, s.Command)

	// Append the positional arguments
	args = append(args, s.Arguments...)

	return args
}

func (s Step) GetFlags() builder.Flags {
	// Always request json formatted output
	return append(s.Flags, builder.NewFlag("format", "json"))
}

type Groups []string

func (groups *Groups) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var groupMap interface{}
	err := unmarshal(&groupMap)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal yaml into Step.Groups")
	}

	switch t := groupMap.(type) {
	case string:
		*groups = append(*groups, t)
	case []interface{}:
		for i := range t {
			group, ok := t[i].(string)
			if !ok {
				return errors.Errorf("invalid yaml type for group item: %T", t[i])
			}
			*groups = append(*groups, group)
		}
	default:
		return errors.Errorf("invalid yaml type for group item: %T", t)
	}

	return nil
}

type Output struct {
	Name     string `yaml:"name"`
	JsonPath string `yaml:"jsonPath"`
}