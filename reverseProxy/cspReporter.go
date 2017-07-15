package reverseProxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	gsh "github.com/mchudgins/go-service-helper/handlers"
)

type cspReporter struct {
	transaction *transaction
}

type report struct {
	DocumentURI    string `json:"document-uri"`
	Referrer       string `json:"referrer"`
	Violation      string `json:"violated-directive"`
	Directive      string `json:"effective-directive"`
	Disposition    string `json:"disposition"`
	OriginalPolicy string `json:"original-policy"`
	BlockedURI     string `json:"blocked-uri"`
	StatusCode     int    `json:"status-code"`
	ScriptSample   string `json:"script-sample"`
}

type dbRecord struct {
	RemoteIP string
	Received time.Time
}

type msg struct {
	Report report `json:"csp-report"`
}

func NewCSPReporter() *cspReporter {
	t, err := New()
	if err != nil {
		panic(err)
	}
	return &cspReporter{transaction: t}
}

func (c *cspReporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer w.WriteHeader(200)

	if r.Method != "POST" {
		return
	}

	logger, _ := gsh.FromContext(r.Context())

	rc := r.Body
	defer rc.Close()

	body, err := ioutil.ReadAll(rc)
	if err != nil {
		logger.
			WithError(err).
			Error("Unable to read body")
	}
	logger.WithField("body", string(body)).Info("")

	var report msg
	err = json.Unmarshal(body, &report)
	if err != nil {
		logger.
			WithError(err).
			WithField("json", string(body)).
			Error("json.Unmarshal failed")
	}

	var out bytes.Buffer
	json.Indent(&out, body, ">", "\t")
	logger.
		WithField("uri", report.Report.DocumentURI).
		WithField("violation", report.Report.Violation).
		Info(out.String())

	_, err = c.transaction.Create()
	if err != nil {
		logger.WithError(err).Fatal("unable to insert CSP report into datastore")
	}
}
