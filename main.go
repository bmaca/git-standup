package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Commit struct {
	Repo    string
	Hash    string
	Date    string // YYYY-MM-DD
	Subject string
}

type reposFlag []string

func (r *reposFlag) String() string     { return strings.Join(*r, ",") }
func (r *reposFlag) Set(v string) error { *r = append(*r, v); return nil }

func main() {
	var (
		repos    reposFlag
		author   string
		sinceStr string
		untilStr string
		lastWeek bool
		thisWeek bool
		markdown bool
	)

	flag.Var(&repos, "repo", "path to git repository (can be repeated)")
	flag.StringVar(&author, "author", "", "author name or email (required)")
	flag.StringVar(&sinceStr, "since", "", "start date (YYYY-MM-DD)")
	flag.StringVar(&untilStr, "until", "", "end date (YYYY-MM-DD)")
	flag.BoolVar(&lastWeek, "last-week", false, "use last calendar week (Mon–Sun)")
	flag.BoolVar(&thisWeek, "this-week", false, "use this calendar week (Mon–Sun)")
	flag.BoolVar(&markdown, "markdown", false, "output in Markdown format")
	flag.Parse()

	if author == "" {
		fmt.Fprintln(os.Stderr, "error: -author is required")
		os.Exit(1)
	}
	if len(repos) == 0 {
		// default to current dir
		repos = append(repos, ".")
	}

	if lastWeek || thisWeek {
		sinceStr, untilStr = weekRange(lastWeek)
	} else {
		if sinceStr == "" && untilStr == "" {
			now := time.Now()
			sinceStr = now.AddDate(0, 0, -7).Format("2006-01-02")
			untilStr = now.Format("2006-01-02")
		} else if sinceStr == "" || untilStr == "" {
			fmt.Fprintln(os.Stderr, "error: either provide both -since and -until, or neither")
			os.Exit(1)
		}
	}

	// key: date -> repo -> []Commit
	byDayRepo := map[string]map[string][]Commit{}
	foundAny := false

	for _, repo := range repos {
		commits, err := getCommits(repo, author, sinceStr, untilStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: repo %s: %v\n", repo, err)
			continue
		}
		if len(commits) > 0 {
			foundAny = true
		}
		for _, c := range commits {
			if byDayRepo[c.Date] == nil {
				byDayRepo[c.Date] = map[string][]Commit{}
			}
			byDayRepo[c.Date][c.Repo] = append(byDayRepo[c.Date][c.Repo], c)
		}
	}
	if !foundAny {
		fmt.Fprintf(os.Stderr,
			"no commits found for author %q in range %s to %s.\nEnsure you have given it the correct User.\n",
			author, sinceStr, untilStr,
		)
		os.Exit(1)
	}

	if markdown {
		printMarkdownReport(byDayRepo, sinceStr, untilStr, author)
	} else {
		printTextReport(byDayRepo, sinceStr, untilStr, author)
	}
}

func weekRange(lastWeek bool) (since, until string) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	if lastWeek {
		monday = monday.AddDate(0, 0, -7)
	}
	sunday := monday.AddDate(0, 0, 6)
	return monday.Format("2006-01-02"), sunday.Format("2006-01-02")
}

func getCommits(repoPath, author, since, until string) ([]Commit, error) {
	args := []string{
		"-C", repoPath,
		"log",
		"--author=" + author,
		"--since=" + since,
		"--until=" + until,
		`--pretty=format:%H|%ad|%s`,
		"--date=short",
	}

	cmd := exec.Command("git", args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting git: %w", err)
	}

	var commits []Commit
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		commits = append(commits, Commit{
			Repo:    repoPath,
			Hash:    parts[0],
			Date:    parts[1],
			Subject: parts[2],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning git output: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	return commits, nil
}

func printTextReport(byDayRepo map[string]map[string][]Commit, since, until, author string) {
	fmt.Printf("Weekly report for %s (%s to %s)\n\n", author, since, until)

	// sort days
	days := make([]string, 0, len(byDayRepo))
	for d := range byDayRepo {
		days = append(days, d)
	}
	sort.Strings(days)

	for _, day := range days {
		fmt.Printf("%s:\n", day)
		repos := byDayRepo[day]

		// sort repos for deterministic output
		repoNames := make([]string, 0, len(repos))
		for r := range repos {
			repoNames = append(repoNames, r)
		}
		sort.Strings(repoNames)

		for _, repo := range repoNames {
			repoName := filepath.Base(repo)
			fmt.Printf("  [%s]\n", repoName)
			for _, c := range repos[repo] {
				fmt.Printf("    - %s\n", c.Subject)
			}
			fmt.Println()
		}
	}
}

func printMarkdownReport(byDayRepo map[string]map[string][]Commit, since, until, author string) {
	fmt.Printf("# Weekly report for %s\n\n", author)
	fmt.Printf("_Range: %s to %s_\n\n", since, until)

	days := make([]string, 0, len(byDayRepo))
	for d := range byDayRepo {
		days = append(days, d)
	}
	sort.Strings(days)

	for _, day := range days {
		fmt.Printf("## %s\n\n", day)
		repos := byDayRepo[day]

		repoNames := make([]string, 0, len(repos))
		for r := range repos {
			repoNames = append(repoNames, r)
		}
		sort.Strings(repoNames)

		for _, repo := range repoNames {
			repoName := filepath.Base(repo)
			fmt.Printf("### Repo: `%s`\n\n", repoName)
			for _, c := range repos[repo] {
				fmt.Printf("- %s\n", c.Subject)
			}
			fmt.Println()
		}
	}
}
