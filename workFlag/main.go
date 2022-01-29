package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/jlvihv/tools/tgSend"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

var ab AutoBooked
var hd header
var fm form
var uc urlConfig
var configFile = "/home/vihv/Code/tools/workFlag/config.xml"

func main() {
	timeTemplate := "9 %s %s ? * 1-5" // 每周几到周几，每个月，不指定某天，几点，几分，9秒
	inTime := fmt.Sprintf(timeTemplate, strconv.Itoa(getRandNum(3)+15), "8")
	outTime := fmt.Sprintf(timeTemplate, "31", "17")

	readConfig()
	initLogger()
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "query":
			usersLogin()
		case "in":
			usersIn()
		case "out":
			usersOut()
		}
		logrus.Infof("程序退出")
		return
	}

	timer := cron.New()
	if err := timer.AddFunc(inTime, usersIn); err != nil {
		logrus.Errorf("签到未执行: 创建定时任务失败: %s", err)
	} else {
		msg := "已创建签到定时任务，时间：" + inTime
		logrus.Infof(msg)
		push(msg)
	}
	// 从后往前读，每周一到周五，每个月，不指定某天，下午17点，31分，9秒，签出
	if err := timer.AddFunc(outTime, usersOut); err != nil {
		logrus.Errorf("签出未执行: 创建定时任务失败: %s", err)
	} else {
		msg := "已创建签出定时任务, 时间：" + outTime
		logrus.Infof(msg)
		push(msg)
	}
	timer.Start()
	select {}
}

func initHeader() {
	hd.host = "portal.zts.com.cn"
	hd.connection = "keep-alive"
	hd.cacheControl = "max-age=0"
	hd.upgradeInsecureRequests = "1"
	hd.userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
	hd.accept = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"
	hd.referer = "http://10.55.10.13"
	hd.acceptEncoding = "gzip, deflate"
	hd.acceptLanguage = "zh-CN,zh;q=0.9"
	hd.cookie = ""
	hd.contentType = "application/x-www-form-urlencoded"

	fm.uuid = ""
	fm.username = ""
	fm.password = ""
	fm.rememberMe = "true"
	fm.lt = ""
	fm.execution = ""
	fm._eventId = "submit"
}

func readConfig() {
	fmt.Println("读取配置文件中...")
	var configPathFile string
	if filepath.IsAbs(configFile) {
		configPathFile = configFile
	} else {
		exePath, err := getExePath()
		if err != nil {
			fmt.Printf("获取程序绝对路径失败，程序将退出: %s\n", err)
			os.Exit(1)
		}
		configPathFile, err = filepath.Abs(exePath + "/" + configFile)
		if err != nil {
			fmt.Printf("获取配置文件绝对路径失败，程序将退出: %s\n", err)
			os.Exit(1)
		}
	}
	f, err := ioutil.ReadFile(configPathFile)
	if err != nil {
		fmt.Printf("读取配置文件失败: %s\n", err)
		return
	}
	err = xml.Unmarshal(f, &ab)
	if err != nil {
		fmt.Printf("解析xml配置文件失败: %s\n", err)
		return
	}
}

func initLogger() {
	fmt.Println("初始化日志中...")
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		ForceQuote:      true,
		TimestampFormat: "2006/01/02 15:04:05",
		FullTimestamp:   true,
	})
	exePath, err := getExePath()
	if err != nil {
		fmt.Printf("获取程序绝对路径失败，程序将退出: %s\n", err)
		os.Exit(1)
	}
	f, err := os.Create(path.Join(exePath, "workFlag.log"))
	if err != nil {
		fmt.Printf("新建日志文件失败: %s\n", err)
	}
	writer := io.MultiWriter(os.Stdout, f)
	logrus.SetOutput(writer)
	logrus.SetLevel(logrus.DebugLevel)
}

func usersLogin() {
	for _, user := range ab.Users {
		logrus.Infof("用户: %s", user.Username)
		user.login()
		if user.err != nil {
			logrus.Errorf("用户 %s 登陆失败，请前往公司手动操作，失败原因: %s", user.Username, user.err)
			continue
		}
		logrus.Infof(user.result)
	}
}

func usersIn() {
	// 随机休眠1-3分钟，免得每次打卡时间都相同
	sleepTime := getRandNum(3)
	time.Sleep(time.Duration(sleepTime) * time.Minute)
	for _, user := range ab.Users {
		user.login().in()
		if user.err != nil {
			msg := fmt.Sprintf("用户 %s: %s: %s\n", user.Username, user.msg, user.err)
			logrus.Errorf(msg)
			push(msg + user.result)
			continue
		}
		user.login()
		logrus.Infof(user.msg + ": " + user.result)
		logrus.Infof("推送到tg")
		msg := fmt.Sprintf("用户 %s: %s\n", user.Username, user.msg)
		push(msg + user.result)
	}
}

func usersOut() {
	for _, user := range ab.Users {
		user.login().out()
		if user.err != nil {
			msg := fmt.Sprintf("用户 %s: %s: %s\n", user.Username, user.msg, user.err)
			logrus.Errorf(msg)
			push(msg + user.result)
			continue
		}
		user.login()
		logrus.Infof(user.msg + ": " + user.result)
		logrus.Infof("推送到tg")
		msg := fmt.Sprintf("用户 %s: %s\n", user.Username, user.msg)
		push(msg + user.result)
	}
}

func (u *User) login() *User {
	initHeader()
	u.getParameter().postForm().queryTable()
	return u
}

