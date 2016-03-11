package labels

import (
	"google.golang.org/api/gmail/v1"
)

type LabelManager struct {
	Client *gmail.Service
	cache  map[string]string
}

func NewLabelManager(client *gmail.Service) *LabelManager {
	return &LabelManager{client, make(map[string]string)}
}

func (lm *LabelManager) GenerateCache() error {
	ll, err := lm.Client.Users.Labels.List("me").Do()
	if err != nil {
		return err
	}
	for _, l := range ll.Labels {
		lm.cache[l.Name] = l.Id
	}
	return nil
}

func (lm *LabelManager) HasLabel(name string) bool {
	_, ok := lm.cache[name]
	return ok
}

func (lm *LabelManager) Name2Id(name string) string {
	_, ok := lm.cache[name]
	if ok {
		return lm.cache[name]
	}
	return ""
}
