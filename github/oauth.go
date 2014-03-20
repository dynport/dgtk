package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

func OauthToken(clientId string, clientSecret string) (token string, e error) {
	l, e := net.Listen("tcp", ":1112")
	if e != nil {
		return "", e
	}
	defer l.Close()
	go http.Serve(l, &authHandler{clientId: clientId, clientSecret: clientSecret})
	c := exec.Command("open", "https://github.com/login/oauth/authorize?client_id=39e94efd6646b79936b4&scope=read:repo,read:user,gist")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	e = c.Run()

	code := <-finished
	return code, nil
}

type oauthResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type authHandler struct {
	clientId     string
	clientSecret string
}

var finished = make(chan string)

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.NotFound(w, r)
		return
	}
	params := url.Values{
		"client_id":     {h.clientId},
		"client_secret": {h.clientSecret},
		"code":          {code},
	}
	buf := bytes.NewBufferString(params.Encode())

	e := func() error {
		req, e := http.NewRequest("POST", "https://github.com/login/oauth/access_token", buf)
		if e != nil {
			return e
		}
		req.Header.Set("Accept", "application/json")
		rsp, e := http.DefaultClient.Do(req)
		if e != nil {
			return e
		}
		defer rsp.Body.Close()
		b, e := ioutil.ReadAll(rsp.Body)
		if e != nil {
			return e
		}
		if rsp.Status[0] != '2' {
			return fmt.Errorf("expected status 2xx, got %s. body=%q", rsp.Status, string(b))
		}

		oauthResponse := &oauthResponse{}
		e = json.Unmarshal(b, &oauthResponse)
		if e != nil {
			return e
		}
		finished <- oauthResponse.AccessToken
		return nil
	}()
	if e != nil {
		w.WriteHeader(500)
		w.Write([]byte(e.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("all ok, you can close your browser now"))
}
