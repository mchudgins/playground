// Copyright © 2018 Mike Hudgins <mchudgins@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package vault

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fatih/structs"
	"go.uber.org/zap"
)

type vaultSecret struct {
	Password string `json:"password"`
	Token    string `json:"token"`
}

type vaultResponse struct {
	RequestID     string      `json:"request_id"`
	LeaseID       string      `json:"lease_id"`
	Renewable     bool        `json:"renewable"`
	LeaseDuration int         `json:"lease_duraton"`
	Data          vaultSecret `json:"data"`
}

func (v *Vault) GetSecret(secretPath string, secretValue string) (string, error) {
	log := v.Logger
	log.Debug("vault.GetSecret+",
		zap.Any("secret", secretPath))
	defer log.Debug("vault.GetSecret-")

	c := &http.Client{}
	r, err := http.NewRequest("GET", v.Address+"/v1/"+secretPath, http.NoBody)
	if err != nil {
		log.Error("unable to create http.Request", zap.Error(err))
		return "", err
	}
	r.Header.Set("X-Vault-Token", v.Token)

	resp, err := c.Do(r)
	if err != nil {
		log.Error("GETT'ing request",
			zap.String("secretPath", secretPath),
			zap.Error(err))
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unable to retrieve secrets from %s -- expected 200 response, got %d", secretPath, resp.StatusCode)
		log.Error("while calling Vault",
			zap.Error(err),
			zap.String("secretPath", secretPath),
			zap.Int("StatusCode", resp.StatusCode))
		return "", err
	}

	secrets, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("reading response", zap.Error(err))
		return "", err
	}

	output := &vaultResponse{}
	err = json.Unmarshal(secrets, output)
	if err != nil {
		log.Error("unable to Unmarshal", zap.Error(err))
		return "", err
	}

	log.Debug("vaultResponse Unmarshaled",
		zap.Any("output", output))

	m := structs.Map(output.Data)

	s, ok := m[secretValue]
	if !ok {
		return "", fmt.Errorf("Secret '%s' not found", secretValue)
	}

	secret, ok := s.(string)
	if !ok {
		return "", fmt.Errorf("Secret '%s' failed string type assertion (%+v)", secretValue, s)
	}

	log.Debug("done", zap.String("secret", secret))

	return secret, nil
}
