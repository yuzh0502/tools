package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gocolly/colly"
	"github.com/robfig/cron"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type workTool struct {
	InTime  string `toml:"in_time"`
	OutTime string `toml:"out_time"`
	User    []user
}

type user struct {
	Username string
	Password string
	Email    string
	msgTitle string
	msgBody  string
	err      error
	c        *colly.Collector
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

var (
	w                                    workTool
	lt, execution, cookie, token, lastID string
	loginUrl                             = "http://portal.zts.com.cn/cas/login?service=http%3A%2F%2F10.55.10.13%2FCASLogin"
	tableUrl                             = "http://10.55.10.13/UIProcessor?Table=vem_wb_qdqc"
	inUrl                                = "http://10.55.10.13/OperateProcessor?operate=vem_wb_qdqc_M1&amp;Table=vem_wb_qdqc&amp;Token=%s&amp;&amp;ID=%s&amp;WindowType=1&amp;extWindow=true&amp;PopupWin=true"
	outUrl                               = "http://10.55.10.13/OperateProcessor?operate=vem_wb_qdqc_M2&amp;Table=vem_wb_qdqc&amp;Token=%s&amp;&amp;ID=%s&amp;WindowType=1&amp;extWindow=true&amp;PopupWin=true"
)

func main() {
	readConfig()
	usersQueryTable()
	//run()
}

func run() {
	timer := cron.New()
	if err := timer.AddFunc(w.InTime, usersIn); err == nil {
		fmt.Println("签到定时任务添加成功")
	}
	if err := timer.AddFunc(w.OutTime, usersOut); err == nil {
		fmt.Println("签出定时任务添加成功")
	}
	timer.Start()
	select {}
}

func readConfig() {
	fmt.Println("reading config...")
	configFile, err := ioutil.ReadFile(getConfigFilePathName())
	if err != nil {
		fmt.Println("reading config file error:", err)
		return
	}
	_, err = toml.Decode(string(configFile), &w)
	if err != nil {
		fmt.Println("decode toml error:", err)
		return
	}
}

func getExePath() string {
	exePath, err := filepath.Abs(os.Args[0])
	if err != nil {
		fmt.Println("get exe path error:", err)
	}
	exePath = exePath[:strings.LastIndex(exePath, string(os.PathSeparator))]
	return exePath
}

func getConfigFilePathName() string {
	exePath := getExePath()
	return exePath + string(os.PathSeparator) + "config.toml"
}

func usersQueryTable() {
	for _, u := range w.User {
		u.login().queryTable()
		u.msgTitle = "查询结果"
		fmt.Println(u.Username, u.msgTitle, u.msgBody)
		sendEmail(u.Email, u.msgTitle, u.msgBody)
	}
}

func usersIn() {
	for _, u := range w.User {
		for i := 0; i < 10; i++ {
			u.login().in().queryTable()
			fmt.Println(u.Username, u.msgTitle, u.msgBody)
			sendEmail(u.Email, u.msgTitle, u.msgBody)
			if u.err == nil {
				break
			}
		}
	}
}

func usersOut() {
	for _, u := range w.User {
		for i := 0; i < 10; i++ {
			u.login().out().queryTable()
			fmt.Println(u.Username, u.msgTitle, u.msgBody)
			sendEmail(u.Email, u.msgTitle, u.msgBody)
			if u.err == nil {
				break
			}
		}
	}
}

func (u *user) login() *user {
	if u == nil || u.err != nil {
		return u
	}
	fmt.Printf("%s login...\n", u.Username)
	u.getParameter().postForm()
	return u
}

func getColly() *colly.Collector {
	c := colly.NewCollector()
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Host", "portal.zts.com.cn")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Cache-Control", "max-age=0")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36")
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9")
		r.Headers.Set("Cookie", cookie)
		r.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
	})
	return c
}

func (u *user) getParameter() *user {
	if u == nil || u.err != nil {
		return u
	}
	fmt.Println("get parameter...")
	lt, execution, cookie, token, lastID = "", "", "", "", ""
	c := getColly()
	c.OnHTML("input", func(e *colly.HTMLElement) {
		if e.Attr("name") == "lt" {
			lt = e.Attr("value")
			println(lt)
		}
		if e.Attr("name") == "execution" {
			execution = e.Attr("value")
		}
	})
	err := c.Visit(loginUrl)
	if err != nil {
		u.err = err
		u.msgBody += err.Error() + "<br/>"
		return u
	}
	cookie = fmt.Sprintf("%s; %s; %s;", c.Cookies(loginUrl)[0].String(), u.Username, md5sum(u.Password))
	fmt.Println(cookie)
	return u
}

