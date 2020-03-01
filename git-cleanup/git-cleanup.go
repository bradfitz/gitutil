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
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// BUG: cherry-pick branches will be removed as soon as the main CL is merged,
// as cherry-picks have the same Change-Id as the original CL.

func main() {
	var targetBranch string
	out, err := exec.Command("git", "branch").Output()
	if err != nil {
		log.Fatalf("Error running git branch: %v", err)
	}
	var branches []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "* ") {
			line = line[2:]
			targetBranch = line
		}
		if strings.HasPrefix(line, "+ ") {
			// Branch is checked out in a worktree.
			// We can't delete it anyway, so ignore it.
			continue
		}
		if line == "" {
			continue
		}
		branches = append(branches, line)
	}
	if len(os.Args) > 1 {
		targetBranch = os.Args[1]
	}
	if !isMainBranch(targetBranch) {
		log.Fatalf("Selected branch %s; must be a master or a dev branch.", targetBranch)
	}

	for _, br := range branches {
		if isMainBranch(br) {
			continue
		}
		if isSubmitted(targetBranch, branchChangeID(br)) {
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
	out, err := exec.Command("git", "show", br).CombinedOutput()
	if err != nil {
		log.Printf("Error running git show %v: %v: %s", br, err, out)
	}
	if m := changeRx.FindSubmatch(out); m != nil {
		return string(m[1])
	}
	return ""
}

// isMainBranch reports whether br is a shared development branch.
func isMainBranch(br string) bool {
	br = strings.TrimPrefix(br, "origin/")
	return br == "master" || strings.HasPrefix(br, "dev.")
}

func isSubmitted(br, changeID string) bool {
	if changeID == "" {
		return false
	}
	for _, v := range changeIDLog(br) {
		if v == changeID {
			return true
		}
	}
	return false
}

var changeIDLogCache = make(map[string][]string)

func changeIDLog(br string) (ret []string) {
	if v, ok := changeIDLogCache[br]; ok {
		return v
	}
	defer func() { changeIDLogCache[br] = ret }()
	cmd := exec.Command("git", "log", "-F", "--grep", "Change-Id:", br)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Pipe error: %v", err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Error running git log: %v", err)
	}
	bs := bufio.NewScanner(stdout)
	const wantLine = "Change-Id: "
	for bs.Scan() {
		line := strings.TrimSpace(bs.Text())
		if !strings.HasPrefix(line, wantLine) {
			continue
		}
		changeID := strings.TrimPrefix(line, wantLine)
		ret = append(ret, changeID)
	}
	if err := bs.Err(); err != nil {
		log.Fatalf("Error running git log: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatalf("Error running git log: %v", err)
	}
	return
}
