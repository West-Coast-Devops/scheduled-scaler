package main

import "testing"

func NeverFail(t *testing.T) {
	if false {
		t.Errorf("Should never fail!")
	}
}
