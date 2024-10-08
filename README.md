[![Go Report Card](https://goreportcard.com/badge/github.com/nirdosh17/cfn-teardown)](https://goreportcard.com/report/github.com/nirdosh17/cfn-teardown)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/nirdosh17/cfn-teardown/blob/main/LICENSE)
![Latest GitHub Release](https://img.shields.io/github/release/nirdosh17/cfn-teardown)

# CFN Teardown
Cleanup CloudFormation stacks respecting the order of dependencies.

## Features

- Stack name pattern matching for deletion. Finds out dependent/importer/child stacks recursively from a root stack.

- Builds dependency tree for faster teardown. Dependency tree also gives insight on loose/tight coupling of the stacks.

- Multiple safety checks to prevent accidental deletion.

- Supports slack notification for deletion status updates via webhook.

---

### Install
Download binary for the appropriate platform from [HERE](https://github.com/nirdosh17/cfn-teardown/releases).

```bash
✗ wget -q https://github.com/nirdosh17/cfn-teardown/releases/download/v1.0.0/cfn-teardown_Linux_x86_64.tar.gz
✗ tar -xzf cfn-teardown_Linux_x86_64.tar.gz
✗ touch ~/.cfn-teardown.yaml 

✗ ./cfn-teardown version
Using config file:  /Users/nirdosh/.cfn-teardown.yaml
Version:  v1.0.0
```

---
### Usage
Required global flags for all commands: `STACK_PATTERN`, `AWS_REGION`, `AWS_PROFILE`

1. Run `cfn-teardown -h` and see available commands and needed parameters.

2. Listing stack dependencies: `cfn-teardown listDependencies`

	_Generates dependencies in  `stack_teardown_details.json` file (printed in terminal as well)_

2. Tear down stacks: `cfn-teardown deleteStacks`

	_Deletes matching stacks and updates status in the teardown details file as the script is running._

---

### Selecting Stacks For Deletion
**For stacks with consistent naming convention:**

Let's say you have stacks starting with the environment name followed by a hyphen:
- _qa-shared-networks_
- _qa-service-user-management_
- _qa-service-user-search_

In this can, you need to set stack pattern as `^qa-` to match stacks starting with `qa-`.

**For stacks which do not follow any naming pattern:**

Example:
- _qa-shared-networks_
- _service-user-management_ (depends on shared networks stack)
- _user-search-service_ (depends on shared networks stack)

Use the root stack's name as the stack pattern i.e. `^qa-shared-networks`. The script will find out all dependendent stacks from the root stack **recursively** until the leaf nodes have zero importer stacks.

---
### Configuration

Configuration for this command can be set in three different ways in the precedence order defined below:
1. Environment variables(same as flag name)
2. Flags e.g. `cfn-teardown deleteStacks --STACK_PATTERN=qaenv-`
3. Supplied YAML Config file (default: ~/.cfn-teardown.yaml)
    <details>
    <summary><b>Minimal config file</b></summary>

    ```yaml
    AWS_REGION: us-east-1
    AWS_PROFILE: staging
    STACK_PATTERN: qa-
    ```
    </details>
    <details>
    <summary><b>All configs present</b></summary>

    ```yaml
    AWS_REGION: us-east-1
    AWS_PROFILE: staging
    TARGET_ACCOUNT_ID: 121212121212
    STACK_PATTERN: qa-
    ABORT_WAIT_TIME_MINUTES: 20
    STACK_WAIT_TIME_SECONDS: 30
    MAX_DELETE_RETRY_COUNT: 5
    SLACK_WEBHOOK_URL: https://hooks.slack.com/services/dummy/dummy/long_hash
    ROLE_ARN: "<arn>"
    DRY_RUN: "false"
    ```
    </details>

See available configurations via: `cfn-teardown <command> --help`

---
### Stack Teardown Strategy

1. Find matching stacks based on the regex provided

2. Prepare stack dependencies
    <details>
    <summary><b>It looks something like this:</b></summary>

      ```json
      {
        "staging-bucket-archived-items": {
          "StackName": "staging-bucket-archived-items",
          "Status": "CREATE_COMPLETE",
          "StackStatusReason": "",
          "DeleteStartedAt": "2021-02-07T03:35:43Z",
          "DeleteCompletedAt": "",
          "DeletionTimeInMinutes": "",
          "DeleteAttempt": 0,
          "Exports": [
            "staging:ItemsArchiveBucket",
            "staging:ItemsArchiveBucketArn"
          ],
          "ActiveImporterStacks": {
            "staging-products-service": {}
          },
          "CFNConsoleLink": "https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/stackinfo?stackId=staging-bucket-archived-items"
        },
        "staging-products-service": {
          "StackName": "staging-products-service",
          "Status": "CREATE_COMPLETE",
          "StackStatusReason": "",
          "DeleteStartedAt": "2021-02-07T03:30:54Z",
          "DeleteCompletedAt": "",
          "DeletionTimeInMinutes": "",
          "DeleteAttempt": 0,
          "Exports": [
            "staging:ProductsServiceEndpoint"
          ],
          "ActiveImporterStacks": {},
          "CFNConsoleLink": "https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/stackinfo?stackId=staging-products-service"
        }
      }
      ```
    </details>

3. Alert slack channel(if provided) and waits before initiating deletion. Starts deletion immediately if no wait time is provided.

4. Select stacks which are eligible for deletion. A stack is eligible for deletion if it's exports are imported by no other stacks. In simple terms, it should have no dependencies.

5. Send delete requests for all selected stacks.

6. Wait for 30 seconds(configurable) before scanning eligible stacks again. Checks If the stack has been already deleted and if deleted updates stack status in the dependency tree.

7. This process (sending delete requests, waiting, checking stack status) is repeated until all stacks have status `DELETE_COMPLETE`.

8. If a stack is not deleted even after exhausting all retries(default 5), teardown is halted and manual intervention is requested.

---

### AWS Credentials
Only AWS profile based authentication supported at the moment. By default, it tries to use the IAM role of the caller but we can also supply role arn if we want the script to assume a different role.

---

### Safety Flags

- `DRY_RUN` flag must be explicitely set to `false` to activate delete functionality

- `ABORT_WAIT_TIME_MINUTES` flag lets us to decide how much to wait before initiating delete as you might want to confirm the stacks that are about to get deleted

- `TARGET_ACCOUNT_ID`: If provided, this flag confirms that the given aws account id matches with account id in the aws session during runtime to make sure that we are deleting stacks in the desired aws account

---

### Limitation
If a stack can't be deleted from the AWS Console itself due to some dependencies or some error, then it won't be deleted by this tool as well. In such case, manual intervention is required.

---

### Demo
> <details><summary><strong>Deleting Stacks</strong></summary>
> <img src="https://user-images.githubusercontent.com/5920689/130366139-30912d09-7d79-4537-8809-014c75ce38c0.gif" width="600" alt="deleting stacks" />

> <details><summary><strong>Slack Notifications</strong></summary>
> <img src="https://user-images.githubusercontent.com/5920689/130365254-dd2d911d-803b-4c02-93ec-2f78badedb6a.png" width="600" alt="slack notifications sample" />


---
### Caution :warning:
_With great power, comes great responsibility_
- First try within small number of test stacks in dry run mode.
- Use redundant safety flags `DRY_RUN`, `TARGET_ACCOUNT_ID` and `ABORT_WAIT_TIME_MINUTES`.
