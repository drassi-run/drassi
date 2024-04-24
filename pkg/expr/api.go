package expr

type Result interface {
	Value() any
	Kind() ResultKind
}

// ResultKind represent kind of evaluation result
type ResultKind int

const (
	Array ResultKind = iota
	Boolean
	Null
	Number
	Object
	String
)

func (v ResultKind) ToString() string {
	switch v {
	case Array:
		return "Array"
	case Boolean:
		return "Boolean"
	case Null:
		return "Null"
	case Number:
		return "Number"
	case String:
		return "String"
	case Object:
		return "Object"
	default:
		return "Unknown ResultKind"
	}
}
