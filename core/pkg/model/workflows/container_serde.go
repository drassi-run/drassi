package workflows

func (c *Container) DecodeMapstructure(input any) (any, error) {
	if image, ok := input.(string); ok {
		c.Image = image
		return nil, nil
	}
	// process Container normal way
	return input, nil
}
