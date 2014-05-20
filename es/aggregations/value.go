package aggregations

import "fmt"

type Value struct {
	Value interface{} `json:"value"`
}

func loadValueAggregate(i map[string]interface{}) (*Value, error) {
	if len(i) == 1 {
		if v, ok := i["value"]; ok {
			return &Value{Value: v}, nil
		}
	}
	return nil, fmt.Errorf("not a value aggregate")

}
