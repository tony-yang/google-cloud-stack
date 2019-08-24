package main

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	plus "google.golang.org/api/plus/v1"

	"golang.org/x/oauth2"

	uuid "github.com/gofrs/uuid"
	"github.com/tony-yang/google-cloud-stack/bookshelf"
)

const (
	defaultSessionID        = "default"
	googleProfileSessionKey = "google_profile"
	oauthTokenSessionKey    = "oauth_token"
	oauthFlowRedirectKey    = "redirect"
)

func init() {
	// Gob encoding for gorilla/sessions
	gob.Register(&oauth2.Token{})
	gob.Register(&Profile{})
}

type Profile struct {
	ID, DisplayName, ImageURL string
}

// validateRedirectURL checks that the URL provided is valid.
// If the URL is missing, redirect the user to the application's root
// The URL must not be absolute (i.e., the URL must refer to a path within this application).
func validateRedirectURL(path string) (string, error) {
	if path == "" {
		return "/", nil
	}
	// Ensure redirect URL is valid and not pointing to a different server
	parsedURL, err := url.Parse(path)
	if err != nil {
		return "/", err
	}
	if parsedURL.IsAbs() {
		return "/", errors.New("URL must not be absolute")
	}
	return path, nil
}

// loginHandler initiates an OAuth flow to authenticate the user
func loginHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.Must(uuid.NewV4()).String()
	oauthFlowSession, err := bookshelf.SessionStore.New(r, sessionID)
	if err != nil {
		fmt.Printf("auth.go 54: oauth Flow Session error %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}
	oauthFlowSession.Options.MaxAge = 10 * 60 // 10 minutes

	redirectURL, err := validateRedirectURL(r.FormValue("redirect"))
	if err != nil {
		fmt.Printf("auth.go 61: redirectURL error %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	oauthFlowSession.Values[oauthFlowRedirectKey] = redirectURL
	if err := oauthFlowSession.Save(r, w); err != nil {
		fmt.Printf("auth.go 67: oauthFlowSession error %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	// Use the session ID for the "state" param
	// This protects against CSRF
	url := bookshelf.OAuthConfig.AuthCodeURL(sessionID, oauth2.ApprovalForce, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}

// logoutHandler clears the default session
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := bookshelf.SessionStore.New(r, defaultSessionID)
	if err != nil {
		fmt.Printf("auth.go 81: could not get default session %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	session.Options.MaxAge = -1 // Clear session
	if err := session.Save(r, w); err != nil {
		fmt.Printf("auth.go 87: could not save session %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	redirectURL := r.FormValue("redirect")
	if redirectURL == "" {
		redirectURL = "/"
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func fetchProfile(ctx context.Context, tok *oauth2.Token) (*plus.Person, error) {
	client := oauth2.NewClient(ctx, bookshelf.OAuthConfig.TokenSource(ctx, tok))
	plusService, err := plus.New(client)
	if err != nil {
		return nil, err
	}
	return plusService.People.Get("me").Do()
}

func stripProfile(p *plus.Person) *Profile {
	return &Profile{
		ID:          p.Id,
		DisplayName: p.DisplayName,
		ImageURL:    p.Image.Url,
	}
}

// oauthCallbackHandler completes the OAuth flow, retrieves the user's profile
// information and stores it in a session.
func oauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	oauthFlowSession, err := bookshelf.SessionStore.Get(r, r.FormValue("state"))
	if err != nil {
		fmt.Printf("auth.go 121: oauthFlowSession Get error %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	redirectURL, ok := oauthFlowSession.Values[oauthFlowRedirectKey].(string)
	// validate this callback came from the app
	if !ok {
		fmt.Printf("auth.go 128: redirectURL validation failed, try logging again: %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	code := r.FormValue("code")
	tok, err := bookshelf.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("auth.go 135: could not get auth token: %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	session, err := bookshelf.SessionStore.New(r, defaultSessionID)
	if err != nil {
		fmt.Printf("auth.go 141: could not get default session %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	ctx := context.Background()
	profile, err := fetchProfile(ctx, tok)
	if err != nil {
		fmt.Printf("auth.go 148: could not fetch Google profile %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	session.Values[oauthTokenSessionKey] = tok
	// Strip the profile to only the fields we need. Otherwise the struct is too big
	session.Values[googleProfileSessionKey] = stripProfile(profile)
	if err := session.Save(r, w); err != nil {
		fmt.Printf("auth.go 156: could not save session %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/books"), http.StatusFound)
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// profileFromSesson retrieves the Google+ profile from the default session.
// Returnsnil if the profile cannot be retrieved
func profileFromSession(r *http.Request) *Profile {
	session, err := bookshelf.SessionStore.Get(r, defaultSessionID)
	if err != nil {
		return nil
	}
	tok, ok := session.Values[oauthTokenSessionKey].(*oauth2.Token)
	if !ok || !tok.Valid() {
		return nil
	}
	profile, ok := session.Values[googleProfileSessionKey].(*Profile)
	if !ok {
		return nil
	}
	return profile
}
