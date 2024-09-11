package xdig

func Value[V any](v V) func() V {
	return func() V {
		return v
	}
}
