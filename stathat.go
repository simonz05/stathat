// Copyright (C) 2012 Numerotron Inc.
// Modifications Copyright (C) 2014 Simon Zimmermann
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.

// Copyright 2012 Numerotron Inc.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
//
// Developed at www.stathat.com by Patrick Crosby
// Contact us on twitter with any questions:  twitter.com/stat_hat

// The stathat package makes it easy to post any values to your StatHat
// account.
package stathat

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const hostname = "api.stathat.com"

type statKind int

const (
	_                 = iota
	kcounter statKind = iota
	kvalue
)

type statReport struct {
	StatKey   string
	UserKey   string
	Value     float64
	Timestamp int64
	statType  statKind
}

// Reporter is a StatHat client that can report stat values/counts to the servers.
type Reporter struct {
	reports chan *statReport
	done    chan bool
	client  *http.Client
	wg      *sync.WaitGroup
}

// NewReporter returns a new Reporter.  You must specify the channel bufferSize and the
// goroutine poolSize.  You can pass in nil for the transport and it will use the
// default http transport.
func NewReporter(bufferSize, poolSize int, transport http.RoundTripper) *Reporter {
	r := new(Reporter)
	r.client = &http.Client{Transport: transport}
	r.reports = make(chan *statReport, bufferSize)
	r.done = make(chan bool)
	r.wg = new(sync.WaitGroup)
	for i := 0; i < poolSize; i++ {
		r.wg.Add(1)
		go r.processReports()
	}
	return r
}

// DefaultReporter is the default instance of *Reporter.
var DefaultReporter = NewReporter(100000, 10, nil)

var testingEnv = false

type testPost struct {
	url    string
	values url.Values
}

var testPostChannel chan *testPost

// The Verbose flag determines if the package should write verbose output to stdout.
var Verbose = false

func setTesting() {
	testingEnv = true
	testPostChannel = make(chan *testPost)
}

func newStatCount(statName, key string, count int) *statReport {
	return &statReport{StatKey: statName,
		UserKey:  key,
		Value:    float64(count),
		statType: kcounter}
}

func newStatValue(statName, key string, value float64) *statReport {
	return &statReport{StatKey: statName,
		UserKey:  key,
		Value:    value,
		statType: kvalue}
}

func (sr *statReport) values() url.Values {
	switch sr.statType {
	case kcounter:
		return sr.counterValues()
	case kvalue:
		return sr.valueValues()
	}
	return nil
}

func (sr *statReport) commonValues() url.Values {
	result := make(url.Values)
	result.Set("stat", sr.StatKey)
	result.Set("ezkey", sr.UserKey)
	if sr.Timestamp > 0 {
		result.Set("t", sr.timeString())
	}
	return result
}

func (sr *statReport) counterValues() url.Values {
	result := sr.commonValues()
	result.Set("count", sr.valueString())
	return result
}

func (sr *statReport) valueValues() url.Values {
	result := sr.commonValues()
	result.Set("value", sr.valueString())
	return result
}

func (sr *statReport) valueString() string {
	return strconv.FormatFloat(sr.Value, 'g', -1, 64)
}

func (sr *statReport) timeString() string {
	return strconv.FormatInt(sr.Timestamp, 10)
}

func (sr *statReport) url() string {
	return fmt.Sprintf("http://%s/ez", hostname)
}

// PostCountOne posts a count of 1 to a stat using DefaultReporter.
func PostCountOne(statName, key string) error {
	return DefaultReporter.PostCountOne(statName, key)
}

// PostCount posts a count to a stat using DefaultReporter.
func PostCount(statName, key string, count int) error {
	return DefaultReporter.PostCount(statName, key, count)
}

// PostCountTime posts a count to a stat at a specific time using DefaultReporter.
func PostCountTime(statName, key string, count int, timestamp int64) error {
	return DefaultReporter.PostCountTime(statName, key, count, timestamp)
}

// PostValue posts a value to a stat using DefaultReporter.
func PostValue(statName, key string, value float64) error {
	return DefaultReporter.PostValue(statName, key, value)
}

// PostValueTime posts a value to a stat at a specific time using DefaultReporter.
func PostValueTime(statName, key string, value float64, timestamp int64) error {
	return DefaultReporter.PostValueTime(statName, key, value, timestamp)
}

// WaitUntilFinished wait for all stats to be sent, or until timeout. Useful
// for simple command- line apps to defer a call to this in main()
func WaitUntilFinished(timeout time.Duration) bool {
	return DefaultReporter.WaitUntilFinished(timeout)
}

// PostCountOne posts a count of 1 to a stat.
func (r *Reporter) PostCountOne(statName, key string) error {
	return r.PostCount(statName, key, 1)
}

// PostCount posts a count to a stat.
func (r *Reporter) PostCount(statName, key string, count int) error {
	r.reports <- newStatCount(statName, key, count)
	return nil
}

// PostCountTime posts a count to a stat at a specific time.
func (r *Reporter) PostCountTime(statName, key string, count int, timestamp int64) error {
	x := newStatCount(statName, key, count)
	x.Timestamp = timestamp
	r.reports <- x
	return nil
}

// PostValue posts a value to a stat.
func (r *Reporter) PostValue(statName, key string, value float64) error {
	r.reports <- newStatValue(statName, key, value)
	return nil
}

// PostValueTime posts a value to a stat at a specific time.
func (r *Reporter) PostValueTime(statName, key string, value float64, timestamp int64) error {
	x := newStatValue(statName, key, value)
	x.Timestamp = timestamp
	r.reports <- x
	return nil
}

func (r *Reporter) processReports() {
	for {
		sr, ok := <-r.reports

		if !ok {
			if Verbose {
				log.Printf("channel closed, stopping processReports()")
			}
			break
		}

		if Verbose {
			log.Printf("posting stat to stathat: %s, %v", sr.url(), sr.values())
		}

		if testingEnv {
			if Verbose {
				log.Printf("in test mode, putting stat on testPostChannel")
			}
			testPostChannel <- &testPost{sr.url(), sr.values()}
			continue
		}

		resp, err := r.client.PostForm(sr.url(), sr.values())
		if err != nil {
			log.Printf("error posting stat to stathat: %s", err)
			continue
		}

		if Verbose {
			body, _ := ioutil.ReadAll(resp.Body)
			log.Printf("stathat post result: %s", body)
		}

		resp.Body.Close()
	}
	r.wg.Done()
}

func (r *Reporter) finish() {
	close(r.reports)
	r.wg.Wait()
	r.done <- true
}

// Wait for all stats to be sent, or until timeout. Useful for simple command-
// line apps to defer a call to this in main()
func (r *Reporter) WaitUntilFinished(timeout time.Duration) bool {
	go r.finish()
	select {
	case <-r.done:
		return true
	case <-time.After(timeout):
		return false
	}
	return false
}
