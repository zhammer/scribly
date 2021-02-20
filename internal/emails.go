package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/vanng822/go-premailer/premailer"
)

var css string
var websiteURL = os.Getenv("WEBSITE_URL")

type viewData struct {
	Data interface{}
}

func (v viewData) StoryLink(story Story) template.HTML {
	return template.HTML(fmt.Sprintf(`<a href="%s/stories/%d">%s</a>`, websiteURL, story.ID, story.Title))
}

func (v viewData) WebsiteURL() string {
	return websiteURL
}

func (v viewData) Replace(original string, pattern string, replacement string) string {
	return strings.ReplaceAll(original, pattern, replacement)
}

func (v viewData) WhoseTurnText(story Story, recipient User) string {
	subject := story.CurrentWriter.Username + "'s"
	if story.CurrentWriterID == recipient.ID {
		subject = "your"
	}
	return fmt.Sprintf("it's %s turn", subject)
}

func (v viewData) CSS() template.HTML {
	return template.HTML("<style>" + css + "</style>")
}

func BuildNudgeEmail(nudger User, nudgee User, story Story) (*Email, error) {
	subject := fmt.Sprintf("%s nudged you to take your turn on %s", nudger.Username, story.Title)
	data := map[string]interface{}{
		"Story":  story,
		"Nudger": nudger,
	}
	body, err := renderTemplateWithCSS("nudge.tmpl", data)
	if err != nil {
		return nil, err
	}

	return &Email{
		Subject: subject,
		Body:    body,
		To:      nudgee,
	}, nil
}

func BuildAddedToStoryEmails(story Story) ([]Email, error) {
	var recipients []User
	for _, cowriter := range story.Cowriters {
		if cowriter.User.ID != story.CreatedByID && cowriter.User.EmailVerificationStatus == EmailVerificationStateVerified {
			recipients = append(recipients, cowriter.User)
		}
	}

	var emails []Email
	for _, recipient := range recipients {
		data := map[string]interface{}{
			"Story":     story,
			"Recipient": recipient,
		}
		body, err := renderTemplateWithCSS("addedtostory.tmpl", data)
		if err != nil {
			return nil, err
		}
		subject := fmt.Sprintf("%s started the story %s", story.CreatedByU.Username, story.Title)
		if story.CurrentWriterID == recipient.ID {
			subject = subject + " - it's your turn!"
		}
		emails = append(emails, Email{Body: body, Subject: subject, To: recipient})
	}

	return emails, nil
}

func BuildTurnNotificationEmails(story Story, turnNumber int) ([]Email, error) {
	turn := story.Turns[turnNumber-1]
	var recipients []User
	for _, cowriter := range story.Cowriters {
		if cowriter.User.ID != turn.TakenByID && cowriter.User.EmailVerificationStatus == EmailVerificationStateVerified {
			recipients = append(recipients, cowriter.User)
		}
	}

	var emails []Email
	for _, recipient := range recipients {
		data := map[string]interface{}{
			"Story":      story,
			"TurnNumber": turnNumber,
			"Turn":       &turn,
			"Recipient":  recipient,
		}
		body, err := renderTemplateWithCSS("storyturnnotification.tmpl", data)
		if err != nil {
			return nil, err
		}

		subject := ""
		if turn.Finishes() {
			subject = fmt.Sprintf("%s is done!", story.Title)
		} else {
			if story.CurrentWriterID == recipient.ID {
				subject = fmt.Sprintf("It's your turn on %s!", story.Title)
			} else {
				subject = fmt.Sprintf("%s took their turn on %s!", turn.TakenByU.Username, story.Title)
			}
		}

		emails = append(emails, Email{
			Subject: subject,
			Body:    body,
			To:      recipient,
		})
	}

	return emails, nil
}

func BuildEmailVerificationEmail(user User, token string) (*Email, error) {
	data := map[string]interface{}{
		"Recipient":         user,
		"VerificationToken": token,
	}
	body, err := renderTemplateWithCSS("verification.tmpl", data)
	if err != nil {
		return nil, err
	}

	return &Email{
		To:      user,
		Subject: "Verify your email",
		Body:    body,
	}, nil
}

func renderTemplateWithCSS(templateName string, data interface{}) (string, error) {
	template, err := template.ParseFiles("goemailtemplates/_layout.tmpl", path.Join("goemailtemplates", templateName))
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	viewData := viewData{Data: data}
	if err := template.ExecuteTemplate(&buffer, templateName, viewData); err != nil {
		return "", err
	}
	prem, err := premailer.NewPremailerFromBytes(buffer.Bytes(), premailer.NewOptions())
	if err != nil {
		return "", err
	}

	return prem.Transform()
}

func init() {
	styleCSS, err := os.ReadFile("static/style.css")
	if err != nil {
		panic(err)
	}

	css = string(styleCSS)
}
