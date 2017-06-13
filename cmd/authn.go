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

	"github.com/mchudgins/playground/pkg/cmd/authn"
	"github.com/spf13/cobra"
)

// authnCmd represents the authn command
var authnCmd = &cobra.Command{
	Use:   "authn",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.PersistentFlags().GetString("port")
		if err != nil {
			fmt.Println("error:  %s", err)
			return
		}
		host, err := cmd.PersistentFlags().GetString("host")
		if err != nil {
			fmt.Println("error: %s", err)
		}
		err = authn.Run(port, host)
		if err != nil {
			fmt.Println("error:  %s", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(authnCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// authnCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// authnCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	authnCmd.PersistentFlags().StringP("port", "p", ":8080", "http listen port")
	authnCmd.PersistentFlags().StringP("host", "n", "", "Canonical Host Name (e.g., http://domain.com)")

}
