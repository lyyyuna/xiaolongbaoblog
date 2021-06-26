use clap::{App, SubCommand};

fn main() {
    let matches = App::new("xiaolongbaoblog")
                .version("0.1.0")
                .author("lyyyuna")
                .about("It is a tool to generate static blog site.")
                .subcommand(SubCommand::with_name("g")
                    .about("generate blog files"))
                .subcommand(SubCommand::with_name("s")
                    .about("serve blog in local"))   
                .get_matches();

    if let Some(_) = matches.subcommand_matches("g") {
        println!("generate")
    } else if let Some(_) = matches.subcommand_matches("s") {
        println!("serve")
    } else {
        println!("unknown sub command")
    }
}
