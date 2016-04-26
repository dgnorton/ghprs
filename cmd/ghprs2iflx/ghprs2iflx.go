package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/influxdata/influxdb/client/v2"
)

//var team = map[string]struct{}{
//	"121watts":             struct{}{},
//	"alexpaxton":           struct{}{},
//	"becketsean":           struct{}{},
//	"benbjohnson":          struct{}{},
//	"benstronaut":          struct{}{},
//	"corylanou":            struct{}{},
//	"DanielMorsing":        struct{}{},
//	"dgnorton":             struct{}{},
//	"e-dard":               struct{}{},
//	"ELKster":              struct{}{},
//	"fred-influx":          struct{}{},
//	"glogic":               struct{}{},
//	"gunnaraasen":          struct{}{},
//	"influxdb-denver-pair": struct{}{},
//	"jackzampolin":         struct{}{},
//	"jademcgough":          struct{}{},
//	"jboursiquot":          struct{}{},
//	"jginfluxdata":         struct{}{},
//	"joelegasse":           struct{}{},
//	"jsternberg":           struct{}{},
//	"jwilder":              struct{}{},
//	"kfitzpatrick":         struct{}{},
//	"kostasb":              struct{}{},
//	"mark-rushakoff":       struct{}{},
//	"markbates":            struct{}{},
//	"michael-influxdb":     struct{}{},
//	"mjdesa":               struct{}{},
//	"nathanielc":           struct{}{},
//	"pauldix":              struct{}{},
//	"rkuchan":              struct{}{},
//	"rossmcdonald":         struct{}{},
//	"rothrock":             struct{}{},
//	"ShubhraKar":           struct{}{},
//	"sparrc":               struct{}{},
//	"timraymond":           struct{}{},
//	"toddboom":             struct{}{},
//	"wfro":                 struct{}{},
//	"willpiers":            struct{}{},
//	"goller":               struct{}{},
//	"nhaugo":               struct{}{},
//	"otoolep":              struct{}{},
//	"jvshahid":             struct{}{},
//}

func loadTeam(filename string) (map[string]struct{}, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	members := strings.Split(string(b), "\n")
	team := map[string]struct{}{}
	for _, member := range members {
		team[member] = struct{}{}
	}
	return team, nil
}

func main() {
	infile := flag.String("i", "ghprs.json", "input file")
	teamfile := flag.String("t", "", "file containing list of team members")
	flag.Parse()

	var team map[string]struct{}
	var err error
	if *teamfile != "" {
		team, err = loadTeam(*teamfile)
		fatalIfErr(err)
	}

	f, err := os.Open(*infile)
	fatalIfErr(err)
	defer f.Close()

	prs := []github.PullRequest{}
	fatalIfErr(json.NewDecoder(f).Decode(&prs))

	c, err := client.NewHTTPClient(client.HTTPConfig{Addr: "http://localhost:8086"})
	fatalIfErr(err)

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "ghprs",
		Precision: "s",
	})
	fatalIfErr(err)

	for _, pr := range prs {
		tags := map[string]string{"user": *pr.User.Login, "state": *pr.State}
		if team != nil {
			if _, ok := team[*pr.User.Login]; !ok {
				tags["community"] = "yes"
			}
		}

		fields := map[string]interface{}{
			"number": *pr.Number,
			"title":  *pr.Title,
		}

		pt, err := client.NewPoint("pr", tags, fields, *pr.CreatedAt)
		fatalIfErr(err)

		bp.AddPoint(pt)
	}

	fatalIfErr(c.Write(bp))
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
