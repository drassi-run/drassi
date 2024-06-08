package command

import (
	"gotest.tools/v3/assert"
	"testing"
)

func TestParseCommandV1(t *testing.T) {
	tests := map[string]*Command{
		"##[do-something k1=v1;]msg": {
			Name: "do-something",
			Params: map[string]string{
				"k1": "v1",
			},
			Value: "msg",
		},
		"##[do-something]": {
			Name: "do-something",
		},
		"##[do-something k1=%3B=%0D=%0A=%5D;]%3B-%0D-%0A-%5D": {
			Name: "do-something",
			Params: map[string]string{
				"k1": ";=\r=\n=]",
			},
			Value: ";-\r-\n-]",
		},
		"##[do-something k1=%253B=%250D=%250A=%255D;]%253B-%250D-%250A-%255D": {
			Name: "do-something",
			Params: map[string]string{
				"k1": "%3B=%0D=%0A=%5D",
			},
			Value: "%3B-%0D-%0A-%5D",
		},
		"##[do-something k1=;k2=;]": {
			Name: "do-something",
		},
		">>>   ##[do-something k1=v1;]msg": {
			Name: "do-something",
			Params: map[string]string{
				"k1": "v1",
			},
			Value: "msg",
		},
	}

	mgr := NewConsoleCommandManager(nil)
	_ = mgr.RegisterCommand("do-something", true, func(cmd *Command) error { return nil })

	for input, expected := range tests {
		t.Run(input, func(tt *testing.T) {
			actual := mgr.parseCommandV1(input)
			assert.DeepEqual(tt, actual, expected)
		})
	}
}

func TestParseCommandV2(t *testing.T) {
	tests := map[string]*Command{
		"::do-something k1=v1,::msg": {
			Name: "do-something",
			Params: map[string]string{
				"k1": "v1",
			},
			Value: "msg",
		},
		"::do-something::": {
			Name: "do-something",
		},
		"::do-something k1=;=%2C=%0D=%0A=]=%3A,::;-%0D-%0A-]-:-,": {
			Name: "do-something",
			Params: map[string]string{
				"k1": ";=,=\r=\n=]=:",
			},
			Value: ";-\r-\n-]-:-,",
		},
		"::do-something k1=;=%252C=%250D=%250A=]=%253A,::;-%250D-%250A-]-:-,": {
			Name: "do-something",
			Params: map[string]string{
				"k1": ";=%2C=%0D=%0A=]=%3A",
			},
			Value: ";-%0D-%0A-]-:-,",
		},
		"::do-something k1=,k2=,::": {
			Name: "do-something",
		},
		"::do-something k1=v1::": {
			Name: "do-something",
			Params: map[string]string{
				"k1": "v1",
			},
		},
		"   ::do-something k1=v1,::msg": {
			Name: "do-something",
			Params: map[string]string{
				"k1": "v1",
			},
			Value: "msg",
		},
		"   >>>   ::do-something k1=v1,::msg": nil,
	}

	mgr := NewConsoleCommandManager(nil)
	_ = mgr.RegisterCommand("do-something", true, func(cmd *Command) error { return nil })

	for input, expected := range tests {
		t.Run(input, func(tt *testing.T) {
			actual := mgr.parseCommandV2(input)
			assert.DeepEqual(tt, actual, expected)
		})
	}
}

func TestIsProcessingCommand(t *testing.T) {
	mgr := NewConsoleCommandManager(nil)
	_ = mgr.RegisterCommand("foobar", true, func(cmd *Command) error { return nil })

	assert.DeepEqual(t, true, mgr.isProcessingCommand("foobar"))
	assert.DeepEqual(t, false, mgr.isProcessingCommand("xxx"))

	mgr.resumeCmdToken = "xxx"
	assert.DeepEqual(t, false, mgr.isProcessingCommand("foobar"))
	assert.DeepEqual(t, true, mgr.isProcessingCommand("xxx"))
}

func TestStopCommands(t *testing.T) {
	mgr := NewConsoleCommandManager(nil)
	_ = mgr.RegisterCommand("do-something", true, func(cmd *Command) error { return nil })
	v1line := "##[do-something k1=v1;]msg"
	v2line := "::do-something k1=v1,::msg"
	cmd := &Command{
		Name: "do-something",
		Params: map[string]string{
			"k1": "v1",
		},
		Value: "msg",
	}

	assert.DeepEqual(t, cmd, mgr.ParseCommand(v2line))
	assert.DeepEqual(t, cmd, mgr.ParseCommand(v1line))

	mgr.resumeCmdToken = "xxx"
	assert.Assert(t, mgr.ParseCommand(v2line) == nil)
	assert.Assert(t, mgr.ParseCommand(v1line) == nil)

	resumeCmd := &Command{Name: "xxx"}
	assert.DeepEqual(t, resumeCmd, mgr.ParseCommand("##[xxx]"))
	assert.DeepEqual(t, resumeCmd, mgr.ParseCommand("::xxx::"))
}
