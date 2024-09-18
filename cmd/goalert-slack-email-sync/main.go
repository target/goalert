/*
goalert-slack-email-sync will create/update AuthSubject entries for users by matching the user's GoAlert email to the corresponding Slack user.
*/
package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/target/goalert/pkg/sysapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	api := flag.String("api", "localhost:1234", "Target address of GoAlert SysAPI server.")
	cert := flag.String("cert-file", "", "Path to PEM-encoded certificate for gRPC auth.")
	key := flag.String("key-file", "", "Path to PEM-encoded key for gRPC auth.")
	ca := flag.String("ca-file", "", "Path to PEM-encoded CA certificate for gRPC auth.")
	token := flag.String("token", "", "Slack API token for looking up users.")
	domain := flag.String("domain", "", "Limit requests to users with an email at the provided domain.")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	creds := insecure.NewCredentials()
	if *cert+*key+*ca != "" {
		cfg, err := sysapi.NewTLS(*ca, *cert, *key)
		if err != nil {
			log.Fatal("tls credentials:", err)
		}
		creds = credentials.NewTLS(cfg)
	}

	conn, err := grpc.NewClient(*api, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal("connect to GoAlert:", err)
	}
	defer conn.Close()

	goalertClient := sysapi.NewSysAPIClient(conn)
	slackClient := slack.New(*token)

	getRetry := func(email string) (*slack.User, error) {
		for {
			slackUser, err := slackClient.GetUserByEmail(email)
			var rateLimitErr *slack.RateLimitedError
			if errors.As(err, &rateLimitErr) {
				log.Printf("ERROR: rate-limited, waiting %s", rateLimitErr.RetryAfter.String())
				time.Sleep(rateLimitErr.RetryAfter)
				continue
			}

			return slackUser, err
		}
	}

	ctx := context.Background()

	info, err := slackClient.GetTeamInfoContext(ctx)
	if err != nil {
		log.Fatalln("get team info:", err)
	}

	providerID := "slack:" + info.ID
	users, err := goalertClient.UsersWithoutAuthProvider(ctx, &sysapi.UsersWithoutAuthProviderRequest{ProviderId: providerID})
	if err != nil {
		log.Fatalln("fetch users missing provider:", err)
	}

	var count int
	for {
		u, err := users.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("fetch missing user:", err)
		}
		if !strings.HasSuffix(u.Email, *domain) {
			continue
		}
		slackUser, err := getRetry(u.Email)
		if err != nil {
			if !strings.Contains(err.Error(), "users_not_found") {
				log.Fatalf("lookup Slack user '%s': %v", u.Email, err)
			}
			log.Printf("lookup Slack user '%s': %v", u.Email, err)
			continue
		}

		_, err = goalertClient.SetAuthSubject(ctx, &sysapi.SetAuthSubjectRequest{Subject: &sysapi.AuthSubject{
			ProviderId: providerID,
			UserId:     u.Id,
			SubjectId:  slackUser.ID,
		}})
		if err != nil {
			log.Fatalf("set provider '%s' auth subject for user '%s' to '%s': %v", providerID, u.Id, slackUser.ID, err)
		}
		count++
	}

	log.Printf("Updated %d users.", count)
}
