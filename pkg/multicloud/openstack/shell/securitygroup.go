// Copyright 2019 Yunion
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

package shell

import (
	"yunion.io/x/pkg/util/shellutils"

	"yunion.io/x/cloudmux/pkg/multicloud/openstack"
)

func init() {
	type SecurityGroupListOptions struct {
		ProjectId string
		Name      string
	}
	shellutils.R(&SecurityGroupListOptions{}, "security-group-list", "List security groups", func(cli *openstack.SRegion, args *SecurityGroupListOptions) error {
		secgroup, err := cli.GetSecurityGroups(args.ProjectId, args.Name)
		if err != nil {
			return err
		}
		printList(secgroup, 0, 0, 0, nil)
		return nil
	})

	type SecurityGroupShowOptions struct {
		ID        string `help:"ID of security group"`
		ShowRules bool   `help:"Show rules"`
	}
	shellutils.R(&SecurityGroupShowOptions{}, "security-group-show", "Show security group", func(cli *openstack.SRegion, args *SecurityGroupShowOptions) error {
		secgroup, err := cli.GetSecurityGroup(args.ID)
		if err != nil {
			return err
		}
		printObject(secgroup)
		if args.ShowRules {
			rules, err := secgroup.GetRules()
			if err != nil {
				return err
			}
			for _, r := range rules {
				printObject(r)
			}
		}
		return nil
	})

	type SecurityGroupCreateOptions struct {
		ProjectId string
		NAME      string `help:"Name of security group"`
		Desc      string `help:"Description of security group"`
	}

	shellutils.R(&SecurityGroupCreateOptions{}, "security-group-create", "Create security group", func(cli *openstack.SRegion, args *SecurityGroupCreateOptions) error {
		secgroup, err := cli.CreateSecurityGroup(args.ProjectId, args.NAME, args.Desc)
		if err != nil {
			return err
		}
		printObject(secgroup)
		return nil
	})

}
