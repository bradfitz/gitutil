// Copyright 2015 Google. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The git-clpatch command fetches a CL
// from gerrit and cherry-picks it without commit.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s <cl number>\n", os.Args[0])
	os.Exit(2)
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}
	cl, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		usage()
	}

	// find the remote ref for the CL
	resp, err := http.Get(fmt.Sprintf("https://go-review.googlesource.com/changes/%d/?o=CURRENT_REVISION", cl))
	if err != nil {
		log.Fatal(err)
	}

	// Work around https://code.google.com/p/gerrit/issues/detail?id=3540
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	body = bytes.TrimPrefix(body, []byte(")]}'"))

	var parse struct {
		CurrentRevision string `json:"current_revision"`
		Revisions       map[string]struct {
			Fetch struct {
				HTTP struct {
					URL string
					Ref string
				}
			}
		}
	}

	if err := json.Unmarshal(body, &parse); err != nil {
		log.Fatal(err)
	}

	ref := parse.Revisions[parse.CurrentRevision].Fetch.HTTP

	// git fetch and cherry-pick
	if err := exec.Command("git", "fetch", ref.URL, ref.Ref).Run(); err != nil {
		log.Fatal(err)
	}
	if err := exec.Command("git", "cherry-pick", "-n", "FETCH_HEAD").Run(); err != nil {
		log.Fatal(err)
	}
}
