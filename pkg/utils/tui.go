// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"

	"github.com/AstroProfundis/tabby"
	"golang.org/x/crypto/ssh/terminal"
)

type table [][]string

func (p table) Len() int {
	return len(p)
}

func (p table) Less(i, j int) bool {
	return p[i][1] > p[j][1]
}

func (p table) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// PrintTable accepts a matrix of strings and print them as ASCII table to terminal
func PrintTable(rows table, header bool) {
	// Print the table
	t := tabby.New()
	if header {
		addRow(t, rows[0], header)
		rows = rows[1:]
	}
	sort.Sort(rows)
	for _, row := range rows {
		addRow(t, row, false)
	}
	t.Print()
}

func addRow(t *tabby.Tabby, rawLine []string, header bool) {
	// Convert []string to []interface{}
	row := make([]interface{}, len(rawLine))
	for i, v := range rawLine {
		row[i] = v
	}

	// Add line to the table
	if header {
		t.AddHeader(row...)
	} else {
		t.AddLine(row...)
	}
}

// Prompt accepts input from console by user
func Prompt(prompt string) string {
	if prompt != "" {
		prompt += " " // append a whitespace
	}
	fmt.Printf(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(input, "\n")
}

// GetPasswd reads a password input from console
func GetPasswd(prompt string) string {
	if prompt != "" {
		prompt += " " // append a whitespace
	}
	fmt.Printf(prompt)

	input, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.Trim(string(input), "\n"))
}
