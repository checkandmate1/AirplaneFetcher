package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"unicode"

	"github.com/joho/godotenv"
	"golang.org/x/net/html"
)

type DepartureAirline struct {
	ICAO  string
	Fleet string
}

type Departure struct {
	Exit                string             `json:"exit"`
	Destination         string             `json:"destination"`
	Altitude            int                `json:"altitude"`
	Route               string             `json:"route"`
	Airlines            []DepartureAirline `json:"airlines"`
	Scratchpad          string             `json:"scratchpad,omitempty"`
	SecondaryScratchpad string             `json:"secondary_scratchpad,omitempty"`
}

type CallsignOutput struct {
	Airline      string
	ICAOCallsign string
}

func main() {
	// Define flags
	printFlag := flag.String("airport", "", "airport to fetch")

	flag.Parse()
	if *printFlag == "" {
		flag.Usage()
		os.Exit(1)
	}
	getDepartureCallsigns(*printFlag)

}

func flightAwareNonsense(callsigns []CallsignOutput) {
	departures := []Departure{}
	scRules := ScratchpadRules{}
	var scratchpads bool = true
	file, err := os.ReadFile("scratchpadRules.json")
	if err != nil {
		scratchpads = false
	}
	json.Unmarshal(file, &scRules)
	for _, aircraft := range callsigns {
		url := fmt.Sprintf("https://www.flightaware.com/live/flight/%v", aircraft.ICAOCallsign)
		resp, err := http.Get(url)
		fmt.Printf("http request for %v done\n", aircraft.ICAOCallsign)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		doc, err := html.Parse(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Printf("http parsed for %v\n", aircraft.ICAOCallsign)
		r := renderNode(doc)
		r, err = cleanUpString(r)
		if err != nil {
			fmt.Println(err)
			continue
		}

		c := strings.Index(r, `"activityLog":{`)
		r = r[c+14:]
		r += "]}"
		f := FlightAwareResponse{}

		json.Unmarshal([]byte(r), &f)
		openscope, _ := parseAirlines()

		for _, flight := range f.Flights {

			if flight.FlightStatus != "" {

				d := Departure{}
				fleet := getFleet(openscope, flight.Aircraft.Type, aircraft.Airline)
				if fleet == "" {
					continue
				}
				d.Airlines = []DepartureAirline{
					DepartureAirline{
						ICAO:  aircraft.Airline,
						Fleet: fleet,
					},
				}

				if flight.FlightPlan.Altitude == nil {
					continue
				}

				d.Altitude = int(flight.FlightPlan.Altitude.(float64) * 100)
				d.Destination = flight.Destination.Icao
				d.Route = flight.FlightPlan.Route
				waypointArray := strings.Split(flight.FlightPlan.Route, " ")
				if unicode.IsDigit(rune(waypointArray[0][len(waypointArray[0])-1])) {
					waypointArray = waypointArray[1:]
				}
				d.Exit = waypointArray[0]
				if scratchpads {
					for _, rule := range scRules.Rules {
						if rule.Exit == d.Exit {
							if rule.Scratchpad != "" {
								d.Scratchpad = rule.Scratchpad
							}
							if rule.SecondaryScratchpad != "" {
								d.Scratchpad = rule.SecondaryScratchpad
							}
						}
					}
				}
				departures = append(departures, d)
				break
			}
		}
	}
	f, err := json.Marshal(departures)
	if err != nil {
		panic(err)
	}
	e, _ := os.Create("output.json")
	e.Write(f)
}

func getFleet(ac map[string]Airlines, acType, airline string) string {
	info := ac[airline]
	for fleet, x := range info.Fleets {
		for _, aircraft := range x {
			if acType == aircraft.ICAO {
				return fleet
			}
		}
	}
	return ""
}

func renderNode(node *html.Node) string {
	var result string

	if node.Type == html.TextNode {
		result += node.Data
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		result += renderNode(c)
	}

	return result
}

func getDepartureCallsigns(airport string) {
	var err error
	var apiKey string
	err = godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	apiKey = os.Getenv("API_KEY")
	resp, err := http.Get(fmt.Sprintf("http://api.aviationstack.com/v1/flights?access_key=%v&dep_icao=%v", apiKey, airport))
	if err != nil {
		panic(err)
	}
	// Turn into JSON

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	r := Response{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		panic(err)
	}
	output := []CallsignOutput{}
	for _, f := range r.Data {
		output = append(output, CallsignOutput{
			Airline:      f.Airline.ICAO,
			ICAOCallsign: f.Flight.ICAO,
		})
	}
	fmt.Println("callsigns obtained. \n ", output)
	flightAwareNonsense(output)
}

func cleanUpString(text string) (string, error) {
	h := strings.Index(text, `"version"`)
	if h == -1 {
		fmt.Println("whhops")
		os.Exit(1)
	}
	text = text[h-1:]
	// Count occurrences of "origin"
	count := strings.Count(text, "origin")

	// Check if there are at least three occurrences
	if count < 3 {
		return "", fmt.Errorf("there are less than three occurrences of 'origin' in the string")
	}
	for i := 0; i < 3; i++ {
		n := strings.LastIndex(text, "origin")
		text = text[:n]
	}
	text = text[:len(text)-3]
	return text, nil
}

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Total  int `json:"total"`
}

