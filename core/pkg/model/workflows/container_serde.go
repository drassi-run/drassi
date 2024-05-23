package workflows

func (c *Container) DecodeMapstructure(input any) (any, error) {
	if image, ok := input.(string); ok {
		m := map[string]any{"image": image}
		return m, nil
	}
	// process Container normal way
	return input, nil
}
