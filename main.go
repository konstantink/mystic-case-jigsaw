// Sample run-helloworld is a minimal Cloud Run service.
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"github.com/jordan-wright/email"
)

const emailTemplate = `
<p>Hello there,</p>

<p>Somebody has left a feedback for the jigsaw. Here are the answers:
<ul>
        <li>Quality: %d</li>
        <li>Artwork: %d</li>
        <li>Quest: %d</li>
        <li>Overall: %d</li>
        <li>Buy next box: %s</li>
        <li>The reason to buy: %q</li>
        <li>Anything to add: %q</li>
</ul></p>

<p>That's all for now.</p>

<p>Have a nice day!</p>

<p>Kind regards,<br/>
Your jigsaw.mystic-case.co.uk</p>
`

var emailChan chan []byte

func main() {
	log.Print("starting server...")
	http.HandleFunc("^/$", handler)
	http.HandleFunc("/favicon.ico", fileHandler("./images/favicon.ico"))
	http.HandleFunc("/sitemap.xml", fileHandler("./sitemap.xml"))
	http.HandleFunc("/robots.txt", fileHandler("./robots.txt"))
	http.HandleFunc("/welcome", townFestival)
	http.HandleFunc("/hints/town-festival", hintsTownFestival)
	http.HandleFunc("/feedback", feedback)
	http.Handle("/static/", http.StripPrefix("/static/", folderHandler("./static")))
	http.Handle("/images/", http.StripPrefix("/images/", folderHandler("./images")))

	emailChan = make(chan []byte, 10)

	go listenForEmail()

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func debug(format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("[DEBUG] %s", format), v...)
}

func check(err error, w http.ResponseWriter) bool {
	if err != nil {
		http.Error(w, err.Error(), 500)
		return false
	}
	return true
}

func fileHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("NAME")
	if name == "" {
		name = "World!"
	}
	fmt.Fprintf(w, "Hello %s!\n", name)
}

func townFestival(w http.ResponseWriter, r *http.Request) {
	debug("Town Festival")
	var context = struct {
		IsMobile  bool
		GtmID     string
		BoxImages []string
	}{
		IsMobile: false,
		GtmID:    os.Getenv("MYSTIC_CASE_GTM_ID"),
		BoxImages: []string{"./images/haunted-castle.webp", "./images/ufo-crash.webp", "./images/time-machine.webp",
			"./images/school-of-magic.webp", "./images/national-treasure.webp", "./images/unfinished-case-of-holmes.webp",
			"./images/simulation-theory.webp", "./images/lost-island.webp", "./images/dragula.webp", "./images/planet.webp"},
	}
	var files = []string{
		"./templates/base.html",
		"./templates/views/home.html",
		"./templates/views/quality_form.html",
		"./templates/views/artwork_form.html",
		"./templates/views/quest_form.html",
		"./templates/views/buy_form.html",
		"./templates/views/overall_form.html",
		"./templates/views/footer.html",
	}

	tpl, err := template.ParseFiles(files...)
	if !check(err, w) {
		return
	}

	tpl.ExecuteTemplate(w, "base", context)
}

func hintsTownFestival(w http.ResponseWriter, r *http.Request) {
	debug("Hints Town Festival")
	var context = struct {
		IsMobile bool
		GtmID    string
	}{
		IsMobile: false,
		GtmID:    os.Getenv("MYCTIC_CASE_GTM_ID"),
	}

	var files = []string{
		"./templates/base.html",
		"./templates/views/hints.html",
		"./templates/views/footer.html",
	}

	tpl, err := template.ParseFiles(files...)
	if !check(err, w) {
		return
	}

	tpl.ExecuteTemplate(w, "base", context)
}

func feedback(w http.ResponseWriter, r *http.Request) {
	check := func(err error) bool {
		if err != nil {
			var resp = struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			}{
				Success: false,
				Message: err.Error(),
			}
			respBytes, _ := json.Marshal(resp)
			fmt.Fprintf(w, string(respBytes))
			return false
		}

		return true
	}

	if r.Method == http.MethodPost {
		defer r.Body.Close()
		var payload struct {
			Quest       int8
			Quality     int8
			Artwork     int8
			Overall     int8
			BuyNext     string
			ReasonToBuy string
			Optional    string
		}
		body, err := io.ReadAll(r.Body)
		if !check(err) {
			return
		}
		err = json.Unmarshal(body, &payload)
		if !check(err) {
			return
		}

		message := []byte(fmt.Sprintf(emailTemplate, payload.Quality, payload.Artwork, payload.Quest,
			payload.Overall, payload.BuyNext, payload.ReasonToBuy, payload.Optional))

		emailChan <- message
		//err = sendEmail(message)
		if !check(err) {
			log.Printf("Failed to send email: %s", err.Error())
			return
		}
		log.Print("Email sent successfully")
		resp := struct {
			Success bool `json:"success"`
		}{
			Success: true,
		}
		respJson, _ := json.Marshal(resp)
		fmt.Fprint(w, string(respJson))
	} else {
		fmt.Fprintf(w, "Method %q is not supported", r.Method)
	}
}

func listenForEmail() {
	host := os.Getenv("MYSTIC_CASE_SMTP_HOST")
	port := os.Getenv("MYSTIC_CASE_SMTP_PORT")
	username := os.Getenv("MYSTIC_CASE_USERNAME")
	password := os.Getenv("MYSTIC_CASE_PASSWORD")
	to := os.Getenv("MYSTIC_CASE_TO")
	from := os.Getenv("MYSTIC_CASE_FROM")

	for message := range emailChan {
		e := email.Email{
			From:    from,
			To:      []string{to},
			Subject: "User feedback on the jigsaw",
			HTML:    message,
		}

		log.Print("Authorising on SMTP server...")
		auth := smtp.PlainAuth("", username, password, host)

		log.Print("Sending email...")
		err := e.Send(fmt.Sprintf("%s:%s", host, port), auth)
		if err != nil {
			log.Printf("Something went wrong: %s", err.Error())
		}
		// return smtp.SendMail(fmt.Sprintf("%s:%s", host, port), auth, from, []string{to}, message)
	}
}

func folderHandler(path string) http.Handler {
	return http.FileServer(http.Dir(path))
}
