#![deny(warnings)]
#![allow(dead_code)]

use std::{fs::File, io::Read, path::{PathBuf}};

#[derive(Debug)]
pub struct Templates {
    pub index_template: String,
    pub about_template: String,
    pub post_template: String,
    pub categories_template: String,
    pub series_template: String,
}

impl Templates {
    pub fn new(p: PathBuf) -> Self {

        let base_path = p;

        let mut index_ss = String::new();
        File::open(base_path.join("index.html")).
            and_then(|mut f| f.read_to_string(&mut index_ss)).unwrap();

        let mut about_ss = String::new();
        File::open(base_path.join("about.html")).
            and_then(|mut f| f.read_to_string(&mut about_ss)).unwrap();

        let mut post_ss = String::new();
        File::open(base_path.join("post.html")).
            and_then(|mut f| f.read_to_string(&mut post_ss)).unwrap();

        let mut categories_ss = String::new();
        File::open(base_path.join("categories.html")).
            and_then(|mut f| f.read_to_string(&mut categories_ss)).unwrap();

        let mut series_ss = String::new();
        File::open(base_path.join("series.html")).
            and_then(|mut f| f.read_to_string(&mut series_ss)).unwrap();

        return Templates {
            index_template: index_ss,
            about_template: about_ss,
            post_template: post_ss,
            categories_template: categories_ss,
            series_template: series_ss,
        }
    }
}