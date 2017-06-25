package authn

import (
	"html/template"
	"net/http"

	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	gsh "github.com/mchudgins/go-service-helper/handlers"
)

const (
	html = `
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
  <title>Login</title>
  <h1>Login</h1>

  <form autocomplete="on" method="POST">
  	<fieldset>
  	<legend>User Credentials</legend>
  	<label for="user-id">Username:</label>
  	<input id="user-id" type="text" name="user-id" autocomplete="email" autofocus="true" placeholder="anyone@example.com" required="true">
  	<input type="submit" value="Login">

		<fieldset>
		<legend>Options</legend>
		<label for="broken-Token">Broken Token</label>
		<input id="broken-Token" name="broken-Token" type="checkbox">
		</fieldset>
  	</fieldset>
  </form>
</body>
</html>`
)

var (
	indexTemplate *template.Template
)

func init() {
	indexTemplate = template.Must(template.New("login").Parse(html))
}

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		Hostname string
		URL      string
		Handler  string
	}

	logger, _ := gsh.FromContext(r.Context())
	logger.Info("loginGetHandler")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "max-age=86400") // one day
	err := indexTemplate.Execute(w, data{Hostname: r.Host, URL: r.URL.Path, Handler: "login"})
	if err != nil {
		logger.WithError(err).
			WithField("template", indexTemplate.Name()).
			WithField("path", r.URL.Path).
			Error("Unable to execute template")
	}
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {

	logger, _ := gsh.FromContext(r.Context())
	logger.Info("loginPostHandler")

	err := r.ParseForm()
	if err != nil {
		logger.WithError(err).WithField("url", r.URL.Path).Warn("error while parsing login form")
	}

	fields := log.Fields{}
	for key, value := range r.PostForm {
		fields[key] = value[0]
	}
	logger.WithFields(fields).Info("posted data")

	cookieToken := &http.Cookie{
		Name:     "Authentication",
		Value:    "token " + r.PostFormValue("user-id"),
		Domain:   strings.Split(r.Host, ":")[0],
		Expires:  time.Now().Add(time.Duration(1) * time.Hour),
		MaxAge:   3600 * 24,
		HttpOnly: true,
		//		Secure:   true,
	}

	http.SetCookie(w, cookieToken)
	w.Header().Set("Location", "http://localhost:8080/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
