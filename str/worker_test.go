package str_test

import (
	"testing"

	"github.com/sspencer/str"
)

type emptyWorker int
type fullWorker int
type partialWorker int

func (w emptyWorker) StringWork(s string) string {
	return ""
}

func (w fullWorker) StringWork(s string) string {
	return s
}

func (w partialWorker) StringWork(s string) string {
	if s == "a" {
		return ""
	}

	return s
}

func TestEmptyWorker(t *testing.T) {
	var w emptyWorker
	input := []string{"a", "b", "c", "d", "e", "f", "g"}
	output := str.Worker(0, input, w)
	if len(output) != 0 {
		t.Errorf("Expected no results, received %d", len(output))
	}
}

func TestFullWorker(t *testing.T) {
	var w fullWorker
	input := []string{"a", "b", "c", "d", "e", "f", "g"}
	output := str.Worker(200, input, w)
	expected := len(input)
	if len(output) != expected {
		t.Errorf("Expected %d results, received %d", expected, len(output))
	}
}

func TestPartialWorker(t *testing.T) {
	var w partialWorker
	input := []string{"a", "b", "c", "d", "e", "f", "g"}
	output := str.Worker(2, input, w)
	expected := len(input) - 1
	if len(output) != expected {
		t.Errorf("Expected %d results, received %d", expected, len(output))
	}
}
