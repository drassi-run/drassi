package executor

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/executor/command"
	"github.com/dungdm93/drassi/core/pkg/executor/secret"
)

func (e *JobExecutor) addSecretMaskHandler(cmd *command.Command) error {
	if cmd.Value == "" {
		return errors.New("can't add secret mask for empty string in ##[add-mask] command")
	}

	s := secret.NewValueSecret(cmd.Value)
	e.AddSecretMask(s)
	for _, mask := range strings.FieldsFunc(cmd.Value, func(c rune) bool { return c == '\n' || c == '\r' }) {
		if mask != cmd.Value {
			s = secret.NewValueSecret(mask)
			e.AddSecretMask(s)
		}
	}

	return nil
}

func (e *JobExecutor) addProblemMatcherHandler(cmd *command.Command) error {
	file := cmd.Value
	if file == "" {
		return errors.New("file path must be specified in ##[add-matcher] command")
	}

	matchers, err := e.readProblemMatcherFile(file)
	if err != nil {
		return err
	}

	for _, m := range matchers.ProblemMatcher {
		e.AddProblemMatcher(m)
	}
	return nil
}

func (e *JobExecutor) removeProblemMatcher(cmd *command.Command) error {
	file := cmd.Value
	owner := cmd.Params["owner"]
	if (file == "") == (owner == "") {
		return errors.New("either an owner name or a file path must be specified in ##[remove-matcher] command")
	}

	var owners []string
	if owner != "" {
		owners = []string{owner}
	} else {
		matchers, err := e.readProblemMatcherFile(file)
		if err != nil {
			return err
		}
		owners = make([]string, len(matchers.ProblemMatcher))
		for i, m := range matchers.ProblemMatcher {
			owners[i] = m.Owner
		}
	}

	for _, o := range owners {
		e.RemoveProblemMatcher(o)
	}

	return nil
}

func (e *JobExecutor) addPath(cmd *command.Command) error {
	paths := []string{cmd.Value}
	return e.AddPath(paths)
}

func (e *JobExecutor) setEnv(cmd *command.Command) error {
	name, ok := cmd.Params["name"]
	if !ok || name == "" {
		return errors.New("required field 'name' is missing in ##[set-output] command")
	}

	env := map[string]string{
		name: cmd.Value,
	}
	return e.SetEnv(env)
}

func (e *JobExecutor) groupingLog(cmd *command.Command) error {
	e.Log("", "##[group]"+cmd.Value)
	return nil
}

func (e *JobExecutor) endGroupingLog(cmd *command.Command) error {
	e.Log("", "##[endgroup]")
	return nil
}

func (e *JobExecutor) logDebug(cmd *command.Command) error {
	return nil
}

func (e *JobExecutor) logMessage(level string, cmd *command.Command) error {
	return nil
}

func (e *JobExecutor) readProblemMatcherFile(file string) (*ProblemMatchers, error) {
	if filepath.IsLocal(file) {
		ws := e.Sandbox().GetWorkspaceDir()
		file = filepath.Join(ws, file)
	}
	reader, err := e.Sandbox().CopyOut(context.Background(), file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	matchers := new(ProblemMatchers)
	if err = json.NewDecoder(reader).Decode(matchers); err != nil {
		return nil, err
	}
	return matchers, nil
}
