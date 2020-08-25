package checks

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/cli/cli/internal/ghrepo"
	"github.com/cli/cli/pkg/cmdutil"
	"github.com/cli/cli/pkg/httpmock"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdChecks(t *testing.T) {
	tests := []struct {
		name  string
		cli   string
		wants ChecksOptions
	}{
		{
			name:  "no arguments",
			cli:   "",
			wants: ChecksOptions{},
		},
		{
			name: "pr argument",
			cli:  "1234",
			wants: ChecksOptions{
				SelectorArg: "1234",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: io,
			}

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *ChecksOptions
			cmd := NewCmdChecks(f, func(opts *ChecksOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			assert.NoError(t, err)

			assert.Equal(t, tt.wants.SelectorArg, gotOpts.SelectorArg)
		})
	}
}

func Test_checksRun_tty(t *testing.T) {
	tests := []struct {
		name    string
		payload checkRunsPayload
		stubs   func(*httpmock.Registry)
		wantOut string
	}{
		{
			name: "no commits",
			stubs: func(reg *httpmock.Registry) {
				reg.StubResponse(200, bytes.NewBufferString(`
					{ "data": { "repository": {
						"pullRequest": { "number": 123 }
					} } }
				`))
			},
		},
		{
			name: "no checks",
			payload: checkRunsPayload{
				CheckRuns: []checkRunPayload{},
			},
		},
		// TODO some failing
		// TODO some pending
		// TODO all passing
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO extract to runCommand
			io, _, stdout, _ := iostreams.Test()
			io.SetStdoutTTY(true)

			opts := &ChecksOptions{
				IO: io,
				BaseRepo: func() (ghrepo.Interface, error) {
					return ghrepo.New("OWNER", "REPO"), nil
				},
				SelectorArg: "123",
			}

			reg := &httpmock.Registry{}
			if tt.stubs != nil {
				tt.stubs(reg)
			} else {
				reg.StubResponse(200, bytes.NewBufferString(`
				{ "data": { "repository": {
					"pullRequest": { "number": 123, "commits": { "nodes": [{"commit": {"oid": "abc"}}]} }
				} } }
			`))
			}
			reg.Register(httpmock.REST("GET", "repos/OWNER/REPO/commits/abc/check-runs"),
				httpmock.JSONResponse(tt.payload))

			opts.HttpClient = func() (*http.Client, error) {
				return &http.Client{Transport: reg}, nil
			}

			err := checksRun(opts)
			assert.NoError(t, err)

			assert.Equal(t, tt.wantOut, stdout.String())
		})
	}
}

func Test_checksRun_nontty(t *testing.T) {
	tests := []struct {
		name    string
		http    func(*httpmock.Registry)
		wantOut string
	}{
		// TODO some checks
		// TODO no checks
		// TODO no commits
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO

		})
	}
}
