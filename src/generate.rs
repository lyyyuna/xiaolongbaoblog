#![deny(warnings)]
#![allow(dead_code)]

use crate::config;
use config::BlogConfig;

pub struct Site<'a> {
    config: &'a BlogConfig,
}

impl<'a> Site<'a> {
    pub fn new(config: &'a BlogConfig) -> Self {
        Site {
            config: config,
        }
    }
}
