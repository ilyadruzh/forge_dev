extern crate sha3;
extern crate web3;
extern crate rustc_hex;
pub extern crate futures;

use rustc_hex::FromHex;
use std::env;
use sha3::{Digest, Keccak256};

// pub use web3::api::Web3Main as Web3;
// pub use web3::api::ErasedWeb3;

use web3::*;
use web3::api;
use web3::api::Personal;
use web3::api::Namespace;
use web3::types::{Address, U256};
use web3::types::TransactionRequest;
use web3::Error;
use web3::error::ErrorKind;
use web3::futures::Future;

extern crate docopt;
extern crate env_logger;
extern crate ethkey;
extern crate panic_hook;
extern crate parity_wordlist;
extern crate rustc_hex;
extern crate serde;
extern crate threadpool;

#[macro_use]
extern crate serde_derive;

use std::num::ParseIntError;
use std::{env, fmt, process, io, sync};

use docopt::Docopt;
use ethkey::{KeyPair, Random, Brain, BrainPrefix, Prefix, Error as EthkeyError, Generator, sign, verify_public, verify_address, brain_recover};
use rustc_hex::{ToHex, FromHex, FromHexError};

pub const USAGE: &'static str = r#"
Ethereum keys generator.
  Copyright 2016, 2017 Parity Technologies (UK) Ltd

Usage:
    ethkey info <secret-or-phrase> [options]
    ethkey generate random [options]
    ethkey generate prefix <prefix> [options]
    ethkey sign <secret> <message>
    ethkey verify public <public> <signature> <message>
    ethkey verify address <address> <signature> <message>
    ethkey recover <address> <known-phrase>
    ethkey [-h | --help]

Options:
    -h, --help         Display this message and exit.
    -s, --secret       Display only the secret.
    -p, --public       Display only the public.
    -a, --address      Display only the address.
    -b, --brain        Use parity brain wallet algorithm.

Commands:
    info               Display public and address of the secret.
    generate random    Generates new random ethereum key.
    generate prefix    Random generation, but address must start with a prefix.
    sign               Sign message using secret.
    verify             Verify signer of the signature.
    recover            Try to find brain phrase matching given address from partial phrase.
"#;

#[derive(Debug, Deserialize)]
struct Args {
	cmd_info: bool,
	cmd_generate: bool,
	cmd_random: bool,
	cmd_prefix: bool,
	cmd_sign: bool,
	cmd_verify: bool,
	cmd_public: bool,
	cmd_address: bool,
	cmd_recover: bool,
	arg_prefix: String,
	arg_secret: String,
	arg_secret_or_phrase: String,
	arg_known_phrase: String,
	arg_message: String,
	arg_public: String,
	arg_address: String,
	arg_signature: String,
	flag_secret: bool,
	flag_public: bool,
	flag_address: bool,
	flag_brain: bool,
}

#[derive(Debug)]
enum Error {
	Ethkey(EthkeyError),
	FromHex(FromHexError),
	ParseInt(ParseIntError),
	Docopt(docopt::Error),
	Io(io::Error),
}

impl From<EthkeyError> for Error {
	fn from(err: EthkeyError) -> Self {
		Error::Ethkey(err)
	}
}

impl From<FromHexError> for Error {
	fn from(err: FromHexError) -> Self {
		Error::FromHex(err)
	}
}

impl From<ParseIntError> for Error {
	fn from(err: ParseIntError) -> Self {
		Error::ParseInt(err)
	}
}

impl From<docopt::Error> for Error {
	fn from(err: docopt::Error) -> Self {
		Error::Docopt(err)
	}
}

impl From<io::Error> for Error {
	fn from(err: io::Error) -> Self {
		Error::Io(err)
	}
}

impl fmt::Display for Error {
	fn fmt(&self, f: &mut fmt::Formatter) -> Result<(), fmt::Error> {
		match *self {
			Error::Ethkey(ref e) => write!(f, "{}", e),
			Error::FromHex(ref e) => write!(f, "{}", e),
			Error::ParseInt(ref e) => write!(f, "{}", e),
			Error::Docopt(ref e) => write!(f, "{}", e),
			Error::Io(ref e) => write!(f, "{}", e),
		}
	}
}

enum DisplayMode {
	KeyPair,
	Secret,
	Public,
	Address,
}

impl DisplayMode {
	fn new(args: &Args) -> Self {
		if args.flag_secret {
			DisplayMode::Secret
		} else if args.flag_public {
			DisplayMode::Public
		} else if args.flag_address {
			DisplayMode::Address
		} else {
			DisplayMode::KeyPair
		}
	}
}

fn main() {

    let my_account: Address = "0x1D495256B26893A351b1f185B16ee314d7749dCb".parse().unwrap();

    env::set_var("RUST_BACKTRACE", "1");

    let (_eloop, transport) = web3::transports::Http::new("http://127.0.0.1:8545").unwrap();
    let web3 = web3::Web3::new(transport);
    
    let accounts = web3.eth().accounts().wait().unwrap();
    println!("Accounts: {:?}", accounts);

    let mut hasher = Keccak256::default();
    hasher.input(b"abc");
    let out = hasher.result();
    println!("HASH = {:x}", out);
    
    while true{
		
		let mut brain = BrainPrefix::new(vec![0], usize::max_value(), BRAIN_WORDS);
		let keypair = brain.generate();//?;
		let phrase = format!("recovery phrase: {}", brain.phrase());
		println!("{:?}", keypair);
	}


}


fn display(result: (KeyPair, Option<String>), mode: DisplayMode) -> String {
	let keypair = result.0;
	match mode {
		DisplayMode::KeyPair => match result.1 {
			Some(extra_data) => format!("{}\n{}", extra_data, keypair),
			None => format!("{}", keypair)
		},
		DisplayMode::Secret => format!("{}", keypair.secret().to_hex()),
		DisplayMode::Public => format!("{:?}", keypair.public()),
		DisplayMode::Address => format!("{:?}", keypair.address()),
	}
}

