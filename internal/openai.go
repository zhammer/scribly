package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

const (
	// https://stackoverflow.com/questions/492090/least-used-delimiter-character-in-normal-text-ascii-128
	asciiGS                = rune(0x1D) //      Group Separator
	asciiFS                = rune(0x1C) //      Field Separator
	turnSeparator          = string(asciiGS)
	turnSeparatorSanitized = string(asciiFS)
)

type HTTPOpenAIGatewayOption func(o *HTTPOpenAIGateway)

type HTTPOpenAIGateway struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

type requestPayload struct {
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float32 `json:"temperature"`
	Stop        string  `json:"stop"`
}

type responsePayload struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func NewHTTPOpenAIGateway(apiKey string, opts ...HTTPOpenAIGatewayOption) *HTTPOpenAIGateway {
	o := HTTPOpenAIGateway{
		apiKey:  apiKey,
		client:  &http.Client{},
		baseURL: "https://api.openai.com",
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &o
}

func WithBaseURL(baseURL string) HTTPOpenAIGatewayOption {
	return func(o *HTTPOpenAIGateway) {
		o.baseURL = baseURL
	}
}

func (o *HTTPOpenAIGateway) PredictText(ctx context.Context, story Story) (string, error) {
	payload := requestPayload{
		Prompt:      buildPrompt(story),
		MaxTokens:   1024,
		Temperature: .9,
		Stop:        turnSeparator,
	}
	body := bytes.Buffer{}
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", o.baseURL+"/v1/engines/davinci/completion", &body)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+o.apiKey)
	req.Header.Add("content-type", "application/json")
	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Received non-200 response from openai %d - '%s'", resp.StatusCode, string(body))
	}

	responseBody := responsePayload{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", err
	}

	return responseBody.Choices[0].Text, nil
}

func buildPrompt(story Story) string {
	prompt := ""
	for _, turn := range story.Turns {
		if turn.Text == "" {
			continue
		}
		sanitized := strings.Replace(turn.Text, turnSeparator, turnSeparatorSanitized, -1)
		prompt += sanitized + turnSeparator
	}

	// there's no existing text to make a prompt, let's choose a random prompt
	if prompt == "" {
		prompt = randomPrompt() + turnSeparator
	}

	// last 1024 chars
	return prompt[int(math.Max(float64(len(prompt)-1024), float64(0))):]
}

// openings to some nice pieces of literature
// these are just used to give openai something to go off of if there's
// no previous text in the scribly story (eg: scribbot takes first turn.)
const (
	// beloved - toni morrison
	beloved = `124 was spiteful. Full of a baby’s venom. The women in the house knew it and so did the children. For years each put up with the spite in his own way, but by 1873 Sethe and her daughter Denver were its only victims. The grandmother, Baby Suggs, was dead, and the sons, Howard and Buglar, had run away by the time they were thirteen years old — as soon as merely looking in a mirror shattered it (that was the signal for Buglar); as soon as two tiny hand prints appeared in the cake (that was it for Howard). Neither boy waited to see more; another kettleful of chickpeas smoking in a heap on the floor; soda crackers crumbled and strewn in a line next to the doorsill. Nor did they wait for one of the relief periods: the weeks, months even, when nothing was disturbed. No. Each one fled at once — the moment the house committed what was for him the one insult not to be born or witnessed a second time. Within two months, in the dead of winter, leaving their grandmother, Baby Suggs; Sethe, their mother; and their little sister, Denver, all by themselves in the gray and white house on Bluestone Road. It didn’t have a number then, because Cincinnati didn’t stretch that far. In fact, Ohio had been calling itself a state only seventy years when first one brother and then the next stuffed quilt packing into his hat, snatched up his shoes, and crept away from the lively spite the house felt for them.`

	// the water dancer - ta-nehisi coates
	waterDancer = `And I could only have seen her there on the stone bridge, a dancer wreathed in ghostly blue, because that was the way they would have taken her back when I was young, back when the Virginia earth was still red as brick and red with life, and though there were other bridges spanning the river Goose, they would have bound her and brought her across this one, because this was the bridge that fed into the turnpike that twisted its way through the green hills and down the valley before bending in one direction, and that direction was south.`

	// 100 years of solitude - gabriel garcía márquez
	soledad = `Many years later, as he faced the firing squad, Colonel Aureliano Buendía was to remember that distant afternoon when his father took him to discover ice.`

	// the idiot - dusky
	theIdiot = `Towards the end of November, during a thaw, at nine o’clock one morning, a train on the Warsaw and Petersburg railway was approaching the latter city at full speed. The morning was so damp and misty that it was only with great difficulty that the day succeeded in breaking; and it was impossible to distinguish anything more than a few yards away from the carriage windows.`

	// pachinko - min jin lee
	pachinko = `History has failed us, but no matter.
At the turn of the century, an aging fisherman and his wife decided to take in lodgers for extra money. Both were born and raised in the fishing village of Yeongdo—a five-mile-wide islet beside the port city of Busan. In their long marriage, the wife gave birth to three sons, but only Hoonie, the eldest and the weakest one, survived. Hoonie was born with a cleft palate and a twisted foot; he was, however, endowed with hefty shoulders, a squat build, and a golden complexion. Even as a young man, he retained the mild, thoughtful temperament he’d had as a child. When Hoonie covered his misshapen mouth with his hands, something he did out of habit meeting strangers, he resembled his nice-looking father, both having the same large, smiling eyes. Inky eyebrows graced his broad forehead, perpetually tanned from outdoor work. Like his parents, Hoonie was not a nimble talker, and some made the mistake of thinking that because he could not speak quickly there was something wrong with his mind, but that was not true.`

	// the balloon hoax - edgar allen poe
	balloon = `THE GREAT problem is at length solved! The air, as well as the earth and the ocean, has been subdued by science, and will become a common and convenient highway for mankind. The Atlantic has been actually crossed in a Balloon! and this too without difficulty- without any great apparent danger- with thorough control of the machine- and in the inconceivably brief period of seventy-five hours from shore to shore!`
)

func randomPrompt() string {
	prompts := []string{beloved, waterDancer, soledad, theIdiot, pachinko, balloon}
	return prompts[rand.Intn(len(prompts))]
}
