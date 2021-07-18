#![deny(warnings)]
#![allow(dead_code)]

use crate::config;
use crate::blog;
use crate::tpl;

use config::BlogConfig;
use blog::Blog;
use tpl::Templates;
use std::fs;
use std::path::Path;
use std::rc::Rc;
use std::collections::HashMap;

#[derive(Debug)]
pub struct Site {
    config: BlogConfig,
    blogs: Vec<Rc<Blog>>,
    categories: HashMap<String, Vec<Rc<Blog>>>,
    series: HashMap<String, Vec<Rc<Blog>>>,
    tpls: Option<Templates>,
}

impl Site {
    pub fn new(conf: BlogConfig) -> Self {
        Site {
            config: conf,
            blogs: Vec::with_capacity(50),
            categories: HashMap::with_capacity(10),
            series: HashMap::with_capacity(10),
            tpls: None,
        }
    }

    pub fn parse(&mut self) {
        let source_dir = self.config.source_dir.as_ref().unwrap();
        let post_dir = self.config.post_dir.as_ref().unwrap();
        let template_dir = self.config.template_dir.as_ref().unwrap();

        let post_sources = Path::new("./").join(source_dir).join(post_dir);
        let template_sources = Path::new("./").join(source_dir).join(template_dir);

        self.tpls = Some(Templates::new(template_sources));

        let files = fs::read_dir(post_sources).expect("cannot read the blog source directory");

        // 解析所有 md，得到原始 html
        for file in files.into_iter() {
            let dir_entry = file.expect("cannot get the item");
            if true == dir_entry.file_type().map(|ft| ft.is_file()).unwrap() {
                let mut one_blog = Blog::new(dir_entry.path());
                one_blog.parse();
                self.blogs.push(Rc::new(one_blog));
            } else {
                continue
            }
        }

        // 解析 category & series
        for blog in self.blogs.iter() {
            let ref category= blog.blog_meta.as_ref().unwrap().category;
            
            if self.categories.contains_key(category) {
                let category_slice = self.categories.get_mut(category).unwrap();
                category_slice.push(blog.clone());
            } else {
                self.categories.insert(category.clone(), vec![blog.clone()]);
            }

            match blog.blog_meta.as_ref().unwrap().series {
                Some(ref series) => {
                    if self.series.contains_key(series) {
                        let series_slice = self.series.get_mut(series).unwrap();
                        series_slice.push(blog.clone());
                    } else {
                        self.series.insert(series.clone(), vec![blog.clone()]);
                    }
                },
                None => {},
            };
        }
    }

    pub fn print_categories(&self) {
        println!("{:?}", self.categories);
    }
}
