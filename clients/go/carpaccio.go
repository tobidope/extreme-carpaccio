package main

type Order struct {
	DepartureDate string `json:"departureDate"`
	ReturnDate string `json:"returnDate"`
	TravellerAges []int `json:"travellerAges"`
	Cover string `json:"cover"`
	Country string `json:"country"`
	Options []string `json:"options"`
}

type Reply struct {
	Quote float64 `json:"quote"`
}
