package main

import (
	"flag"
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
	"log"
	"math/rand"
	"os"
	"text/tabwriter"
	"time"
)

var poolName = flag.String("pool", "default", "Pool to connect to")
var bucketName = flag.String("bucket", "default", "Bucket to connect to")

const myfmt = "2006-02-01-15:04:05.000000000"

var names = []string{
	"Jan Lehnardt",
	"John Christopher Anderson",
	"Noah Slater",
	"Filipe David Borba Manana",
	"Adam Kocoloski",
	"Paul Joseph Davis",
	"Christopher Lenz",
	"Damien F. Katz",
	"Robert Newson",
	"Benoit Chesneau",
	"Jason David Davies",
	"Mark Hammond",
	"Randall Leeds",
	"Bin Cui",
	"Benjamin Young",
	"Dustin Sallings",
	"Steve Yen",
	"Joe Schaefer",
}

var actions = []string{
	"submitted", "aborted", "approved", "declined",
}

var projects = []string{
	"ep-engine", "couchdb", "ns_server", "moxi", "libcouchbase",
}

type Record struct {
	Author   string `json:"author"`
	Reviewer string `json:"reviewer"`
	Action   string `json:"action"`
	Project  string `json:"project"`
	Score    int    `json:"score"`
}

func report(c couchbase.Client, b couchbase.Bucket) {
	fmt.Printf("-----------------------------------------------------\n")
	fmt.Printf("Got %d success messages, %d not-my-vbucket\n",
		c.Statuses[0], c.Statuses[7])
	fmt.Printf("-----------------------------------------------------\n")
	return
	tr := tabwriter.NewWriter(os.Stdout, 8, 8, 1, ' ', 0)
	defer tr.Flush()
	params := map[string]interface{}{
		"group_level":        1,
		"stale":              "update_after",
		"connection_timeout": 60000,
	}
	vres, err := b.View("test", "test", params)
	if err != nil {
		log.Printf("Error executing view:  %v", err)
	}

	for _, e := range vres.Errors {
		fmt.Printf(" * Error from %s:  %s\n", e.From, e.Reason)
	}

	for _, r := range vres.Rows {
		fmt.Fprintf(tr, "%v:\t%v\n", r.Key, r.Value)
	}
}

func harass(c couchbase.Client, b couchbase.Bucket) {
	fmt.Printf("Doing stuff\n")

	go func() {
		for {
			time.Sleep(2 * time.Second)
			report(c, b)
		}
	}()

	for {
		r := Record{
			Author:   names[rand.Intn(len(names))],
			Reviewer: names[rand.Intn(len(names))],
			Action:   actions[rand.Intn(len(actions))],
			Project:  projects[rand.Intn(len(projects))],
			Score:    rand.Intn(4) - 2,
		}

		k := time.Now().Format(myfmt)

		if err := b.Set(k, r); err != nil {
			log.Fatalf("Oops, failed a store of %s:  %v", k, err)
		}
	}
}

func main() {
	flag.Parse()
	c, err := couchbase.Connect(flag.Arg(0))
	if err != nil {
		log.Fatalf("Error connecting:  %v", err)
	}

	pool, err := c.GetPool(*poolName)
	if err != nil {
		log.Fatalf("Error getting pool:  %v", err)
	}

	bucket, err := pool.GetBucket(*bucketName)
	if err != nil {
		log.Fatalf("Error getting bucket:  %v", err)
	}

	harass(c, bucket)
}
