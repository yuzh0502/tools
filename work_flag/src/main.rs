use std::fs::File;
use std::io::Read;

use serde_derive::Deserialize;

fn main() {
    read_config(); // 读取配置文件
}

#[derive(Debug, Deserialize)]
struct WorkFlag {
    #[serde(rename = "BaseInfo")]
    base_info: BaseInfo,
    #[serde(rename = "User")]
    users: Vec<User>,
}

#[derive(Debug, Deserialize)]
struct BaseInfo {
    login_url: String,
    user_url: String,
    table_url: String,
    in_url: String,
    out_url: String,
}

#[derive(Debug, Deserialize)]
struct User {
    username: String,
    password: String,
}

static CONFIG_FILE_PATH: &str = "/home/vihv/Code/tools/work_flag/src/config.toml";

fn read_config() {
    // 先读取配置文件，得到其文本字符串
    let mut config_str = String::new();
    File::open(CONFIG_FILE_PATH).expect("config.toml not found").read_to_string(&mut config_str).expect("config.toml read error");
    let config: WorkFlag = toml::from_str(&config_str).expect("config.toml parse error");
    println!("{:#?}", config);
}