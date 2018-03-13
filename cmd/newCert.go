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

package cmd

import (
	"context"
	"fmt"

	"os"

	repo "github.com/mchudgins/playground/gitWrapper"
	"github.com/mchudgins/playground/log"
	"github.com/mchudgins/playground/vault"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	repository     string = "https://gitlab.com/dstcorp/testRepo.git"
	gitlabUsername string = "dst_certificate_management"
	vaultGitlabURL string = "secret/aws-lambda/certificateManagementBot/gitlab"
)

// newCertCmd represents the newCert command
var newCertCmd = &cobra.Command{
	Use:   "newCert <domain name> <alternative names>",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		logger := log.GetLogger("vault")
		ctx := context.Background()
		logger.Debug("newCertCmd+")
		defer logger.Debug("newCertCmd-")

		if len(args) < 1 {
			cmd.Usage()
			return
		}

		v := vault.New(logger, vaultAddress, vaultToken)
		cert, key, err := v.NewCert(ctx, args[0], args[1:])
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error creating certificate: %s\n\n", err)
			os.Exit(1)
			cert = `-----BEGIN CERTIFICATE-----
MIIEKzCCAhOgAwIBAgIRANyRKap3ZqPd8TPVFaFCv4wwDQYJKoZIhvcNAQELBQAw
gYwxCzAJBgNVBAYTAlVTMRkwFwYDVQQKDBBEU1QgU3lzdGVtcywgSW5jMUUwQwYD
VQQLDDxEU1QgSW50ZXJuYWwgVXNlIE9ubHkgLS0gVW5jb25zdHJhaW5lZCBDbG91
ZCBDZXJ0IFNpZ25pbmcgQ0ExGzAZBgNVBAMMEnVjYXAtY2EuZHN0Y29ycC5pbzAe
Fw0xODAzMDYxNzI4MTdaFw0xODA2MDQxNzI4MTdaMDsxGTAXBgNVBAoTEERTVCBT
eXN0ZW1zLCBJbmMxHjAcBgNVBAMTFXRlc3QubG9jYWwuZHN0Y29ycC5pbzBZMBMG
ByqGSM49AgEGCCqGSM49AwEHA0IABIjD58J0cSDMjmIusAn3hO8X2MNgyf48LDt4
3mNs0my11MWHU2wgoz2h3EWgVmUmsdyU5oZp0oxlCWiJnQh65RGjgaIwgZ8wDgYD
VR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNV
HRMBAf8EAjAAMB8GA1UdIwQYMBaAFCZPDmy9K6uj32tzXMnBQ5RpK3taMD8GA1Ud
EQQ4MDaCFXRlc3QubG9jYWwuZHN0Y29ycC5pb4IJbG9jYWxob3N0ghIqLmxvY2Fs
LmRzdGNvcnAuaW8wDQYJKoZIhvcNAQELBQADggIBAGGt31PtFaW1gV2VoH6ANH2C
JskECCY3Mnj1OK1FaYpFp5t2G5kr0gkmNyyd2L7hKT8ugKQTtPwpK1614TSmjCf7
d7X5V+6vZXymWJKYdKk/0c91bqnfDG/Cb9BKG89rxLc3hv+sdHHzUDT6NBasOcV8
yjwjGzEFGS52f3Llv4RadVlCdTBSCH9lZLgA+fy9caKXIf4hyhrPjmYhdFUS6KTO
095d9URNe2lEWjDGTU3uQ0qqT+JfzVJ2hXa4AacetoQgJKvY40UpaPe+Ix5sH890
RtXFZN570PbURqJy5/HkRmEFVxg6XbbbG0eSPxISQEJsJwfaEszij6g7Cb5krPQw
AsUMYaeNV0z1O8N+3JFpQrkEhHMHC0i/9M7O19vgmlCh8YEpquR/kGP8Mq2Z/JXn
5nBeioDRDeN85/gYKubz1PkQ+CW9kgcW1BULbSy0SN+j/FlU8ZV/ZQ8p0tXi9RdW
RFJR+sduXcS2WUruDpIeyc9Ix8ggvaVokTldKSr9yhfqKud2W+5tEbcBtCaeK2+b
n4XlgGEKorMQ9gFJ+Kj9RkQ4sm3o1FmeuDbZHKwXs8nw7Bk1hfBwnbNVlyJcDFyn
272X41MIBW/vxRPPoIZAjja6QHQ1LJ1aCo218Unf4mA9oFYPUVO9oBFwXdeBi23h
GQAkLE59fvxQs8A11mNL
-----END CERTIFICATE-----
`
		}

		if os.Getenv("LOG_LEVEL") == "DEBUG" {
			fmt.Printf("%s\n%s\n\n", cert, key)
		}

		type feedback struct {
			result string
			thread string
			err    error
		}
		c := make(chan feedback)

		// go get password & token for gitlab (in parallel)

		secretGetter := func(ctx context.Context, c chan feedback, secret string) {
			val, err := v.GetSecret(ctx, vaultGitlabURL, secret) // secretValue MUST start with Uppercase
			if err != nil {
				fmt.Errorf("unable to retrieve the gitlab %s for the service agent -- %s", secret, err)
			}
			c <- feedback{
				thread: secret,
				result: val,
				err:    err,
			}
		}

		var gitlabPassword, gitlabToken string
		go secretGetter(ctx, c, "Password")
		go secretGetter(ctx, c, "Token")

		for i := 0; i < 2; i++ {
			f := <-c
			if f.err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %s\n", err)
				os.Exit(2)
			}
			if f.thread == "Token" {
				gitlabToken = f.result
			} else {
				gitlabPassword = f.result
			}
		}

		// store the public key in a new git branch & create a merge request
		go func(t string) {
			repo := &repo.GitWrapper{
				Logger:         logger,
				Repository:     repository,
				GitlabUsername: gitlabUsername,
				GitlabPassword: gitlabPassword,
				GitlabToken:    gitlabToken,
			}

			err := repo.AddOrUpdateFile(ctx, args[0], args, "somebody@example.com", cert)
			if err != nil {
				err = fmt.Errorf("unable to commit certificate to git repository: %s", err)
			}
			c <- feedback{
				thread: t,
				err:    err,
			}
		}("createMergeRequest")

		// store the private key in Vault
		go func(t string) {
			err := v.StoreSecret("secret/certificates/"+args[0], "key", key)
			if err != nil {
				logger.Error("unable to store secret in vault",
					zap.Error(err))
			}
			c <- feedback{
				thread: t,
				err:    err,
			}
		}("storeSecret")

		for i := 0; i < 2; i++ {
			f := <-c
			if f.err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %s\n", err)
				os.Exit(3)
			}
		}
	},
}

func init() {
	vaultCmd.AddCommand(newCertCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCertCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
