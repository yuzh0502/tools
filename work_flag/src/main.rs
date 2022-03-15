use std::collections::HashMap;
use std::fs::File;
use std::io::Read;

use crypto::digest::Digest;
use crypto::md5::Md5;
use reqwest::header::{HeaderMap, HeaderValue};
use serde_derive::Deserialize;
use toml::value::Index;
use visdom::Vis;

fn main() {
    run();
}

#[derive(Debug, Deserialize)]
struct Config {
    in_time: String,
    out_time: String,

    #[serde(rename = "User")]
    users: Vec<User>,
}

#[derive(Debug, Deserialize)]
struct User {
    username: String,
    password: String,
}

struct Kqr {
    text: String,
    kqr_employee_dept: String,
}

struct Record {
    kqr: Kqr,
    id: String,
    kqrq: String,
    qdsj: String,
    qcsj: String,
    r#type: String,
}

static CONFIG_FILE_PATH: &str = "/home/vihv/Code/tools/work_flag/src/config.toml";

static LOGIN_URL: &str = "http://portal.zts.com.cn/cas/login?service=http%3A%2F%2F10.55.10.13%2FCASLogin";
static TABLE_URL: &str = "http://10.55.10.13/UIProcessor?Table=vem_wb_qdqc";
static TABLE_URL2: &str = "http://10.55.10.13/UserProject.do?project=itzhgl_sc";
static IN_URL: &str = "http://10.55.10.13/OperateProcessor?operate=vem_wb_qdqc_M1&amp;Table=vem_wb_qdqc&amp;Token=%s&amp;&amp;ID=%s&amp;WindowType=1&amp;extWindow=true&amp;PopupWin=true";
static OUT_URL: &str = "http://10.55.10.13/OperateProcessor?operate=vem_wb_qdqc_M2&amp;Table=vem_wb_qdqc&amp;Token=%s&amp;&amp;ID=%s&amp;WindowType=1&amp;extWindow=true&amp;PopupWin=true";

fn run() {
    let config = read_config(); // 读取配置文件
    users_login(&config); // 登录
}


fn read_config() -> Config {
    // 先读取配置文件，得到其文本字符串
    let mut config_str = String::new();
    File::open(CONFIG_FILE_PATH).expect("config.toml not found").read_to_string(&mut config_str).expect("config.toml read error");
    let config: Config = toml::from_str(&config_str).expect("config.toml parse error");
    config
}

fn users_login(config: &Config) {
    for user in config.users.iter() {
        println!("用户{}登录中...", user.username);
        user.login();
        // match user.login(config: &config) {
        //     Ok(()) => println!("用户{}登录成功", user.username),
        //     Err(e) => println!("用户{}登录失败，错误信息：{}", user.username, e),
        // }
    }
}

impl User {
    fn login(&self) {
        if let Some((lt, execution, cookie)) = self.get_parameter() {
            if let Some(req_cli) = self.post_form(lt, execution, cookie) {
                self.query_table(req_cli);
                // self.work_in(req_cli);
            }
        }
    }

    // 获取lt和execution,cookie三个参数
    fn get_parameter(&self) -> Option<(String, String, String)> {
        println!("获取所需要的参数");
        let req_cli = reqwest::blocking::Client::new();
        let res = req_cli.get(LOGIN_URL).headers(get_header("")).send();
        let mut result = ("".to_string(), "".to_string(), "".to_string());
        return match res {
            Ok(resp) => {
                let headers = resp.headers();
                if let Some(cookie) = headers.get("set-cookie") {
                    result.2 = cookie.to_str().unwrap().to_string();
                }
                let html = resp.text().unwrap();
                println!("{}", html);
                let root = Vis::load(html).unwrap();
                let inputs = root.find("input[name=lt][value]");
                if let Some(lt) = inputs.attr("value") {
                    let lt = lt.to_string();
                    result.0 = lt;
                }
                let inputs = root.find("input[name=execution][value]");
                if let Some(execution) = inputs.attr("value") {
                    let execution = execution.to_string();
                    result.1 = execution;
                }
                println!("获取所需要的参数成功");
                Some(result)
            }
            Err(_) => {
                println!("获取所需要的参数失败");
                None
            }
        };
    }

    // 提交登录表单
    fn post_form(&self, lt: String, execution: String, cookie: String) -> Option<reqwest::blocking::Client> {
        println!("提交登录表单");
        let mut md5sum = Md5::new();
        md5sum.input_str(&self.password);
        let password = md5sum.result_str();
        let mut params = HashMap::new();
        params.insert("username", self.username.clone());
        params.insert("password", password.clone());
        params.insert("lt", lt);
        params.insert("execution", execution);
        params.insert("_eventId", "submit".to_string());
        params.insert("rememberMe", "true".to_string());
        let cookie_split: Vec<&str> = cookie.split(";").collect();
        let cookie = format!("{}; {}; {};", cookie_split[0], self.username.clone(), password);

        let req_cli = reqwest::blocking::Client::new();
        let header = get_header(&cookie);
        let res = req_cli.post(LOGIN_URL).headers(header).form(&params).send();
        return match res {
            Ok(resp) => {
                let html = resp.text().unwrap();
                println!("{}", html);
                let root = Vis::load(html).unwrap();
                let msg = root.find("#msg.login_errors");
                if msg.text() != "" {
                    println!("登录失败，错误信息：{}", msg.text());
                } else {
                    println!("登录成功");
                }
                Some(req_cli)
            }
            Err(_) => {
                println!("提交登录表单失败");
                None
            }
        };
    }

    fn query_table(&self, req_cli: reqwest::blocking::Client) {
        println!("查询表格");
        // let res = req_cli.get(LOGIN_URL).headers(get_header("")).send();
        let res = req_cli.get(TABLE_URL2).send();
        return match res {
            Ok(resp) => {
                println!("{:?}", resp);
                let html = resp.text().unwrap();
                println!("{}", html);
            }
            Err(_) => {
                println!("查询表格失败");
            }
        };
    }

    fn work_in(&self, req_cli: reqwest::blocking::Client) {
        println!("尝试打卡");
        let res = req_cli.get(IN_URL).send();
        match res {
            Ok(resp) => {
                println!("{:?}", resp);
                let html = resp.text().unwrap();
                println!("{}", html);
            }
            Err(_) => {
                println!("进入工作失败");
            }
        }
    }
}

// 获取header
fn get_header(cookie: &str) -> HeaderMap {
    let mut headers = HeaderMap::new();
    headers.insert("Host", HeaderValue::from_str("portal.zts.com.cn").unwrap());
    headers.insert("Connection", HeaderValue::from_str("keep-alive").unwrap());
    headers.insert("Cache-Control", HeaderValue::from_str("max-age=0").unwrap());
    headers.insert("Upgrade-Insecure-Requests", HeaderValue::from_str("1").unwrap());
    headers.insert("User-Agent", HeaderValue::from_str("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36").unwrap());
    headers.insert("Accept", HeaderValue::from_str("text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9").unwrap());
    headers.insert("Referer", HeaderValue::from_str("http://10.55.10.13").unwrap());
    headers.insert("Accept-Language", HeaderValue::from_str("zh-CN,zh;q=0.9").unwrap());
    headers.insert("Cookie", HeaderValue::from_str(cookie).unwrap());
    headers
}