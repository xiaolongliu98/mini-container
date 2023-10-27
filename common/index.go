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
