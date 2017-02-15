package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func main() {

	var influxAddr string
	var influxUsername string
	var influxPassword string
	var influxDatabase string
	var influxMeasurement string
	var influxPrecision string
	var postgresDriver string
	var postgresSource string
	var postgresQuery string

	viper.SetConfigName("app")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	} else {
		influxAddr = viper.GetString("influxdb.addr")
		influxUsername = viper.GetString("influxdb.username")
		influxPassword = viper.GetString("influxdb.password")
		influxDatabase = viper.GetString("influxdb.database")
		influxMeasurement = viper.GetString("influxdb.measurement")
		influxPrecision = viper.GetString("influxdb.precision")

		postgresDriver = viper.GetString("postgresql.driver")
		postgresSource = viper.GetString("postgresql.source")
		postgresQuery = viper.GetString("postgresql.query")
	}

	httpClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxAddr,
		Username: influxUsername,
		Password: influxPassword,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer httpClient.Close()

	q := client.Query{
		Command:  "SELECT * FROM " + influxMeasurement + " ORDER BY time DESC LIMIT 1",
		Database: influxDatabase,
	}

	response, err := httpClient.Query(q)
	if err != nil {
		log.Fatal(err)
		return
	}

	lastDate := "20010101010101"
	if len((response.Results[0].Series)) > 0 {
		t, err := time.Parse(time.RFC3339, response.Results[0].Series[0].Values[0][0].(string))
		if err != nil {
			log.Fatal(err)
		}
		lastDate = t.Format("20060102150405")
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  influxDatabase,
		Precision: influxPrecision,
	})
	if err != nil {
		log.Fatal(err)
	}

	tags := map[string]string{influxMeasurement: influxMeasurement}

	db, err := sql.Open(postgresDriver, postgresSource)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf(postgresQuery, lastDate))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	for rows.Next() {
		for i, _ := range columns {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)

		measure := make(map[string]interface{})

		for i, col := range columns {
			measure[col] = values[i]
		}

		pt, err := client.NewPoint(influxMeasurement, tags, measure, measure["time"].(time.Time))
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)

	}

	err = httpClient.Write(bp)
	if err != nil {
		log.Fatal(err)
	}

}
