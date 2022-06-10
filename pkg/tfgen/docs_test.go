// Copyright 2016-2018, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// nolint: lll
package tfgen

import (
	"bytes"
	"sort"
	"strings"
	"testing"
	"text/template"

	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
	"github.com/stretchr/testify/assert"
)

type testcase struct {
	Input    string
	Expected string
}

func TestURLRewrite(t *testing.T) {
	tests := []testcase{
		{
			Input:    "The DNS name for the given subnet/AZ per [documented convention](http://docs.aws.amazon.com/efs/latest/ug/mounting-fs-mount-cmd-dns-name.html).", // nolint: lll
			Expected: "The DNS name for the given subnet/AZ per [documented convention](http://docs.aws.amazon.com/efs/latest/ug/mounting-fs-mount-cmd-dns-name.html).", // nolint: lll
		},
		{
			Input:    "It's recommended to specify `create_before_destroy = true` in a [lifecycle][1] block to replace a certificate which is currently in use (eg, by [`aws_lb_listener`](lb_listener.html)).", // nolint: lll
			Expected: "It's recommended to specify `createBeforeDestroy = true` in a [lifecycle][1] block to replace a certificate which is currently in use (eg, by `awsLbListener`).",                         // nolint: lll
		},
		{
			Input:    "The execution ARN to be used in [`lambda_permission`](/docs/providers/aws/r/lambda_permission.html)'s `source_arn`",                       // nolint: lll
			Expected: "The execution ARN to be used in [`lambdaPermission`](https://www.terraform.io/docs/providers/aws/r/lambda_permission.html)'s `sourceArn`", // nolint: lll
		},
		{
			Input:    "See google_container_node_pool for schema.",
			Expected: "See google.container.NodePool for schema.",
		},
	}

	g, err := NewGenerator(GeneratorOptions{
		Package:  "google",
		Version:  "0.1.2",
		Language: "nodejs",
		ProviderInfo: tfbridge.ProviderInfo{
			Name: "google",
			Resources: map[string]*tfbridge.ResourceInfo{
				"google_container_node_pool": {Tok: "google:container/nodePool:NodePool"},
			},
		},
	})
	assert.NoError(t, err)

	for _, test := range tests {
		text, _ := reformatText(g, test.Input, nil)
		assert.Equal(t, test.Expected, text)
	}
}

