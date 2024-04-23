package expression

type IReadOnlyObj interface {
	Count() int
	Keys() []string
	Values() []any
	ContainsKey(key string) (exist bool)
	GetValue(key string) (exist bool, value any)
	Enumerator() *Enumerator
}
