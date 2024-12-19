package main

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"

	"demoGolangGitCreation/env"
)

var (
	prAzureToken = env.Var("AZURE_DEVOPS_EXT_PAT").FallsbackTo(env.Var("AZURE_DEVOPS_TOKEN"))
)

// AzureHost Represents an Azure Devops client config and information for the Azure Devops target
type AzureHost struct {
	client   *scm.Client
	host     string
	org      string
	token    string
	username string
	ca_cert  string
}

func main() {
	var fullOrg *url.URL
	var err error
	azureURLString := "https://dev.azure.com/tanzu-scc/catalog_test"

	fullOrg, err = url.Parse(azureURLString)
	if err != nil {
		panic(fmt.Errorf("failed to parse url '%s': %w", azureURLString, err))
	}

	host := &AzureHost{
		token:    prAzureToken.MustResolve(),
		host:     fullOrg.Scheme + "://" + fullOrg.Host,
		org:      strings.TrimPrefix(fullOrg.Path, "/"),
		username: "_token",
	}

	client, err := factory.NewClient("azure", host.host, host.token)
	host.client = client
	if err != nil {
		panic(fmt.Errorf("could not create azure client: '%w'", err))
	}

	timedName := getTimedName()
	repo, err := host.CleanRepository(context.TODO(), timedName)
	if err != nil {
		panic(fmt.Errorf("could not clean azure repo: '%w'", err))
	}

	fmt.Printf("Success! Visit %s/%s\n", host.host, repo.FullName)
}

func getTimedName() string {
	dateTime := time.Now().Format(time.DateTime)

	// regular expression to match punctuation and spaces
	re := regexp.MustCompile(`[[:punct:]\s]`)
	// Replace all matches with underscores
	return re.ReplaceAllString(dateTime, "_")
}

func (h AzureHost) CleanRepository(ctx context.Context, name string) (*scm.Repository, error) {
	repoPath := fmt.Sprintf("%s/%s", h.org, name)
	res, err := h.client.Repositories.Delete(ctx, repoPath)
	if err != nil {
		if res != nil {
			if res.Status == 404 {
				fmt.Printf("repository does not exist, not an error. repository_name: %s\n", name)
			} else {
				return nil, fmt.Errorf("error deleting repository '%s': %w", name, err)
			}
		} else {
			return nil, fmt.Errorf("error deleting repository '%s': %w", name, err)
		}
	}

	var retryErr error
	maxRetries := 2

	var repo *scm.Repository

	// TODO does azure need retries?
	for retries := 0; retries < maxRetries; retries++ {
		repo, retryErr = h.CreateRepository(ctx, name)
		if retryErr == nil {
			break
		}
		fmt.Printf(fmt.Errorf("create repo failed, will retry. retryError: %w, repository_name: %s\n", retryErr.Error(), name).Error())
		time.Sleep(3 * time.Second)
	}
	if retryErr != nil {
		return nil, retryErr
	}

	// this special info log makes it possible to locate the repository created for the test. eg:
	// INFO[0005] created shortened repository logger=outerloop-basic-with-pr.azure-clean-repository scenario= repository_name=catalog-test.rash2.ootb-supply-chain-basic-outer-pr-azure shortened_name=catalog-test.rash2.FfVLfj8oEcu
	fmt.Printf("created shortened repository\n")

	return repo, nil
}

func (h AzureHost) CreateRepository(ctx context.Context, name string) (*scm.Repository, error) {
	ri := &scm.RepositoryInput{
		Namespace:   h.org,
		Name:        name,
		Description: "catalog gitops test repository",
		Private:     true,
	}

	r, _, err := h.client.Repositories.Create(ctx, ri)
	if err != nil {
		return nil, fmt.Errorf("could not create azure repository '%s': %w", name, err)
	}

	return r, nil
}
