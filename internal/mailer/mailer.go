package mailer

import (
	"bytes"
	"embed"
	ht "html/template"
	"strconv"
	tt "text/template"
	"time"

	"github.com/wneessen/go-mail"
)

// Below we declare a new variable with the type embed.FS (embedded file system) to hold our email templates. This has a comment directive in the format `//go:embed <path>` IMMEDIATELY ABOVE it, which indicates to Go that we want to store the contents of the ./templates directory in the templateFS embedded file system variable.

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	client *mail.Client
	sender string
}

func New(host, port, username, password, sender string) (*Mailer, error) {

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	// Initialize options with default local development behaviors
	opts := []mail.Option{
		mail.WithPort(portInt),
		mail.WithTimeout(5 * time.Second),
	}

	if username == "" {
		// local developmet (mailpit doesn't requires username and password in local setup)
		opts = append(opts, mail.WithTLSPolicy(mail.NoTLS)) // required for local mailpit
	} else {
		// Production configuration
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthLogin),
			mail.WithUsername(username),
			mail.WithPassword(password),
			mail.WithTLSPolicy(mail.TLSMandatory),
		)
	}

	client, err := mail.NewClient(host, opts...)

	if err != nil {
		return nil, err
	}

	mailer := &Mailer{
		client: client,
		sender: sender,
	}

	return mailer, nil
}

func (m *Mailer) Send(recipient string, templateFile string, data any) error {

	textTmpl, err := tt.New("").ParseFS(templateFS, "templates/"+templateFile)

	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)

	err = textTmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)

	err = textTmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlTmpl, err := ht.New("").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)

	err = htmlTmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMsg()

	err = msg.To(recipient)
	if err != nil {
		return err
	}

	err = msg.From(m.sender)
	if err != nil {
		return err
	}

	msg.Subject(subject.String())
	msg.SetBodyString(mail.TypeTextPlain, plainBody.String())
	msg.AddAlternativeString(mail.TypeTextHTML, htmlBody.String())

	return m.client.DialAndSend(msg)
}
