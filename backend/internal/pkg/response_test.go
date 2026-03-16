package pkg

import (
	"encoding/json"
	"testing"
)

func TestSuccessResponse(t *testing.T) {
	resp := SuccessResponse(map[string]string{"key": "value"})

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if result["code"].(float64) != 0 {
		t.Fatalf("Expected code 0, got %v", result["code"])
	}
	if result["message"] != "success" {
		t.Fatalf("Expected message 'success', got %v", result["message"])
	}
	if result["data"] == nil {
		t.Fatal("Expected data to be present")
	}
}

func TestErrorResponse(t *testing.T) {
	resp := ErrorResponse(400, "bad request")

	data, _ := json.Marshal(resp)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if result["code"].(float64) != 400 {
		t.Fatalf("Expected code 400, got %v", result["code"])
	}
	if result["message"] != "bad request" {
		t.Fatalf("Expected message 'bad request', got %v", result["message"])
	}
	if result["data"] != nil {
		t.Fatal("Expected data to be nil")
	}
}

func TestSuccessResponseNilData(t *testing.T) {
	resp := SuccessResponse(nil)

	data, _ := json.Marshal(resp)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	if result["code"].(float64) != 0 {
		t.Fatalf("Expected code 0, got %v", result["code"])
	}
}
