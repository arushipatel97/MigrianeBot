/*
Many times migraines are caused/escalated by changes in the weather, & migriane
sufferers are told to make a log of factors such as the weather to help further
predict when the person will face a migriane.

Migraine-bot is a slackbot that tells one the weather for a given city,Country,
by formatting a slack message as "weather/city,country".
The user can specify whether or not the user experienced a migrane, by
formatting a slack message as "weather/city,country/migraine:(yes/no)".
Furthermore the user can ask, based on previous logging, if a migraine is
expected for a given day, by formatting a slack message as
"weather/city,country/migraine?" The more a user inputs data, the more likely
it will be able to produce accurate results, given the migrianes are in fact
weather related.

There is no guarantee on the accuracy of the prediction. This is solely meant
as a means to better personal logging and predicitng for migraine sufferers.

The following code can really be used for any condition by changing the
condition variable
*/

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/briandowns/openweathermap"
)

//The data regarding the weather/prediction is stored in two maps(one for
//experincing symptoms and the other for not). Each map has the floor division
// of the temperature (ex: 60-69deg stored as key 6), for which the value will
// be a set of the description (ex: sunny, rainy, etc).
var TempMapYes = make(map[int]ThreadUnsafeSet)
var TempMapNo = make(map[int]ThreadUnsafeSet)

const condition = "migraine"

//The following code it constantly reading slack messages waiting for a request
//for the weather (& prediction), will parse the request and will post the
//weather and prediction (based on the request).
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
		parts := strings.Split(message.Text, "/")
		//only continue functionality if related to weather/condition(migraine)
		if parts[0] == "weather" {
			description, temp, post, err := getWeather(parts[1])
			if err != nil {
				log.Println("Could not get weather:", err)
				return
			}
			message.Text = post
			err = postMessage(ws, message)
			if err != nil {
				log.Println("error posting message")
			}
			if len(parts) > 2 {

				if parts[2] == fmt.Sprintf("%s?", condition) {
					//estbalish prediction based on "weather/city,country/condition(migraine)?" slack message
					prediction := predict((temp / 10), description)
					message.Text = fmt.Sprintf("%s is %s", condition, prediction)
					err = postMessage(ws, message)
					if err != nil {
						log.Println("error posting message")
					}
				}
				//further populates data structures based on "weather/city,Country/condition(migraine):(yes/no)" slack message
				recordCondition(parts[2], (temp / 10), description)

			}
		}
	}
}

//updates map/sets with new information to make more accurate predictions
func recordCondition(conditionBool string, temp int, description string) {
	if conditionBool == fmt.Sprintf("%s:yes", condition) {
		if TempMapYes[temp] == nil {
			TempMapYes[temp] = New()
		}
		currentSet := TempMapYes[temp]
		currentSet.Add(description)
	}
	if conditionBool == fmt.Sprintf("%s:no", condition) {
		if TempMapNo[temp] == nil {
			TempMapNo[temp] = New()
		}
		currentSet := TempMapNo[temp]
		currentSet.Add(description)
	}
}

//Looks through both maps to predict likeliness of condition, based on if the
//temperature is in either map and moreover if the specific description is in
//set for the specified temperature.
func predict(temp int, description string) (resp string) {
	yesNum := predictSpecific(TempMapYes[temp], description)
	noNum := predictSpecific(TempMapNo[temp], description)
	fmt.Println(yesNum)
	fmt.Println(noNum)
	total := yesNum - noNum
	switch total {
	case 2:
		resp = "Very Likely"
	case 1:
		resp = "Likely"
	case 0:
		fmt.Println(yesNum)
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

//calls openweathermap to retrieve weather based on city,Country specified
//returns parsed format of information obtained
func getWeather(place string) (string, int, string, error) {
	w, err := openweathermap.NewCurrent("F", "en")
	if err != nil {
		return "", 0, "", fmt.Errorf("Could not get weather: %s", err)
	}
	err = w.CurrentByName(place)
	if err != nil {
		return "", 0, "", fmt.Errorf("Weather fetch fail: %s", err)
	}
	var description string
	if len(w.Weather) > 0 {
		description = string(w.Weather[0].Description)
	}
	temp := int(w.Main.Temp)
	post := fmt.Sprintf("Current temperature in %s is %d degrees farenheight (%s)", w.Name, temp, description)
	return description, temp, post, nil
}
