/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"drassi.run/core/pkg/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

type containerTestStruct struct {
	Con          Container             `actions:"con"`
	ConPtr       *Container            `actions:"conPtr"`
	ListOfCon    []Container           `actions:"listOfCon"`
	MapOfCon     map[string]Container  `actions:"mapOfCon"`
	ListOfConPtr []*Container          `actions:"listOfConPtr"`
	MapOfConPtr  map[string]*Container `actions:"mapOfConPtr"`
}

func TestDecodeContainer(t *testing.T) {
	t.Run("string", func(tt *testing.T) {
		con := Container{Image: "ubuntu:22.04"}
		testDecodeContainer(tt, "ubuntu:22.04", con)
	})

	t.Run("map", func(tt *testing.T) {
		con := Container{
			Image: "ubuntu:22.04",
			Credentials: &ContainerCredentials{
				Username: "username",
				Password: "password",
			},
		}
		val := map[string]any{
			"image": "ubuntu:22.04",
			"credentials": map[string]any{
				"username": "username",
				"password": "password",
			},
		}
		testDecodeContainer(tt, val, con)
	})
}

func testDecodeContainer[C any](tt *testing.T, value C, con Container) {
	data := map[string]any{
		"con":       value,
		"conPtr":    value,
		"listOfCon": []C{value},
		"mapOfCon": map[string]any{
			"key": value,
		},
		"listOfConPtr": []C{value},
		"mapOfConPtr": map[string]any{
			"key": value,
		},
	}

	actual := containerTestStruct{}
	err := model.Decode(data, &actual)
	assert.NoError(tt, err)

	expected := containerTestStruct{
		Con:          con,
		ConPtr:       &con,
		ListOfCon:    []Container{con},
		ListOfConPtr: []*Container{&con},
		MapOfCon:     map[string]Container{"key": con},
		MapOfConPtr:  map[string]*Container{"key": &con},
	}
	assert.EqualValues(tt, actual, expected)
}
