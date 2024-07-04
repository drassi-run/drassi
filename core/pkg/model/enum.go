package model

type Architecture string

const (
	ARM32 Architecture = "ARM32"
	X64   Architecture = "X64"
	X86   Architecture = "X86"
)

type Machine string

const (
	Linux   Machine = "linux"
	MacOS   Machine = "macos"
	Windows Machine = "windows"
)

func (m *Machine) FileSeparator() rune {
	switch *m {
	case Linux, MacOS:
		return '/'
	case Windows:
		return '\\'
	default:
		panic("Unknown machine")
	}
}

func (m *Machine) PathSeparator() rune {
	switch *m {
	case Linux, MacOS:
		return ':'
	case Windows:
		return ';'
	default:
		panic("Unknown machine")
	}
}
