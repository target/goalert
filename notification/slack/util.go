package slack

func getSlackLink(url, text string) string {
	return "<" + url + "|" + text + ">"
}