func TestParseArgReferenceSection(t *testing.T) {
	tests := []struct {
		input    []string
		expected map[string]*argumentDocs
	}{
		{
			input: []string{
				"* `iam_instance_profile` - (Optional) The IAM Instance Profile to",
				"launch the instance with. Specified as the name of the Instance Profile. Ensure your credentials have the correct permission to assign the instance profile according to the [EC2 documentation](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2.html#roles-usingrole-ec2instance-permissions), notably `iam:PassRole`.",
				"* `ipv6_address_count`- (Optional) A number of IPv6 addresses to associate with the primary network interface. Amazon EC2 chooses the IPv6 addresses from the range of your subnet.",
				"* `ipv6_addresses` - (Optional) Specify one or more IPv6 addresses from the range of the subnet to associate with the primary network interface",
				"* `tags` - (Optional) A mapping of tags to assign to the resource.",
			},
			expected: map[string]*argumentDocs{
				"iam_instance_profile": {
					description: "The IAM Instance Profile to" + "\n" +
						"launch the instance with. Specified as the name of the Instance Profile. Ensure your credentials have the correct permission to assign the instance profile according to the [EC2 documentation](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2.html#roles-usingrole-ec2instance-permissions), notably `iam:PassRole`.",
					arguments: map[string]*argumentDocs{},
				},
				"ipv6_address_count": {
					description: "A number of IPv6 addresses to associate with the primary network interface. Amazon EC2 chooses the IPv6 addresses from the range of your subnet.",
					arguments:   map[string]*argumentDocs{},
				},
				"ipv6_addresses": {
					description: "Specify one or more IPv6 addresses from the range of the subnet to associate with the primary network interface",
					arguments:   map[string]*argumentDocs{},
				},
				"tags": {
					description: "A mapping of tags to assign to the resource.",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
		{
			input: []string{
				"* `jwt_configuration` - (Optional) The configuration of a JWT authorizer. Required for the `JWT` authorizer type.",
				"Supported only for HTTP APIs.",
				"",
				"The `jwt_configuration` object supports the following:",
				"",
				"* `audience` - (Optional) A list of the intended recipients of the JWT. A valid JWT must provide an aud that matches at least one entry in this list.",
				"* `issuer` - (Optional) The base domain of the identity provider that issues JSON Web Tokens, such as the `endpoint` attribute of the [`aws_cognito_user_pool`](/docs/providers/aws/r/cognito_user_pool.html) resource.",
			},
			expected: map[string]*argumentDocs{
				"jwt_configuration": {
					description: "The configuration of a JWT authorizer. Required for the `JWT` authorizer type." + "\n" +
						"Supported only for HTTP APIs.",
					arguments: map[string]*argumentDocs{
						"audience": {
							description: "A list of the intended recipients of the JWT. A valid JWT must provide an aud that matches at least one entry in this list.",
							arguments:   map[string]*argumentDocs{},
						},
						"issuer": {
							description: "The base domain of the identity provider that issues JSON Web Tokens, such as the `endpoint` attribute of the [`aws_cognito_user_pool`](/docs/providers/aws/r/cognito_user_pool.html) resource.",
							arguments:   map[string]*argumentDocs{},
						},
					},
				},
			},
		},
		{
			input: []string{
				"* `website` - (Optional) A website object (documented below).",
				"~> **NOTE:** You cannot use `acceleration_status` in `cn-north-1` or `us-gov-west-1`",
				"",
				"The `website` object supports the following:",
				"",
				"* `index_document` - (Required, unless using `redirect_all_requests_to`) Amazon S3 returns this index document when requests are made to the root domain or any of the subfolders.",
				"* `routing_rules` - (Optional) A json array containing [routing rules](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-websiteconfiguration-routingrules.html)",
				"describing redirect behavior and when redirects are applied.",
			},
			expected: map[string]*argumentDocs{
				"website": {
					description: "A website object (documented below)." + "\n" +
						"~> **NOTE:** You cannot use `acceleration_status` in `cn-north-1` or `us-gov-west-1`",
					arguments: map[string]*argumentDocs{
						"index_document": {
							description: "Amazon S3 returns this index document when requests are made to the root domain or any of the subfolders.",
							arguments:   map[string]*argumentDocs{},
						},
						"routing_rules": {
							description: "A json array containing [routing rules](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-websiteconfiguration-routingrules.html)" + "\n" +
								"describing redirect behavior and when redirects are applied.",
							arguments: map[string]*argumentDocs{},
						},
					},
				},
			},
		},
		{
			input: []string{
				"* `action` - (Optional) The action that CloudFront or AWS WAF takes when a web request matches the conditions in the rule. Not used if `type` is `GROUP`.",
				"  * `type` - (Required) valid values are: `BLOCK`, `ALLOW`, or `COUNT`",
				"* `override_action` - (Optional) Override the action that a group requests CloudFront or AWS WAF takes when a web request matches the conditions in the rule. Only used if `type` is `GROUP`.",
				"  * `type` - (Required) valid values are: `BLOCK`, `ALLOW`, or `COUNT`",
			},
			// Note: This is the existing behavior and is indeed a bug. The type field should be nested within action and override_action.
			expected: map[string]*argumentDocs{
				"action": {
					description: "The action that CloudFront or AWS WAF takes when a web request matches the conditions in the rule. Not used if `type` is `GROUP`.",
					arguments:   map[string]*argumentDocs{},
				},
				"override_action": {
					description: "Override the action that a group requests CloudFront or AWS WAF takes when a web request matches the conditions in the rule. Only used if `type` is `GROUP`.",
					arguments:   map[string]*argumentDocs{},
				},
				"type": {
					description: "valid values are: `BLOCK`, `ALLOW`, or `COUNT`",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
		{
			input: []string{
				"* `priority` - (Optional) The priority associated with the rule.",
				"",
				"* `priority` is optional (with a default value of `0`) but must be unique between multiple rules",
			},
			expected: map[string]*argumentDocs{
				"priority": {
					description: "is optional (with a default value of `0`) but must be unique between multiple rules",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
		{
			input: []string{
				"* `allowed_audiences` (Optional) Allowed audience values to consider when validating JWTs issued by Azure Active Directory.",
				"* `retention_policy` - (Required) A `retention_policy` block as documented below.",
				"",
				"---",
				"* `retention_policy` supports the following:",
			},
			expected: map[string]*argumentDocs{
				"retention_policy": {
					description: "A `retention_policy` block as documented below.",
					arguments:   map[string]*argumentDocs{},
				},
				"allowed_audiences": {
					description: "Allowed audience values to consider when validating JWTs issued by Azure Active Directory.",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
		{
			input: []string{
				"* `launch_template_config` - (Optional) Launch template configuration block. See [Launch Template Configs](#launch-template-configs) below for more details. Conflicts with `launch_specification`. At least one of `launch_specification` or `launch_template_config` is required.",
				"* `spot_maintenance_strategies` - (Optional) Nested argument containing maintenance strategies for managing your Spot Instances that are at an elevated risk of being interrupted. Defined below.",
				"* `spot_price` - (Optional; Default: On-demand price) The maximum bid price per unit hour.",
				"* `wait_for_fulfillment` - (Optional; Default: false) If set, Terraform will",
				"  wait for the Spot Request to be fulfilled, and will throw an error if the",
				"  timeout of 10m is reached.",
				"* `target_capacity` - The number of units to request. You can choose to set the",
				"  target capacity in terms of instances or a performance characteristic that is",
				"  important to your application workload, such as vCPUs, memory, or I/O.",
				"* `allocation_strategy` - Indicates how to allocate the target capacity across",
				"  the Spot pools specified by the Spot fleet request. The default is",
				"  `lowestPrice`.",
				"* `instance_pools_to_use_count` - (Optional; Default: 1)",
				"  The number of Spot pools across which to allocate your target Spot capacity.",
				"  Valid only when `allocation_strategy` is set to `lowestPrice`. Spot Fleet selects",
				"  the cheapest Spot pools and evenly allocates your target Spot capacity across",
				"  the number of Spot pools that you specify.",
			},
			expected: map[string]*argumentDocs{
				"launch_template_config": {
					description: "Launch template configuration block. See [Launch Template Configs](#launch-template-configs) below for more details. Conflicts with `launch_specification`. At least one of `launch_specification` or `launch_template_config` is required.",
					arguments:   map[string]*argumentDocs{},
				},
				"spot_maintenance_strategies": {
					description: "Nested argument containing maintenance strategies for managing your Spot Instances that are at an elevated risk of being interrupted. Defined below.",
					arguments:   map[string]*argumentDocs{},
				},
				"spot_price": {
					description: "The maximum bid price per unit hour.",
					arguments:   map[string]*argumentDocs{},
				},
				"wait_for_fulfillment": {
					description: "If set, Terraform will\nwait for the Spot Request to be fulfilled, and will throw an error if the\ntimeout of 10m is reached.",
					arguments:   map[string]*argumentDocs{},
				},
				"target_capacity": {
					description: "The number of units to request. You can choose to set the\ntarget capacity in terms of instances or a performance characteristic that is\nimportant to your application workload, such as vCPUs, memory, or I/O.",
					arguments:   map[string]*argumentDocs{},
				},
				"allocation_strategy": {
					description: "Indicates how to allocate the target capacity across\nthe Spot pools specified by the Spot fleet request. The default is\n`lowestPrice`.",
					arguments:   map[string]*argumentDocs{},
				},
				"instance_pools_to_use_count": {
					description: "\nThe number of Spot pools across which to allocate your target Spot capacity.\nValid only when `allocation_strategy` is set to `lowestPrice`. Spot Fleet selects\nthe cheapest Spot pools and evenly allocates your target Spot capacity across\nthe number of Spot pools that you specify.",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
	}

	for _, tt := range tests {
		parser := &tfMarkdownParser{
			ret: entityDocs{
				Arguments: make(map[string]*argumentDocs),
			},
		}
		parser.parseArgReferenceSection(tt.input, "", "")

		assert.Equal(t, tt.expected, parser.ret.Arguments)
	}
}

func TestParseArgReferenceSection_WithParentArg(t *testing.T) {
	input := []string{
		"### dead_letter_config",
		"",
		"blah blah",
		"",
		"* `target_arn` - (Required) target_arn_desc",
	}

	parser := &tfMarkdownParser{
		ret: entityDocs{
			Arguments: map[string]*argumentDocs{
				"dead_letter_config": {
					description: "This was previously parsed.",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
	}

	expected := map[string]*argumentDocs{
		"dead_letter_config": {
			description: "This was previously parsed.",
			arguments: map[string]*argumentDocs{
				"target_arn": {
					description: "target_arn_desc",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
	}

	parser.parseArgReferenceSection(input, "dead_letter_config", "")

	assert.Equal(t, expected, parser.ret.Arguments)
}

func TestParseArgReferenceSection_NestedArgumentsSameName(t *testing.T) {
	input := []string{
		"## Arguments Reference",
		"",
		"The following arguments are supported:",
		"* `param_1` - (Optional) param_1_desc",
		"",
		// "nested_block_1" will not match the regex because of the digit:
		"* `nested_block_one` - (Optional) nested_block_1_desc",
		"",
		"* `nested_block_two` - (Optional) nested_block_2_desc",
		"",
		"The optional `nested_block_one` subblock supports:",
		"",
		"* `nested_param` - (Required) nested_block_1.nested_param_desc",
		"",
		"The optional `nested_block_two` subblock supports:",
		"",
		"* `nested_param` - (Required) nested_block_2.nested_param_desc",
		"",
	}

	parser := &tfMarkdownParser{
		ret: entityDocs{
			Arguments: make(map[string]*argumentDocs),
		},
	}

	expected := map[string]*argumentDocs{
		"param_1": {
			description: "param_1_desc",
			arguments:   map[string]*argumentDocs{},
		},
		"nested_block_one": {
			description: "nested_block_1_desc",
			arguments: map[string]*argumentDocs{
				"nested_param": {
					description: "nested_block_1.nested_param_desc",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
		"nested_block_two": {
			description: "nested_block_2_desc",
			arguments: map[string]*argumentDocs{
				"nested_param": {
					description: "nested_block_2.nested_param_desc",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
	}

	parser.parseArgReferenceSection(input, "", "")

	assert.Equal(t, expected, parser.ret.Arguments)
}

func TestParseArgReferenceSection_NestedSubsectionWithSameNameAsArg(t *testing.T) {
	// Tests parsing of sections like https://github.com/hashicorp/terraform-provider-aws/blob/471ca4e25a732b0c0d566dbc645ea712b48e1a56/website/docs/r/lambda_function.html.markdown#dead_letter_config

	parser := &tfMarkdownParser{
		ret: entityDocs{
			Arguments: map[string]*argumentDocs{
				"dead_letter_config": {
					description: "dead_letter_config_desc",
					arguments:   make(map[string]*argumentDocs),
				},
			},
		},
	}

	input := []string{
		"### dead_letter_config",
		"",
		"Dead letter queue configuration that specifies the queue or topic where Lambda sends asynchronous events when they fail processing. For more information, see [Dead Letter Queues](https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html#dlq).",
		"",
		"",
		"* `target_arn` - (Required) target_arn_desc",
	}

	expected := map[string]*argumentDocs{
		"dead_letter_config": {
			description: "dead_letter_config_desc",
			arguments: map[string]*argumentDocs{
				"target_arn": {
					description: "target_arn_desc",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
	}

	parser.parseArgReferenceSection(input, "dead_letter_config", "")

	assert.Equal(t, expected, parser.ret.Arguments)
}

func TestParseArgReferenceSection_DoubleNestedSubsectionWithArgNameAndConfigurationBlock(t *testing.T) {
	// Integration test for parsing of docs like https://github.com/hashicorp/terraform-provider-aws/blob/main/website/docs/r/mskconnect_connector.html.markdown#basic-configuration

	input1 := []string{
		"## Argument Reference",
		"",
		"The following arguments are supported:",
		"",
		"* `capacity` - (Required) Information about the capacity allocated to the connector. See below.",
	}

	input2 := []string{
		"### capacity Configuration Block",
		"",
		"* `autoscaling` - (Optional) Information about the auto scaling parameters for the connector. See below.",
	}

	input3 := []string{
		"### autoscaling Configuration Block",
		"",
		"",
		"* `max_worker_count` - (Required) The maximum number of workers allocated to the connector.",
	}

	expected := map[string]*argumentDocs{
		"capacity": {
			description: "Information about the capacity allocated to the connector. See below.",
			arguments: map[string]*argumentDocs{
				"autoscaling": {
					description: "Information about the auto scaling parameters for the connector. See below.",
					arguments: map[string]*argumentDocs{
						"max_worker_count": {
							description: "The maximum number of workers allocated to the connector.",
							arguments:   map[string]*argumentDocs{},
						},
					},
				},
			},
		},
	}

	parser := &tfMarkdownParser{
		ret: entityDocs{
			Arguments: map[string]*argumentDocs{},
		},
	}

	// This emulates the behavior of parseSection(), assuming that parseArgNameFromHeader() and getMatchingArgNames() are doing their jobs correctly:
	parser.parseArgReferenceSection(input1, "", "")
	parser.parseArgReferenceSection(input2, "capacity", "")
	parser.parseArgReferenceSection(input3, "capacity.autoscaling", "")

	assert.Equal(t, expected, parser.ret.Arguments)
}

func TestParseArgReferenceSection_NonFullyQualifiedSubBlocks(t *testing.T) {
	// Sometimes sub-blocks are not fully-qualified, e.g. "conditions" in https://registry.terraform.io/providers/hashicorp/google-beta/latest/docs/resources/access_context_manager_access_level#argument-reference

	input := []string{
		"## Argument Reference",
		"",
		"The following arguments are supported:",
		"",
		"* `basic` - (Optional) basic_desc",
		"",
		"<a name=\"nested_basic\"></a>The `basic` block supports:",
		"",
		"* `conditions` - (Required) conditions_desc",
		"",
		"<a name=\"nested_conditions\"></a>The `conditions` block supports:",
		"",
		"* `device_policy` - (Optional) device_policy_desc",
	}

	expected := map[string]*argumentDocs{
		"basic": {
			description: "basic_desc",
			arguments: map[string]*argumentDocs{
				"conditions": {
					description: "conditions_desc",
					arguments: map[string]*argumentDocs{
						"device_policy": {
							description: "device_policy_desc",
							arguments:   map[string]*argumentDocs{},
						},
					},
				},
			},
		},
	}

	parser := &tfMarkdownParser{
		ret: entityDocs{
			Arguments: map[string]*argumentDocs{},
		},
	}

	parser.parseArgReferenceSection(input, "", "")

	assert.Equal(t, expected, parser.ret.Arguments)
}

func TestGetFooterLinks(t *testing.T) {
	input := `## Attributes Reference

For **environment** the following attributes are supported:

[1]: https://docs.aws.amazon.com/lambda/latest/dg/welcome.html
[2]: https://docs.aws.amazon.com/lambda/latest/dg/walkthrough-s3-events-adminuser-create-test-function-create-function.html
[3]: https://docs.aws.amazon.com/lambda/latest/dg/walkthrough-custom-events-create-test-function.html`

	expected := map[string]string{
		"[1]": "https://docs.aws.amazon.com/lambda/latest/dg/welcome.html",
		"[2]": "https://docs.aws.amazon.com/lambda/latest/dg/walkthrough-s3-events-adminuser-create-test-function-create-function.html",
		"[3]": "https://docs.aws.amazon.com/lambda/latest/dg/walkthrough-custom-events-create-test-function.html",
	}

	actual := getFooterLinks(input)

	assert.Equal(t, expected, actual)
}

func TestReplaceFooterLinks(t *testing.T) {
	inputText := `# Resource: aws_lambda_function

	Provides a Lambda Function resource. Lambda allows you to trigger execution of code in response to events in AWS, enabling serverless backend solutions. The Lambda Function itself includes source code and runtime configuration.

	For information about Lambda and how to use it, see [What is AWS Lambda?][1]
	* (Required) The function [entrypoint][3] in your code.`
	footerLinks := map[string]string{
		"[1]": "https://docs.aws.amazon.com/lambda/latest/dg/welcome.html",
		"[2]": "https://docs.aws.amazon.com/lambda/latest/dg/walkthrough-s3-events-adminuser-create-test-function-create-function.html",
		"[3]": "https://docs.aws.amazon.com/lambda/latest/dg/walkthrough-custom-events-create-test-function.html",
	}

	expected := `# Resource: aws_lambda_function

	Provides a Lambda Function resource. Lambda allows you to trigger execution of code in response to events in AWS, enabling serverless backend solutions. The Lambda Function itself includes source code and runtime configuration.

	For information about Lambda and how to use it, see [What is AWS Lambda?](https://docs.aws.amazon.com/lambda/latest/dg/welcome.html)
	* (Required) The function [entrypoint](https://docs.aws.amazon.com/lambda/latest/dg/walkthrough-custom-events-create-test-function.html) in your code.`
	actual := replaceFooterLinks(inputText, footerLinks)
	assert.Equal(t, expected, actual)

	// Test when there are no footer link.
	actual = replaceFooterLinks(inputText, nil)
	assert.Equal(t, inputText, actual)
}

func TestFixExamplesHeaders(t *testing.T) {
	codeFence := "```"
	t.Run("WithCodeFences", func(t *testing.T) {
		markdown := `
# digitalocean\_cdn

Provides a DigitalOcean CDN Endpoint resource for use with Spaces.

## Example Usage

#### Basic Example

` + codeFence + `typescript
// Some code.
` + codeFence + `
## Argument Reference`

		var processedMarkdown string
		groups := splitGroupLines(markdown, "## ")
		for _, lines := range groups {
			fixExampleTitles(lines)
			for _, line := range lines {
				processedMarkdown += line
			}
		}

		assert.NotContains(t, processedMarkdown, "#### Basic Example")
		assert.Contains(t, processedMarkdown, "### Basic Example")
	})

	t.Run("WithoutCodeFences", func(t *testing.T) {
		markdown := `
# digitalocean\_cdn

Provides a DigitalOcean CDN Endpoint resource for use with Spaces.

## Example Usage

#### Basic Example

Misleading example title without any actual code fences. We should not modify the title.

## Argument Reference`

		var processedMarkdown string
		groups := splitGroupLines(markdown, "## ")
		for _, lines := range groups {
			fixExampleTitles(lines)
			for _, line := range lines {
				processedMarkdown += line
			}
		}

		assert.Contains(t, processedMarkdown, "#### Basic Example")
	})
}

func TestExtractExamples(t *testing.T) {
	basic := `Previews a CIDR from an IPAM address pool. Only works for private IPv4.

~> **NOTE:** This functionality is also encapsulated in a resource sharing the same name. The data source can be used when you need to use the cidr in a calculation of the same Root module, count for example. However, once a cidr range has been allocated that was previewed, the next refresh will find a **new** cidr and may force new resources downstream. Make sure to use Terraform's lifecycle ignore_changes policy if this is undesirable.

## Example Usage
Basic usage:`
	assert.Equal(t, "## Example Usage\nBasic usage:", extractExamples(basic))

	noExampleUsages := `Something mentioning Terraform`
	assert.Equal(t, "", extractExamples(noExampleUsages))

	// This use case is not known to exist in the wild, but we want to make sure our handling here is conservative given that there's no strictly defined schema to TF docs.
	multipleExampleUsages := `Something mentioning Terraform

	## Example Usage
	Some use case

	## Example Usage
	Some other use case
`
	assert.Equal(t, "", extractExamples(multipleExampleUsages))
}

func TestReformatExamples(t *testing.T) {
	runTest := func(input string, expected [][]string) {
		inputSections := splitGroupLines(input, "## ")
		output := reformatExamples(inputSections)

		assert.ElementsMatch(t, expected, output)
	}

	// This is a simple use case. We expect no changes to the original doc:
	simpleDoc := `description

## Example Usage

example usage content`

	simpleDocExpected := [][]string{
		{
			"description",
			"",
		},
		{
			"## Example Usage",
			"",
			"example usage content",
		},
	}

	runTest(simpleDoc, simpleDocExpected)

	// This use case demonstrates 2 examples at the same H2 level: a canonical Example Usage and another example
	// for a specific use case. We expect these to be transformed into a canonical H2 "Example Usage" with an H3 for
	// the specific use case.
	// This scenario is common in the pulumi-gcp provider:
	gcpDoc := `description

## Example Usage

example usage content

## Example Usage - Specific Case

specific case content`

	gcpDocExpected := [][]string{
		{
			"description",
			"",
		},
		{
			"## Example Usage",
			"",
			"example usage content",
			"",
			"### Specific Case",
			"",
			"specific case content",
		},
	}

	runTest(gcpDoc, gcpDocExpected)

	// This use case demonstrates 2 no canonical Example Usage/basic case and 2 specific use cases. We expect the
	// function to add a canonical Example Usage section with the 2 use cases as H3's beneath the canonical section.
	// This scenario is common in the pulumi-gcp provider:
	gcpDoc2 := `description

## Example Usage - 1

content 1

## Example Usage - 2

content 2`

	gcpDoc2Expected := [][]string{
		{
			"description",
			"",
		},
		{
			"## Example Usage",
			"### 1",
			"",
			"content 1",
			"",
			"### 2",
			"",
			"content 2",
		},
	}

	runTest(gcpDoc2, gcpDoc2Expected)
}

func TestFormatEntityName(t *testing.T) {
	assert.Equal(t, "'prov_entity'", formatEntityName("prov_entity"))
	assert.Equal(t, "'prov_entity' (aliased or renamed)", formatEntityName("prov_entity_legacy"))
}

func TestHclConversionsToString(t *testing.T) {
	input := map[string]string{
		"typescript": "var foo = bar;",
		"java":       "FooFactory fooFactory = new FooFactory();",
		"go":         "foo := bar",
		"python":     "foo = bar",
		"yaml":       "# Good enough YAML example",
		"csharp":     "var fooFactory = barProvider.Baz();",
		"pcl":        "# Good enough PCL example",
		"haskell":    "", // i.e., a language we could not convert, which should not appear in the output
	}

	// We use a template because we cannot escape backticks within a herestring, and concatenating this output would be
	// very difficult without using a herestring.
	expectedOutputTmpl := `{{ .CodeFences }}typescript
var foo = bar;
{{ .CodeFences }}
{{ .CodeFences }}python
foo = bar
{{ .CodeFences }}
{{ .CodeFences }}csharp
var fooFactory = barProvider.Baz();
{{ .CodeFences }}
{{ .CodeFences }}go
foo := bar
{{ .CodeFences }}
{{ .CodeFences }}java
FooFactory fooFactory = new FooFactory();
{{ .CodeFences }}
{{ .CodeFences }}pcl
# Good enough PCL example
{{ .CodeFences }}
{{ .CodeFences }}yaml
# Good enough YAML example
{{ .CodeFences }}`

	outputTemplate, _ := template.New("dummy").Parse(expectedOutputTmpl)
	data := struct {
		CodeFences string
	}{
		CodeFences: "```",
	}

	var buf = bytes.Buffer{}
	_ = outputTemplate.Execute(&buf, data)

	assert.Equal(t, buf.String(), hclConversionsToString(input))
}

func TestGroupLines(t *testing.T) {
	input := `description

## subtitle 1

subtitle 1 content

## subtitle 2

subtitle 2 content
`
	expected := [][]string{
		{
			"description",
			"",
		},
		{
			"## subtitle 1",
			"",
			"subtitle 1 content",
			"",
		},
		{
			"## subtitle 2",
			"",
			"subtitle 2 content",
			"",
		},
	}

	assert.Equal(t, expected, groupLines(strings.Split(input, "\n"), "## "))
}

func TestParseArgFromMarkdownLine(t *testing.T) {
	// nolint:lll
	tests := []struct {
		input         string
		expectedName  string
		expectedDesc  string
		expectedFound bool
	}{
		{"* `name` - (Required) A unique name to give the role.", "name", "A unique name to give the role.", true},
		{"* `key_vault_key_id` - (Optional) The Key Vault key URI for CMK encryption. Changing this forces a new resource to be created.", "key_vault_key_id", "The Key Vault key URI for CMK encryption. Changing this forces a new resource to be created.", true},
		// In rare cases, we may have a match where description is empty like the following, taken from https://github.com/hashicorp/terraform-provider-aws/blob/main/website/docs/r/spot_fleet_request.html.markdown
		{"* `instance_pools_to_use_count` - (Optional; Default: 1)", "instance_pools_to_use_count", "", true},
		{"", "", "", false},
		{"Most of these arguments directly correspond to the", "", "", false},
	}

	for _, test := range tests {
		name, desc, found := parseArgFromMarkdownLine(test.input)
		assert.Equal(t, test.expectedName, name)
		assert.Equal(t, test.expectedDesc, desc)
		assert.Equal(t, test.expectedFound, found)
	}
}

func TestGetNestedBlockName(t *testing.T) {
	var tests = []struct {
		input, expected string
	}{
		{"", ""},

		// AWS examples:
		{"The `website` object supports the following:", "website"},
		{"#### result_configuration Argument Reference", "result_configuration"},
		{"The `encryption_configuration` block supports the following argument:", "encryption_configuration"},

		// This is a common starting line of base arguments, so should result in zero value:
		{"The following arguments are supported:", ""},

		// GCP examples, from https://raw.githubusercontent.com/hashicorp/terraform-provider-google-beta/main/website/docs/r/sql_database_instance.html.markdown
		{"The `settings` block supports:", "settings"},
		{"The optional `settings.database_flags` sublist supports:", "settings.database_flags"},
		{"The optional `settings.backup_configuration` subblock supports:", "settings.backup_configuration"},
		{"The optional `replica_configuration` block must have `master_instance_name` set to work, cannot be updated, and supports:", "replica_configuration"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, getNestedBlockName(tt.input))
	}
}

func TestOverlayAttributesToAttributes(t *testing.T) {
	source := entityDocs{
		Attributes: map[string]string{
			"overwrite_me": "overwritten_desc",
			"source_only":  "source_only_desc",
		},
	}

	dest := entityDocs{
		Attributes: map[string]string{
			"overwrite_me": "original_desc",
			"dest_only":    "dest_only_desc",
		},
	}

	expected := entityDocs{
		Attributes: map[string]string{
			"overwrite_me": "overwritten_desc",
			"source_only":  "source_only_desc",
			"dest_only":    "dest_only_desc",
		},
	}

	overlayAttributesToAttributes(source, dest)

	assert.Equal(t, expected, dest)
}

func TestOverlayArgsToAttributes(t *testing.T) {
	source := entityDocs{
		Arguments: map[string]*argumentDocs{
			"overwrite_me": {
				description: "overwritten_desc",
			},
			"source_only": {
				description: "source_only_desc",
			},
		},
	}

	dest := entityDocs{
		Attributes: map[string]string{
			"overwrite_me": "original_desc",
			"dest_only":    "dest_only_desc",
		},
	}

	expected := entityDocs{
		Attributes: map[string]string{
			"overwrite_me": "overwritten_desc",
			"source_only":  "source_only_desc",
			"dest_only":    "dest_only_desc",
		},
	}

	overlayArgsToAttributes(source, dest)

	assert.Equal(t, expected, dest)
}

func TestOverlayArgsToArgs(t *testing.T) {
	source := entityDocs{
		Arguments: map[string]*argumentDocs{
			"overwrite_me": {
				description: "overwritten_desc",
				arguments: map[string]*argumentDocs{
					"nested_source_only": {
						description: "nested_source_only_desc",
						arguments: map[string]*argumentDocs{
							"double_nested_source_only": {
								description: "double_nested_source_only_desc",
							},
						},
					},
					"nested_overwrite_me": {description: "nested_overwrite_me_overwritten_desc"},
				},
			},
			"source_only": {
				description: "source_only_desc",
				arguments:   map[string]*argumentDocs{},
			},
		},
	}

	dest := entityDocs{
		Arguments: map[string]*argumentDocs{
			"overwrite_me": {
				description: "original_desc",
				arguments: map[string]*argumentDocs{
					"nested_dest_only": {
						description: "nested_dest_only_desc",
						arguments:   map[string]*argumentDocs{},
					},
					"nested_overwrite_me": {
						description: "nested_overwrite_me original desc",
						arguments:   map[string]*argumentDocs{},
					},
				},
			},
			"dest_only": {
				description: "dest_only_desc",
				arguments:   map[string]*argumentDocs{},
			},
		},
	}

	expected := entityDocs{
		Arguments: map[string]*argumentDocs{
			"overwrite_me": {
				description: "overwritten_desc",
				arguments: map[string]*argumentDocs{
					"nested_dest_only": {
						description: "nested_dest_only_desc",
						arguments:   map[string]*argumentDocs{},
					},
					"nested_source_only": {
						description: "nested_source_only_desc",
						arguments: map[string]*argumentDocs{
							"double_nested_source_only": {
								description: "double_nested_source_only_desc",
								arguments:   map[string]*argumentDocs{},
							},
						},
					},
					"nested_overwrite_me": {
						description: "nested_overwrite_me_overwritten_desc",
						arguments:   map[string]*argumentDocs{},
					},
				},
			},
			"source_only": {
				description: "source_only_desc",
				arguments:   map[string]*argumentDocs{},
			},
			"dest_only": {
				description: "dest_only_desc",
				arguments:   map[string]*argumentDocs{},
			},
		},
	}

	overlayArgsToArgs(source, dest)

	assert.Equal(t, expected, dest)
}

func TestArgFromNestedPath(t *testing.T) {
	args := map[string]*argumentDocs{
		"a": {
			description: "a desc",
			arguments: map[string]*argumentDocs{
				"aa": {
					description: "aa desc",
					arguments: map[string]*argumentDocs{
						"aaa": {
							description: "aaa desc",
						},
					},
				},
			},
		},
	}

	newNode := argumentDocs{
		arguments: make(map[string]*argumentDocs),
	}

	assert.Equal(t, args["a"], ensureArgFromNestedPath("a", args))
	assert.Equal(t, args["a"].arguments["aa"], ensureArgFromNestedPath("a.aa", args))
	assert.Equal(t, args["a"].arguments["aa"].arguments["aaa"], ensureArgFromNestedPath("a.aa.aaa", args))
	assert.Equal(t, &newNode, ensureArgFromNestedPath("b", args))
	assert.Equal(t, &newNode, ensureArgFromNestedPath("a.ab", args))
}

func TestFlattenKeys(t *testing.T) {
	input := map[string]*argumentDocs{
		"a": {
			arguments: map[string]*argumentDocs{
				"a": {
					arguments: map[string]*argumentDocs{
						"a": {},
						"b": {},
					},
				},
			},
		},
		"b": {
			arguments: map[string]*argumentDocs{},
		},
	}

	expected := []string{
		"a",
		"a.a",
		"a.a.a",
		"a.a.b",
		"b",
	}

	actual := flattenKeys(input, "")

	sort.Strings(expected)
	sort.Strings(actual)

	assert.Equal(t, expected, actual)
}

func TestFindMatchingKeys(t *testing.T) {
	fqKeys := []string{
		"not_a_match",
		"nested_1.settings",
		"settings",
		"settings.not_a_match", // it should only report the key above
		"nested_1.nested_2.settings",
	}

	expected := []string{
		"nested_1.settings",
		"settings",
		"nested_1.nested_2.settings",
	}

	actual := findMatchingKeys("settings", fqKeys)

	sort.Strings(expected)
	sort.Strings(actual)

	assert.Equal(t, expected, actual)
}

func TestParseArgNameFromHeader(t *testing.T) {
	assert.Equal(t, "dead_letter_config", parseArgNameFromHeader("### dead_letter_config"))
	assert.Equal(t, "capacity", parseArgNameFromHeader("### capacity Configuration Block"))
}

func TestMatchingArgNamesFromSectionHeader(t *testing.T) {
	args := map[string]*argumentDocs{
		"dead_letter_config": {
			arguments: map[string]*argumentDocs{},
		},
		"subblock": {
			arguments: map[string]*argumentDocs{
				"subblock_arg": {
					arguments: map[string]*argumentDocs{},
				},
			},
		},
		"multiple_matches": {
			arguments: map[string]*argumentDocs{},
		},
		"subblock_two": {
			arguments: map[string]*argumentDocs{
				"multiple_matches": {
					arguments: map[string]*argumentDocs{},
				},
			},
		},
	}

	tests := []struct {
		input    string
		expected []string
	}{
		{"notfound", []string{}},
		{"dead_letter_config", []string{"dead_letter_config"}},
		{"subblock_arg", []string{"subblock.subblock_arg"}},
		{"multiple_matches", []string{"multiple_matches", "subblock_two.multiple_matches"}},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, getMatchingArgNames(test.input, args, ""))
	}

	assert.Equal(t, []string{}, getMatchingArgNames("dummy", nil, ""))
}

func TestCleanupArgs_WithElided(t *testing.T) {
	reformatFunc := func(text string) (string, bool) {
		if strings.Contains(text, "badtext") {
			return "", true
		}

		return text, false
	}

	actualWarnings := []string{}
	warnFunc := func(msg string, args ...interface{}) {
		actualWarnings = append(actualWarnings, msg)
	}

	input := map[string]*argumentDocs{
		"ok_arg": {
			description: "ok_arg_desc",
			arguments: map[string]*argumentDocs{
				"bad_nested_1": {
					description: "this has some badtext",
					arguments: map[string]*argumentDocs{
						"bad_double_nested": {
							description: "more badtext - YIPES",
							arguments:   map[string]*argumentDocs{},
						},
					},
				},
			},
		},
		"bad_arg": {
			description: "this has some badtext",
			arguments: map[string]*argumentDocs{
				"ok_nested": {
					description: "ok_nested_desc",
					arguments:   map[string]*argumentDocs{},
				},
				"bad_nested_2": {
					description: "badtext also found here",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
	}

	expected := map[string]*argumentDocs{
		"ok_arg": {
			description: "ok_arg_desc",
			arguments: map[string]*argumentDocs{
				"bad_nested_1": {
					description: "",
					arguments: map[string]*argumentDocs{
						"bad_double_nested": {
							description: "",
							arguments:   map[string]*argumentDocs{},
						},
					},
				},
			},
		},
		"bad_arg": {
			description: "",
			arguments: map[string]*argumentDocs{
				"ok_nested": {
					description: "ok_nested_desc",
					arguments:   map[string]*argumentDocs{},
				},
				"bad_nested_2": {
					description: "",
					arguments:   map[string]*argumentDocs{},
				},
			},
		},
	}

	expectedWarnings := []string{
		"Found <elided> in docs for entity 'dummy' argument 'ok_arg.bad_nested_1'. The argument's description will be dropped in the Pulumi provider.",
		"Found <elided> in docs for entity 'dummy' argument 'ok_arg.bad_nested_1.bad_double_nested'. The argument's description will be dropped in the Pulumi provider.",
		"Found <elided> in docs for entity 'dummy' argument 'bad_arg'. The argument's description will be dropped in the Pulumi provider.",
		"Found <elided> in docs for entity 'dummy' argument 'bad_arg.bad_nested_2'. The argument's description will be dropped in the Pulumi provider.",
	}

	actual, elided := cleanupArgs(input, "dummy", reformatFunc, warnFunc, "")
	assert.Equal(t, true, elided)
	assert.Equal(t, expected, actual)
	assert.Equal(t, expectedWarnings, actualWarnings)
}
