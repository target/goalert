package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/target/goalert/devtools/mockslack"
)

func main() {
	addr := flag.String("addr", "localhost:8085", "Address to listen on.")
	prefix := flag.String("prefix", "", "API URL prefix.")
	appName := flag.String("app-name", "GoAlert", "Name of the initial app.")
	clientID := flag.String("client-id", "", "Default client ID.")
	clientSecret := flag.String("client-secret", "", "Default client secret.")
	accessToken := flag.String("access-token", "", "Default access token.")
	channels := flag.String("channels", "general,test,foobar", "Comma-delimited list of initial channels.")
	autoChannel := flag.Bool("auto-channel", false, "Automatically create missing channels on chat.postMessage calls.")
	scopes := flag.String("scopes", "bot", "Comma-delimited list of scopes to add for the initial app.")
	singleUser := flag.String("single-user", "", "If set, all requests will be implicitly authenticated.")
	userGroups := flag.String("user-groups", "test,foobar", "Comma-delimited list of initial user groups.")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	srv := mockslack.NewServer()

	app, err := srv.InstallStaticApp(mockslack.AppInfo{
		Name:         *appName,
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		AccessToken:  *accessToken,
	}, strings.Split(*scopes, ",")...)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("AppName      = %s", app.Name)
	log.Printf("ClientID     = %s", app.ClientID)
	log.Printf("ClientSecret = %s", app.ClientSecret)
	log.Printf("AccessToken  = %s", app.AccessToken)

	if *channels != "" {
		for _, ch := range strings.Split(*channels, ",") {
			srv.NewChannel(ch)
		}
	}

	if *userGroups != "" {
		for _, ug := range strings.Split(*userGroups, ",") {
			srv.NewUserGroup(ug)
		}
	}

	srv.SetAutoCreateChannel(*autoChannel)

	h := http.Handler(srv)
	if *prefix != "" {
		h = http.StripPrefix(*prefix, h)
	}

	if *singleUser != "" {
		usr := srv.NewUser(*singleUser)
		log.Printf("S. UserID    = %s", usr.ID)
		log.Printf("S. UserName  = %s", usr.Name)
		log.Printf("S. UserAuth  = %s", usr.AuthToken)

		next := h
		h = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.AddCookie(&http.Cookie{
				Name:  "slack_token",
				Value: usr.AuthToken,
			})
			next.ServeHTTP(w, req)
		})
	}

	log.Println("Listening:", *addr)
	err = http.ListenAndServe(*addr, h)
	if err != nil {
		log.Fatal(err)
	}
}
