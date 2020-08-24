package checks

import (
	"fmt"
	"time"

	"github.com/cli/cli/api"
	"github.com/cli/cli/internal/ghrepo"
)

type checkRun struct {
	Name    string
	Status  string
	Link    string
	Elapsed time.Duration
}

type checkRunList struct {
	Passing   int
	Failing   int
	Pending   int
	checkRuns []checkRun
}

func checkRuns(client *api.Client, repo ghrepo.Interface, pr *api.PullRequest) (checkRunList, error) {
	list := checkRunList{}
	path := fmt.Sprintf("repos/%s/%s/commits/%s/check-runs",
		repo.RepoOwner(), repo.RepoName(), pr.Commits.Nodes[0].Commit.Oid)
	var response struct {
		CheckRuns []struct {
			Name        string
			Status      string
			Conclusion  string
			StartedAt   time.Time `json:"started_at"`
			CompletedAt time.Time `json:"completed_at"`
			HtmlUrl     string    `json:"html_url"`
		} `json:"check_runs"`
	}

	err := client.REST(repo.RepoHost(), "GET", path, nil, &response)
	if err != nil {
		return list, err
	}

	for _, cr := range response.CheckRuns {
		elapsed := cr.CompletedAt.Sub(cr.StartedAt)

		run := checkRun{
			Elapsed: elapsed,
			Name:    cr.Name,
			Link:    cr.HtmlUrl,
		}

		if cr.Status == "in_progress" || cr.Status == "queued" {
			list.Pending++
			run.Status = "pending"
		} else if cr.Status == "completed" {
			switch cr.Conclusion {
			case "neutral", "success":
				list.Passing++
				run.Status = "pass"
			case "canceled", "timed_out", "failed":
				list.Failing++
				run.Status = "fail"
			}
		} else {
			panic(fmt.Errorf("unsupported status: %q", cr.Status))
		}

		list.checkRuns = append(list.checkRuns, run)
	}

	return list, nil
}
