package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/smithy-go"
)

// createECRRepository creates an ECR repository if it does not exist.
func createECRRepository(repoName, region string) error {
	// Create a new session using the default credentials and region from the environment.
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)

	// Check if the repository exists
	_, err = ecrClient.DescribeRepositories(
		context.TODO(),
		&ecr.DescribeRepositoriesInput{
			RepositoryNames: []string{repoName},
		},
	)

	if err == nil {
		log.Infof("Repository %s already exists.", repoName)
		return nil
	} else {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if ae.ErrorCode() != "RepositoryNotFoundException" {
				log.Infof("code: %s, message: %s, fault: %s", ae.ErrorCode(), ae.ErrorMessage(), ae.ErrorFault().String())
				return err
			}
		} else {
			return err
		}
	}

	// If the repository does not exist, create it
	_, err = ecrClient.CreateRepository(
		context.TODO(),
		&ecr.CreateRepositoryInput{
			RepositoryName: &repoName,
		},
	)

	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if ae.ErrorCode() != "RepositoryAlreadyExistsException" {
				log.Infof("failed to create repository code: %s, message: %s, fault: %s", ae.ErrorCode(), ae.ErrorMessage(), ae.ErrorFault().String())
				return err
			} else {
				log.Warnf("repository seems to have been created in the mean time, race condition ? code: %s, message: %s, fault: %s", ae.ErrorCode(), ae.ErrorMessage(), ae.ErrorFault().String())
				return nil
			}
		} else {
			return err
		}
	}

	log.Infof("repository %s created successfully", repoName)

	repositoryPolicy := os.Getenv("REPOSITORY_POLICY")
	if repositoryPolicy != "" {
		_, err = ecrClient.SetRepositoryPolicy(context.TODO(),
			&ecr.SetRepositoryPolicyInput{
				PolicyText:     &repositoryPolicy,
				RepositoryName: &repoName,
			},
		)
		if err != nil {
			var oe *smithy.OperationError
			if errors.As(err, &oe) {
				log.Warnf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
			} else {
				log.Warnf("error while applying repository policy: %v", err)
			}
		} else {
			log.Info("repository policy applied")
		}
	} else {
		log.Info("no repository policy provided")
	}

	lifecyclePolicy := os.Getenv("LIFECYCLE_POLICY")
	if lifecyclePolicy != "" {
		_, err = ecrClient.PutLifecyclePolicy(context.TODO(),
			&ecr.PutLifecyclePolicyInput{
				LifecyclePolicyText: &lifecyclePolicy,
				RepositoryName:      &repoName,
			},
		)
		if err != nil {
			var oe *smithy.OperationError
			if errors.As(err, &oe) {
				log.Warnf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
			} else {
				log.Warnf("error while applying lifecycle policy: %v", err)
			}
		} else {
			log.Info("lifecycle policy applied")
		}
	} else {
		log.Info("no lifecycle policy provided")
	}

	return nil
}

// extractRepoName strips the ECR URL to get the repository name.
func extractRepoName(ecrURL string) string {
	for strings.HasSuffix(ecrURL, "/") {
		ecrURL = ecrURL[:len(ecrURL)-1]
	}

	split_url := strings.Split(ecrURL, "/")

	last_part := split_url[len(split_url)-1]
	if strings.Contains(last_part, ":") {
		split_url[len(split_url)-1] = strings.Split(last_part, ":")[0]
	}

	if strings.Contains(split_url[0], ".amazonaws.com") {
		return strings.Join(split_url[1:], "/")
	}

	return strings.Join(split_url, "/")
}

func getVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "couldn't read build info"
	}

	return fmt.Sprintf("%s version %s", bi.Path, bi.Main.Version)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\nUsage: %s [--region region] <repository-url-or-name>\n\nEnvironment variable:\n REPOSITORY_POLICY: repository policy to set upon repository creation\n LIFECYCLE_POLICY: lifecycle policy to set upon repository creation\n\nargs:\n", getVersion(), os.Args[0])
		flag.PrintDefaults()
	}

	region := flag.String("region", "eu-west-1", "AWS region to use")

	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalf("missing repository name")
		flag.PrintDefaults()
	}

	repoInput := flag.Arg(0)

	repoName := extractRepoName(repoInput)

	err := createECRRepository(repoName, *region)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
