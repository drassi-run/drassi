package model

import "github.com/google/shlex"

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

func (s Shell) SupportedPlatform(platform Machine) bool {
	switch s {
	case "": // unspecified `shell` parameter
		return platform == Linux || platform == MacOS
	case Bash, Pwsh, Python:
		return true // Support all platforms
	case Sh:
		return platform == Linux || platform == MacOS
	case Cmd, Powershell:
		return platform == Windows
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
