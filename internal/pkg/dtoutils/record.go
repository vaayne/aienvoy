package dtoutils

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/types"
)

type BaseModel struct {
	Id      string         `json:"id,omitempty" mapstructure:"id,omitempty"`
	Created types.DateTime `json:"created,omitempty" mapstructure:"created,omitempty"`
	Updated types.DateTime `json:"updated,omitempty" mapstructure:"updated,omitempty"`
}

func FromRecord(r *models.Record, output any) error {
	jsonData, err := r.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, output)
}

func ToRecord(r *models.Record, input any) error {
	var mapData map[string]any
	if err := mapstructure.Decode(input, &mapData); err != nil {
		return err
	}
	for k, v := range mapData {
		r.Set(k, v)
	}
	return nil
}
