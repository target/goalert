package alert

const TypeAlertMetaV1 = "alert_meta_v1"

type AlertMeta struct {
	Type        string        `json:"type"`
	AlertMetaV1 AlertMetaData `json"alert_meta_v1"`
}

type AlertMetaData map[string]string

type AlertMetaInput []struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func ToAlertMeta(meta AlertMetaInput) AlertMeta {
	alertMeta := AlertMetaData{}
	for _, v := range meta {
		alertMeta[v.Key] = v.Value
	}
	return AlertMeta{
		Type:        TypeAlertMetaV1,
		AlertMetaV1: alertMeta,
	}
}
