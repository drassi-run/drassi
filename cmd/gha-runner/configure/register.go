package configure

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	utilhttp "github.com/dungdm93/drasi/pkg/util/http"
	"github.com/spf13/cobra"
)

type registerOptions struct {
	url   string
	token string
	group string
	name  string
}

type GitHubAuthResult struct {
	TenantUrl   string `json:"url"`
	TokenSchema string `json:"token_schema"`
	Token       string `json:"token"`
	UseV2Flow   bool   `json:"use_v2_flow"`
}

type GHAResponse[T any] struct {
	Count int32 `json:"count"`
	Value []T   `json:"value"`
}

func NewRegisterCommand() *cobra.Command {
	var opts registerOptions

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register new runner to the GitHub Actions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRegister(cmd.Context(), &opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.url, "url", "", "GitHub Actions URL")
	flags.StringVar(&opts.token, "token", "", "Actions registration token")
	flags.StringVar(&opts.group, "group", "", "GitHub Actions RunnerReference group")
	flags.StringVar(&opts.name, "name", "", "GitHub Actions RunnerReference name")

	return cmd
}

func runRegister(ctx context.Context, opts *registerOptions) error {
	if opts.url == "" {
		err := huh.NewInput().
			Title("What is the URL of your repository?").
			Value(&opts.url).
			Placeholder("https://github.com/[ORG]/[REPO]").
			Validate(IsNotEmpty).
			Run()
		if err != nil {
			return err
		} else {
			fmt.Printf("GitHub Actions URL: %s\n", opts.url)
		}
	}
	if opts.token == "" {
		err := huh.NewInput().
			Title("What is your runner register token?").
			Password(true).
			Value(&opts.token).
			Validate(IsNotEmpty).
			Run()
		if err != nil {
			return err
		} else {
			fmt.Printf("Actions registration token: %s\n", opts.token)
		}
	}

	auth, err := retrieveAuthResult(ctx, opts)
	if err != nil {
		return err
	}
	if err = save(".credentials", auth); err != nil {
		return err
	}

	group, err := selectRunnerGroup(ctx, auth, opts)
	if err != nil {
		return err
	}

	if opts.name == "" {
		if hostname, err := os.Hostname(); err == nil {
			opts.name = hostname // default runner name
		}
		err := huh.NewInput().
			Title("Enter the name of runner?").
			Value(&opts.name).
			Validate(IsNotEmpty).
			Run()
		if err != nil {
			return err
		} else {
			fmt.Printf("GitHub Actions Runner name: %s\n", opts.name)
		}
	}

	runners, err := getRunners(ctx, opts.name, auth)
	if err != nil {
		return err
	}
	if len(runners) > 0 {
		// TODO replace runner
		return fmt.Errorf("a runner exists with the same name %s", opts.name)
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	if err = saveRSA(key); err != nil {
		return err
	}

	runner := &Runner{
		RunnerReference: RunnerReference{
			Name:    opts.name,
			Version: "1.2.3",
		},
		MaxParallelism: 10,
		Labels: []RunnerLabel{
			{Name: "self-hosted", Type: RunnerLabelTypeSystem},
			{Name: "Linux", Type: RunnerLabelTypeSystem},
			{Name: "X64", Type: RunnerLabelTypeSystem},
		},
		Authorization: RunnerAuthorization{
			PublicKey: NewRunnerPublicKey(&key.PublicKey),
		},
	}

	runner, err = addRunner(ctx, group.ID, runner, auth)
	if err != nil {
		return err
	}
	if err = save(".runner", runner); err != nil {
		return err
	}

	return nil
}

// IsNotEmpty requires a non-empty string.
func IsNotEmpty(value string) error {
	if value == "" {
		return fmt.Errorf("required value")
	}

	return nil
}

func getApiUrl(u *url.URL) string {
	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Host)
	if host == "github.com" || host == "www.github.com" || host == "github.localhost" ||
		strings.HasSuffix(host, ".ghe.com") ||
		strings.HasSuffix(host, ".ghe.localhost") {
		apiUrl := url.URL{
			Scheme: scheme,
			Host:   "api." + host,
			Path:   "/actions/runner-registration",
		}
		return apiUrl.String()
	} else {
		apiUrl := url.URL{
			Scheme: scheme,
			Host:   host,
			Path:   "/api/v3/actions/runner-registration",
		}
		return apiUrl.String()
	}
}

