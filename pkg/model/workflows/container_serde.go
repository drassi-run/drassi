package workflows

func (c *Container) DecodeMapstructure(input any) (any, error) {
	if s, ok := input.(string); ok {
		if image, err := NewEvaluable(s, toString); err != nil {
			return nil, err
		} else {
			c.Image = image
			return nil, nil
		}
	}
	// process ContainerBase normal way
	return input, nil
}
