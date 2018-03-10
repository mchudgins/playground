package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"
)

type createCertInput struct {
	CommonName       string `json:"common_name"`
	AlternativeNames string `json:"alt_names"`
}

type createCertResponseData struct {
	Certificate  string `json:"certificate"`
	Issuer       string `json:"issuing_ca"`
	Key          string `json:"private_key"`
	SerialNumber string `json:"serial_number"`
}

type createCertResponse struct {
	RequestID string                 `json:"request_id"`
	LeaseID   string                 `json:"lease_id"`
	Renewable bool                   `json"renewable"`
	Data      createCertResponseData `json:"data"`
}

func (v *Vault) NewCert(ctx context.Context, commonName string, alternativeNames []string) (cert string, key string, err error) {
	v.Logger.Debug("vault.NewCert+",
		zap.String("commonName", commonName),
		zap.Any("alternativeNames", alternativeNames))
	defer v.Logger.Debug("vault.NewCert-")
	log := v.Logger

	var alt string
	for _, name := range alternativeNames {
		if len(alt) > 0 {
			alt = alt + ","
		}
		alt = alt + name
		log.Debug("alternative", zap.String("name", name))
	}

	input := &createCertInput{
		CommonName:       commonName,
		AlternativeNames: alt,
	}

	buf, err := json.Marshal(input)
	if err != nil {
		log.Error("unable to Marshall input", zap.Error(err))
		return
	}

	log.Debug("marshal", zap.ByteString("buf", buf))

	// need to POST the data to the vault api

	c := &http.Client{}
	body := bytes.NewReader(buf)
	r, err := http.NewRequest("POST", v.Address+"/v1/ucap/issue/dst-cloud", body)
	if err != nil {
		log.Error("unable to create http.Request", zap.Error(err))
		return
	}
	r.Header.Set("X-Vault-Token", v.Token)

	resp, err := c.Do(r)
	if err != nil {
		log.Error("POST'ing request", zap.Error(err))
		return
	}

	certBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("reading response", zap.Error(err))
		return
	}

	output := &createCertResponse{}
	err = json.Unmarshal(certBody, output)
	if err != nil {
		log.Error("unable to Unmarshal", zap.Error(err))
		return
	}

	log.Debug("createCertResponse",
		zap.String("certificate", output.Data.Certificate),
		zap.String("issuer", output.Data.Issuer),
		zap.String("key", output.Data.Key))

	cert = output.Data.Certificate
	key = output.Data.Key
	return
}
