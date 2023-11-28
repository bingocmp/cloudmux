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

package bingocloud

import (
	"fmt"

	"yunion.io/x/cloudmux/pkg/cloudprovider"
	"yunion.io/x/pkg/util/secrules"
)

type SSecurityGroupRule struct {
	group     *SSecurityGroup
	direction secrules.TSecurityRuleDirection

	BoundType   string `json:"boundType"`
	Description string `json:"description"`
	FromPort    int    `json:"fromPort"`
	IPProtocol  string `json:"ipProtocol"`
	Groups      []struct {
		GroupId   string
		GroupName string
	} `json:"groups"`
	IPRanges []struct {
		CIDRIP string `json:"cidrIp"`
	} `json:"ipRanges"`
	L2Accept     string `json:"l2Accept"`
	PermissionId string `json:"permissionId"`
	Policy       string `json:"policy"`
	ToPort       int    `json:"toPort"`
}

func (self *SSecurityGroupRule) GetGlobalId() string {
	return self.PermissionId
}

func (self *SSecurityGroupRule) GetDescription() string {
	return self.Description
}

func (self *SSecurityGroupRule) GetAction() secrules.TSecurityRuleAction {
	if self.Policy == "DROP" {
		return secrules.SecurityRuleDeny
	}
	return secrules.SecurityRuleAllow
}

func (self *SSecurityGroupRule) GetProtocol() string {
	protocol := secrules.PROTO_ANY
	if self.IPProtocol != "all" {
		protocol = self.IPProtocol
	}
	return protocol
}

func (self *SSecurityGroupRule) GetPorts() string {
	if self.FromPort > 0 && self.ToPort > 0 {
		if self.FromPort == self.ToPort {
			return fmt.Sprintf("%d", self.FromPort)
		}
		return fmt.Sprintf("%d-%d", self.FromPort, self.ToPort)
	}
	return ""
}

func (self *SSecurityGroupRule) GetPriority() int {
	return 0
}

func (self *SSecurityGroupRule) GetCIDRs() []string {
	nets := []string{}
	for _, ip := range self.IPRanges {
		nets = append(nets, ip.CIDRIP)
	}
	return nets
}

func (self *SSecurityGroupRule) GetDirection() secrules.TSecurityRuleDirection {
	return self.direction
}

func (self *SSecurityGroupRule) Delete() error {
	return self.group.region.DeleteSecurityGroupRule(self)
}

func (self *SSecurityGroupRule) Update(opts *cloudprovider.SecurityGroupRuleUpdateOptions) error {
	self.group.region.DeleteSecurityGroupRule(self)
	return self.group.region.CreateSecurityGroupRules(self.group.GroupId, &cloudprovider.SecurityGroupRuleCreateOptions{
		Desc:      opts.Desc,
		Protocol:  opts.Protocol,
		Ports:     opts.Ports,
		Direction: self.GetDirection(),
		CIDR:      opts.CIDR,
		Action:    opts.Action,
	})
}
