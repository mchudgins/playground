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
	"log"

	echo "github.com/dstcorp/rpc-golang/service"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// echoClientCmd represents the echoClient command
var echoClientCmd = &cobra.Command{
	Use:   "echoClient",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		creds, err := credentials.NewClientTLSFromFile("/usr/local/share/ca-certificates/dst-root.crt",
			"")
		if err != nil {
			panic(err)
		}
		echoServer, err := cmd.Flags().GetString("hostname")
		if err != nil {
			panic(err)
		}
		conn, err := grpc.Dial(echoServer,
			grpc.WithTransportCredentials(creds),
			grpc.WithCompressor(grpc.NewGZIPCompressor()),
			grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
			grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor))
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		client := echo.NewEchoServiceClient(conn)

		request := &echo.EchoRequest{
			Message: args[0],
		}
		response, err := client.Echo(context.Background(), request)
		if err != nil {
			panic(err)
		}
		log.Printf("response: %s\n", response.Message)
	},
}

func init() {
	RootCmd.AddCommand(echoClientCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// echoClientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// echoClientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	echoClientCmd.Flags().StringP("hostname", "H", "echo.local.dstcorp.io:50050", "host:port of echo server")
}
