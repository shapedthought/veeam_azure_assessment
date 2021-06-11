package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"
)

type TokenLoginSpec struct {
	Username  string `json:"Username"`
	Password  string `json:"Password"`
	GrantType string `json:"grant_type"`
}

type TokenModel struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
	Issued       string `json:".issued"`
	Expires      string `json:".expires"`
	UserName     string `json:"username"`
	RoleName     string `json:"rolename"`
	MFA          bool   `json:"mfa_enabled"`
}

type AboutServer struct {
	ServerVersion string `json:"serverVersion"`
	WorkerVersion string `json:"workerVersion"`
	DatabaseId    string `json:"databaseId"`
	Copyright     string `json:"copyright"`
}

type ServerInfo struct {
	SubId             string `json:"subscriptionId"`
	ServerName        string `json:"serverName"`
	AzureRegion       string `json:"azureRegion"`
	AzureVmId         string `json:"azureVmId"`
	ResourceGroup     string `json:"resourceGroup"`
	AzureEnvironment  string `json:"azureEnvironment"`
	VirtualMachineUId string `json:"virtualMachineUniqueId"`
}

type SessionLogger struct {
	JobSessionId string
	Log          []struct {
		LogTime            string `json:"logTime"`
		Status             string `json:"status"`
		Message            string `json:"message"`
		ExecutionStartTime string `json:"executionStartTime"`
		ExecutionDuration  string `json:"executionDuration"`
	} `json:"log"`
}

type OutputData struct {
	Version       string
	WorkerVersion string
	AzureRegion   string
	ServerName    string
	StartTime     string
	EndTime       string
	// Duration      string
	SessionInfo SessionInfo
	SessionLog  []SessionLog
}

var tpl *template.Template

func main() {
	// uname := "ed"
	// server := "51.105.4.34"
	// pass := "Jiu^1^jitsu-"

	// 51.105.4.34 Jiu^1^jitsu- ed
	// backup vm = azureuser

	var uname string
	var server string
	var pass string

	flag.StringVar(&uname, "u", "uname", "username")
	flag.StringVar(&server, "s", "server", "server address")
	flag.StringVar(&pass, "p", "password", "server password")

	flag.Parse()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	connstring := fmt.Sprintf("https://%s/api/oauth2/token", server)

	data := url.Values{}
	data.Set("grant_type", "Password")
	data.Set("Username", uname)
	data.Set("Password", pass)

	r, err := http.NewRequest("POST", connstring, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Accept", "application/json")

	res, err := client.Do(r)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(res.Status)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var token TokenModel

	json.Unmarshal(body, &token)

	var to = token.AccessToken

	// Channel

	ch := make(chan []byte)

	// about
	ab := fmt.Sprintf("https://%s/api/v2/system/about", server)

	getData(to, ab, client, tr, ch)

	d := <-ch

	var sd AboutServer

	json.Unmarshal(d, &sd)

	// overview
	ov := fmt.Sprintf("https://%s/api/v2/system/serverInfo", server)

	c := getData(to, ov, client, tr)

	version := sd.ServerVersion
	workerVersion := sd.WorkerVersion

	var si ServerInfo

	json.Unmarshal(c, &si)
	serverName := si.ServerName
	azureRegion := si.AzureRegion

	// Azure sessions
	// reportDateTo=$(date "+%Y-%m-%d")
	// reportDateFrom=$(date -d "$reportDateTo - 1 day" '+%Y-%m-%d')
	tTo := time.Now()
	tFrom := tTo.AddDate(0, 0, 1)
	tString := tTo.Format("2006-01-02")
	fString := tFrom.Format("2006-01-02")

	// ses := fmt.Sprintf("https://%s/api/v2/jobSessions?Types=PolicyBackup&Types=PolicySnapshot&FromUtc=%s&ToUtc=%s", server, fString, tString)
	ses := fmt.Sprintf("https://%s/api/v2/jobSessions?Types=PolicyBackup&Types=PolicySnapshot", server)

	var sin SessionInfo

	e := getData(to, ses, client, tr)

	json.Unmarshal(e, &sin)

	var sesId []string

	// Get the session ID from each session
	for _, s := range sin.Results {
		sesId = append(sesId, s.ID)
	}

	var sessLog []SessionLog

	for _, s := range sesId {
		var sessl SessionLog
		sesl := fmt.Sprintf("https://%s/api/v2/jobSessions/%s/log", server, s)
		f := getData(to, sesl, client, tr)
		json.Unmarshal(f, &sessl)
		sessLog = append(sessLog, sessl)
	}

	output := OutputData{
		Version:       version,
		WorkerVersion: workerVersion,
		AzureRegion:   azureRegion,
		ServerName:    serverName,
		StartTime:     tString,
		EndTime:       fString,
		SessionInfo:   sin,
		SessionLog:    sessLog,
	}

	nf, err := os.Create("index.html")
	if err != nil {
		log.Fatal(err)
	}
	tpl = template.Must(template.New("").ParseFiles("tpl.gohtml"))
	err = tpl.ExecuteTemplate(nf, "tpl.gohtml", output)
	if err != nil {
		log.Fatal(err)
	}
}

func getData(t string, cs string, cl *http.Client, tr *http.Transport, ch chan []byte) {

	req, _ := http.NewRequest("GET", cs, nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+t)

	res, err := cl.Do(req)

	if err != nil {
		log.Panicln(err)
	}
	defer res.Body.Close()
	bo, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panicln(err)
	}

	ch <- bo
	close(ch)
}
