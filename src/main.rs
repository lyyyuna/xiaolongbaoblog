use clap::{App, SubCommand};
mod config;
mod site;
mod blog;
mod tpl;

use site::Site;
use config::{*};

fn main() {
    let matches = App::new("xiaolongbaoblog")
                .version("0.1.0")
                .author("lyyyuna")
                .about("It is a tool to generate static blog site.")
                .subcommand(SubCommand::with_name("g")
                    .about("generate blog files"))
                .subcommand(SubCommand::with_name("s")
                    .about("serve blog in local"))   
                .subcommand(SubCommand::with_name("d")
                    .about("deploy the blog to some git repo"))   
                .get_matches();

    let blog_cfg = read_config();

    if let Some(_) = matches.subcommand_matches("g") {
        println!("generate");
        let mut s = Site::new(blog_cfg);
        s.parse();
        s.print_categories();
    } else if let Some(_) = matches.subcommand_matches("s") {
        println!("serve")
    } else if let Some(_) = matches.subcommand_matches("d") {
        println!("deploy")
    } else {
        println!("unknown sub command")
    }
}