func retrieveAuthResult(ctx context.Context, opts *registerOptions) (*GitHubAuthResult, error) {
	u, err := url.Parse(opts.url)
	if err != nil {
		return nil, err
	}
	apiUrl := getApiUrl(u)

	// Request body
	data, err := json.Marshal(map[string]string{
		"url":          opts.url,
		"runner_event": "register",
	})
	if err != nil {
		return nil, err
	}

	// Construct request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiUrl, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("RemoteAuth %s", opts.token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("UserAgent", "gha-runner")

	// Make a request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if !utilhttp.IsSuccess(res.StatusCode) {
		return nil, fmt.Errorf("http response code %d", res.StatusCode)
	}

	// Extract the response body
	var result GitHubAuthResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func getRunnerGroups(ctx context.Context, result *GitHubAuthResult) ([]RunnerGroup, error) {
	// Construct request
	endpoint, err := url.JoinPath(result.TenantUrl, "/_apis/distributedtask/pools?poolType=Automation")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", result.Token))

	// Make a request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if !utilhttp.IsSuccess(res.StatusCode) {
		return nil, fmt.Errorf("http response code %d", res.StatusCode)
	}

	// Extract the response body
	var groups GHAResponse[RunnerGroup]
	if err := json.NewDecoder(res.Body).Decode(&groups); err != nil {
		return nil, err
	}
	return groups.Value, nil
}

func selectRunnerGroup(ctx context.Context, result *GitHubAuthResult, opts *registerOptions) (*RunnerGroup, error) {
	groups, err := getRunnerGroups(ctx, result)
	if err != nil {
		return nil, err
	}

	if opts.group != "" {
		for _, g := range groups {
			if g.IsHosted {
				continue
			}
			if g.Name == opts.group {
				return &g, nil
			}
		}
		return nil, fmt.Errorf("could not find any self-hosted runner group named %s", opts.group)
	} else {
		var selectOptions []huh.Option[RunnerGroup]
		for _, g := range groups {
			if g.IsHosted {
				continue
			}
			selectOptions = append(selectOptions, huh.NewOption(g.Name, g))
		}
		var group RunnerGroup
		err = huh.NewSelect[RunnerGroup]().
			Title("Select the runner group to add this runner to?").
			Options(selectOptions...).
			Value(&group).
			Run()
		if err != nil {
			return nil, err
		}
		return &group, nil
	}
}

func getRunners(ctx context.Context, name string, result *GitHubAuthResult) ([]RunnerReference, error) {
	// Construct request
	endpoint, err := url.Parse(result.TenantUrl)
	if err != nil {
		return nil, err
	}
	endpoint = endpoint.JoinPath("/_apis/distributedtask/pools/0/agents") // pools = 0 mean search in all agentPools
	q := endpoint.Query()
	q.Add("agentName", name)
	q.Add("includeCapabilities", "False")
	endpoint.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", result.Token))
	req.Header.Set("Accept", "application/json")

	// Make a request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if !utilhttp.IsSuccess(res.StatusCode) {
		return nil, fmt.Errorf("http response code %d", res.StatusCode)
	}

	// Extract the response body
	var runners GHAResponse[RunnerReference]
	if err := json.NewDecoder(res.Body).Decode(&runners); err != nil {
		return nil, err
	}
	return runners.Value, nil
}

func addRunner(ctx context.Context, groupId int32, runner *Runner, auth *GitHubAuthResult) (*Runner, error) {
	// Construct request
	var data []byte
	data, err := json.Marshal(runner)
	if err != nil {
		return nil, err
	}
	endpoint, err := url.JoinPath(auth.TenantUrl, fmt.Sprintf("/_apis/distributedtask/pools/%d/agents", groupId))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.Token))
	req.Header.Set("Content-Type", "application/json; charset=utf-8; api-version=6.0-preview.2")
	req.Header.Set("Accept", "application/json; api-version=6.0-preview.2")

	// Make a request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if !utilhttp.IsSuccess(res.StatusCode) {
		return nil, fmt.Errorf("http response code %d", res.StatusCode)
	}

	// Extract the response body
	var result Runner
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func save(file string, object any) error {
	if f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, os.ModePerm); err != nil {
		return err
	} else {
		defer f.Close()

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		return enc.Encode(object)
	}
}

func saveRSA(key *rsa.PrivateKey) error {
	pub := x509.MarshalPKCS1PublicKey(&key.PublicKey)
	if f, err := os.OpenFile("rsa.pub", os.O_CREATE|os.O_WRONLY, os.ModePerm); err != nil {
		return err
	} else {
		err = pem.Encode(f, &pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pub,
		})
		if err != nil {
			return err
		}
	}

	pri := x509.MarshalPKCS1PrivateKey(key)
	if f, err := os.OpenFile("rsa", os.O_CREATE|os.O_WRONLY, os.ModePerm); err != nil {
		return err
	} else {
		err = pem.Encode(f, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: pri,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
