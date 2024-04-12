package executor

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/model"
	"github.com/google/shlex"
)

// Shell You can override the default shell settings in the runner's operating system using the shell keyword.
// You can use built-in shell keywords, or you can define a custom set of shell options.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsshell
type Shell string

const (
	Bash       Shell = "bash"
	Pwsh       Shell = "pwsh"
	Python     Shell = "python"
	Sh         Shell = "sh"
	Cmd        Shell = "cmd"
	Powershell Shell = "powershell"
)

var (
	defaultCommand    = []string{"bash", "-e", "{0}"}
	bashCommand       = []string{"bash", "--noprofile", "--norc", "-eo", "pipefail", "{0}"}
	pwshCommand       = []string{"pwsh", "-command", ". '{0}'"}
	pythonCommand     = []string{"python", "{0}"}
	shCommand         = []string{"sh", "-e", "{0}"}
	cmdCommand        = []string{"%ComSpec%", "/D", "/E:ON", "/V:OFF", "/S", "/C", `CALL "{0}"`}
	powershellCommand = []string{"powershell", "-command", ". '{0}'"}
)

func (s Shell) SupportedPlatform(platform model.Machine) bool {
	switch s {
	case "": // unspecified `shell` parameter
		return platform == model.Linux || platform == model.MacOS
	case Bash, Pwsh, Python:
		return true // Support all platforms
	case Sh:
		return platform == model.Linux || platform == model.MacOS
	case Cmd, Powershell:
		return platform == model.Windows
	default:
		return true // Custom shell
	}
}

func (s Shell) Command() ([]string, error) {
	switch s {
	case "": // unspecified `shell` parameter
		return defaultCommand, nil
	case Bash:
		return bashCommand, nil
	case Pwsh:
		return pwshCommand, nil
	case Python:
		return pythonCommand, nil
	case Sh:
		return shCommand, nil
	case Cmd:
		return cmdCommand, nil
	case Powershell:
		return powershellCommand, nil
	default:
		return shlex.Split(string(s)) // Custom shell
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/ScriptHandlerHelpers.cs#L23-L31
func (s Shell) Extension() string {
	switch s {
	case "": // unspecified `shell` parameter
		return ".sh"
	case Bash:
		return ".sh"
	case Pwsh:
		return ".ps1"
	case Python:
		return ".py"
	case Sh:
		return ".sh"
	case Cmd:
		return ".cmd"
	case Powershell:
		return ".ps1"
	default:
		return ""
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/ScriptHandlerHelpers.cs#L51-L68
func (s Shell) FixupScript(script string) string {
	switch s {
	case Cmd:
		return fmt.Sprintf("@echo off\n%s", script)
	case Pwsh, Powershell:
		scriptPrepend := "$ErrorActionPreference = 'stop'"
		scriptAppend := `if ((Test-Path -LiteralPath variable:\LASTEXITCODE)) { exit $LASTEXITCODE }`
		return fmt.Sprintf("%s\n%s\n%s", scriptPrepend, script, scriptAppend)
	default:
		return script
	}
}
