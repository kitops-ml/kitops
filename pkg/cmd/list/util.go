// Copyright 2024 The KitOps Authors.
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
//
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"fmt"

	"github.com/kitops-ml/kitops/pkg/kit"
)

const (
	listTableHeader = "REPOSITORY\tTAG\tMAINTAINER\tNAME\tSIZE\tDIGEST"
	listTableFmt    = "%s\t%s\t%s\t%s\t%s\t%s"
)

type modelInfo = kit.ModelInfo

func formatModelInfo(m *modelInfo) []string {
	if len(m.Tags) == 0 {
		line := fmt.Sprintf(listTableFmt, m.Repo, "<none>", m.Author, m.ModelName, m.Size, m.Digest)
		return []string{line}
	}
	var lines []string
	for _, tag := range m.Tags {
		line := fmt.Sprintf(listTableFmt, m.Repo, tag, m.Author, m.ModelName, m.Size, m.Digest)
		lines = append(lines, line)
	}
	return lines
}
