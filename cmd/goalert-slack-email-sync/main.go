/*
	goalert-slack-email-sync will create/update AuthSubject entries for users by matching the user's GoAlert email to the corresponding Slack user.
*/
package main

import "flag"

func main() {
	flag.String("goalert-api", "", "Target address of GoAlert SysAPI server.")
	flag.String("cert-file", "", "Path to PEM-encoded certificate for gRPC auth.")
	flag.String("key-file", "", "Path to PEM-encoded key for gRPC auth.")
	flag.String("ca-file", "", "Path to PEM-encoded CA certificate for gRPC auth.")
	flag.String("slack-token", "", "Slack API token for looking up users.")
	flag.Parse()

	// get list of missing subjects
	// hit user.lookupByEmail Slack API with token for each user's email
	// call SetAuthSubject for each found
}