fn execute<S, I>(command: I) -> Result<String, Error> where I: IntoIterator<Item=S>, S: AsRef<str> {
	let args: Args = Docopt::new(USAGE)
		.and_then(|d| d.argv(command).deserialize())?;

	return if args.cmd_info {
		let display_mode = DisplayMode::new(&args);

		let result = if args.flag_brain {
			let phrase = args.arg_secret_or_phrase;
			let phrase_info = validate_phrase(&phrase);
			let keypair = Brain::new(phrase).generate().expect("Brain wallet generator is infallible; qed");
			(keypair, Some(phrase_info))
		} else {
			let secret = args.arg_secret_or_phrase.parse().map_err(|_| EthkeyError::InvalidSecret)?;
			(KeyPair::from_secret(secret)?, None)
		};
		Ok(display(result, display_mode))
	} else if args.cmd_generate {
		let display_mode = DisplayMode::new(&args);
		let result = if args.cmd_random {
			if args.flag_brain {
				let mut brain = BrainPrefix::new(vec![0], usize::max_value(), BRAIN_WORDS);
				let keypair = brain.generate()?;
				let phrase = format!("recovery phrase: {}", brain.phrase());
				(keypair, Some(phrase))
			} else {
				(Random.generate()?, None)
			}
		} else if args.cmd_prefix {
			let prefix = args.arg_prefix.from_hex()?;
			let brain = args.flag_brain;
			in_threads(move || {
				let iterations = 1024;
				let prefix = prefix.clone();
				move || {
					let prefix = prefix.clone();
					let res = if brain {
						let mut brain = BrainPrefix::new(prefix, iterations, BRAIN_WORDS);
						let result = brain.generate();
						let phrase = format!("recovery phrase: {}", brain.phrase());
						result.map(|keypair| (keypair, Some(phrase)))
					} else {
						let result = Prefix::new(prefix, iterations).generate();
						result.map(|res| (res, None))
					};

					Ok(res.map(Some).unwrap_or(None))
				}
			})?
		} else {
			return Ok(format!("{}", USAGE))
		};
		Ok(display(result, display_mode))
	} else if args.cmd_sign {
		let secret = args.arg_secret.parse().map_err(|_| EthkeyError::InvalidSecret)?;
		let message = args.arg_message.parse().map_err(|_| EthkeyError::InvalidMessage)?;
		let signature = sign(&secret, &message)?;
		Ok(format!("{}", signature))
	} else if args.cmd_verify {
		let signature = args.arg_signature.parse().map_err(|_| EthkeyError::InvalidSignature)?;
		let message = args.arg_message.parse().map_err(|_| EthkeyError::InvalidMessage)?;
		let ok = if args.cmd_public {
			let public = args.arg_public.parse().map_err(|_| EthkeyError::InvalidPublic)?;
			verify_public(&public, &signature, &message)?
		} else if args.cmd_address {
			let address = args.arg_address.parse().map_err(|_| EthkeyError::InvalidAddress)?;
			verify_address(&address, &signature, &message)?
		} else {
			return Ok(format!("{}", USAGE))
		};
		Ok(format!("{}", ok))
	} else if args.cmd_recover {
		let display_mode = DisplayMode::new(&args);
		let known_phrase = args.arg_known_phrase;
		let address = args.arg_address.parse().map_err(|_| EthkeyError::InvalidAddress)?;
		let (phrase, keypair) = in_threads(move || {
			let mut it = brain_recover::PhrasesIterator::from_known_phrase(&known_phrase, BRAIN_WORDS);
			move || {
				let mut i = 0;
				while let Some(phrase) = it.next() {
					i += 1;

					let keypair = Brain::new(phrase.clone()).generate().unwrap();
					if keypair.address() == address {
						return Ok(Some((phrase, keypair)))
					}

					if i >= 1024 {
						return Ok(None)
					}
				}

				Err(EthkeyError::Custom("Couldn't find any results.".into()))
			}
		})?;
		Ok(display((keypair, Some(phrase)), display_mode))
	} else {
		Ok(format!("{}", USAGE))
	}
}

const BRAIN_WORDS: usize = 12;

fn validate_phrase(phrase: &str) -> String {
	match Brain::validate_phrase(phrase, BRAIN_WORDS) {
		Ok(()) => format!("The recovery phrase looks correct.\n"),
		Err(err) => format!("The recover phrase was not generated by Parity: {}", err)
	}
}

fn in_threads<F, X, O>(prepare: F) -> Result<O, EthkeyError> where
	O: Send + 'static,
	X: Send + 'static,
	F: Fn() -> X,
	X: FnMut() -> Result<Option<O>, EthkeyError>,
{
	let pool = threadpool::Builder::new().build();

	let (tx, rx) = sync::mpsc::sync_channel(1);
	let is_done = sync::Arc::new(sync::atomic::AtomicBool::default());

	for _ in 0..pool.max_count() {
		let is_done = is_done.clone();
		let tx = tx.clone();
		let mut task = prepare();
		pool.execute(move || {
			loop {
				if is_done.load(sync::atomic::Ordering::SeqCst) {
					return;
				}

				let res = match task() {
					Ok(None) => continue,
					Ok(Some(v)) => Ok(v),
					Err(err) => Err(err),
				};

				// We are interested only in the first response.
				let _ = tx.send(res);
			}
		});
	}

	if let Ok(solution) = rx.recv() {
		is_done.store(true, sync::atomic::Ordering::SeqCst);
		return solution;
	}

	Err(EthkeyError::Custom("No results found.".into()))
}