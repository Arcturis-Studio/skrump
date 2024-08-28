package utils

import "log"


func AssertStringValue(d any) string {
	s, ok := d.(string)
	if !ok {
		log.Fatal("expected value as string")
	}
	return s
}
