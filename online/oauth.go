package online

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

	cv "github.com/jimlambrt/go-oauth-pkce-code-verifier"
	"github.com/skratchdot/open-golang/open"
)

//TODO: check if token is expired

func do_oauth() {
	err := checkForJSON()
	if err != nil {
		//fmt.Printf("can't read/write to JSON: %s\n", err)
		return
		//os.Exit(1)
	}
	if UserInfo.ApiToken == "" {
		AuthUser()
	}

	//close(Ready)
	return
}

func AuthUser() {
	//get redirect url from http server
	redirectPort, err := findPort()

	if err != nil {
		//fmt.Printf("can't find port on localhost: %s\n", err)
		return
		//os.Exit(1)
	}
	//fmt.Printf("got port %d\n", redirectPort)
	RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/", redirectPort)
	//RedirectURL = fmt.Sprintf("http://localhost:%d", redirectPort)
	//fmt.Printf("RedirectURL %v\n", RedirectURL)

	// initialize the code verifier
	var CodeVerifier, _ = cv.CreateCodeVerifier()

	// Create code_challenge with S256 method
	codeChallenge := CodeVerifier.CodeChallengeS256()
	//fmt.Printf("challenge:%s\n", codeChallenge)

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
			//fmt.Println("Url Param 'code' is missing")
			io.WriteString(w, "Error: could not find 'code' URL parameter\n")

			// close the HTTP server and return
			cleanup(srv)
			return
		}
		//fmt.Printf("code is %v\n", code)

		// trade the authorization code and the code verifier for an access token
		// codeVerifier := CodeVerifier.String()
		token, err := getAccessToken(code, CodeVerifier.String(), RedirectURL, ClientID)
		if err != nil {
			//fmt.Printf("could not get access token: %v\n", err)
			// io.WriteString(w, "Error: could not retrieve access token\n")

			io.WriteString(w, `
			<html>
				<body>
					<h1>Error</h1>
					<h2>could not retrieve access token</h2>
				</body>
			</html>`)

			// close the HTTP server and return
			cleanup(srv)
			return
		}

		UserInfo.ApiToken = token
		b, err := json.Marshal(&UserInfo)
		if err != nil {
			//fmt.Println("could not write config file")
			io.WriteString(w, `
			<html>
				<body>
					<h1>Error</h1>
					<h2>could not store access token</h2>
				</body>
			</html>`)

			// close the HTTP server and return
			cleanup(srv)
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

			// close the HTTP server and return
			cleanup(srv)
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

		// close the HTTP server
		cleanup(srv)
	})

	//fmt.Printf("opening browser\n")
	err = open.Start(fullUrl)
	if err != nil {
		//fmt.Printf("can't open browser to URL %s: %s\n", AuthURL, err)
		return
		//os.Exit(1)
	}
	//fmt.Printf("started server\n")
	//start http server
	srv.ListenAndServe()
	//log.Fatal(srv.ListenAndServe())
	return
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

	err := fmt.Errorf("Unable to find an open port, failing")
	return 0, err

}

func cleanup(server *http.Server) {
	// we run this as a goroutine so that this function falls through and
	// the socket to the browser gets flushed/closed before the server goes away
	go server.Close()
}

func checkForJSON() error {
	if _, err := os.Stat(json_path); err == nil {

		jsonFile, err := os.Open(json_path)
		defer jsonFile.Close()
		// if we os.Open returns an error then handle it
		if err != nil {
			return err
		}
		byteValue, _ := io.ReadAll(jsonFile)
		err = json.Unmarshal(byteValue, &UserInfo)
		if err != nil {
			return err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		//fmt.Printf("json does not exist\n")
		// path/to/whatever does *not* exist

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
	//fmt.Printf("json checking done no errors\n")
	return nil
}
