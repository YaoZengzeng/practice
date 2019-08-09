package main

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	user     = "root"
	password = "123456"
	address  = "127.0.0.1"
	port     = "3306"
	db       = "alert"
)

type AlertItem struct {
	Id           string
	Alertname    string
	Serverity    string
	Resourcetype string
	Source       string
	Info         string
	Start        time.Time
	End          time.Time

	// Optional fields
	Organization string
	Project      string
	Cluster      string
	Namespace    string
	Node         string
	Pod          string
	Deployment   string
	Statefulset  string

	// Extend fileds, all other labels will be marshalled into this field.
	Extend string
}

// The alert item stored in db will include `counter` field, so wrap it with AlertDBItem.
type AlertDBItem struct {
	AlertItem
	Count int
}

type DB struct {
	*sqlx.DB
}

// updateAlert update the alert in db directly.
func (db *DB) updateAlert(alert AlertDBItem) error {
	_, err := db.NamedExec("UPDATE alerts SET count=:count, end=:end WHERE id=:id", alert)
	if err != nil {
		return err
	}
	return nil
}

// insertAlert insert the alert to db directly.
func (db *DB) insertAlert(alert AlertDBItem) error {
	_, err := db.NamedExec("INSERT INTO alerts VALUES (:id, :alertname, :serverity, :resourcetype, :source, :info, :count, :start, :end, :organization, :project, :cluster, :namespace, :node, :pod, :deployment, :statefulset, :extend)", alert)
	if err != nil {
		return err
	}
	return nil
}

// queryAlert query the matching alerts from db directly.
func (db *DB) queryAlert(alert AlertDBItem) ([]AlertDBItem, error) {
	res := []AlertDBItem{}
	nstmt, err := db.PrepareNamed(`SELECT * FROM alerts WHERE alertname REGEXP :alertname and serverity REGEXP :serverity and resourcetype REGEXP :resourcetype and source REGEXP :source and organization REGEXP :organization
					and project REGEXP :project and cluster REGEXP :cluster and namespace REGEXP :namespace and node REGEXP :node and pod REGEXP :pod and deployment REGEXP :deployment and statefulset REGEXP :statefulset
					and extend REGEXP :extend`)
	err = nstmt.Select(&res, alert)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (db *DB) InsertAlert(alert AlertItem) error {
	item := AlertDBItem{}

	// Check if the AlertItem exists.
	err := db.Get(&item, "SELECT * from alerts WHERE id = ? and end > ?", alert.Id, alert.Start)
	// If return error, we assume that there is no overlapping item.
	if err != nil {
		item = AlertDBItem{AlertItem: alert, Count: 1}
		return db.insertAlert(item)
	}

	// Aggregate the alerts, then update.
	item.End = alert.End
	item.Count++

	return db.updateAlert(item)
}

var querylists = []AlertItem{
	// Return all alerts.
	{
		Alertname: ".*alert.*",
	},
	// Return major alerts.
	{
		Serverity: "major",
	},
	// Return alert1.
	{
		Alertname: ".*ert.*",
		Source:    "kubernetes",
		Cluster:   "c1",
	},
	// Return no alerts.
	{
		Alertname:    "alert1",
		Resourcetype: "tenantmanager",
	},
}

var alertlists = []AlertItem{
	{
		Id:           "1",
		Alertname:    "alert1",
		Serverity:    "critical",
		Resourcetype: "pod",
		Source:       "kubernetes",
		Info:         "info1",
		Start:        time.Now(),
		End:          time.Now().Add(1 * time.Hour),
		Cluster:      "c1",
		Namespace:    "n1",
		Pod:          "p1",
	},
	{
		Id:           "2",
		Alertname:    "alert2",
		Serverity:    "warning",
		Resourcetype: "node",
		Source:       "clustermanager",
		Info:         "info2",
		Start:        time.Now(),
		End:          time.Now().Add(2 * time.Hour),
		Cluster:      "c2",
		Node:         "n1",
	},
	{
		Id:           "3",
		Alertname:    "alert3",
		Serverity:    "major",
		Resourcetype: "cluster",
		Source:       "tenantmanager",
		Info:         "info3",
		Start:        time.Now(),
		End:          time.Now().Add(15 * time.Minute),
		Cluster:      "c3",
	},
	{
		Id:           "3",
		Alertname:    "alert3",
		Serverity:    "major",
		Resourcetype: "cluster",
		Source:       "tenantmanager",
		Info:         "info3",
		Start:        time.Now(),
		End:          time.Now().Add(30 * time.Minute),
		Cluster:      "c3",
	},
}

func main() {
	d, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@(%s:%s)/mysql?parseTime=true", user, password, address, port))
	if err != nil {
		fmt.Printf("Connect database failed: %v\n", err)
		return
	}

	db := &DB{d}

	err = db.Ping()
	if err != nil {
		fmt.Printf("Ping DB failed: %v\n", err)
		return
	} else {
		fmt.Printf("Ping DB succeeded\n")
	}

	schema := `CREATE DATABASE IF NOT EXISTS alertdb;`
	result, err := db.Exec(schema)
	if err != nil {
		fmt.Printf("Create database failed: %v\n", err)
		return
	}

	schema = `USE alertdb;`
	result, err = db.Exec(schema)
	if err != nil {
		fmt.Printf("Use database failed: %v\n", err)
		return
	}

	schema = `CREATE TABLE IF NOT EXISTS alerts (
			id text,
			alertname text,
			serverity text,
			resourcetype text,
			source text,
			info text,
			count integer,
			start timestamp,
			end timestamp,
			organization text,
			project text,
			cluster text,
			namespace text,
			node text,
			pod text,
			deployment text,
			statefulset text,
			extend text);`
	result, err = db.Exec(schema)
	if err != nil {
		fmt.Printf("Create table failed: %v\n", err)
		return
	}
	fmt.Printf("Exec result is %v\n", result)

	for _, alert := range alertlists {
		err := db.InsertAlert(alert)
		if err != nil {
			fmt.Printf("Insert alert failed: %v\n", err)
		}
	}

	for _, alert := range querylists {
		list, err := db.queryAlert(AlertDBItem{AlertItem: alert})
		if err != nil {
			fmt.Printf("Query alert failed: %v\n", err)
		} else {
			fmt.Printf("alerts:\n%v\n\n", list)
		}
	}
}
