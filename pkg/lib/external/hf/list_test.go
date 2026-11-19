// Copyright 2025 The KitOps Authors.
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

package hf

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalkRepoTree(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"type":"file","oid":"o1","size":4,"path":"a.json"}]`))
	}))
	defer srv.Close()

	base, err := url.Parse(srv.URL)
	require.NoError(t, err)

	listing, err := walkRepoTree(context.Background(), srv.Client(), "", base, ".")
	require.NoError(t, err)
	require.Len(t, listing.Files, 1)
	assert.Equal(t, "a.json", listing.Files[0].Path)
}

func TestWalkRepoTreeFollowsPagination(t *testing.T) {
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/tree" && r.URL.Query().Get("cursor") == "":
			w.Header().Set("Link", fmt.Sprintf(`<%s/tree?cursor=page2>; rel="next"`, srvURL))
			_, _ = w.Write([]byte(`[{"type":"file","path":"a.json"}]`))
		case r.URL.Path == "/tree" && r.URL.Query().Get("cursor") == "page2":
			_, _ = w.Write([]byte(`[{"type":"file","path":"b.json"},{"type":"directory","path":"sub"}]`))
		case r.URL.Path == "/tree/sub" && r.URL.Query().Get("cursor") == "":
			w.Header().Set("Link", `</tree/sub?cursor=page2>; rel="next"`)
			_, _ = w.Write([]byte(`[{"type":"file","path":"sub/c.json"}]`))
		case r.URL.Path == "/tree/sub" && r.URL.Query().Get("cursor") == "page2":
			_, _ = w.Write([]byte(`[{"type":"file","path":"sub/d.json"}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"unexpected request"}`))
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	base, err := url.Parse(srv.URL + "/tree")
	require.NoError(t, err)

	listing, err := walkRepoTree(context.Background(), srv.Client(), "", base, ".")
	require.NoError(t, err)
	require.Len(t, listing.Files, 2)
	assert.Equal(t, "a.json", listing.Files[0].Path)
	assert.Equal(t, "b.json", listing.Files[1].Path)
	require.Len(t, listing.Subdirs, 1)
	require.Len(t, listing.Subdirs[0].Files, 2)
	assert.Equal(t, "sub/c.json", listing.Subdirs[0].Files[0].Path)
	assert.Equal(t, "sub/d.json", listing.Subdirs[0].Files[1].Path)
}

func TestNextPageUrl(t *testing.T) {
	cur, err := url.Parse("https://huggingface.co/api/models/org/repo/tree/main")
	require.NoError(t, err)

	tests := []struct {
		name    string
		link    string
		want    string
		wantErr bool
	}{
		{name: "no link header", link: ""},
		{name: "no next link", link: `<https://huggingface.co/api/models/org/repo/tree/main?cursor=a>; rel="prev"`},
		{
			name: "next link",
			link: `<https://huggingface.co/api/models/org/repo/tree/main?limit=1000&cursor=abc%3D%3D>; rel="next"`,
			want: "https://huggingface.co/api/models/org/repo/tree/main?limit=1000&cursor=abc%3D%3D",
		},
		{
			name: "next link among others",
			link: `<https://huggingface.co/x?cursor=a>; rel="prev", <https://huggingface.co/x?cursor=b>; rel="next"`,
			want: "https://huggingface.co/x?cursor=b",
		},
		{name: "relative next link", link: `</api/x?cursor=b>; rel="next"`, want: "https://huggingface.co/api/x?cursor=b"},
		{name: "next link on another host", link: `<https://example.com/x?cursor=b>; rel="next"`, wantErr: true},
		{name: "next link is the current page", link: `<https://huggingface.co/api/models/org/repo/tree/main>; rel="next"`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := http.Header{}
			if tt.link != "" {
				header.Set("Link", tt.link)
			}
			next, err := nextPageUrl(cur, header)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.want == "" {
				assert.Nil(t, next)
				return
			}
			require.NotNil(t, next)
			assert.Equal(t, tt.want, next.String())
		})
	}
}
