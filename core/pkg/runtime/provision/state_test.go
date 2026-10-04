/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestStateSuite(t *testing.T) {
	suite.Run(t, new(StateTestSuite))
}

type StateTestSuite struct {
	suite.Suite
}

func (s *StateTestSuite) TestGenericState() {
	st := new(State)

	const keyStr = StateKey[string]("str_key")
	st.Set(keyStr, "hello")
	val, ok := st.Get(keyStr)
	s.Require().True(ok)
	s.Require().Equal("hello", val)
	s.Require().Equal("hello", st.MustGet(keyStr))

	const keyMissing = StateKey[int]("missing_key")
	intVal, ok := st.Get(keyMissing)
	s.Require().False(ok)
	s.Require().Equal(0, intVal)
	s.Require().Panics(func() {
		st.MustGet(keyMissing)
	})

	const keyMismatch = StateKey[int]("str_key")
	mismatchVal, ok := st.Get(keyMismatch)
	s.Require().False(ok)
	s.Require().Equal(0, mismatchVal)
	s.Require().Panics(func() {
		st.MustGet(keyMismatch)
	})
}
