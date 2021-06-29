#![deny(warnings)]

use std::{fs::File, io::Read, collections::BTreeMap};
use serde::Deserialize;

const NAME: &str = "config.toml";

#[derive(Debug, Deserialize)]
pub struct BlogConfig {
    // Site
    pub title: Option<String>,
    pub subtitle: Option<String>,
    pub description: Option<String>,
    pub author: Option<String>,
    pub language: Option<String>,
    // URL
    pub url: Option<String>,
    pub root: Option<String>,
    pub permalink: Option<String>,
    // Directory
    pub source: Option<String>,
    pub public_dir: Option<String>,
    // pub tag_dir: Option<String>,
    pub archive_dir: Option<String>,
    pub categories: Option<String>,
    pub i18n_dir: Option<String>,
    // Date / Time format
    pub data_format: Option<String>,
    pub time_format: Option<String>,
    // Pagination
    pub per_page: Option<String>,
    pub pagination_dir: Option<String>,
    // Deploy
    pub deploys: Option<Vec<GitConfig>>,

    // Menu
    pub menu: Option<BTreeMap<String, String>>
}

#[derive(Debug, Deserialize)]
pub struct GitConfig {
    pub repository: Option<String>,
    pub branch: Option<String>,
}

pub fn read_config() -> BlogConfig {
    let mut input = String::new();
    File::open(NAME)
        .and_then(|mut f| f.read_to_string(&mut input))
        .unwrap();

    let decoded: BlogConfig = toml::from_str(&input).unwrap();

    decoded
}
