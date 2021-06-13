# Go Veeam Azure Assessment

This is a project that builds off the working that my friend and collegue started in creating a report for Veeam Backup for Azure.

https://jorgedelacruz.uk/2021/06/04/veeam-html-daily-report-for-veeam-backup-for-azure-is-now-available-community-project/

I'm learning Golang and found it a useful project to try out a few things including go routines and templating.

Some of the benefits of using Go are:

1. Single compiled binary file to make it easy to run (./main.exe)
2. The ability to use a configuration yaml file
3. Output to HTML via a template
4. Native SMTP support to send emails

My hope is that others can foke this project and make it better. There is also plenty of scope to extend this to other services like AWS.

A config.yaml file is required for this to work and it needs to be in the root of the directory where the binary is run from. 

The format is below:

server: "5.5.5.5"
username: "username" 
password: "password"
serverConfig:
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;from: "example@email.com"
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;to: "example2@email.com"
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;smtpHost: "smtp.server.com"
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;smtpPort: "587"
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;smtpPass: "smtpPass"

Note that the top section is for your instance of VBAzure.

The programe doesn't include any schedule but can be tied into Windows schedular or if compiled on Linux, a Cronjob. 