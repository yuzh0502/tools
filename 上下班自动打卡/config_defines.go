package main

import "github.com/gocolly/colly"

type AutoBooked struct {
	BaseInfo BaseInfo `xml:"BaseInfo"`
	Users    []User   `xml:"User"`
}

// BaseInfo 基本信息
type BaseInfo struct {
	LoginUrl string `xml:"login_url,attr"`
	UserUrl  string `xml:"user_url,attr"`
	TableUrl string `xml:"table_url,attr"`
	InUrl    string `xml:"in_url,attr"`
	OutUrl   string `xml:"out_url,attr"`
}

// User 定义单个用户的信息
type User struct {
	Username string `xml:"username,attr"`
	Password string `xml:"password,attr"`
	err      error
	c        *colly.Collector
	msg      string
	result   string
}

type header struct {
	host                    string
	connection              string
	cacheControl            string
	upgradeInsecureRequests string
	userAgent               string
	accept                  string
	referer                 string
	acceptEncoding          string
	acceptLanguage          string
	cookie                  string
	contentType             string
}

type form struct {
	uuid       string
	username   string
	password   string
	rememberMe string
	lt         string
	execution  string
	_eventId   string
}

type Kqr struct {
	Text            string `json:"text,omitempty"`
	KqrEmployeeDept string `json:"kqr_employee_dept,omitempty"`
}

type Record struct {
	Kqr  Kqr    `json:"kqr"` // 考勤人
	ID   string `json:"id,omitempty"`
	Kqrq string `json:"kqrq,omitempty"` // 考勤日期
	Qdsj string `json:"qdsj,omitempty"` // 签到时间
	Qcsj string `json:"qcsj,omitempty"` // 签出时间
	Type string `json:"type,omitempty"` // 工作日性质
}

type urlConfig struct {
	lastID string
	token  string
}