// 获取需要的参数，主要是cookie，lt和execution，cookie里包含了session
func (u *User) getParameter() *User {
	if u == nil || u.err != nil {
		return u
	}
	logrus.Debugf("准备获取所需要的参数")
	c := getColly()
	c.OnHTML("input", func(e *colly.HTMLElement) {
		if e.Attr("name") == "lt" {
			fm.lt = e.Attr("value")
		}
		if e.Attr("name") == "execution" {
			fm.execution = e.Attr("value")
		}
	})
	err := c.Visit(ab.BaseInfo.LoginUrl)
	if err != nil {
		u.err = err
		u.msg = fmt.Sprintf("获取参数过程中: 访问 %s 错误: %s", ab.BaseInfo.LoginUrl, err)
		logrus.Errorf(u.msg)
		return u
	}
	hd.cookie = fmt.Sprintf("%s; %s; %s;", c.Cookies(ab.BaseInfo.LoginUrl)[0].String(), u.Username, md5sum(u.Password))
	return u
}

// post表单，其实就是登陆，这一步如果成功了，就会顺利登陆
func (u *User) postForm() *User {
	if u == nil || u.err != nil {
		return u
	}
	logrus.Debugf("准备登陆，正在提交数据")
	c := getColly()
	fm.username = u.Username
	fm.password = md5sum(u.Password)
	err := c.Post(ab.BaseInfo.LoginUrl, map[string]string{
		"uuid":       fm.uuid,
		"username":   fm.username,
		"password":   fm.password,
		"rememberMe": fm.rememberMe,
		"lt":         fm.lt,
		"execution":  fm.execution,
		"_eventId":   fm._eventId,
	})
	if err != nil {
		u.err = err
		u.msg = fmt.Sprintf("登陆过程中，提交数据失败: %s", err)
		logrus.Errorf(u.msg)
		return u
	}
	u.c = c
	return u
}

func (u *User) queryTable() *User {
	if u == nil || u.err != nil {
		return u
	}
	logrus.Debugf("准备获取考勤记录")
	c := u.c
	c.OnHTML("script", func(e *colly.HTMLElement) {
		scriptContext := e.Text
		reRecords, err := regexp.Compile(`\[{"id":"ID=.*?]`)
		if err != nil {
			u.err = err
			u.msg = fmt.Sprintf("获取考勤记录过程中: 正则编译错误: %s", err)
			logrus.Errorf(u.msg)
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
		u.result = first + "\n" + second

		reTokenConfig, err := regexp.Compile(`Token=.*?'`)
		if err != nil {
			u.err = err
			u.msg = fmt.Sprintf("获取考勤记录过程中: 正则编译错误: %s", err)
			logrus.Errorf(u.msg)
			return
		}
		tokenString := reTokenConfig.FindString(scriptContext)
		uc.token = tokenString[7 : len(tokenString)-1]
		uc.lastID = records[0].ID[3:]
	})
	err := c.Visit(ab.BaseInfo.TableUrl)
	if err != nil {
		u.err = err
		u.msg = fmt.Sprintf("获取考勤记录过程中: 访问 %s 错误: %s", ab.BaseInfo.TableUrl, err)
		logrus.Errorf(u.msg)
		return u
	}
	return u
}

func (u *User) in() *User {
	if u == nil || u.err != nil {
		return u
	}
	c := u.c
	inUrl := fmt.Sprintf(ab.BaseInfo.InUrl, uc.token, uc.lastID)

	c.OnHTML("html", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "签到一天只允许签一次") {
			u.msg = "已签到，不必重复签到"
			return
		}
		u.msg = "签到成功"
	})

	err := c.Visit(inUrl)
	if err != nil {
		u.err = err
		u.msg = fmt.Sprintf("签到过程中: 访问 %s 错误: %s", inUrl, err)
		logrus.Errorf(u.msg)
		return u
	}
	return u
}

func (u *User) out() *User {
	if u == nil || u.err != nil {
		return u
	}
	c := u.c
	outUrl := fmt.Sprintf(ab.BaseInfo.OutUrl, uc.token, uc.lastID)
	hd.referer = ab.BaseInfo.UserUrl

	c.OnHTML("html", func(_ *colly.HTMLElement) {
		u.msg = "签出成功"
	})

	err := c.Visit(outUrl)
	if err != nil {
		u.err = err
		u.msg = fmt.Sprintf("签出过程中: 访问 %s 错误: %s", outUrl, err)
		logrus.Errorf(u.msg)
		return u
	}
	return u
}

func push(msg string) {
	err := tgSend.Send("http://localhost:7890", 956772010, msg)
	if err != nil {
		logrus.Errorf("推送到 telegram 失败: %s", err)
	}
}

func md5sum(str string) string {
	sum := md5.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

func getColly() *colly.Collector {
	c := colly.NewCollector()
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Host", hd.host)
		r.Headers.Set("Connection", hd.connection)
		r.Headers.Set("Cache-Control", hd.cacheControl)
		r.Headers.Set("Upgrade-Insecure-Requests", hd.upgradeInsecureRequests)
		r.Headers.Set("User-Agent", hd.userAgent)
		r.Headers.Set("Accept", hd.accept)
		r.Headers.Set("Accept-Encoding", hd.acceptEncoding)
		r.Headers.Set("Accept-Language", hd.acceptLanguage)
		r.Headers.Set("Cookie", hd.cookie)
		r.Headers.Set("Content-Type", hd.contentType)
	})
	return c
}

// 获取一个1-n之间的随机数
func getRandNum(n int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(n) + 1
}

// 获取程序所在路径
func getExePath() (string, error) {
	exePath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "", err
	}
	exePath = exePath[:strings.LastIndex(exePath, string(os.PathSeparator))]
	return exePath, nil
}
