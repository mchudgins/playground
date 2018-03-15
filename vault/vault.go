package vault

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"

	"go.uber.org/zap"
)

const dstRoot string = `-----BEGIN CERTIFICATE-----
MIIFrjCCA5agAwIBAgIQCPz2Tojb7tbnbFBAuJFikzANBgkqhkiG9w0BAQsFADBw
MQswCQYDVQQGEwJVUzEZMBcGA1UECgwQRFNUIFN5c3RlbXMsIEluYzEpMCcGA1UE
CwwgRFNUIEludGVybmFsIFVzZSBPbmx5IC0tIFJPT1QgQ0ExGzAZBgNVBAMMEnJv
b3QtY2EuZHN0Y29ycC5pbzAgFw0xNzA5MDgxODU0MzlaGA8yMDQyMDkwODEyMDAw
MFowcDELMAkGA1UEBhMCVVMxGTAXBgNVBAoMEERTVCBTeXN0ZW1zLCBJbmMxKTAn
BgNVBAsMIERTVCBJbnRlcm5hbCBVc2UgT25seSAtLSBST09UIENBMRswGQYDVQQD
DBJyb290LWNhLmRzdGNvcnAuaW8wggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIK
AoICAQDaYlhv+GH3OG2PIBVFNKnESPQdoWdAfdOZUCHKBzYxHkbZcvSQ7NZ56JWZ
4PFoPc3iSbytzEAz03TpvPS3snrdrfiBxEPWbAFgyYdMeZRTlg9zmRc5tVmPMawm
Q6gQpJqAyrYxkm/+rjiXXDIrKdeMFrKAst+MGNJz8v2EjqN2vOZ9jCcOHj5/WPLC
OJFmAZsO43m77RUsihoR0IP+TlDxMsY5HXl7LC3mNfOQsoXsINLSBV6aJ4MXuyAi
B2XMB/q67q71YysoldYxL8JXxHxEqxBirvW/H3nqITgOUAuTEbhm58VnfcMjwNwj
9H3pIjNnNCaphEDq7prsQmCMbs+fkH1awPmD9nHaJZf9Glxf3Yuql84/0l+KhjZe
sTLg8uTMmzzpp6em2aqeqmqYszdnnfR+k+ah1gbnXbQCQnLw3QoJe4K/FPWeg73S
/U7mY0HCffFYQzuVk5o5wOvqBCwmOSgBnIrTp8JYGKcT9ceBoFPxEc1e53FHRvRG
iHsKs36X6ZDIZL73bF8N3APQgGAsuQ+pHIUkAAsutjIKJamoMGOvb8vf7om21tag
aPw8NDSZf9ZcAXNR3FXcfV4h8xY3HC1g6eRqO3zQYXCohBnXl2mem8PssqV+07d7
GC/Qpjz7Jj3Hr0Eg6r6B/ibLbfuOv92jfAgkE80WRiGmNouERQIDAQABo0IwQDAP
BgNVHRMBAf8EBTADAQH/MA4GA1UdDwEB/wQEAwIBBjAdBgNVHQ4EFgQUsZZNyBKw
H6iDfngI8juQWiSVGQMwDQYJKoZIhvcNAQELBQADggIBAF9X2iHRlNheso2GZSZm
3eXjHQ2WuLewSxniYD/4BKGrno6evLmWr8QmJAWCYiUkKgNVhaAVGN0KnoUhlx3i
Q8FsJjN3Pr3g43j7IfkZ/RlUR2IbGN83uw5pll17zMh11w9Lp9xSkl6uuMQ6oMhd
ugs6p9s0zexkBeHFxjig97A86ZzwRXbLZdSko6FQUSqxvk8bQswLcy1OrTKaf+RE
OfoBPkq8Eg8l6kFBDEj38jpoEsnB6TnY+2LDdeJOyXZQljO+7J0mgTdTqIYuYbZP
Ow8iH/CnqExjJgRaetYl2+yC5aAUHgMMKpkV3+NBYVUdq/L0eaTIoEvxNvZsfvfD
b00LrQcd6sT/CdEU2MHmGCy+XcVw78VR5p/OeYRbV3CgXcgoFFzl+YJ8xCIvdSBC
/xR/7E31xhEln0/6uogw8uluqgViDmHhPmxk4/xT9/2TGhzDmjN/EwY4GN7VfC9W
SYzbGs+zypxoik/mzr1IbhR6RSNL734yzeagim0+BaTYZDiRAAj/jjKNHO2dBH4C
SLgGkj5TMmcV0d7ktjbZ+MP3oFN6BRgvNvyjJf1BCxx2bSQ59FDdGgoY8UEen7ME
LJpSVHU2wunQ3vFTntzvjzWetIlZHd7scrHiPcNNIGnxEDMymO2oSTzpP/pUWgmI
OjvrfL86QP8xM84dC57Mt1I5
-----END CERTIFICATE-----
`

type Vault struct {
	address string
	token   string
	logger  *zap.Logger
	client  *http.Client
}

func New(logger *zap.Logger, address, token string) *Vault {

	return &Vault{
		logger:  logger,
		address: address,
		token:   token,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: getRootCAPool(),
				},
			},
		},
	}
}

func getRootCAPool() *x509.CertPool {
	pool, err := x509.SystemCertPool()
	if err != nil {
		panic(err)
	}

	ok := pool.AppendCertsFromPEM([]byte(dstRoot))
	if !ok {
		panic("unable to parse DST Root CA certificate")
	}

	return pool
}
