package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
	"fmt"
)

var lastRequest []byte
var lastOrder Order

func main() {
	http.HandleFunc("/quote", handler)
	http.HandleFunc("/feedback", func (rw http.ResponseWriter, req *http.Request) {
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
			fmt.Printf("Feedback: %s\n", feedback)	
		}

		rw.WriteHeader(200)
	})
	http.ListenAndServe(":3000", nil)
}

var CoverRisk = map[string]float64{
	"Basic":  1.8,
	"Extra": 2.4,
	"Premier": 4.2,
}

var CountryRisk = map[string]float64{
	"FR": 1.0,
	"ES": 1.0,
	"BE": 1.1,
	"FI": 0.8,
	"EL": 0.9, //Greece
}

const SkiingPrice = 24.0

func isFamilyDiscount(order Order) bool {
	if len(order.TravellerAges) != 4 {
		return false
	}

	adults := 0
	children := 0
	for _, age := range order.TravellerAges {
		if age < 18 {
			children += 1
		} else {
			adults += 1
		}
	}

	return adults == 2 && children == 2
}

func getAgeRisk(order Order) float64 {
	sum := 0.0
	for _, age := range order.TravellerAges {
		if  age < 18 {
			sum += 0.5
		} else if age <=45 {
			sum += 1.0
		} else if  age <= 65 {
			sum += 1.2
		} else if age <= 75 {
			sum +=  1.4 
		} else {
			sum += 2.0
		}
	}

	return sum
}

func calculateQuote(data []byte) Reply {
	var order Order
	json.Unmarshal(data, &order)
	timeLayout := "2006-01-02"
	returnDate, _ :=  time.Parse(timeLayout, order.ReturnDate)
	departureDate, _ := time.Parse(timeLayout, order.DepartureDate)
	countryRisk := 1.0
	if order.Country != "" {
		countryRisk = CountryRisk[order.Country]
	}

	numberOfDays := returnDate.Sub(departureDate).Hours()/24
	fmt.Printf("Number of days %v: \n", numberOfDays)
	quote := CoverRisk[order.Cover] * countryRisk * numberOfDays * getAgeRisk(order)
	if len(order.Options) > 0 && order.Options[0] == "Skiing" {
		quote += SkiingPrice
	}

	if isFamilyDiscount(order) {
		quote = 0.8 * quote
	}
	//quote := 0.0

	fmt.Printf("Raw data %s\n", data)
	fmt.Printf("Got order: %#v\n", order)

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

	reply := calculateQuote(body)

	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(200)
	encoder := json.NewEncoder(rw)
	encoder.Encode(reply)
}
