package mocktwilio

import (
	"fmt"
	"strings"

	"github.com/ttacon/libphonenumber"
)

type numberDB struct {
	svc    map[string]*MsgService
	num    map[string]*Number
	svcNum map[string]*MsgService
}

func newNumberDB() *numberDB {
	return &numberDB{
		svc:    make(map[string]*MsgService),
		num:    make(map[string]*Number),
		svcNum: make(map[string]*MsgService),
	}
}

func (db *numberDB) clearMsgSvc(id string) {
	s := db.svc[id]
	if s == nil {
		return
	}

	for _, n := range s.Numbers {
		delete(db.svcNum, n)
	}
	delete(db.svc, id)
}

func (db *numberDB) AddUpdateMsgService(ms MsgService) error {
	if !strings.HasPrefix(ms.ID, "MG") {
		return fmt.Errorf("invalid MsgService SID %s", ms.ID)
	}

	if ms.SMSWebhookURL != "" {
		err := validateURL(ms.SMSWebhookURL)
		if err != nil {
			return err
		}
	}
	for _, nStr := range ms.Numbers {
		_, err := libphonenumber.Parse(nStr, "")
		if err != nil {
			return fmt.Errorf("invalid phone number %s: %v", nStr, err)
		}
	}

	db.clearMsgSvc(ms.ID)
	db.svc[ms.ID] = &ms

	for _, n := range ms.Numbers {
		db.svcNum[n] = &ms
	}

	return nil
}

func (db *numberDB) AddUpdateNumber(n Number) error {
	_, err := libphonenumber.Parse(n.Number, "")
	if err != nil {
		return fmt.Errorf("invalid phone number %s: %v", n.Number, err)
	}
	if n.SMSWebhookURL != "" {
		err = validateURL(n.SMSWebhookURL)
		if err != nil {
			return err
		}
	}
	if n.VoiceWebhookURL != "" {
		err = validateURL(n.VoiceWebhookURL)
		if err != nil {
			return err
		}
	}

	db.num[n.Number] = &n

	return nil
}

func (db *numberDB) MsgSvcExists(id string) bool { _, ok := db.svc[id]; return ok }
func (db *numberDB) NumberExists(s string) bool {
	if _, ok := db.svcNum[s]; ok {
		return true
	}

	if _, ok := db.num[s]; ok {
		return true
	}

	return false
}

func (db *numberDB) SMSWebhookURL(number string) string {
	if s, ok := db.svcNum[number]; ok && s.SMSWebhookURL != "" {
		return s.SMSWebhookURL
	}
	if n, ok := db.num[number]; ok && n.SMSWebhookURL != "" {
		return n.SMSWebhookURL
	}
	return ""
}

func (db *numberDB) VoiceWebhookURL(number string) string {
	if n, ok := db.num[number]; ok && n.VoiceWebhookURL != "" {
		return n.VoiceWebhookURL
	}
	return ""
}

func (db *numberDB) MsgSvcNumbers(id string) []string {
	if s, ok := db.svc[id]; ok {
		return s.Numbers
	}
	return nil
}
