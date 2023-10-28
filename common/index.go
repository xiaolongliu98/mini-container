package common

import "log"

func Must(err ...error) {
	for i, e := range err {
		if e != nil {
			log.Fatalf("ERROR[%d]: %v\n", i, e)
		}
	}
}

func Err(rets ...any) error {
	for _, ret := range rets {
		if err, ok := ret.(error); ok {
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

func ErrGroupThrough(err ...error) (int, error) {
	var targetErr error = nil
	var idx int = -1

	for i, e := range err {
		if e != nil && idx == -1 {
			targetErr = e
			idx = i
		}
	}

	return idx, targetErr
}
