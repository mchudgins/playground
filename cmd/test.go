// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"strings"

	//log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	log "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	yaml2 "gopkg.in/yaml.v2"
)

type Authn struct {
	Audience string `yaml:"audience"`
	Issuer   string `yaml:"issuer"`
}

type Backend struct {
	Port          string `yaml:"port"`
	CanonicalName string `yaml:"canonicalName"`
}

type config struct {
	Authn    Authn   `yaml:"authn"`
	Backend  Backend `yaml:"backend"`
	LogLevel string  `yaml:"logLevel"`
}

var (
	defaultAuthnConfig   = Authn{Audience: "fubar.dstcorp.net", Issuer: "authn.dstcorp.net"}
	defaultBackendConfig = Backend{Port: ":8080", CanonicalName: "localhost"}
	defaultConfig        = config{Authn: defaultAuthnConfig, Backend: defaultBackendConfig, LogLevel: "Debug"}
)

func GetLogger() *log.Logger {
	//config := log.NewProductionConfig()
	config := log.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()

	return logger.With(log.String("x-request-id", "01234"))
}

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//		logger, _ := log.NewProduction()
		logger := GetLogger()
		defer logger.Sync()

		strarray, err := cmd.PersistentFlags().GetStringArray("define")
		if err != nil {
			logger.Fatal("unable to GetStringArray", log.Error(err))
			//			log.WithError(err).Fatal("unable to GetStringArray")
			panic(err)
		}
		for _, val := range strarray {
			//log.WithField("element", val).Info()
			logger.Info("", log.String("element", val))
			amendStruct(val)
		}

		y, err := yaml.Marshal(defaultConfig)
		if err != nil {
			//log.WithError(err).Fatal("while marshaling to yaml")
			logger.Fatal("while marshaling to yaml", log.Error(err))
			panic(err)
		}
		logger.Info("\n" + string(y))
		//log.Info("\n" + string(y))
		logger.Info("pre-exit")
		logger.Info("exit", log.String("status", "ok"))
	},
}

func buildYAMLKey(key string) (string, int) {
	items := strings.Split(key, ".")
	yamlKey := items[0] + ":"
	for i := 1; i < len(items); i++ {
		yamlKey += "\n"
		for j := i; j > 0; j-- {
			yamlKey += "  "
		}
		yamlKey += items[i] + ":"
	}
	return yamlKey, len(items)
}

func amendStruct(val string) {
	//logger, _ := log.NewProduction()
	logger := GetLogger()
	defer logger.Sync()

	elements := strings.Split(val, "=")
	for i, el := range elements {
		//log.WithField("i", i).WithField("val", el).Info()
		logger.Info("", log.Int("i", i), log.String("val", el))
	}

	//y := elements[0] + ": " + elements[1]
	y, _ := buildYAMLKey(elements[0])
	y += " " + elements[1]

	//log.WithField("y", y).Info("yaml.Unmarshal")
	logger.Info("yaml.Unmarshal", log.String("y", y))

	//err = yaml.Unmarshal([]byte(y), &defaultConfig)
	err := yaml2.Unmarshal([]byte(y), &defaultConfig)
	if err != nil {
		logger.Fatal("Unmarshal", log.Error(err), log.String("y", y))
		//log.WithError(err).Infof("yaml.Unmarshal('%s')", y)
	}
}

func init() {
	RootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	testCmd.PersistentFlags().StringArrayP("define", "D", []string{}, "configuration overrides")
	testCmd.PersistentFlags().StringP("config", "c", "app.yaml", "configuration source, e.g., https://config/....")
}
