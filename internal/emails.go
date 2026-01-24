package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"scribly/embed"
	"strings"

	"github.com/vanng822/go-premailer/premailer"
)

type viewData struct {
	Data       interface{}
	websiteURL string
}

type viewDataOption func(*viewData)

func newViewData(websiteURL string, opts ...viewDataOption) viewData {
	vd := viewData{websiteURL: websiteURL}
	for _, opt := range opts {
		opt(&vd)
	}
	return vd
}

func withData(data interface{}) viewDataOption {
	return func(vd *viewData) {
		vd.Data = data
	}
}

func (v viewData) StoryLink(story Story) template.HTML {
	return template.HTML(fmt.Sprintf(`<a href="%s/stories/%d">%s</a>`, v.websiteURL, story.ID, story.Title))
}

func (v viewData) WebsiteURL() string {
	return v.websiteURL
}

func (v viewData) Replace(original string, pattern string, replacement string) string {
	return strings.ReplaceAll(original, pattern, replacement)
}

func (v viewData) NewLineify(str string) template.HTML {
	return template.HTML(strings.ReplaceAll(str, "\n", "<br>"))
}

func (v viewData) WhoseTurnText(story Story, recipient User) string {
	subject := story.CurrentWriter.Username + "'s"
	if story.CurrentWriterID == recipient.ID {
		subject = "your"
	}
	return fmt.Sprintf("it's %s turn", subject)
}

func (v viewData) CSS() template.HTML {
	return template.HTML("<style>" + embed.CSS + "</style>")
}

func BuildNudgeEmail(websiteURL string, nudger User, nudgee User, story Story) (*Email, error) {
	subject := fmt.Sprintf("%s nudged you to take your turn on %s", nudger.Username, story.Title)
	data := map[string]interface{}{
		"Story":  story,
		"Nudger": nudger,
	}
	body, err := renderTemplateWithCSS("nudge.tmpl", websiteURL, data)
	if err != nil {
		return nil, err
	}

	return &Email{
		Subject: subject,
		Body:    body,
		To:      nudgee,
	}, nil
}

func BuildAddedToStoryEmails(websiteURL string, story Story) ([]Email, error) {
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
		body, err := renderTemplateWithCSS("addedtostory.tmpl", websiteURL, data)
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

func BuildTurnNotificationEmails(websiteURL string, story Story, turnNumber int) ([]Email, error) {
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
		body, err := renderTemplateWithCSS("storyturnnotification.tmpl", websiteURL, data)
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

func BuildEmailVerificationEmail(websiteURL string, user User, token string) (*Email, error) {
	data := map[string]interface{}{
		"Recipient":         user,
		"VerificationToken": token,
	}
	body, err := renderTemplateWithCSS("verification.tmpl", websiteURL, data)
	if err != nil {
		return nil, err
	}

	return &Email{
		To:      user,
		Subject: "Verify your email",
		Body:    body,
	}, nil
}

func renderTemplateWithCSS(templateName string, websiteURL string, data interface{}) (string, error) {
	var buffer bytes.Buffer
	vd := newViewData(websiteURL, withData(data))
	if err := embed.EmailTemplates.ExecuteTemplate(&buffer, templateName, vd); err != nil {
		return "", err
	}
	prem, err := premailer.NewPremailerFromBytes(buffer.Bytes(), premailer.NewOptions())
	if err != nil {
		return "", err
	}

	return prem.Transform()
}
