extern crate ring;
extern crate hex;

use ring::{digest, hmac};
use std::time::{SystemTime, UNIX_EPOCH};

/// Stores the data provided by the user in a format compatible with the telegram_web_login_verifier::LoginVerifier
#[derive(Debug)]
pub struct RequestData {
    pub auth_date: u64,
    pub first_name: String,
    pub hash: String,
    pub id: i32,
    pub photo_url: String,
    pub username: String
}

/// Verify the provided user data with the bot token
pub struct LoginVerifier {
    key: hmac::SigningKey
}

impl LoginVerifier {
    /**
    Returns a new LoginVerifier using the provided bot token as the key

    # Arguments
    * `token` - A &str containing the bot token provided by Telegram's Botfather
    # Examples
    ```
    extern crate telegram_web_login_verifier;

    use telegram_web_login_verifier::{LoginVerifier, RequestData};

    let verifier = LoginVerifier::new("123456789:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA");
    ```

    */
    pub fn new(token: &str) -> LoginVerifier {
        let key_value = digest::digest(&digest::SHA256, token.as_bytes());


        LoginVerifier {
            key: hmac::SigningKey::new(&digest::SHA256, key_value.as_ref())
        }
    }
    fn generate_data_check_string(data: &RequestData) -> String {
        format!("auth_date={}\nfirst_name={}\nid={}\nphoto_url={}\nusername={}",
        data.auth_date,
        data.first_name,
        data.id,
        data.photo_url,
        data.username
        )
    }

    fn parse_data_check_string(id: &str, first_name: &str, last_name: &str, username: &str, 
                                photo_url: &str, auth_date: &str, hash: &str) -> Option<String> {  

    }
    /**
        Verifies if the provided login data is valid

        # Arguments
        * `data` - A reference to a telegram_web_login_verifier::RequestData struct.
        * `check_time_stamp` - If true the method will check if _auth_date_ is older than a day, in that case it will return _Err("The login request expired")_

        # Remarks
        The method will return Ok(true) if the verification succeeds, it will return an error otherwise.

        # Examples
        ```
        extern crate telegram_web_login_verifier;

        use telegram_web_login_verifier::{LoginVerifier, RequestData};

        let verifier = LoginVerifier::new("123456789:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA");

        let data = RequestData {
            auth_date: 1234567890,
            first_name: "First name".to_string(),
            hash: "d029f87e3d80f8fd9b1be67c7426b4cc1ff47b4a9d0a8461c826a59d8c5eb6cd".to_string(),
            id: 1234567,
            photo_url: "https://t.me/i/userpic/320/username.jpg".to_string(),
            username: "username".to_string()
        };

        let result = verifier.verify(&data, true);

        match result {
            Ok(_) => println!("Ok!"),
            Err(e) => println!("{}", e)
        }
        ```
    */
    pub fn verify(&self, data: &RequestData, check_time_stamp: bool) -> Result<bool, &'static str> {
        if check_time_stamp {
            let system_time = SystemTime::now().duration_since(UNIX_EPOCH).expect("UNIX_EPOCH can not be earlier than systemtime");
            if (system_time.as_secs() - data.auth_date) > 86400 {
                return Err("The login request expired")
            }
        }

        let data_check_string = LoginVerifier::generate_data_check_string(&data);
        let signature = hmac::sign(&self.key, data_check_string.as_bytes());
        let signature_string = hex::encode(signature.as_ref());

        if data.hash == signature_string {
            return Ok(true);
        } else {
            return Err("Invalid login data");
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn valid_data() {
        let v = LoginVerifier::new("537813868:AAEF2YTOOXNAWpnO1J9p6B5xs-SfU3S0lQI");
        let data = RequestData {
            auth_date: 1519574536,
            first_name: "Claudio4".to_string(),
            hash: "b4771ead3d50c8712cdded9ce5f7166eb90c630e3d43de812b7a0bc5f2885bc2".to_string(),
            id: 4039441,
            photo_url: "https://t.me/i/userpic/320/claudio4.jpg".to_string(),
            username: "claudio4".to_string()
        };
        match v.verify(&data, false) {
            Ok(b) => assert!(b, "The result should be true"),
            Err(e) => panic!("The result should be Ok but it returned the following error: {}", e)
        };
    }
    #[test]
    fn invalid_token() {
        let v = LoginVerifier::new("537813868:AAEF2YTOOXNAWpnO1J9p6B5xs-SfU3S0lQI");
        let data = RequestData {
            auth_date: 1519574536,
            first_name: "Claudio4".to_string(),
            hash: "123456789:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA".to_string(),
            id: 4039441,
            photo_url: "https://t.me/i/userpic/320/claudio4.jpg".to_string(),
            username: "claudio4".to_string()
        };
        match v.verify(&data, false) {
            Ok(_) => panic!("The result should be an error, but is true"),
            Err(e) => assert!(e == "Invalid login data", "The error should be \"Invalid login data\", but \"{}\" was returned", e)
        };
    }
    #[test]
    fn expired_data() {
        let v = LoginVerifier::new("537813868:AAEF2YTOOXNAWpnO1J9p6B5xs-SfU3S0lQI");
        let data = RequestData {
            auth_date: 946684800,
            first_name: "Claudio4".to_string(),
            hash: "b4771ead3d50c8712cdded9ce5f7166eb90c630e3d43de812b7a0bc5f2885bc2".to_string(),
            id: 4039441,
            photo_url: "https://t.me/i/userpic/320/claudio4.jpg".to_string(),
            username: "claudio4".to_string()
        };
        match v.verify(&data, true) {
            Ok(_) => panic!("The result should be an error, but is true"),
            Err(e) => assert!(e == "The login request expired", "The error should be \"The login request expired\", but \"{}\" was returned", e)
        };
    }
}
