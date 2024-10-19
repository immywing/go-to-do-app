package models_test

import (
	"testing"

	"go-to-do-app/to-do-lib/models"
)

func TestParsePriorityWithValidStrings(t *testing.T) {
	inputs := []string{"Low", "Medium", "High"}
	for _, s := range inputs {
		_, err := models.ParsePriority(s)
		if err != nil {
			t.Errorf("parser failed with %s error", err)
		}
	}
}

func TestParsePriorityWithValidLowercaseStrings(t *testing.T) {
	inputs := []string{"low", "medium", "high"}
	for _, s := range inputs {
		_, err := models.ParsePriority(s)
		if err != nil {
			t.Errorf("parser failed with %s error", err)
		}
	}
}

func TestParsePriorityErrorsWithInvalidString(t *testing.T) {
	input := "Critical"
	ret, err := models.ParsePriority(input)
	if err == nil {
		t.Errorf("Expected parser to fail given an input of %s, but returned %s", input, ret)
	}
}
