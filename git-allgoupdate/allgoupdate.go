// Copyright 2015 Google. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The git-allgoupdate command runs "go get -u" on all Go-related
// repos. This is only related to git in that Go uses git now
// and I had no better place to put this tool.
package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func main() {
	if os.Getenv("GOPATH") == "" {
		log.Fatalf("No GOPATH set.")
	}
	repos := []string{
		"golang.org/x/benchmarks",
		"golang.org/x/blog",
		"golang.org/x/build",
		"golang.org/x/crypto",
		"golang.org/x/debug",
		"golang.org/x/example",
		"golang.org/x/exp",
		"golang.org/x/image",
		"golang.org/x/mobile",
		"golang.org/x/net",
		"golang.org/x/oauth2",
		"golang.org/x/playground",
		"golang.org/x/review",
		"golang.org/x/sys",
		"golang.org/x/talks",
		"golang.org/x/text",
		"golang.org/x/tools",
		"google.golang.org/api",
		"google.golang.org/cloud",
		"github.com/golang/groupcache",
		"github.com/golang/winstrap",
	}
	var wg sync.WaitGroup
	for _, repo := range repos {
		repo := repo
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("Updating %s ...", repo)
			err := exec.Command("go", "get", "-u", repo).Run()
			log.Printf("Updated %s: %v", repo, err)
			if e, ok := userEmail[os.Getenv("USER")]; ok {
				cmd := exec.Command("git", "config", "user.email", e)
				cmd.Dir = filepath.Join(os.Getenv("GOPATH"), "src", repo)
				if err := cmd.Run(); err != nil {
					log.Printf("Failed to update user.email in %s: %v", cmd.Dir, err)
				}
			}
		}()
	}
	wg.Wait()
}

// email address to use for Go repos, as a function of $USER
var userEmail = map[string]string{
	"bradfitz": "bradfitz@golang.org",
}
