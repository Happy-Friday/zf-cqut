package cqut

import (
	"testing"
	"encoding/json"
	"log"
)

var (
	USERNAME = ""
	PASSWORD = ""
)

func stringJson(m interface{}) string {
	buf, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(buf)
}

func TestGrades(t *testing.T) {
	c := NewCqut(USERNAME, PASSWORD)
	c.Initialize()
	m := c.GetCoursesTable(true)
	if m == nil {
		t.Error("nil data")
	}
	log.Println(stringJson(m))
}

func TestCoursesTable(t *testing.T) {
	c := NewCqut(USERNAME, PASSWORD)
	c.Initialize()
	m := c.GetCoursesTable(true)
	if m == nil {
		t.Error("nil data")
	}
	log.Println(stringJson(m))
}

func TestGradesPoints(t *testing.T) {
	c := NewCqut(USERNAME, PASSWORD)
	c.Initialize()
	m := c.GetGradesPoints(true)
	if m == nil {
		t.Error("nil data")
	}
	log.Println(stringJson(m))
}