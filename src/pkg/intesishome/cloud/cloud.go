package cloud

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	_statusCommand string = `{"status":{"hash":"x"},"config":{"hash":"x"}}`
	_statusVersion string = "1.8.5"
	Endpoint       string = "/api.php/get/control"
	Hostname       string = "https://user.intesishome.com"
)

type Response struct {
	Config struct {
		Token          int     `json:"token"`
		PushToken      string  `json:"pushToken"`
		LastAppVersion string  `json:"lastAppVersion"`
		ForceUpdate    int     `json:"forceUpdate"`
		SetDelay       float64 `json:"setDelay"`
		ServerIP       string  `json:"serverIP"`
		ServerPort     int     `json:"serverPort"`
		Hash           string  `json:"hash"`
		Inst           []struct {
			ID      int    `json:"id"`
			Order   int    `json:"order"`
			Name    string `json:"name"`
			Devices []struct {
				ID             string `json:"id"`
				Name           string `json:"name"`
				FamilyID       int    `json:"familyId"`
				ModelID        int    `json:"modelId"`
				InstallationID int    `json:"installationId"`
				ZoneID         int    `json:"zoneId"`
				Order          int    `json:"order"`
				Widgets        []int  `json:"widgets"`
			} `json:"devices"`
		} `json:"inst"`
	} `json:"config"`
	Status struct {
		Hash   string `json:"hash"`
		Status []struct {
			DeviceID int64 `json:"deviceId"`
			UID      int   `json:"uid"`
			Value    int   `json:"value"`
			Name     string
		} `json:"status"`
	} `json:"status"`
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type Cloud struct {
	username string
	password string
	hostname string
	logger   *log.Logger
}

type Option func(c *Cloud)

func New(username, password string, opts ...Option) *Cloud {
	c := &Cloud{
		username: username,
		password: password,
		hostname: Hostname,
		logger:   log.New(os.Stdout, "" /* prefix */, log.Ldate|log.Ltime|log.Lshortfile),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithHostname(host string) Option {
	return func(c *Cloud) {
		c.hostname = Hostname
		if host == "" {
			return
		}
		if !strings.Contains(host, "http://") {
			c.hostname = "http://" + host
			return
		}
		c.hostname = host
	}
}

func WithLogger(l *log.Logger) Option {
	return func(c *Cloud) {
		c.logger = l
	}
}

func (c *Cloud) Status() (r Response, err error) {
	return status(c)
}

func (c *Cloud) Token() (int, error) {
	var token int
	r, err := status(c)
	if err != nil {
		return token, err
	}
	return r.Config.Token, err
}

func (c *Cloud) Command() (string, error) {
	r, err := status(c)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", r.Config.ServerIP, r.Config.ServerPort), nil
}

func status(c *Cloud) (r Response, err error) {
	form := statusForm(c.username, c.password)
	uri := c.hostname + Endpoint
	resp, err := http.PostForm(uri, form)
	if err != nil {
		return
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected response code: %v body: %s", resp.StatusCode, body)
		return
	}
	if len(body) < 10 {
		err = fmt.Errorf("unexpected response body: %s", string(body))
		return
	}
	if err = json.Unmarshal(body, &r); err != nil {
		err = fmt.Errorf("malformed payload: %v", err.Error())
		return
	}
	if r.ErrorCode != 0 {
		err = fmt.Errorf("unexpected response error: %v message: %v", r.ErrorCode, r.ErrorMessage)
		return
	}
	return r, err
}

func statusForm(user, pass string) (ret url.Values) {
	ret = url.Values{}
	ret.Set("username", user)
	ret.Add("password", pass)
	ret.Add("version", _statusVersion)
	ret.Add("cmd", _statusCommand)
	return
}
