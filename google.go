package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/fitness/v1"
	"google.golang.org/api/option"
)

func unixMilliseconds(time time.Time) int64 {
	return time.UnixNano() / 1000000
}

type FitnessPoint struct {
	Start time.Time
	End   time.Time
	Value int64
}

func getFitnessData(token *oauth2.Token, conf *oauth2.Config, start time.Time, end time.Time) (outdata []FitnessPoint, err error) {
	fitnessService, err := fitness.NewService(oauth2.NoContext, option.WithTokenSource(conf.TokenSource(oauth2.NoContext, token)))
	if err != nil {
		return nil, err
	}

	results, err := fitnessService.Users.Dataset.Aggregate(
		"me",
		&fitness.AggregateRequest{
			AggregateBy: []*fitness.AggregateBy{
				{
					DataSourceId: "derived:com.google.step_count.delta:com.google.android.gms:estimated_steps",
				},
			},
			BucketByTime: &fitness.BucketByTime{
				DurationMillis: 3600000 / 2,
			},
			EndTimeMillis: unixMilliseconds(end),
			StartTimeMillis: unixMilliseconds(start),
		},
	).Do()
	if err != nil {
		fmt.Println("boom")
		return nil, err
	}

	for _, b := range results.Bucket {
		for _, ds := range b.Dataset {
			for _, p := range ds.Point {
				for _, v := range p.Value {
					outdata = append(outdata, FitnessPoint{
						Start: time.Unix(0, p.StartTimeNanos),
						End: time.Unix(0, p.EndTimeNanos),
						Value: v.IntVal,
					})
				}
			}
		}
	}

	return outdata, nil
}

type User struct {
    Email string `json:"email"`
}

func fetchEmail(token *oauth2.Token) (*string, error) {
	client := conf.Client(oauth2.NoContext, token)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	if err != nil {
        return nil, err
	}
    defer email.Body.Close()
    data, _ := ioutil.ReadAll(email.Body)
	var user User
	json.Unmarshal(data, &user)
	return &user.Email, nil
}

func fetchAndSaveFitnessData(token *oauth2.Token, conf *oauth2.Config) error {
	end := time.Now().Add(time.Hour * -6)
	start := end.Add(time.Hour * -1)
	return fetchAndSaveFitnessDataWithDates(token, conf, start, end)
}

func fetchAndSaveFitnessDataWithDates(token *oauth2.Token, conf *oauth2.Config, start time.Time, end time.Time) error {
	outdata, err := getFitnessData(token, conf, start, end)
	if err != nil {
		return err
	}
	writePoints(outdata)
	return nil
}
