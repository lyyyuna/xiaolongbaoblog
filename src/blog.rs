#![deny(warnings)]
#![allow(dead_code)]

use std::fmt;
use std::str::FromStr;
use std::{fs, path::PathBuf};
use chrono::prelude::*;
use serde::{Deserialize, de};
use pulldown_cmark::{html, Options, Parser};

#[derive(Debug)]
pub struct Blog {
    path: PathBuf,
    pub url: Option<String>,
    raw_contents: String,

    pub blog_meta: Option<BlogMeta>,
    md_contents: Option<String>,
    pub html_contents: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct BlogMeta {
    pub title: String,
    #[serde(deserialize_with = "from_timestr")]
    date: DateTime<Utc>,
    pub category: String,
    pub series: Option<String>,
}

struct DateTimeVisitor;

impl<'de> de::Visitor<'de> for DateTimeVisitor {
    type Value = DateTime<Utc>;

    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        write!(formatter, "a string represents chrono::Datetime")
    }

    fn visit_str<E>(self, v: &str) -> Result<Self::Value, E>
    where E: de::Error {
        match Utc.datetime_from_str(v, "%Y-%m-%d %H:%M:%S") {
            Ok(t) => Ok(t),
            Err(_) => Err(de::Error::invalid_value(de::Unexpected::Str(v), &self)),
        }
    }
}

fn from_timestr<'de, D>(d: D) -> Result<DateTime<Utc>, D::Error>
where
    D: de::Deserializer<'de>,
{
    d.deserialize_str(DateTimeVisitor)
}

impl Blog {
    pub fn new(p: PathBuf) -> Self {
        let raw_contents = fs::read_to_string(&p).expect("fail to read the blog souce");

        Blog{
            path: p.clone(),
            raw_contents: raw_contents,
            blog_meta: None,
            url: None,
            md_contents: None,
            html_contents: None,
        }
    }

    pub fn parse(&mut self) {
        self.parse_meta_data();
        self.parse_markdown();
    }

    fn parse_meta_data(&mut self) {
        let directives: Vec<&str> = self.raw_contents.split("---").collect();
        assert!(directives.len() == 2);

        let meta_str = directives[0];
        let meta: BlogMeta = serde_yaml::from_str(meta_str).unwrap();
        self.blog_meta = Some(meta);

        self.md_contents = Some(String::from_str(directives[1]).unwrap());

        // 生成 url
        let date = self.blog_meta.as_ref().unwrap().date;
        let year = date.year();
        let month = date.month();
        let day = date.day();
        let filename = self.path.file_stem().unwrap().to_str().unwrap();

        self.url = Some(format!("{}/{}/{}/{}", year, month, day, filename));
    }

    fn parse_markdown(&mut self) {
        // 生成 html 原始内容
        let options = Options::empty();
        let parser = Parser::new_ext(self.md_contents.as_ref().unwrap(), options);

        let mut html_output = String::with_capacity(self.md_contents.as_ref().map(|s| s.len()).unwrap() * 3 / 2);
        html::push_html(&mut html_output, parser);

        self.html_contents = Some(html_output);
    }
}
