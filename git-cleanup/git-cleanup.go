// Copyright 2014 Google. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The git-cleanup command deletes branches which have already
// been merged to the Gerrit server.
package main

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	curBranch := ""
	out, err := exec.Command("git", "branch").Output()
	if err != nil {
		log.Fatalf("Error running git branch: %v", err)
	}
	var branches []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		cur := false
		if strings.HasPrefix(line, "* ") {
			line = line[2:]
			cur = true
		}
		if line == "" {
			continue
		}
		branches = append(branches, line)
		if cur {
			curBranch = line
		}
	}
	if curBranch != "master" {
		log.Fatalf("On branch %s; must be on master", curBranch)
	}

	for _, br := range branches {
		if br == "master" {
			continue
		}
		if isSubmitted(branchChangeID(br)) {
			// Display a ref for the branch we're about to delete,
			// so that if we screw up (never!), the user can get it back easily.
			short, err := exec.Command("git", "rev-parse", "--short", "refs/heads/"+br).Output()
			if err != nil {
				log.Fatalf("Error running git rev-parse: %v", err)
			}
			short = bytes.TrimSpace(short)
			log.Printf("Removing branch %s (%s) ...", br, short)
			if out, err := exec.Command("git", "branch", "-D", br).CombinedOutput(); err != nil {
				log.Printf("Error removing branch %s: %v, %s", br, err, out)
			}

			// Remove corresponding tag, if there is one.
			tag := br + ".mailed"
			short, err = exec.Command("git", "rev-parse", "--short", "refs/tags/"+tag).Output()
			if err != nil {
				continue
			}
			short = bytes.TrimSpace(short)
			log.Printf("Removing tag %s (%s) ...", tag, short)
			if out, err := exec.Command("git", "tag", "-d", tag).CombinedOutput(); err != nil {
				log.Printf("Error removing tag %s: %v, %s", tag, err, out)
			}
		}
	}
}

var changeRx = regexp.MustCompile(`(?m)^\s*Change-Id: (I[0-9a-f]+)`)

// returns change-id or the empty string
func branchChangeID(br string) string {
	out, _ := exec.Command("git", "show", br).CombinedOutput()
	if m := changeRx.FindSubmatch(out); m != nil {
		return string(m[1])
	}
	return ""
}

func isSubmitted(changeID string) bool {
	if changeID == "" {
		return false
	}
	for _, v := range changeIDLog() {
		if v == changeID {
			return true
		}
	}
	return false
}

var changeIDLogCache []string

func changeIDLog() (ret []string) {
	if v := changeIDLogCache; v != nil {
		return v
	}
	defer func() { changeIDLogCache = ret }()
	cmd := exec.Command("git", "log")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("pipe error: %v", err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("error running git log: %v", err)
	}
	bs := bufio.NewScanner(stdout)
	wantLine := []byte("Change-Id: ")
	for bs.Scan() {
		if !bytes.Contains(bs.Bytes(), wantLine) {
			continue
		}
		changeID := strings.TrimPrefix(strings.TrimSpace(bs.Text()), "Change-Id: ")
		ret = append(ret, changeID)
	}
	if err := bs.Err(); err != nil {
		log.Fatalf("error running git log: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatalf("error running git log: %v", err)
	}
	return
}
