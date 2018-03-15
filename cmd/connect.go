// Copyright © 2017 Mike Hudgins <mchudgins@gmail.com>
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
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("connect called")

		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		// if you want to change the loading rules (which files in which order), you can do so here

		configOverrides := &clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				LocationOfOrigin:      "dev",
				Server:                "https://dev.dstcorp.io:8443/",
				InsecureSkipTLSVerify: true,
			},
			AuthInfo: clientcmdapi.AuthInfo{
				LocationOfOrigin: "dev",
				Token:            "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJtY2giLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlY3JldC5uYW1lIjoiY2VydC1tZ3ItbGFtYmRhLXRva2VuLWx4NWg2Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6ImNlcnQtbWdyLWxhbWJkYSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImU5YjVkNWFiLTI4NjctMTFlOC05YjU3LTEyNDc0N2YyM2RiYyIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDptY2g6Y2VydC1tZ3ItbGFtYmRhIn0.oGrJUG7NIn0SBPuwA-woZx1YLa2bKbDErj0nCK2xz_He5y7r1JxkFKvRcmhEtdSr-yCLARYiNXUWxwAT5XAw7Vqv1kUK9ZpUydnVJMk9I7-xIDq7Z-6eIkMyudrh4vGL8ssTILG9EoKcs-Xk6housCRbCJXBNIF-ewb2ml-233terW7xyUZnTbXnfbSfnWrx76TfHayiWx9JRWMSo5ZzlAV0st0WpmJyJS3x6umf7c75oK-_29iDGRmk8ZgP4dVPRi8YYqRpMGilGIb07Yu4p4ac4xnSgAlLcER6AHXUq-lrlec_NxWXSX6Thn43SGfAtQ2d06iL9-4DY92yT3m7Fg",
			},
		}

		// if you want to change override values or bind them to flags, there are methods to help you

		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err := kubeConfig.ClientConfig()
		if err != nil {
			panic(err)
		}

		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}

		ns, err := client.Pods("mch").List(v1.ListOptions{})
		if err != nil {
			panic(err)
		}
		for _, n := range ns.Items {
			fmt.Printf("%s/%s/%s\n", n.Name, n.ClusterName, n.Namespace)
		}
	},
}

func init() {
	k8sCmd.AddCommand(connectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// connectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// connectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
