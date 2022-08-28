package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/leekchan/timeutil"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

type Event struct {
	Summary     string
	Description string
	date        string
	startTime   string
	endTime     string
}

func createEventData(eventData *Event) *calendar.Event {
	start_datatime := eventData.date + "T" + eventData.startTime + ":00+09:00"
	end_datatime := eventData.date + "T" + eventData.endTime + ":00+09:00"

	event := &calendar.Event{
		Summary:     eventData.Summary,
		Description: eventData.Description,
		Start: &calendar.EventDateTime{
			DateTime: start_datatime,
			TimeZone: "Asia/Tokyo",
		},
		End: &calendar.EventDateTime{
			DateTime: end_datatime,
			TimeZone: "Asia/Tokyo",
		},
	}

	return event
}

func main() {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	t := time.Now()
	var worktime string
	var eventData Event
	print("Click here to enter a Start time and End Time (HH:mm-HH:mm): ")
	fmt.Scan(&worktime)
	print("Click here to enter a Summary: ")
	stdin := bufio.NewScanner(os.Stdin)
	stdin.Scan()
	summary := stdin.Text()

	eventData.Summary = summary
	eventData.Description = "Working Log"
	eventData.date = timeutil.Strftime(&t, "%Y-%m-%d")
	worktimeValue := strings.Split(worktime, "-")
	eventData.startTime = worktimeValue[0]
	eventData.endTime = worktimeValue[1]

	_, err = srv.Events.Insert("primary", createEventData(&eventData)).Do()
	if err != nil {
		log.Fatalf("Unable to create event: %v", err)
	} else {
		fmt.Printf("\x1b[32m%s\x1b[0m", "SUCCESS")
	}
}
