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
	template "html/template"

	"io/ioutil"
	"os"
	"sort"

	"github.com/ghodss/yaml"
	"github.com/mchudgins/playground/pkg/cmd/backend/htmlGen"
	"github.com/spf13/cobra"
)

type ApiEndpoint struct {
	Name       string
	Desc       string
	Version    string
	SwaggerURL string
	Status     string
	Public     bool
}

type apiEntry struct {
	IconURL   string
	Name      string
	Desc      string
	Endpoints []ApiEndpoint
}

var (
	htmlTemplate = template.Must(template.New("html").Parse(htmlGen.DefaultHTML))
)

// read stdin, process the file, write to stdout
func Run(cmd *cobra.Command, args []string) {

	var buf []byte
	var err error

	if len(args) > 0 {
		buf, err = ioutil.ReadFile(args[0])

	} else {
		// read stdin into buf
		in := os.Stdin

		buf, err = ioutil.ReadAll(in)
	}
	if err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "unable to read file -- %s\n", err)
		return
	}

	// Unmarshal the buf
	apiList := make([]apiEntry, 0, 100)
	err = yaml.Unmarshal(buf, &apiList)
	if err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "unable to parse configuration -- %s\n", err)
		return
	}

	// fill in the blanks and validate
	for i, api := range apiList {

		sort.Slice(apiList[i].Endpoints, func(j, k int) bool {
			return api.Endpoints[j].Name < api.Endpoints[k].Name
		})

		for j, endpt := range api.Endpoints {
			if len(endpt.Status) == 0 {
				apiList[i].Endpoints[j].Status = "active"
			}

			if len(endpt.Version) == 0 {
				apiList[i].Endpoints[j].Version = "v1"
			}

			if len(endpt.SwaggerURL) == 0 {
				// no test endpoint? Then you're 'unavailable'!
				apiList[i].Endpoints[j].Status = "unavailable"
			}
		}
	}

	// write it out

	sort.Slice(apiList, func(i, j int) bool { // want a sorted apiList
		return apiList[i].Name < apiList[j].Name
	})

	type data struct {
		APIList []apiEntry
	}
	err = htmlTemplate.Execute(cmd.OutOrStdout(), data{APIList: apiList})
	if err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "Error executing template: %s\n", err)
		return
	}
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "htmlGen",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: Run,
}

func init() {
	RootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
