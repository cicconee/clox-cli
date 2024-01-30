package prompt

import (
	"fmt"
	"strings"
)

// InString prints msg and takes a string input from the user. The input value will
// be stored in dst. The prompt is formatted as "msg: ".
func InString(msg string, dst *string) {
	fmt.Printf("%s: ", msg)
	fmt.Scanln(dst)
}

func Password() string {
	var password string
	InString("Password", &password)
	return password
}

// ConfigureAPIToken will prompt the user to enter an API token. If an empty value is
// entered, it will loop until user enters a value. Once a valid API token is
// entered, it will return it.
func ConfigureAPIToken() string {
	var token string

	for {
		InString("API Token", &token)
		token = strings.TrimSpace(token)
		if token != "" {
			break
		}

		fmt.Println("Token cannot be empty")
	}

	return token
}

// ConfigurePassword will prompt the user to enter and confirm a password. If
// passwords do not match, it will loop until user confirms a valid password. Once a
// password is confirmed, it will be returned.
func ConfigurePassowrd() string {
	var pass string
	var confirmPass string

	for {
		InString("Password", &pass)
		InString("Confirm Password", &confirmPass)

		if pass == confirmPass {
			break
		}

		fmt.Println("Passwords do not match")
		pass = ""
		confirmPass = ""
	}

	return pass
}
