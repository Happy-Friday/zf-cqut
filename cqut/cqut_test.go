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
	if err := c.Initialize(); err != nil {
		t.Error(err)
		return
	}
	m := c.GetGrades(true, "2017-2018")
	if m == nil {
		t.Error("nil data")
		return
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

func TestUserInfo(t *testing.T) {
	c := NewCqut(USERNAME, PASSWORD)
	err := c.Initialize()
	if err != nil {
		t.Error(err)
		return
	}
	m := c.GetUserInfo()
	if m == nil {
		t.Error("nil data")
		return
	}
	log.Println(stringJson(m))
}

func TestAllInfos(t *testing.T) {
	c := NewCqut(USERNAME, PASSWORD)
	err := c.Initialize()
	if err != nil {
		t.Error(err)
		return
	}
	userInfos := c.GetUserInfo()
	grades := c.GetGrades(true, "2017-2018")
	gps := c.GetGradesPoints(false)
	cs := c.GetCoursesTable(true)
	m := map[string]interface{} {
		"userInfos" : userInfos,
		"grades" : grades,
		"gps" : gps,
		"cs" : cs,
	}
	log.Println(stringJson(m))
}