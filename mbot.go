package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/briandowns/openweathermap"
)

var TempMapYes = make(map[int]ThreadUnsafeSet)
var TempMapNo = make(map[int]ThreadUnsafeSet)

func main() {

	ws, _ := slackConnect(SLACK_TOKEN)
	fmt.Println("Ctrl^C to quit")

	//main loop: will continuely run and montior messages, looking for a
	// weather/migriane request
	for {
		//reads message on slack
		message, err := getMessage(ws)
		if err != nil {
			log.Println("error getting message")
		}
		parts := strings.Split(message.Text, " ")
		//only continue functionality if related to weather/migraine
		if parts[0] == "weather" {
			weather, err := getWeather(parts[1])
			if err != nil {
				log.Println("Could not get weather:", err)
				return
			}
			description := ""
			if len(weather.Weather) > 0 {
				description = weather.Weather[0].Description
			}
			temp := int(weather.Main.Temp)
			message.Text = fmt.Sprintf("The current temperature for %s is %d degrees farenheight (%s)", weather.Name, temp, description)
			err = postMessage(ws, message)
			if err != nil {
				log.Println("error posting message")
			}
			fmt.Println(len(parts))
			if len(parts) > 2 {
				if parts[2] == "migraine:?" {
					//estbalish prediction based on "weather zipcode ?" slack message
					prediction := predict((temp / 10), description)
					message.Text = fmt.Sprintf("Migraine is %s", prediction)
					err = postMessage(ws, message)
					if err != nil {
						log.Println("error posting message")
					}
				}
				//further populates data structures based on "weather zipcode yes/no" slack message
				recordMigraine(parts[2], (temp / 10), description)
			}
		}
	}
}

//updates map/sets with new information to make more accurate predictions
func recordMigraine(migraineBool string, temp int, description string) {
	if migraineBool == "migraine:yes" {
		if TempMapYes[temp] == nil {
			TempMapYes[temp] = New()
		}
		currentSet := TempMapYes[temp]
		currentSet.Add(description)
	}
	if migraineBool == "migriane:no" {
		if TempMapYes[temp] == nil {
			TempMapYes[temp] = New()
		}
		currentSet := TempMapNo[temp]
		currentSet.Add(description)
	}
}

//looks through both maps to predict likeliness of migraine
func predict(temp int, description string) (resp string) {
	yesNum := predictSpecific(TempMapYes[temp], description)
	noNum := predictSpecific(TempMapNo[temp], description)
	total := yesNum + noNum
	switch total {
	case 2:
		resp = "Very Likely"
	case 1:
		resp = "Likely"
	case 0:
		if yesNum == 0 {
			resp = "Unkown"
		} else {
			resp = "Mixed Data- could go either way"
		}
	case -1:
		resp = "Unlikely"
	case -2:
		resp = "Very Unlikely"
	}
	return resp
}

//calculates value based on if the temp is in the map specified and it the
//set of the temp specified contains the description string
func predictSpecific(set ThreadUnsafeSet, description string) (resp int) {
	if set.Contains(description) {
		resp = 2 //temp in map & description in set
	} else if set.Cardinality() == 0 {
		resp = 0 //temp hasn't caused before
	} else {
		resp = 1 //at least temp has come up before (in map)
	}
	return
}

//calls openweathermap to retrieve weather based on zipcode specified
func getWeather(place string) (*openweathermap.CurrentWeatherData, error) {
	w, err := openweathermap.NewCurrent("F", "en")
	if err != nil {
		return nil, fmt.Errorf("Could not get weather: %s", err)
	}
	err = w.CurrentByName(place)
	if err != nil {
		return nil, fmt.Errorf("Weather fetch fail: %s", err)
	}
	return w, nil
}
