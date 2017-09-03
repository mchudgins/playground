// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"os"

	"io/ioutil"

	"encoding/pem"

	"crypto/x509"

	"github.com/spf13/cobra"
)

type CertificateDescription struct {
	Subject          string
	AlternativeNames []string
	Issuer           string
	CA               bool
}

func (cd *CertificateDescription) ToString() string {
	ca := "server"
	if cd.CA {
		ca = "CA"
	}
	return fmt.Sprintf("%s, %s, %s, %s", cd.Subject, cd.AlternativeNames, cd.Issuer, ca)
}

// describe-certCmd represents the describe-cert command
var describeCertCmd = &cobra.Command{
	Use:   "describe-cert [certificate filenames]",
	Short: "display important details of an x509 certificate",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var fileList []*os.File

		filecount := len(args)
		if filecount == 0 {
			fileList = make([]*os.File, 1, 1)
			fileList[0] = os.Stdin
		} else {
			fileList = make([]*os.File, filecount, filecount)
			for i, filename := range args {
				var err error
				fileList[i], err = os.OpenFile(filename, os.O_RDONLY, 0)
				defer fileList[i].Close()
				if err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "%s while attempting to open %s.\n", err, filename)
				}
			}
		}

		for i, file := range fileList {
			cd, err := describe(file)
			if err != nil {
				fmt.Fprint(cmd.OutOrStderr(), "while accessing %s -- %s", args[i], err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s, %s\n", args[i], cd.ToString())
		}

	},
}

func describe(file io.Reader) (*CertificateDescription, error) {
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	p, _ := pem.Decode(buf)
	if p == nil {
		panic("failed to decode certificate")
	}
	cert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		panic(err)
	}

	return &CertificateDescription{
		Subject:          cert.Subject.CommonName,
		AlternativeNames: cert.DNSNames,
		Issuer:           cert.Issuer.CommonName,
		CA:               cert.IsCA,
	}, nil
}

func init() {
	RootCmd.AddCommand(describeCertCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// describe-certCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// describe-certCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
