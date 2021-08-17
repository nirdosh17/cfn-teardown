/*
Copyright © 2021 Nirdosh Gautam

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts"
)

type S3Manager struct {
	TargetAccountId string
	NukeRoleARN     string
	AWSProfile      string
	AWSRegion       string
}

func (sm S3Manager) EmptyBucket(bucketName string) error {
	svc, err := sm.Session()
	if err != nil {
		return err
	}

	fmt.Printf("Emptying bucket '%v'...\n", bucketName)

	// Setup BatchDeleteIterator to iterate through a list of objects
	iterator := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{Bucket: aws.String(bucketName)})
	err = s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iterator)
	if err != nil {
		fmt.Printf("Unable to delete objects from bucket '%v': %v\n", bucketName, err)
		return err
	}

	// check if the bucket is empty
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	if err != nil {
		return fmt.Errorf("Error listing objects from bucket '%v': %v", bucketName, err)
	}

	if len(resp.Contents) != 0 {
		return fmt.Errorf("Failed to empty bucket. Number of items left: %v", len(resp.Contents))
	}

	fmt.Printf("Bucket '%v' emptied successfully\n", bucketName)

	return nil
}

// assumes staging nuke role
func (sm S3Manager) Session() (*s3.S3, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(sm.AWSRegion)},
		SharedConfigState: session.SharedConfigEnable,
		Profile:           sm.AWSProfile,
	}))

	desiredAccount, err := sm.IsDesiredAWSAccount(sess)
	if err != nil {
		return nil, err
	}

	// to make things easy while running this script locally
	if desiredAccount {
		return s3.New(sess), err
	} else {
		// Create the credentials from AssumeRoleProvider if nuke role arn is provided
		creds := stscreds.NewCredentials(sess, sm.NukeRoleARN)
		// Create service client value configured for credentials from assumed role
		return s3.New(sess, &aws.Config{Credentials: creds, MaxRetries: &AWS_SDK_MAX_RETRY}), err
	}
}

func (sm S3Manager) IsDesiredAWSAccount(sess *session.Session) (bool, error) {
	svc := sts.New(sess)
	result, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		fmt.Printf("Error requesting AWS caller identity: %v", err.Error())
		return false, err
	}

	if *result.Account == sm.TargetAccountId {
		return true, err
	}
	return false, err
}
