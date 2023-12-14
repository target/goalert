package graphql2

func (d DestinationInput) FieldValue(id string) string {
	for _, f := range d.Values {
		if f.FieldID == id {
			return f.Value
		}
	}
	return ""
}
