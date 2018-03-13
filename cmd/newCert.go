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

		// v is the vault client
		v := vault.New(logger, vaultAddress, vaultToken)

		// type 'feedback' allows go routines to communicate with the main routine
		type feedback struct {
			result string
			thread string
			err    error
		}
		c := make(chan feedback)

		// go get the gitlab password & token (in parallel)
		//
		// gitlab uses the password for checkin/checkout operations
		// and the token for merge requests.

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

		// ask Vault to create the public key & private key
		cert, key, err := v.NewCert(ctx, args[0], args[1:])
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error creating certificate: %s\n\n", err)
			os.Exit(1)
		}

		//  wait for the go routines to fetch the pword and token
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

		// wait for public & private keys to be persisted
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
