// Copyright (C) 2012 Numerotron Inc.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.

package stathat

import (
	"testing"
)

func TestNewStatCount(t *testing.T) {
	setTesting()
	x := newStatCount("abc", "pc@pc.com", 1)
	if x == nil {
		t.Fatalf("expected a StatReport object")
	}
	if x.statType != kcounter {
		t.Errorf("expected counter")
	}
	if x.StatKey != "abc" {
		t.Errorf("expected abc")
	}
	if x.UserKey != "pc@pc.com" {
		t.Errorf("expected pc@pc.com")
	}
	if x.Value != 1.0 {
		t.Errorf("expected 1.0")
	}
	if x.Timestamp != 0 {
		t.Errorf("expected 0")
	}
}

func TestNewStatValue(t *testing.T) {
	setTesting()
	x := newStatValue("abc", "pc@pc.com", 3.14159)
	if x == nil {
		t.Fatalf("expected a StatReport object")
	}
	if x.statType != kvalue {
		t.Errorf("expected value")
	}
	if x.StatKey != "abc" {
		t.Errorf("expected abc")
	}
	if x.UserKey != "pc@pc.com" {
		t.Errorf("expected pc@pc.com")
	}
	if x.Value != 3.14159 {
		t.Errorf("expected 3.14159")
	}
}

func TestURLValues(t *testing.T) {
	setTesting()
	x := newStatCount("abc", "pc@pc.com", 1)
	v := x.values()
	if v == nil {
		t.Fatalf("expected url values")
	}
	if v.Get("stat") != "abc" {
		t.Errorf("expected abc")
	}
	if v.Get("ezkey") != "pc@pc.com" {
		t.Errorf("expected pc@pc.com")
	}
	if v.Get("count") != "1" {
		t.Errorf("expected count of 1")
	}

	y := newStatValue("abc", "pc@pc.com", 3.14159)
	v = y.values()
	if v == nil {
		t.Fatalf("expected url values")
	}
	if v.Get("stat") != "abc" {
		t.Errorf("expected abc")
	}
	if v.Get("ezkey") != "pc@pc.com" {
		t.Errorf("expected pc@pc.com")
	}
	if v.Get("value") != "3.14159" {
		t.Errorf("expected value of 3.14159")
	}
}

func TestPosts(t *testing.T) {
	setTesting()
	Verbose = true
	PostCountOne("a stat", "pc@pc.com")
	p := <-testPostChannel
	if p.url != "http://api.stathat.com/ez" {
		t.Errorf("expected ez url")
	}
	if p.values.Get("stat") != "a stat" {
		t.Errorf("expected a stat")
	}
	if p.values.Get("ezkey") != "pc@pc.com" {
		t.Errorf("expected pc@pc.com")
	}
	if p.values.Get("count") != "1" {
		t.Errorf("expected count of 1")
	}

	PostCount("a stat", "pc@pc.com", 213)
	p = <-testPostChannel
	if p.url != "http://api.stathat.com/ez" {
		t.Errorf("expected ez url")
	}
	if p.values.Get("stat") != "a stat" {
		t.Errorf("expected a stat")
	}
	if p.values.Get("ezkey") != "pc@pc.com" {
		t.Errorf("expected pc@pc.com")
	}
	if p.values.Get("count") != "213" {
		t.Errorf("expected count of 213")
	}

	PostValue("a stat", "pc@pc.com", 2.13)
	p = <-testPostChannel
	if p.url != "http://api.stathat.com/ez" {
		t.Errorf("expected ez url")
	}
	if p.values.Get("stat") != "a stat" {
		t.Errorf("expected a stat")
	}
	if p.values.Get("ezkey") != "pc@pc.com" {
		t.Errorf("expected pc@pc.com")
	}
	if p.values.Get("value") != "2.13" {
		t.Errorf("expected value of 2.13")
	}

	PostCountTime("a stat", "pc@pc.com", 213, 300000)
	p = <-testPostChannel
	if p.values.Get("t") != "300000" {
		t.Errorf("expected t value of 300000, got %s", p.values.Get("t"))
	}

	PostValueTime("a stat", "pc@pc.com", 2.13, 400000)
	p = <-testPostChannel
	if p.values.Get("t") != "400000" {
		t.Errorf("expected t value of 400000, got %s", p.values.Get("t"))
	}
}
