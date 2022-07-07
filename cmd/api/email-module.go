package MaoApi

var (
	EmailModuleRegisterName = "email-module"
)

type EmailMessage struct {
	Subject string
	Content string
}

type EmailModule interface {
	SendEmail(message *EmailMessage)
}