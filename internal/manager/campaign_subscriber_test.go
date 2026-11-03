package manager

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/knadh/listmonk/models"
)

// Templates must be able to branch on list names with sprig's `has`.
func TestCampaignSubscriberLists(t *testing.T) {
	m := &Manager{}
	tpl, err := template.New("t").Funcs(m.makeGnericFuncMap()).Parse(`{{ if has "Pledge" .Subscriber.Lists }}pledged{{ else }}no{{ end }} {{ .Subscriber.Email }} {{ .Subscriber.FirstName }}`)
	if err != nil {
		t.Fatal(err)
	}

	for lists, want := range map[string]string{"Pledge": "pledged", "News": "no"} {
		var b bytes.Buffer
		msg := CampaignMessage{Subscriber: CampaignSubscriber{
			Subscriber: models.Subscriber{Email: "a@b.com", Name: "Alice Smith"},
			Lists:      []string{lists},
		}}
		if err := tpl.Execute(&b, msg); err != nil {
			t.Fatal(err)
		}
		if got := b.String(); got != want+" a@b.com Alice" {
			t.Fatalf("lists=%s: got %q", lists, got)
		}
	}
}
