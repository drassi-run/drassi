/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ref

type Type int

const (
	TypeInvalid Type = iota
	TypeNull
	TypeBoolean
	TypeInteger
	TypeFloat
	TypeString
	TypeList
	TypeMap
	TypeStruct
)

func (t Type) String() string {
	switch t {
	case TypeNull:
		return "null"
	case TypeBoolean:
		return "boolean"
	case TypeInteger:
		return "number"
	case TypeFloat:
		return "number"
	case TypeString:
		return "string"
	case TypeList:
		return "array"
	case TypeMap:
		return "object"
	case TypeStruct:
		return "object"
	default:
		return "<invalid>"
	}
}
