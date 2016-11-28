package main

import (
	"testing"
	"encoding/json"
)

func TestCalculateQuote(t *testing.T) {
	data := `{"country":"FR","departureDate":"2016-12-04","returnDate":"2017-01-09","travellerAges":[7,50,67],"options":[],"cover":"Extra"}`
	reply := calculateQuote([]byte(data))
	if reply.Quote != 267.8399999999999 {
		t.Fail()
	}

	data = []byte(`{"country":"ES","departureDate":"2016-12-02","returnDate":"2016-12-31","travellerAges":[73,54],"options":[],"cover":"Basic"}`)
	if reply.Quote != 108,58 {
		t.Fail()
	}
}

func TestParseFeedback(t *testing.T) {
	data := []byte(`{"message":"Congrats MrRobot, your answer ({quote=159.12}) was right !-> You just earned 100.0","type":"WIN"}`)
	var feedback Feedback
	json.Unmarshal(data, &feedback)
	if feedback.Message != "Congrats MrRobot, your answer ({quote=159.12}) was right !-> You just earned 100.0" ||
		feedback.Type != "WIN" {
			t.Fail()
	}
}