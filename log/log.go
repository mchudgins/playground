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

/*
log provides the logger utilities & interfaces
*/

package log

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetLogger(name string) *zap.Logger {
	// See the documentation for Config and zapcore.EncoderConfig for all the
	// available options.
	rawJSON := []byte(`{
	  "level": "debug",
	  "encoding": "json",
	  "outputPaths": ["stdout"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var config zap.Config
	if err := json.Unmarshal(rawJSON, &config); err != nil {
		panic(err)
	}
	config.InitialFields = make(map[string]interface{}, 1)
	config.InitialFields["cmd"] = name

	level := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if len(level) == 0 {
		level = "INFO"
	}

	switch level {
	case "DEBUG":
	case "TRACE":
		config.Level.SetLevel(zapcore.DebugLevel)
		break

	case "INFO":
		config.Level.SetLevel(zapcore.InfoLevel)
		break

	case "WARN":
		config.Level.SetLevel(zapcore.WarnLevel)

	default:
		fmt.Printf("Unknown LOG_LEVEL value %s.  Log Level set to INFO.", level)
	}

	devMode := strings.ToUpper(os.Getenv("DEVMODE"))
	if len(devMode) > 0 && devMode != "off" && devMode != "false" {
		config.Encoding = "console"
		config.EncoderConfig = zap.NewDevelopmentEncoderConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	//	config := log.NewDevelopmentConfig()
	//	config.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return logger //.With(log.String("x-request-id", "01234"))
}
