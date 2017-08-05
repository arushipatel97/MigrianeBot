package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/briandowns/openweathermap"
)

var TempMapYes map[float64]ThreadUnsafeSet
var TempMapNo map[float64]ThreadUnsafeSet

func main() {
	ws, _ := slackConnect("xoxb-222678512498-5NgtV8IJ7Nkjqq3j6Jc1zKdp")
	fmt.Println("Ctrl^C to quit")

	// //make call to slack API to get lists of users on channel
	// url := "https://slack.com/api/users.list?token=784QiOKi30552hslZ4WeqFC3"
	// resp, err := http.Get(url)
	// if err != nil {
	// 	log.Println("error received for Slack GET", err.Error())
	// 	return
	// }

	// error-handling api request failure
	// if resp.StatusCode != 200 {
	// 	log.Println("non-200 status code received")
	// 	return
	// }
	//
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Println("error received decoding JSON", err.Error())
	// 	return
	// }
	// err = resp.Body.Close()
	// if err != nil {
	// 	log.Println("error closing body", err.Error())
	// 	return
	// }
	//
	// var respUser respMembers
	// if err = json.Unmarshal(body, &respUser); err != nil {
	// 	log.Println("error on JSON unmarshal", err.Error())
	// 	return
	// }

	TempMapYes = make(map[float64]ThreadUnsafeSet)
	TempMapNo = make(map[float64]ThreadUnsafeSet)
	//yes, no := New(), New()

	//main loop: will continuely run and montior number of messages
	for {
		message, err := getMessage(ws)
		fmt.Printf(message.Text)
		if err != nil {
			log.Println("error getting message")
		}
		parts := strings.Split(message.Text, " ")
		fmt.Printf(message.Text)
		if parts[0] == "weather" {
			weather, err := GetWeather(parts[1])
			if err != nil {
				fmt.Println("Could not get weather:", err)
				return
			}
			description := ""
			if len(weather.Weather) > 0 {
				description = weather.Weather[0].Description
			}
			message.Text = fmt.Sprintf("The current temperature for %s is %.0f degrees farenheight (%s)", weather.Name, weather.Main.Temp, description)
			err = postMessage(ws, message)
			if err != nil {
				log.Println("error posting message")
			}
			if parts[2] == "?" {
				prediction := predict(weather.Main.Temp, description)
				message.Text = fmt.Sprintf("Migraine is %s", prediction)
				err = postMessage(ws, message)
				if err != nil {
					log.Println("error posting message")
				}
			}
			recordMigraine(parts[2], weather.Main.Temp, description)
		}
	}
}

func recordMigraine(migraineBool string, temp float64, description string) {
	if migraineBool == "yes" {
		currentSet := TempMapYes[temp]
		currentSet.Add(description)
	}
	if migraineBool == "no" {
		currentSet := TempMapNo[temp]
		currentSet.Add(description)
	}
}

func predict(temp float64, description string) (resp string) {
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
			resp = "Mixed Data"
		}
	case -1:
		resp = "Unlikely"
	case -2:
		resp = "Very Unlikely"
	}
	return resp
}

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

func GetWeather(place string) (*openweathermap.CurrentWeatherData, error) {
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
