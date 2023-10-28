package common

import (
	"encoding/json"
	"log"
	"os"
)

func Must(err ...error) {
	for i, e := range err {
		if e != nil {
			log.Fatalf("ERROR[%d]: %v\n", i, e)
		}
	}
}

func Err(rets ...any) error {
	for i := len(rets) - 1; i >= 0; i-- {
		if err, ok := rets[i].(error); ok {
			return err
		}
	}
	return nil
}

func ErrGroup(err ...error) (int, error) {
	for i, e := range err {
		if e != nil {
			return i, e
		}
	}
	return -1, nil
}

func ErrGroupCount(err ...error) (int, error) {
	var targetErr error = nil
	var count int = -1

	for _, e := range err {
		if e != nil {
			count++
		}

		if e != nil && targetErr == nil {
			targetErr = e
		}
	}

	return count, targetErr
}

func WriteJSON(name string, obj any) error {
	data, _ := json.Marshal(obj)
	return os.WriteFile(name, data, 0644)
}

func ReadJSON(name string, objPtr any) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, objPtr)
}
