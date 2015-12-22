package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var count int = 0

type GoogleMapApiResult struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

type NewLocationRequest struct {
	Address string `json:"address"`
	City    string `json:"city"`
	Name    string `json:"name"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

type NewLocationResponse struct {
	ID         int    `json:"id" bson:"_id"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	Zip        string `json:"zip"`
	Coordinate struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinate"`
}

type GetLocationResponse struct {
	ID         int    `json:"id" bson:"_id"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	Zip        string `json:"zip"`
	Coordinate struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinate"`
}

type PutLocationRequest struct {
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}


func createLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	
	locreq := NewLocationRequest{}

	
	json.NewDecoder(req.Body).Decode(&locreq)

	
	lat, lng := getLatLong(getAPIUrl(locreq))

	
	locres := NewLocationResponse{}
	locres.Address = locreq.Address
	locres.City = locreq.City
	locres.State = locreq.State
	locres.ID = getCount()
	locres.Name = locreq.Name
	locres.Zip = locreq.Zip
	locres.Coordinate.Lat = lat
	locres.Coordinate.Lng = lng

	
	resJson, _ := json.Marshal(locres)

	
	addLocationToDB(locres)

	
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(201)
	fmt.Fprintf(rw, "%s", resJson)
}

//GET
func getLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	
	id := p.ByName("location_id")
	locationId, _ := strconv.Atoi(id)

	
	getres := GetLocationResponse{}

	

	getres = getLocationFromDB(locationId)

	
	resJson, _ := json.Marshal(getres)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(200)
	fmt.Fprintf(rw, "%s", resJson)

}

//PUT
func putLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	
	id := p.ByName("location_id")
	locationId, _ := strconv.Atoi(id)

	
	putreq := PutLocationRequest{}

	
	json.NewDecoder(req.Body).Decode(&putreq)

	putres := updateLocationInDB(locationId, putreq)

	
	resJson, _ := json.Marshal(putres)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(201)
	fmt.Fprintf(rw, "%s", resJson)

}

//DELETE
func deleteLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	
	id := p.ByName("location_id")
	locationId, _ := strconv.Atoi(id)

	
	session, err1 := mgo.Dial(getDatabaseURL())

	
	if err1 != nil {
		fmt.Println("Error in database connection")
		os.Exit(1)
	} else {
		fmt.Println("Session Created")
	}

	
	err2 := session.DB("locationdb").C("locationCollection").RemoveId(locationId)
	if err2 != nil {
		panic(err2)
	} else {
		fmt.Println("Location deleted : " + id)
	}

	rw.WriteHeader(200)
}

func addLocationToDB(locres NewLocationResponse) {

	
	session, err1 := mgo.Dial(getDatabaseURL())

	
	if err1 != nil {
		fmt.Println("Error in database connection")
		os.Exit(1)
	} else {
		fmt.Println("Session Created")
	}

	
	session.DB("locationdb").C("locationCollection").Insert(locres)

	
	session.Close()
	fmt.Println("Session Closed")

}

func getLocationFromDB(locationId int) GetLocationResponse {

	
	session, err1 := mgo.Dial(getDatabaseURL())

	
	if err1 != nil {
		fmt.Println("Error in database connection")
		os.Exit(1)
	} else {
		fmt.Println("Session Created")
	}

	getres := GetLocationResponse{}

	err2 := session.DB("locationdb").C("locationCollection").FindId(locationId).One(&getres)

	if err2 != nil {
		panic(err2)
	} else {
		fmt.Println("Location retrieved from Location DB")
	}

	
	session.Close()
	fmt.Println("Session Closed")

	return getres

}

func updateLocationInDB(locationId int, plr PutLocationRequest) GetLocationResponse {

	plres := GetLocationResponse{}

	//connect to mongo db
	session, err1 := mgo.Dial(getDatabaseURL())

	
	if err1 != nil {
		fmt.Println("Error in database connection")
		os.Exit(1)
	} else {
		fmt.Println("Session Created")
	}

	err2 := session.DB("locationdb").C("locationCollection").FindId(locationId).One(&plres)

	if err2 != nil {
		panic(err2)
	} else {
		fmt.Println("Location retrieved from TripPlanner DB")
	}

	plres.Address = plr.Address
	plres.City = plr.City
	plres.State = plr.State
	plres.Zip = plr.Zip

	nltemp := NewLocationRequest{}
	nltemp.Address = plres.Address
	nltemp.City = plres.City
	nltemp.State = plres.State
	nltemp.Zip = plres.Zip
	nltemp.Name = plres.Name

	
	lat, lng := getLatLong(getAPIUrl(nltemp))

	
	plres.Coordinate.Lat = lat
	plres.Coordinate.Lng = lng

	err3 := session.DB("locationdb").C("locationCollection").UpdateId(locationId, plres)

	if err3 != nil {
		panic(err3)
	} else {
		fmt.Println("Location updated in Location DB")
	}

	
	session.Close()
	fmt.Println("Session Closed")

	return plres

}

func getAPIUrl(locreq NewLocationRequest) string {

	var address string = locreq.Address
	address = strings.Replace(address, " ", "+", -1)

	var city string = locreq.City
	city = strings.Replace(city, " ", "+", -1)
	city = ",+" + city

	var state string = locreq.State
	state = strings.Replace(state, " ", "+", -1)
	state = ",+" + state

	var zip string = locreq.Zip
	zip = strings.Replace(zip, " ", "+", -1)
	zip = "+" + zip

	
	var urlPart1 string = "http://maps.google.com/maps/api/geocode/json?address="
	var urlPart2 string = address + city + state + zip
	var urlPart3 string = "&sensor=false"

	var url string = urlPart1 + urlPart2 + urlPart3
	

	return url
}

func getLatLong(url string) (float64, float64) {

	result := GoogleMapApiResult{}

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error in getting response from google maps api", err.Error())
		os.Exit(1)
	}

	json.NewDecoder(response.Body).Decode(&result)


	lat := result.Results[0].Geometry.Location.Lat
	lng := result.Results[0].Geometry.Location.Lng

	return lat, lng

}

func getCount() int {
	count = count + 1
	return count
}

func getDatabaseURL() string {
	var dburl string = "mongodb://Pulkit:12345@ds045464.mongolab.com:45464/locationdb"
	return dburl
}

func main() {
	mux := httprouter.New()
	mux.GET("/locations/:location_id", getLocation)
	mux.POST("/locations", createLocation)
	mux.PUT("/locations/:location_id", putLocation)
	mux.DELETE("/locations/:location_id", deleteLocation)
	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