func (u *user) postForm() *user {
	if u == nil || u.err != nil {
		return u
	}
	fmt.Println("post form...")
	c := getColly()
	err := c.Post(loginUrl, map[string]string{
		"uuid":      "",
		"username":  u.Username,
		"password":  md5sum(u.Password),
		"lt":        lt,
		"execution": execution,
		"_eventId":  "submit",
	})
	if err != nil {
		u.err = err
		u.msgBody += err.Error() + "<br/>"
		fmt.Println("post form error:", err)
	}
	u.c = c
	return u
}

func (u *user) queryTable() *user {
	if u == nil || u.err != nil {
		return u
	}
	fmt.Println("query table...")
	c := u.c
	c.OnHTML("script", func(e *colly.HTMLElement) {
		scriptContext := e.Text
		fmt.Println(scriptContext)
		reRecords, err := regexp.Compile(`\[{"id":"ID=.*?]`)
		if err != nil {
			u.err = err
			u.msgBody += err.Error() + "<br/>"
			fmt.Printf("正则编译错误: %s\n", err)
			return
		}
		recordsString := reRecords.FindString(scriptContext)
		var records []Record
		err = json.Unmarshal([]byte(recordsString), &records)
		if err != nil {
			return
		}
		first := fmt.Sprintf("最后一次记录：姓名：%s，考勤日期：%s，签到时间：%s，签出时间：%s", records[0].Kqr.Text, records[0].Kqrq, records[0].Qdsj, records[0].Qcsj)
		second := fmt.Sprintf("倒数第二记录：姓名：%s，考勤日期：%s，签到时间：%s，签出时间：%s", records[1].Kqr.Text, records[1].Kqrq, records[1].Qdsj, records[1].Qcsj)
		u.msgBody += first + "<br/>" + second + "<br/>"
		reTokenConfig, err := regexp.Compile(`Token=.*?'`)
		if err != nil {
			u.err = err
			u.msgBody += err.Error() + "<br/>"
			fmt.Printf("正则编译错误: %s\n", err)
			return
		}
		tokenString := reTokenConfig.FindString(scriptContext)
		token = tokenString[7 : len(tokenString)-1]
		lastID = records[0].ID[3:]
	})
	err := c.Visit(tableUrl)
	if err != nil {
		u.err = err
		u.msgBody += err.Error() + "<br/>"
		fmt.Printf("visit tableUrl error: %s", err)
	}
	return u
}

func (u *user) in() *user {
	if u == nil || u.err != nil {
		return u
	}
	c := u.c
	inUrl := fmt.Sprintf(inUrl, token, lastID)
	c.OnHTML("html", func(e *colly.HTMLElement) {
		u.msgTitle = "签到成功"
	})
	err := c.Visit(inUrl)
	if err != nil {
		u.err = err
		u.msgBody += err.Error() + "<br/>"
		u.msgTitle = "签到失败"
	}
	return u
}

func (u *user) out() *user {
	if u == nil || u.err != nil {
		return u
	}
	c := u.c
	outUrl := fmt.Sprintf(outUrl, token, lastID)
	c.OnHTML("html", func(_ *colly.HTMLElement) {
		u.msgTitle = "签出成功"
	})
	err := c.Visit(outUrl)
	if err != nil {
		u.err = err
		u.msgBody += err.Error() + "<br/>"
		u.msgTitle = "签出失败"
	}
	return u
}

func md5sum(str string) string {
	sum := md5.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

func sendEmail(to, title, msg string) {
	if to == "" {
		return
	}
	m := gomail.NewMessage()
	m.SetHeader("From", "vihv@qq.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", title)
	m.SetBody("text/html", msg)
	email := gomail.NewDialer("smtp.qq.com", 587, "vihv@qq.com", "kpyuadxmymptbhbb")
	err := email.DialAndSend(m)
	if err != nil {
		fmt.Printf("send email to %s err: %s", to, err)
	}
}
