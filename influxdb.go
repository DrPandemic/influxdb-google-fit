package main

import (
	"context"
	"fmt"
	"os"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func writePoints(data []FitnessPoint) {
    client := influxdb2.NewClient(os.Getenv("INFLUXDB_URL"), os.Getenv("INFLUXDB_TOKEN"))
    writeAPI := client.WriteAPIBlocking(os.Getenv("INFLUXDB_ORG"), os.Getenv("INFLUXDB_BUCKET"))

	for _, element := range data {
		p := influxdb2.NewPointWithMeasurement("fit").
			AddField("count", element.Value).
			AddTag("user", *currentEmail).
			SetTime(element.Start)
		err := writeAPI.WritePoint(context.Background(), p)
		if err != nil {
			fmt.Println("Failed to save point", err)
		}
	}
}
