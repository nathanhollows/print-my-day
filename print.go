package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/jubnzv/go-taskwarrior"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// IndexHandler serves the front page
func IndexHandler(w http.ResponseWriter, r *http.Request) error {
	data := make(map[string]interface{})
	data["url"] = os.Getenv("PRINT_URL")
	data["data"] = getTasks()
	data["events"] = getEvents()
	data["date"] = time.Now().Format("Mon January 2")
	return render(w, data, "index.html")
}

// PrintHandler serves the json to print
func PrintHandler(w http.ResponseWriter, r *http.Request) error {
	data := make(map[string]interface{})
	w.Header().Set("Content-Type", "application/json")
	data["data"] = getTasks()
	data["events"] = getEvents()
	data["date"] = time.Now().Format("Mon January 2")
	return render(w, data, "print.html")
}

type event struct {
	Time    string
	Summary string
}

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

func getTasks() []taskwarrior.Task {
	tw, _ := taskwarrior.NewTaskWarrior("~/.taskrc")
	tw.FetchAllTasks()
	tasks := tw.Tasks
	for i := len(tasks) - 1; i >= 0; i-- {
		task := tasks[i]
		if task.Status != "pending" {
			tasks = append(tasks[:i],
				tasks[i+1:]...)
		}
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Urgency > tasks[j].Urgency
	})
	return tasks
}

func getEvents() []event {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	year, month, day := time.Now().Date()
	t := time.Date(year, month, day, 0, 0, 0, 0, time.Local).Format(time.RFC3339)
	end := time.Date(year, month, day, 23, 0, 0, 0, time.Local).Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).TimeMax(end).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	eventList := []event{}
	if len(events.Items) != 0 {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = "    *"
			} else {
				eventTime, _ := time.Parse(time.RFC3339, item.Start.DateTime)
				date = eventTime.Format("03:04")
			}
			eventList = append(eventList, event{Time: date, Summary: item.Summary})

		}
	}
	return eventList
}
