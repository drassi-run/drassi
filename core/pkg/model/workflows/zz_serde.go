package workflows

import "encoding/json/v2"

var unmarshalers []*json.Unmarshalers

func JsonUnmarshalers() *json.Unmarshalers {
	return json.JoinUnmarshalers(unmarshalers...)
}
