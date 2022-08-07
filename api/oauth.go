package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	pkce "github.com/jimlambrt/go-oauth-pkce-code-verifier"
	"github.com/skratchdot/open-golang/open"
)

func PerformOAuth() (string, error) {
	var token string
	err := checkForJSON() //check JSON exists, if not, make
	if err != nil {
		return "", err
	}
	if UserInfo.ApiToken == "" {
		token, err = AuthUser() //get token
		if err != nil {
			return "", err
		}
	} else {
		token = UserInfo.ApiToken
	}
	return token, nil
}

func TimeCheck() {
	if time.Until(UserInfo.TokenExpirationDate) <= 0 {
		UserInfo.ApiToken = ""
		AuthUser()
	}
}

func AuthUser() (string, error) {
	var token string
	UserInfo.TokenCreationDate = time.Now()
	UserInfo.TokenExpirationDate = UserInfo.TokenCreationDate.AddDate(1, 0, 0) //add one year
	redirectPort, err := findPort()
	if err != nil {
		return "", err
	}

	RedirectURL := fmt.Sprintf("http://127.0.0.1:%d/", redirectPort)

	// initialize the code verifier
	var CodeVerifier, _ = pkce.CreateCodeVerifier()

	// Create code_challenge with S256 method
	codeChallenge := CodeVerifier.CodeChallengeS256()

	params := fmt.Sprintf(
		"&response_type=code"+
			"&client_id=%s"+
			"&redirect_uri=%s"+
			"&code_challenge_method=S256"+
			"&code_challenge=%s"+
			"&scope=%s",
		//"&state=2",
		ClientID, RedirectURL, codeChallenge, strings.Join(Scopes[:], " "))

	fullUrl := fmt.Sprintf("%s?%s", AuthURL, params)

	//set http server
	srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", redirectPort)}

	// define a handler that will get the authorization code, call the token endpoint, and close the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// get the authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			io.WriteString(w, "Error: could not find 'code' URL parameter\n")
			cleanup(srv) // close the HTTP server and return
			return
		}
		// trade the authorization code and the code verifier for an access token
		token, err = getAccessToken(code, CodeVerifier.String(), RedirectURL, ClientID)
		if err != nil {
			io.WriteString(w, `
			<html>
				<body>
					<h1>Error</h1>
					<h2>could not retrieve access token</h2>
				</body>
			</html>`)
			cleanup(srv) // close the HTTP server and return
			return
		}
		UserInfo.ApiToken = token
		b, err := json.Marshal(&UserInfo)
		if err != nil {
			io.WriteString(w, `
			<html>
				<body>
					<h1>Error</h1>
					<h2>could not store access token</h2>
				</body>
			</html>`)
			cleanup(srv) // close the HTTP server and return
			return
		}
		err = os.WriteFile(json_path, b, 0644)
		if err != nil {
			//fmt.Println("could not write config file")
			io.WriteString(w, `
			<html>
				<body>
					<h1>Error</h1>
					<h2>could not store access token</h2>
				</body>
			</html>`)
			cleanup(srv) // close the HTTP server and return
			return
		}
		// return an indication of success to the caller
		io.WriteString(w, `
		<html>
			<body>
				<h1>Login successful!</h1>
				<h2>Success, you may close this page.</h2>
			</body>
		</html>`)
		cleanup(srv) // close the HTTP server
	})
	err = open.Start(fullUrl)
	if err != nil {
		return "", err
	}
	srv.ListenAndServe() //start http server
	return token, nil
}

// getAccessToken trades the authorization code retrieved from the first OAuth2 leg for an access token
func getAccessToken(code string, codeVerifier string, RedirectURL string, ClientID string) (string, error) {
	tokenParameters := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {codeVerifier},
		"redirect_uri":  {RedirectURL},
		"client_id":     {ClientID},
	}
	//application/x-www-form-urlencoded

	//fmt.Println(RedirectURL)

	// create the request and execute it
	res, err := http.PostForm(TokenURL, tokenParameters)

	body, _ := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	//fmt.Println("HTTP Response Status:", res.StatusCode, http.StatusText(res.StatusCode))
	if res.StatusCode == 400 {
		//log.Printf(string(body))
		return "", fmt.Errorf("wrong http code")
	}
	// process the response

	// unmarshal the json into a string map
	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		//fmt.Printf("JSON error: %s", err)
		return "", err
	}

	// retrieve the access token out of the map, and return to caller
	if !isNil(responseData["access_token"]) {
		accessToken := responseData["access_token"].(string)
		return accessToken, nil
	}
	return "", fmt.Errorf("interface is nil")
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func findPort() (int, error) {
	port := 8000
	// foundOpenPort := false

	for port < 8100 {

		host := fmt.Sprintf("127.0.0.1:%d", port)

		////fmt.Printf("Trying %s\n", host)
		ln, err := net.Listen("tcp", host)
		if err != nil {
			//fmt.Fprintf(os.Stderr, "Can't listen on port %d: %s", port, err)
			////fmt.Printf("TCP Port %d is not available\n", port)
			// move to next port
			port++
			continue
		} else {
			_ = ln.Close()
			//fmt.Printf("TCP Port %d is available\n", port)
			return port, nil
		}
	}

	err := fmt.Errorf("unable to find an open port, failing")
	return 0, err

}

func cleanup(server *http.Server) {
	// we run this as a goroutine so that this function falls through and
	// the socket to the browser gets flushed/closed before the server goes away
	go server.Close()
}

func checkForJSON() error {
	TimeCheck()
	_, err := os.Stat(json_path)
	if err == nil {
		jsonFile, err := os.Open(json_path)
		if err != nil {
			return err
		}
		defer jsonFile.Close()
		byteValue, _ := io.ReadAll(jsonFile)
		err = json.Unmarshal(byteValue, &UserInfo)
		if err != nil {
			return err
		}
	} else if errors.Is(err, os.ErrNotExist) { // path/to/whatever does *not* exist
		b, err := json.Marshal(&UserInfo)
		if err != nil {
			return err
		}
		err = os.WriteFile(json_path, b, 0644)
		if err != nil {
			return err
		}
	} else {
		return err // file may or may not exist. See err for details.
	}
	return nil
}
