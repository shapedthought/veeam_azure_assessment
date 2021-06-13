package main

import (
	"bytes"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"
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

type CredSpec struct {
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Server       string `yaml:"server"`
	ServerConfig struct {
		From     string `yaml:"from"`
		To       string `yaml:"to"`
		SmtpHost string `yaml:"smtpHost"`
		SmtpPort string `yaml:"smtpPort"`
		SmtpPass string `yaml:"smtpPass"`
	} `yaml:"serverConfig"`
}

type SessionId struct {
	SessionId  string
	PolicyName string
}

var tpl *template.Template

var fm = template.FuncMap{
	"df": dataFormat,
	"dr": durationTime,
}

func dataFormat(c string) string {
	d := strings.Split(c, ".")
	dd := strings.Split(d[0], "T")
	t := fmt.Sprintf("%s %s", dd[1], dd[0])
	return t
}

func durationTime(c string) string {
	d := strings.Split(c, ".")
	return d[0]
}

func main() {

	var creds CredSpec
	yml, err := os.Open("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(yml)
	if err != nil {
		log.Fatal(err)
	}

	yaml.Unmarshal(b, &creds)
	if err != nil {
		log.Fatal(err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	connstring := fmt.Sprintf("https://%s/api/oauth2/token", creds.Server)

	data := url.Values{}
	data.Set("grant_type", "Password")
	data.Set("Username", creds.Username)
	data.Set("Password", creds.Password)

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

	// write to file
	_ = ioutil.WriteFile("creds.json", body, 0644)

	var to = token.AccessToken

	var wg sync.WaitGroup

	// about
	wg.Add(1)
	cha := make(chan []byte)
	ab := fmt.Sprintf("https://%s/api/v2/system/about", creds.Server)

	go getData(to, ab, client, tr, cha, &wg)

	d := <-cha

	var sd AboutServer

	json.Unmarshal(d, &sd)
	version := sd.ServerVersion
	workerVersion := sd.WorkerVersion

	// overview
	wg.Add(1)
	chb := make(chan []byte)
	ov := fmt.Sprintf("https://%s/api/v2/system/serverInfo", creds.Server)

	go getData(to, ov, client, tr, chb, &wg)

	c := <-chb

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
	ses := fmt.Sprintf("https://%s/api/v2/jobSessions?Types=PolicyBackup&Types=PolicySnapshot", creds.Server)

	var sin SessionInfo
	wg.Add(1)
	chc := make(chan []byte)
	go getData(to, ses, client, tr, chc, &wg)

	e := <-chc

	json.Unmarshal(e, &sin)

	// var sesId []string

	var sesIdStruct []SessionId

	// Get the session ID from each session
	for _, s := range sin.Results {
		se := SessionId{
			SessionId:  s.ID,
			PolicyName: s.BackupJobInfo.PolicyName,
		}
		// fmt.Println(se)
		sesIdStruct = append(sesIdStruct, se)
		// sesId = append(sesId, s.ID)
	}

	var sessLog []SessionLog
	chf := make(chan []byte)
	var f []byte

	for _, s := range sesIdStruct {
		var sessl SessionLog
		wg.Add(1)
		sesl := fmt.Sprintf("https://%s/api/v2/jobSessions/%s/log", creds.Server, s.SessionId)
		go getData(to, sesl, client, tr, chf, &wg)
		f = <-chf
		json.Unmarshal(f, &sessl)
		sessLog = append(sessLog, sessl)
	}

	// for _, s := range sesId {
	// 	var sessl SessionLog
	// 	wg.Add(1)
	// 	sesl := fmt.Sprintf("https://%s/api/v2/jobSessions/%s/log", creds.Server, s)
	// 	go getData(to, sesl, client, tr, chf, &wg)
	// 	f = <-chf
	// 	json.Unmarshal(f, &sessl)
	// 	sessLog = append(sessLog, sessl)
	// }

	wg.Wait()

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

	// nf, err := os.Create("index.html")
	if err != nil {
		log.Fatal(err)
	}
	tpl = template.Must(template.New("").Funcs(fm).ParseFiles("tpl.gohtml"))

	sendEmail(
		creds.ServerConfig.From,
		creds.ServerConfig.To,
		creds.ServerConfig.SmtpHost,
		creds.ServerConfig.SmtpPort,
		creds.ServerConfig.SmtpPass,
		tpl,
		output)
	// err = tpl.ExecuteTemplate(nf, "tpl.gohtml", output)
	if err != nil {
		log.Fatal(err)
	}
}

func getData(t string, cs string, cl *http.Client, tr *http.Transport, ch chan []byte, wg *sync.WaitGroup) {
	req, _ := http.NewRequest("GET", cs, nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+t)

	res, err := cl.Do(req)

	if res.StatusCode != 200 {
		fmt.Println(cs, res.Status)
	}

	if err != nil {
		log.Panicln(err)
	}
	defer res.Body.Close()
	bo, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panicln(err)
	}

	ch <- bo
	wg.Done()
}

func sendEmail(from string, to string, serv string, port string, pass string, tpl *template.Template, output OutputData) {
	// https://www.loginradius.com/blog/async/sending-emails-with-golang/

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
	_, err := body.Write([]byte(fmt.Sprintf("To: %s;\nFrom: %s;\nSubject: Azure Backup Report \n%s\n\n", to, from, mimeHeaders)))
	if err != nil {
		log.Fatal(err)
	}
	err = tpl.ExecuteTemplate(&body, "tpl.gohtml", output)
	if err != nil {
		log.Fatal(err)
	}

	tos := []string{
		to,
	}

	auth := smtp.PlainAuth("", from, pass, serv)
	host := fmt.Sprintf("%s:%s", serv, port)
	// fmt.Println(body.String())
	err = smtp.SendMail(host, auth, from, tos, body.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Email Sent!")
}
