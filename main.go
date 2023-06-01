package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/oauth2" //package die vereist is om de Oauth2.0 te gebruiken
)

const (
	clientID        = "9fff275d-98e9-435c-ac69-426203a13c75"                                                         //Client-ID krijg ik wanneer ik de App-regrstratie maak
	clientSecret    = "eO98Q~Ks0cASPwqgFQiqp-p1gs9fqOGbNaek-apv"                                                     //Deze krijg ik wanneer ik een secret-key aanmaak in de App-regristratie
	redirectURL     = "http://localhost:8080/callback"                                                               //Deze moet ik aanmaken wanneer ik de App-regrstratie aanmaakte. Deze is default
	authorizeURL    = "https://login.microsoftonline.com/d58b4d56-7f52-40be-94cd-e6538315af7f/oauth2/v2.0/authorize" //Hierdoor krijg ik de Authorisatie-code
	tokenURL        = "https://login.microsoftonline.com/d58b4d56-7f52-40be-94cd-e6538315af7f/oauth2/v2.0/token"     //Hierdoor krijg ik de Authorisatie-token, hiervoor heb ik wel de code nodig
	siteCreationURL = "https://twsr4-admin.sharepoint.com/_api/SPSiteManager/create"                                 //Via deze link maakt hij de SharePoint site aan, deze is gekoppeld aan mijn eigen tennant
)

// in dit stukje zal ik met de informatie hierboven de oauth2 instellen
func main() {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authorizeURL,
			TokenURL: tokenURL,
		},
		RedirectURL: redirectURL,
		Scopes:      []string{"AllSites.Manage"}, //hier geef ik alle rechten die overeen moeten komen met diegene die ik geef in de app-regristratie
	}

	// Hier start ik de autorisatie
	authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Open deze URL in je browser om in te loggen: \n%s\n", authURL)

	// Start de webserver om de callback te ontvangen
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		_, err := config.Exchange(context.Background(), code)
		if err != nil {
			log.Fatal("Fout bij het uitwisselen van autorisatiecode: ", err)
		}

		// Creëer een nieuwe SharePoint-site
		err = createSite(code)
		if err != nil {
			log.Fatal("Fout bij het maken van de SharePoint-site: ", err)
		}

		fmt.Println("SharePoint-site succesvol aangemaakt!")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Functie om een nieuwe SharePoint-site aan te maken
func createSite(authorizationCode string) error {
	client := &http.Client{}

	siteURL := "https://twsr4.sharepoint.com/sites/my-new-site"

	// Aanvraag maken voor het maken van een nieuwe SharePoint-site
	req, err := http.NewRequest("POST", siteCreationURL, nil)
	if err != nil {
		return err //voorbeeld errorhandling
	}

	req.Header.Add("Authorization", "Bearer "+authorizationCode)
	req.Header.Add("Accept", "application/json;odata=verbose")
	req.Header.Add("Content-Type", "application/json;odata=verbose")

	// Sitegegevens configureren
	body := strings.NewReader(`{"url": "` + siteURL + `", "title": "Mijn SharePoint-site", "description": "Dit is mijn nieuwe SharePoint-site"}`)
	req.Body = ioutil.NopCloser(body)
	req.ContentLength = int64(body.Len())

	// Aanvraag versturen
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Fout bij het maken van de SharePoint-site: %s", resp.Status)
	}

	return nil
}
