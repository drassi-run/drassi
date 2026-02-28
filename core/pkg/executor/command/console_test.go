/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupConsoleCmdMgr() *consoleManager[string] {
	return NewConsoleManager[string]().(*consoleManager[string])
}

func noop(context.Context, string, *Command) error { return nil }

func TestConsoleManager_ParseCommandV1(t *testing.T) {
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

	mgr := setupConsoleCmdMgr()
	_ = mgr.Register(NewConsoleHandler("do-something", true, noop))

	for input, expected := range tests {
		t.Run(input, func(tt *testing.T) {
			actual := mgr.parseCommandV1(input)
			assert.EqualValues(tt, actual, expected, input)
		})
	}
}

func TestConsoleManager_ParseCommandV2(t *testing.T) {
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

	mgr := setupConsoleCmdMgr()
	_ = mgr.Register(NewConsoleHandler("do-something", true, noop))

	for input, expected := range tests {
		t.Run(input, func(tt *testing.T) {
			actual := mgr.parseCommandV2(input)
			assert.EqualValues(tt, actual, expected, input)
		})
	}
}

func TestConsoleManager_IsProcessingCommand(t *testing.T) {
	mgr := setupConsoleCmdMgr()
	_ = mgr.Register(NewConsoleHandler("foobar", true, noop))

	assert.True(t, mgr.isProcessingCommand("foobar"))
	assert.False(t, mgr.isProcessingCommand("xxx"))

	mgr.resumeCmdToken = "xxx"
	assert.False(t, mgr.isProcessingCommand("foobar"))
	assert.True(t, mgr.isProcessingCommand("xxx"))
}

func TestConsoleManager_StopCommands(t *testing.T) {
	mgr := setupConsoleCmdMgr()
	_ = mgr.Register(NewConsoleHandler("do-something", true, noop))
	v1line := "##[do-something k1=v1;]msg"
	v2line := "::do-something k1=v1,::msg"
	cmd := &Command{
		Name: "do-something",
		Params: map[string]string{
			"k1": "v1",
		},
		Value: "msg",
	}

	assert.EqualValues(t, cmd, mgr.ParseCommand(v2line), v2line)
	assert.EqualValues(t, cmd, mgr.ParseCommand(v1line), v1line)

	mgr.resumeCmdToken = "xxx"
	assert.Nil(t, mgr.ParseCommand(v2line), v2line)
	assert.Nil(t, mgr.ParseCommand(v1line), v1line)

	resumeCmd := &Command{Name: "xxx"}
	assert.EqualValues(t, resumeCmd, mgr.ParseCommand("##[xxx]"), "##[xxx]")
	assert.EqualValues(t, resumeCmd, mgr.ParseCommand("::xxx::"), "::xxx::")
}
