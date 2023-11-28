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
	"context"
	"fmt"

	"yunion.io/x/cloudmux/pkg/cloudprovider"
	"yunion.io/x/jsonutils"

	api "yunion.io/x/cloudmux/pkg/apis/compute"
	"yunion.io/x/cloudmux/pkg/multicloud"
)

type SElbBackend struct {
	multicloud.SResourceBase
	BingoTags
	group *SElbBackendGroup

	HealthCheckPort int
	Target          Target
	TargetHealth    TargetHealth
}

type Target struct {
	ID      string
	Port    int
	Weight  int
	Address string
}

type TargetHealth struct {
	Reason string
	State  string
}

func (self *SElbBackend) GetId() string {
	return fmt.Sprintf("%s::%s::%d", self.group.GetId(), self.Target.ID, self.Target.Port)
}

func (self *SElbBackend) GetName() string {
	return self.GetId()
}

func (self *SElbBackend) GetGlobalId() string {
	return self.GetId()
}

func (self *SElbBackend) GetStatus() string {
	return api.LB_STATUS_ENABLED
}

func (self *SElbBackend) Refresh() error {
	new, err := self.group.GetILoadbalancerBackendById(self.GetId())
	if err != nil {
		return err
	}
	return jsonutils.Update(self, new)
}

func (self *SElbBackend) IsEmulated() bool {
	return false
}

func (self *SElbBackend) GetProjectId() string {
	return ""
}

func (self *SElbBackend) GetWeight() int {
	return self.Target.Weight
}

func (self *SElbBackend) GetPort() int {
	return self.Target.Port
}

func (self *SElbBackend) GetBackendType() string {
	switch self.group.TargetType {
	case "instance":
		return api.LB_BACKEND_GUEST
	case "network-interface":
		return api.LB_BACKEND_NETWORK_INTERFACE
	case "ip":
		return api.LB_BACKEND_IP
	default:
		return self.group.TargetType
	}
}

func (self *SElbBackend) GetBackendRole() string {
	return api.LB_BACKEND_ROLE_DEFAULT
}

func (self *SElbBackend) GetBackendId() string {
	return self.Target.ID
}

func (self *SElbBackend) GetGroupId() string {
	return self.group.GetId()
}

func (self *SElbBackend) SyncConf(ctx context.Context, port, weight int) error {
	newest, err := self.group.region.SyncElbBackend(self, port, weight)
	if err != nil {
		return err
	}
	jsonutils.Update(self, newest)
	return nil
}

func (self *SElbBackend) GetIpAddress() string {
	return self.Target.Address
}

func (self *SRegion) SyncElbBackend(backend *SElbBackend, newPort, weight int) (cloudprovider.ICloudLoadbalancerBackend, error) {
	targetGroupId := backend.GetGroupId()
	targetId := backend.GetBackendId()
	oldPort := backend.Target.Port

	err := self.RemoveElbBackend(targetGroupId, targetId, weight, oldPort)
	if err != nil {
		return nil, err
	}

	oldBackends, err := backend.group.GetILoadbalancerBackends()
	if err != nil {
		return nil, err
	}
	err = self.AddElbBackend(targetGroupId, targetId, weight, newPort)
	if err != nil {
		return nil, err
	}
	newBackends, err := backend.group.GetILoadbalancerBackends()
	if err != nil {
		return nil, err
	}
	for _, new := range newBackends {
		exist := false
		for _, old := range oldBackends {
			if new.GetId() == old.GetId() {
				break
			}
		}
		if !exist {
			return new, nil
		}
	}
	return nil, cloudprovider.ErrNotFound
}
