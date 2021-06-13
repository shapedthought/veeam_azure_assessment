# Go Veeam Azure Assessment

This is a project that builds off the excellent work that my friend Jorge De La Cruze has done in puttig together a Azure job report. I'm also doing this in an effort to get Jorge to convert to Go!

https://jorgedelacruz.uk/2021/06/04/veeam-html-daily-report-for-veeam-backup-for-azure-is-now-available-community-project/

I'm learning Golang and found it a useful project to try out a few things including go routines, channels and templating.

Some of the benefits of using Go are:

1. Single compiled binary file to make it easy to distribute and run
2. The ability to use a configuration yaml file
3. Output to HTML
4. Native SMTP support to send emails

My hope is that others can foke this project and make it better. There is also plenty of scope to extend this to other services like AWS.

A config.yaml file is required for this to work and it needs to be in the root of the directory where the binary is run from. 

The format is below:

	server: "5.5.5.5"
	username: "username" 
	password: "password"
	serverConfig:
	  from: "example@email.com"
	  to: "example2@email.com"
	  smtpHost: "smtp.server.com"
	  smtpPort: "587"
	  smtpPass: "smtpPass"

Note that this uses the PlainAuth method which means that TLS is required or it will fail. This means you will need to use <b>port 587</b> though it isn't hardcoded.

https://golang.org/pkg/net/smtp/

The programe doesn't include any schedule but can be tied into Windows schedular or if compiled on Linux, a cron job. 

You can either download the released binary from the releases on the right handside or you can clone the repo, install Golang and then run the following in the terminal.

	go run .

You can also compile it yourself using.

	go build .

If you want to compile it for another architecture please refer to the Go documentation.

### To Do

1. The date filtering isn't working currently so each report sends all job information.
2. General refactoring