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
	"fmt"
	"strings"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	withinFieldMask []int = []int{8, 8, 8, 8}
)

// cidrCmd represents the cidr command
var cidrCmd = &cobra.Command{
	Use:   "cidr <value>",
	Short: "compute the octet notation for a bitmask",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			cmd.Usage()
			return
		}

		mask, err := cmd.Flags().GetString("mask")
		if err != nil {
			panic(err)
		}
		within, err := cmd.Flags().GetString("within")
		if err != nil {
			panic(err)
		}

		str, err := translate(args[0], mask, within)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		fmt.Printf("%s\n", str)
	},
}

func translate(value, mask, within string) (string, error) {
	fmt.Printf("value: %s, mask: %s\n", value, mask)

	//parse the mask
	fields, err := parse(mask)
	if err != nil {
		return "", err
	}

	// make sure the mask sums to 32
	sum := 0
	for _, i := range fields {
		sum += i
	}
	if sum != 32 {
		return "", fmt.Errorf("expected the mask to define 32 bits, only found %d", sum)
	}

	// parse the value
	values, err := parse(value)

	if len(fields) != len(values) {
		return "",
			fmt.Errorf("different number of fields in the mask(%d) and the value(%d)", len(fields), len(values))
	}

	netmask, err := computeCIDR(fields, values)
	if err != nil {
		return "", err
	}

	withinValues, err := parse(within)
	if len(withinFieldMask) != len(withinValues) {
		return "",
			fmt.Errorf("different number of fields in the mask(%d) and the value(%d)",
				len(withinFieldMask), len(withinValues))
	}
	withinCIDR, err := computeCIDR(withinFieldMask, withinValues)

	for i, x := range withinCIDR {
		netmask[i] = netmask[i] | x
	}

	var output string
	output = fmt.Sprintf("%d.%d.%d.%d",
		netmask[0],
		netmask[1],
		netmask[2],
		netmask[3])

	return output, nil
}

func parse(mask string) ([]int, error) {
	var sep string

	for _, c := range mask {
		if c < '0' || c > '9' {
			sep = string(c)
			break
		}
	}

	if len(sep) == 0 {
		return nil, fmt.Errorf("The mask '%s' has only one or no fields", mask)
	}

	str := strings.Split(mask, sep)
	fields := make([]int, len(str))

	for i, s := range str {
		var err error
		fields[i], err = strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("error parsing mask field '%s' -- %s", s, err)
		}
	}

	return fields, nil
}

func computeCIDR(fields, values []int) ([]int, error) {

	var result uint32
	for i, f := range fields {
		var field uint32 = uint32(f)
		var uval uint32 = uint32(values[i])
		field = uval & generateAndMask(f)
		if field != uval {
			return nil, fmt.Errorf("field #%d (%d) exceeds the defined field length of %d", i, uval, f)
		}

		result = result << uint32(f)
		result = result | field
	}

	netmask := make([]int, 4)
	for i, _ := range netmask {
		index := len(netmask) - i - 1
		netmask[index] = int(result & 0x0ff)
		result = result >> 8
	}

	return netmask, nil
}

func generateAndMask(length int) uint32 {
	var mask uint32

	mask = 0
	for i := 0; i < length; i++ {
		mask <<= 1
		mask |= 1
	}
	return mask
}

func init() {
	RootCmd.AddCommand(cidrCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cidrCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cidrCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	cidrCmd.Flags().StringP("mask", "m", "8:13:4:7", "bitmask for translation")
	cidrCmd.Flags().StringP("within", "w", "0.0.0.0", "result is OR'ed with this CIDR")
}
