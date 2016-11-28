package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var lastBody []byte
var lastOrder Order
var successCount = 0
var failureCount = 0

func main() {
	http.HandleFunc("/quote", handler)
	http.HandleFunc("/feedback", func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Printf("error reading body: %v\n", err)
			rw.WriteHeader(204)
			return
		}

		var feedback Feedback
		json.Unmarshal(body, &feedback)
		if feedback.Type != "WIN" {
			fmt.Printf("Request unsuccesfull\n")
			fmt.Printf("Raw Feedback: %s\n", body)
			fmt.Printf("Feedback: %s\n", feedback)
			fmt.Printf("Raw data %s\n", lastBody)
			fmt.Printf("Got order: %#v\n", lastOrder)
			failureCount += 1
		} else {
			fmt.Printf("Request succesfull\n")
			successCount += 1
		}

		//fmt.Printf(">>> Success ratio= %v\n", failureCount/successCount*100)

		rw.WriteHeader(200)
	})
	http.ListenAndServe(":3000", nil)
}

var CoverRisk = map[string]float64{
	"Basic":   1.7,
	"Extra":   2.4,
	"Premier": 4.2,
}

var CountryRisk = map[string]float64{
	"FR": 1.0,
	"ES": 1.0,
	"BE": 1.1,
	"FI": 0.8,
	"EL": 0.9, //Greece
	"CZ": 1.4,
	"UK": 1.1,
}

const SkiingPrice = 28.0

const ChildrenAge = 15
const SeniorAge = 60

func (order *Order) hasFamilyDiscount() bool {
	if len(order.TravellerAges) >= 4 {
		return false
	}

	adults := 0
	children := 0
	for _, age := range order.TravellerAges {
		if age < ChildrenAge {
			children += 1
		} else if age < SeniorAge {
			adults += 1
		}
	}

	return adults >= 2 && children >= 2
}

func moreChildrenThanAdults(order Order) bool {
	adults := 0
	children := 0

	for _, age := range order.TravellerAges {
		if age < ChildrenAge {
			children += 1
		} else if age < SeniorAge {
			adults += 1
		}
	}

	return children > adults
}

func unknownDiscount(order Order) bool {
	return len(order.TravellerAges) == 2
}

func (order *Order) getAgeRisk() float64 {
	sum := 0.0
	for _, age := range order.TravellerAges {
		if age < 18 {
			sum += 0.5
		} else if age <= 45 {
			sum += 1.0
		} else if age <= 65 {
			sum += 1.2
		} else if age <= 75 {
			sum += 1.4
		} else {
			sum += 2.0
		}
	}

	return sum
}

func calculateQuote(data []byte) Reply {
	var order Order
	err := json.Unmarshal(data, &order)
	if err != nil {
		fmt.Printf("Error while parsing order: %v\n", err)
	}
	lastBody = data
	lastOrder = order
	timeLayout := "2006-01-02"
	returnDate, _ := time.Parse(timeLayout, order.ReturnDate)
	departureDate, _ := time.Parse(timeLayout, order.DepartureDate)
	countryRisk := 1.0

	if order.Country != "" {
		countryRisk = CountryRisk[order.Country]
	}

	numberOfDays := returnDate.Sub(departureDate).Hours() / 24
	quote := CoverRisk[order.Cover] * countryRisk * numberOfDays * order.getAgeRisk()
	if len(order.Options) > 0 && order.Options[0] == "Skiing" {
		quote += SkiingPrice
	}

	if order.hasFamilyDiscount() {
		quote = 0.8 * quote
	}

	if unknownDiscount(order) {
		quote = 0.8 * quote
	}

	/*
		if moreChildrenThanAdults(order) {
			quote = 1.15 * quote
		}
	*/

	return Reply{quote}
}

func handler(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("error reading body: %v\n", err)
		rw.WriteHeader(204)
		return
	}

	fmt.Printf("Got request\n")
	reply := calculateQuote(body)

	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(200)
	encoder := json.NewEncoder(rw)
	encoder.Encode(reply)
}
