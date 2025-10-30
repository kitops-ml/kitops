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

package info

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/kitops-ml/kitops/pkg/artifact"
	"github.com/kitops-ml/kitops/pkg/cmd/options"
	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/repo/util"
	"github.com/kitops-ml/kitops/pkg/output"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	shortDesc = `Show the configuration for a modelkit`
	longDesc  = `Print the contents of a modelkit config to the screen.

By default, kit will check local storage for the specified modelkit. To see
the configuration for a modelkit stored on a remote registry, use the
--remote flag.`
	example = `# See configuration for a local modelkit:
kit info mymodel:mytag

# See configuration for a local modelkit by digest:
kit info mymodel@sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a

# See configuration for a remote modelkit:
kit info --remote registry.example.com/my-model:1.0.0`
)

// Currently supported filter syntax: alphanumeric (plus dashes and underscores), dot-delimited fields
var validFilterRegexp = regexp.MustCompile(`^\.?[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)*$`)

type infoOptions struct {
	options.NetworkOptions
	configHome  string
	checkRemote bool
	modelRef    *registry.Reference
	filter      string
}

func InfoCommand() *cobra.Command {
	opts := &infoOptions{}

	cmd := &cobra.Command{
		Use:     "info [flags] MODELKIT",
		Short:   shortDesc,
		Long:    longDesc,
		Example: example,
		RunE:    runCommand(opts),
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if cmd.Flags().Changed("remote") {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) >= 1 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.GetLocalModelKitsCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
	}

	opts.AddNetworkFlags(cmd)
	cmd.Flags().BoolVarP(&opts.checkRemote, "remote", "r", false, "Check remote registry instead of local storage")
	cmd.Flags().StringVarP(&opts.filter, "filter", "f", "", "filter with node selectors")
	cmd.Flags().SortFlags = false

	return cmd
}

func runCommand(opts *infoOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}
		config, err := getInfo(cmd.Context(), opts)
		if err != nil {
			if errors.Is(err, errdef.ErrNotFound) {
				return output.Fatalf("Could not find modelkit %s", util.FormatRepositoryForDisplay(opts.modelRef.String()))
			}
			return output.Fatalf("Error resolving modelkit: %s", err)
		}

		if len(opts.filter) > 0 {
			filteredOutput, err := filterKitfile(config, opts.filter)
			if err != nil {
				return output.Fatalln(err)
			}
			fmt.Print(string(filteredOutput))
		} else {
			yamlBytes, err := config.MarshalToYAML()
			if err != nil {
				return output.Fatalf("Error formatting manifest: %w", err)
			}
			fmt.Print(string(yamlBytes))
		}

		return nil
	}
}

func (opts *infoOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome

	ref, extraTags, err := util.ParseReference(args[0])
	if err != nil {
		return err
	}
	if len(extraTags) > 0 {
		return fmt.Errorf("invalid reference format: extra tags are not supported: %s", strings.Join(extraTags, ", "))
	}
	if ref.Reference == "" {
		return fmt.Errorf("missing tag or digest from ModelKit reference '%s'", args[0])
	}
	opts.modelRef = ref

	if opts.modelRef.Registry == util.DefaultRegistry && opts.checkRemote {
		return fmt.Errorf("can not check remote: %s does not contain registry", util.FormatRepositoryForDisplay(opts.modelRef.String()))
	}

	if err := opts.NetworkOptions.Complete(ctx, args); err != nil {
		return err
	}

	return nil
}

func filterKitfile(config *artifact.KitFile, filter string) ([]byte, error) {
	if err := checkFilterIsValid(filter); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}
	// Accept filters that start (jq-style) and don't start with a '.'; we need to trim as otherwise we start the list
	// with an empty string
	filter = strings.TrimPrefix(filter, ".")

	var filterSlice = strings.Split(filter, ".")
	value := reflect.ValueOf(config)
	for _, str := range filterSlice {
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		field := value.FieldByName(cases.Title(language.Und, cases.NoLower).String(str))
		if !field.IsValid() {
			return nil, fmt.Errorf("error filtering output: cannot find required node")
		}
		value = field
	}

	buf := new(bytes.Buffer)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(value.Interface()); err != nil {
		return nil, fmt.Errorf("error formatting manifest: %w", err)
	}
	return buf.Bytes(), nil
}

func checkFilterIsValid(filter string) error {
	if strings.Contains(filter, "[") {
		return fmt.Errorf("array access using '[]' is not currently supported")
	}
	if !validFilterRegexp.MatchString(filter) {
		return fmt.Errorf("invalid filter: %s", filter)
	}
	return nil
}
