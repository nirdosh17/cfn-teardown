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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/nirdosh17/cfn-teardown/models"
)

type NotificationManager struct {
	StackPattern           string
	NotificationWebHookURL string // Webhook url is specific to channel
}

type AlertMessage struct {
	Message     string // Long message with details about the event
	Event       string // Start | Complete | Error
	FailedStack models.StackDetails
	Attachment  map[string]interface{}
}

// building slack messages: https://app.slack.com/block-kit-builder
type SlackMessage struct {
	Attachments []map[string]interface{} `json:"attachments"`
}

var ColorMapping map[string]string = map[string]string{"Start": "#f0e62e", "Complete": "#25db2e", "Error": "#e81e1e"}

func (nm NotificationManager) StartAlert(am AlertMessage) {
	am.Event = "Start"
	am.Attachment = map[string]interface{}{
		"color": ColorMapping[am.Event],
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": "Stack Deletion Started",
				},
			},
			{
				"type": "context",
				"elements": []map[string]string{
					{
						"type": "mrkdwn",
						"text": am.Message,
					},
				},
			},
			{
				"type": "divider",
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{
						"type": "mrkdwn",
						"text": ("*Stack Pattern* \n " + nm.StackPattern),
					},
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("*Stack Count* \n %v", TOTAL_STACK_COUNT),
					},
				},
			},
		},
	}
	nm.Alert(am)
}

func (nm NotificationManager) ErrorAlert(am AlertMessage) {
	am.Event = "Error"
	am.Attachment = map[string]interface{}{
		"color": ColorMapping[am.Event],
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": "Stack Deletion Failed",
				},
			},
			{
				"type": "context",
				"elements": []map[string]string{
					{
						"type": "mrkdwn",
						"text": "Manual Intervention Required",
					},
				},
			},
			{
				"type": "divider",
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{
						"type": "mrkdwn",
						"text": ("*Stack Pattern* \n " + nm.StackPattern),
					},
					{
						"type": "mrkdwn",
						"text": "*Runtime* \n" + fmt.Sprintf("%.2f Hour/s", NUKE_DURATION_IN_HRS),
					},
				},
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{
						"type": "mrkdwn",
						"text": "*Stacks Deleted* \n" + fmt.Sprintf("%v/%v", DELETED_STACK_COUNT, TOTAL_STACK_COUNT),
					},
					{
						"type": "mrkdwn",
						"text": "*Failed Stack* \n" + fmt.Sprintf("<%v|%v>", am.FailedStack.CFNConsoleLink, am.FailedStack.StackName),
					},
				},
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{
						"type": "mrkdwn",
						"text": "*Reason* \n" + am.FailedStack.StackStatusReason,
					},
				},
			},
		},
	}
	nm.Alert(am)
}

func (nm NotificationManager) StuckAlert(am AlertMessage) {
	am.Event = "Error"
	am.Attachment = map[string]interface{}{
		"color": ColorMapping[am.Event],
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": "Stack Deletion Stuck",
				},
			},
			{
				"type": "context",
				"elements": []map[string]string{
					{
						"type": "mrkdwn",
						"text": am.Message,
					},
				},
			},
			{
				"type": "divider",
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{
						"type": "mrkdwn",
						"text": ("*Stack Pattern* \n " + nm.StackPattern),
					},
					{
						"type": "mrkdwn",
						"text": "*Runtime* \n" + fmt.Sprintf("%.2f Hour/s", NUKE_DURATION_IN_HRS),
					},
				},
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{
						"type": "mrkdwn",
						"text": "*Stacks Deleted* \n" + fmt.Sprintf("%v/%v", DELETED_STACK_COUNT, TOTAL_STACK_COUNT),
					},
				},
			},
		},
	}
	nm.Alert(am)
}

func (nm NotificationManager) SuccessAlert(am AlertMessage) {
	am.Event = "Complete"
	am.Attachment = map[string]interface{}{
		"color": ColorMapping[am.Event],
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": "Stack Deletion Completed",
				},
			},
			{
				"type": "divider",
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{
						"type": "mrkdwn",
						"text": ("*Stack Pattern* \n " + nm.StackPattern),
					},
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("*Deleted Stacks* \n %v", TOTAL_STACK_COUNT),
					},
				},
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{
						"type": "mrkdwn",
						"text": ("*Started At* \n " + NUKE_START_TIME),
					},
					{
						"type": "mrkdwn",
						"text": ("*Completed At* \n " + NUKE_END_TIME + fmt.Sprintf("(%.2f Hour/s)", NUKE_DURATION_IN_HRS)),
					},
				},
			},
		},
	}
	nm.Alert(am)
}

func (nm NotificationManager) GenericAlert(am AlertMessage) {
	am.Event = "Error"
	am.Attachment = map[string]interface{}{
		"color": ColorMapping[am.Event],
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": "Stack Deletion Error",
				},
			},
			{
				"type": "context",
				"elements": []map[string]string{
					{
						"type": "mrkdwn",
						"text": am.Message,
					},
				},
			},
		},
	}
	nm.Alert(am)
}

func (nm NotificationManager) Alert(am AlertMessage) error {
	msgBody := SlackMessage{
		Attachments: []map[string]interface{}{am.Attachment},
	}

	postBody, err := json.Marshal(msgBody)
	if err != nil {
		log.Printf("[Alert] Error marshaling request body: %v", err)
		return err
	}

	resp, err := http.Post(nm.NotificationWebHookURL, "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		log.Printf("Error posting message to Slack: %v", err)
		return err
	}
	defer resp.Body.Close()

	//Read the response body
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Printf("Got %v status code from Slack, Response body: %v\n", resp.StatusCode, string(body))
		log.Printf("Request body: %v\n", string(postBody))
		return fmt.Errorf("Failed to publish message %+v to Slack", msgBody)
	}

	return nil
}