type APIDeparture struct {
	Airport         string      `json:"airport"`
	Timezone        string      `json:"timezone"`
	IATA            string      `json:"iata"`
	ICAO            string      `json:"icao"`
	Terminal        interface{} `json:"terminal"`
	Gate            interface{} `json:"gate"`
	Delay           interface{} `json:"delay"`
	Scheduled       string      `json:"scheduled"`
	Estimated       string      `json:"estimated"`
	Actual          interface{} `json:"actual"`
	EstimatedRunway interface{} `json:"estimated_runway"`
	ActualRunway    interface{} `json:"actual_runway"`
}

type Arrival struct {
	Airport         string      `json:"airport"`
	Timezone        string      `json:"timezone"`
	IATA            string      `json:"iata"`
	ICAO            string      `json:"icao"`
	Terminal        interface{} `json:"terminal"`
	Gate            interface{} `json:"gate"`
	Baggage         interface{} `json:"baggage"`
	Delay           interface{} `json:"delay"`
	Scheduled       string      `json:"scheduled"`
	Estimated       string      `json:"estimated"`
	Actual          interface{} `json:"actual"`
	EstimatedRunway interface{} `json:"estimated_runway"`
	ActualRunway    interface{} `json:"actual_runway"`
}

type Airline struct {
	Name string `json:"name"`
	IATA string `json:"iata"`
	ICAO string `json:"icao"`
}

type Flight struct {
	Number     string      `json:"number"`
	IATA       string      `json:"iata"`
	ICAO       string      `json:"icao"`
	Codeshared interface{} `json:"codeshared"`
}

type Data struct {
	FlightDate   string       `json:"flight_date"`
	FlightStatus string       `json:"flight_status"`
	Departure    APIDeparture `json:"departure"`
	Arrival      Arrival      `json:"arrival"`
	Airline      Airline      `json:"airline"`
	Flight       Flight       `json:"flight"`
	Aircraft     interface{}  `json:"aircraft"`
	Live         interface{}  `json:"live"`
}

type Response struct {
	Pagination Pagination `json:"pagination"`
	Data       []Data     `json:"data"`
}

