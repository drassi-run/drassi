package workflows

import "drassi.run/core/pkg/model"

func (c *Container) DecodeMapstructure(input any) (any, error) {
	if image, ok := model.Stringify(input); ok {
		m := map[string]any{"image": image}
		return m, nil
	}
	// process Container normal way
	return input, nil
}