type FlightAwareResponse struct {
	Flights []struct {
		Origin struct {
			Tz                    string    `json:"TZ"`
			IsValidAirportCode    bool      `json:"isValidAirportCode"`
			IsCustomGlobalAirport bool      `json:"isCustomGlobalAirport"`
			AltIdent              any       `json:"altIdent"`
			Iata                  string    `json:"iata"`
			FriendlyName          string    `json:"friendlyName"`
			FriendlyLocation      string    `json:"friendlyLocation"`
			Coord                 []float64 `json:"coord"`
			IsLatLon              bool      `json:"isLatLon"`
			Icao                  string    `json:"icao"`
			Gate                  any       `json:"gate"`
			Terminal              any       `json:"terminal"`
			Delays                any       `json:"delays"`
		} `json:"origin"`
		Destination struct {
			Tz                    string    `json:"TZ"`
			IsValidAirportCode    bool      `json:"isValidAirportCode"`
			IsCustomGlobalAirport bool      `json:"isCustomGlobalAirport"`
			AltIdent              any       `json:"altIdent"`
			Iata                  string    `json:"iata"`
			FriendlyName          string    `json:"friendlyName"`
			FriendlyLocation      string    `json:"friendlyLocation"`
			Coord                 []float64 `json:"coord"`
			IsLatLon              bool      `json:"isLatLon"`
			Icao                  string    `json:"icao"`
			Gate                  any       `json:"gate"`
			Terminal              any       `json:"terminal"`
			Delays                any       `json:"delays"`
		} `json:"destination"`
		AircraftType         string `json:"aircraftType"`
		AircraftTypeFriendly string `json:"aircraftTypeFriendly"`
		FlightID             string `json:"flightId"`
		TakeoffTimes         struct {
			Scheduled int `json:"scheduled"`
			Estimated int `json:"estimated"`
			Actual    any `json:"actual"`
		} `json:"takeoffTimes"`
		LandingTimes struct {
			Scheduled int `json:"scheduled"`
			Estimated int `json:"estimated"`
			Actual    any `json:"actual"`
		} `json:"landingTimes"`
		GateDepartureTimes struct {
			Scheduled int `json:"scheduled"`
			Estimated any `json:"estimated"`
			Actual    any `json:"actual"`
		} `json:"gateDepartureTimes"`
		GateArrivalTimes struct {
			Scheduled int `json:"scheduled"`
			Estimated any `json:"estimated"`
			Actual    any `json:"actual"`
		} `json:"gateArrivalTimes"`
		Ga                   bool   `json:"ga"`
		FlightStatus         string `json:"flightStatus"`
		FpasAvailable        bool   `json:"fpasAvailable"`
		CanEdit              bool   `json:"canEdit"`
		Cancelled            bool   `json:"cancelled"`
		ResultUnknown        bool   `json:"resultUnknown"`
		Diverted             bool   `json:"diverted"`
		Adhoc                bool   `json:"adhoc"`
		FruOverride          bool   `json:"fruOverride"`
		Timestamp            any    `json:"timestamp"`
		RoundedTimestamp     any    `json:"roundedTimestamp"`
		PermaLink            string `json:"permaLink"`
		TaxiIn               any    `json:"taxiIn"`
		TaxiOut              any    `json:"taxiOut"`
		GlobalIdent          bool   `json:"globalIdent"`
		GlobalFlightFeatures bool   `json:"globalFlightFeatures"`
		GlobalVisualizer     bool   `json:"globalVisualizer"`
		FlightPlan           struct {
			Speed           int    `json:"speed"`
			Altitude        any    `json:"altitude"`
			Route           string `json:"route"`
			DirectDistance  int    `json:"directDistance"`
			PlannedDistance any    `json:"plannedDistance"`
			Departure       int    `json:"departure"`
			Ete             int    `json:"ete"`
			FuelBurn        struct {
				Gallons int `json:"gallons"`
				Pounds  int `json:"pounds"`
			} `json:"fuelBurn"`
		} `json:"flightPlan"`
		Links struct {
			Operated           string `json:"operated"`
			Registration       string `json:"registration"`
			Permanent          string `json:"permanent"`
			TrackLog           string `json:"trackLog"`
			FlightHistory      string `json:"flightHistory"`
			BuyFlightHistory   string `json:"buyFlightHistory"`
			ReportInaccuracies string `json:"reportInaccuracies"`
			Facebook           string `json:"facebook"`
			Twitter            string `json:"twitter"`
		} `json:"links"`
		Aircraft struct {
			Type          string `json:"type"`
			Lifeguard     bool   `json:"lifeguard"`
			Heavy         bool   `json:"heavy"`
			Tail          any    `json:"tail"`
			Owner         any    `json:"owner"`
			OwnerLocation any    `json:"ownerLocation"`
			OwnerType     any    `json:"owner_type"`
			CanMessage    bool   `json:"canMessage"`
			FriendlyType  string `json:"friendlyType"`
			TypeDetails   struct {
				Manufacturer string `json:"manufacturer"`
				Model        string `json:"model"`
				Type         string `json:"type"`
				EngCount     string `json:"engCount"`
				EngType      string `json:"engType"`
			} `json:"typeDetails"`
		} `json:"aircraft"`
		DisplayIdent       string `json:"displayIdent"`
		EncryptedFlightID  string `json:"encryptedFlightId"`
		PredictedAvailable bool   `json:"predictedAvailable"`
		PredictedTimes     struct {
			Out any `json:"out"`
			Off any `json:"off"`
			On  any `json:"on"`
			In  any `json:"in"`
		} `json:"predictedTimes"`
	} `json:"flights"`
}

type ScratchpadRules struct {
	Rules []struct {
		Exit                string `json:"exit,omitempty"`
		Scratchpad          string `json:"scratchpad,omitempty"`
		SecondaryScratchpad string `json:"secondary_scratchpad,omitempty"`
	} `json:"rules,omitempty"`
}
