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
	"strconv"
	"strings"
	"time"

	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/errors"

	api "yunion.io/x/cloudmux/pkg/apis/compute"
	"yunion.io/x/cloudmux/pkg/cloudprovider"
	"yunion.io/x/cloudmux/pkg/multicloud"
)

type SElbBackendGroup struct {
	multicloud.SResourceBase
	BingoTags
	region *SRegion

	CreatedTime                time.Time
	TargetGroupArn             string
	TargetGroupId              string
	LoadBalancerArns           []string
	TargetGroupName            string
	DisplayName                string
	TargetType                 string
	VpcId                      string
	OwnerId                    string
	Protocol                   string
	Port                       int64
	HealthCheckEnabled         bool
	HealthCheckIntervalSeconds int
	HealthCheckMethod          string
	HealthCheckPath            string
	HealthCheckPort            string
	HealthCheckProtocol        string
	HealthCheckTimeoutSeconds  int
	HealthyThresholdCount      int
	UnhealthyThresholdCount    int
	Matcher                    struct {
		HttpCode string
	}
	Description string
}

func (self *SElbBackendGroup) GetId() string {
	return self.TargetGroupId
}

func (self *SElbBackendGroup) GetName() string {
	return self.DisplayName
}

func (self *SElbBackendGroup) GetGlobalId() string {
	return self.GetId()
}

func (self *SElbBackendGroup) GetStatus() string {
	return api.LB_STATUS_ENABLED
}

func (self *SElbBackendGroup) Refresh() error {
	lbbg, err := self.region.GetElbBackendGroup(self.GetId())
	if err != nil {
		return err
	}
	return jsonutils.Update(self, lbbg)
}

func (self *SElbBackendGroup) GetSysTags() map[string]string {
	data := map[string]string{}
	data["port"] = strconv.FormatInt(self.Port, 10)
	data["target_type"] = self.TargetType
	data["health_check_protocol"] = strings.ToLower(self.HealthCheckProtocol)
	data["health_check_interval"] = strconv.Itoa(self.HealthCheckIntervalSeconds)
	return data
}

func (self *SElbBackendGroup) GetProjectId() string {
	return ""
}

func (self *SElbBackendGroup) IsDefault() bool {
	return false
}

func (self *SElbBackendGroup) GetType() string {
	switch self.TargetType {
	case "instance":
		return api.LB_BACKEND_GUEST
	case "network-interface":
		return api.LB_BACKEND_NETWORK_INTERFACE
	case "ip":
		return api.LB_BACKEND_IP
	default:
		return self.TargetType
	}
}

func (self *SElbBackendGroup) GetILoadbalancerBackends() ([]cloudprovider.ICloudLoadbalancerBackend, error) {
	backends, err := self.region.GetELbBackends(self.GetId())
	if err != nil {
		return nil, errors.Wrap(err, "GetELbBackends")
	}

	ibackends := make([]cloudprovider.ICloudLoadbalancerBackend, len(backends))
	for i := range backends {
		backends[i].group = self
		ibackends[i] = &backends[i]
	}

	return ibackends, nil
}

func (self *SElbBackendGroup) GetILoadbalancerBackendById(backendId string) (cloudprovider.ICloudLoadbalancerBackend, error) {
	backend, err := self.region.GetELbBackend(backendId)
	if err != nil {
		return nil, errors.Wrap(err, "GetELbBackend")
	}

	backend.group = self
	return backend, nil
}

func (self *SElbBackendGroup) GetProtocolType() string {
	switch self.Protocol {
	case "TCP":
		return api.LB_LISTENER_TYPE_TCP
	case "UDP":
		return api.LB_LISTENER_TYPE_UDP
	case "HTTP":
		return api.LB_LISTENER_TYPE_HTTP
	case "HTTPS":
		return api.LB_LISTENER_TYPE_HTTPS
	case "TCP_UDP":
		return api.LB_LISTENER_TYPE_TCP_UDP
	default:
		return ""
	}
}

func (self *SElbBackendGroup) GetScheduler() string {
	attrs, err := self.region.GetElbBackendGroupAttributes(self.GetId())
	if err != nil {
		return ""
	}
	if val, exist := attrs["load_balancing.method"]; exist {
		switch val {
		case "round_robin":
			return api.LB_SCHEDULER_RR
		case "weighted_round_robin":
			return api.LB_SCHEDULER_WRR
		case "least_conn":
			return api.LB_SCHEDULER_WLC
		case "ip_hash":
			return api.LB_SCHEDULER_SCH
		}
		return val
	}
	return ""
}

func (self *SElbBackendGroup) GetHealthCheck() (*cloudprovider.SLoadbalancerHealthCheck, error) {
	health := &cloudprovider.SLoadbalancerHealthCheck{}
	health.HealthCheck = api.LB_BOOL_OFF
	if self.HealthCheckEnabled {
		health.HealthCheck = api.LB_BOOL_ON
	}
	health.HealthCheckRise = self.HealthyThresholdCount
	health.HealthCheckFail = self.UnhealthyThresholdCount
	health.HealthCheckInterval = self.HealthCheckIntervalSeconds
	health.HealthCheckURI = self.HealthCheckPath
	health.HealthCheckType = strings.ToLower(self.HealthCheckProtocol)
	health.HealthCheckTimeout = self.HealthCheckTimeoutSeconds
	health.HealthCheckHttpCode = ToOnecloudHealthCode(self.Matcher.HttpCode)
	return health, nil
}

func (self *SElbBackendGroup) GetStickySession() (*cloudprovider.SLoadbalancerStickySession, error) {
	attrs, err := self.region.GetElbBackendGroupAttributes(self.GetId())
	if err != nil {
		return nil, errors.Wrap(err, "GetElbBackendGroupAttributes")
	}

	cookieTime := 0
	if t, ok := attrs["stickiness.lb_cookie.duration_seconds"]; !ok {
		cookieTime, err = strconv.Atoi(t)
	}

	ret := &cloudprovider.SLoadbalancerStickySession{
		StickySession:              attrs["stickiness.enabled"],
		StickySessionCookie:        "",
		StickySessionType:          api.LB_STICKY_SESSION_TYPE_INSERT,
		StickySessionCookieTimeout: cookieTime,
	}

	return ret, nil
}

func (self *SElbBackendGroup) AddBackendServer(targetId string, weight int, port int) (cloudprovider.ICloudLoadbalancerBackend, error) {
	err := self.region.AddElbBackend(self.GetId(), targetId, weight, port)
	if err != nil {
		return nil, errors.Wrap(err, "AddElbBackend")
	}
	target, err := self.region.GetELbBackend(genElbBackendId(self.GetId(), targetId, port))
	if err != nil {
		return nil, errors.Wrap(err, "GetELbBackend")
	}
	target.group = self
	return target, nil
}

func (self *SElbBackendGroup) RemoveBackendServer(targetId string, weight int, port int) error {
	return self.region.RemoveElbBackend(self.GetId(), targetId, weight, port)
}

func (self *SElbBackendGroup) Delete(ctx context.Context) error {
	return self.region.DeleteElbBackendGroup(self.GetId())
}

func (self *SElbBackendGroup) Sync(ctx context.Context, group *cloudprovider.SLoadbalancerBackendGroup) error {
	return nil
}

func (self *SRegion) GetELbBackends(targetGroupId string) ([]SElbBackend, error) {
	params := map[string]string{
		"TargetGroupArn": targetGroupId,
	}
	resp, err := self.invoke("DescribeTargetHealth", params)
	if err != nil {
		return nil, err
	}

	var ret struct {
		Targets []SElbBackend `json:"TargetHealthDescriptions"`
	}
	err = resp.Unmarshal(&ret, "DescribeTargetHealthResult")
	return ret.Targets, err
}

func (self *SRegion) GetELbBackend(id string) (*SElbBackend, error) {
	targetGroupId, targetId, port, err := parseElbBackendId(id)
	if err != nil {
		return nil, errors.Wrap(err, "parseElbBackendId")
	}
	lbbg, err := self.GetElbBackendGroup(targetGroupId)
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"TargetGroupArn":        targetGroupId,
		"Targets.member.1.Id":   targetId,
		"Targets.member.1.Port": fmt.Sprintf("%d", port),
	}
	var ret struct {
		Targets []SElbBackend `json:"TargetHealthDescriptions"`
	}
	resp, err := self.invoke("DescribeTargetHealth", params)
	if err != nil {
		return nil, err
	}
	resp.Unmarshal(&ret, "DescribeTargetHealthResult")
	for i := range ret.Targets {
		ret.Targets[i].group = lbbg
		if ret.Targets[i].GetGlobalId() == id {
			return &ret.Targets[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func parseElbBackendId(id string) (string, string, int, error) {
	segs := strings.Split(id, "::")
	if len(segs) != 3 {
		return "", "", 0, fmt.Errorf("%s is not a valid target id", id)
	}

	port, err := strconv.Atoi(segs[2])
	if err != nil {
		return "", "", 0, fmt.Errorf("%s is not a valid target id, %s", id, err)
	}

	return segs[0], segs[1], port, nil
}

func genElbBackendId(targetGroupId string, targetId string, port int) string {
	return strings.Join([]string{targetGroupId, targetId, strconv.Itoa(port)}, "::")
}

func (self *SRegion) AddElbBackend(targetGroupId, targetId string, weight int, port int) error {
	params := map[string]string{
		"TargetGroupArn":          targetGroupId,
		"Targets.member.1.Id":     targetId,
		"Targets.member.1.Port":   fmt.Sprintf("%d", port),
		"Targets.member.1.Weight": fmt.Sprintf("%d", weight),
	}
	_, err := self.invoke("RegisterTargets", params)
	if err != nil {
		return errors.Wrapf(err, "RegisterTargets")
	}
	return nil
}

func (self *SRegion) RemoveElbBackend(targetGroupId, targetId string, weight int, port int) error {
	if strings.Contains(targetId, "::") {
		_, targetId, _, _ = parseElbBackendId(targetId)
	}
	params := map[string]string{
		"TargetGroupArn":        targetGroupId,
		"Targets.member.1.Id":   targetId,
		"Targets.member.1.Port": fmt.Sprintf("%d", port),
	}
	_, err := self.invoke("DeregisterTargets", params)
	return err
}

func (self *SRegion) DeleteElbBackendGroup(id string) error {
	_, err := self.invoke("DeleteTargetGroup", map[string]string{"TargetGroupArn": id})
	return err
}

func (self *SRegion) RemoveElbBackends(targetGroupId string) error {
	backends, err := self.GetELbBackends(targetGroupId)
	if err != nil {
		return errors.Wrap(err, "GetELbBackends")
	}

	if len(backends) == 0 {
		return nil
	}
	for i := range backends {
		err := self.RemoveElbBackend(targetGroupId, backends[i].GetBackendId(), 0, backends[i].GetPort())
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *SRegion) GetElbBackendGroupAttributes(id string) (map[string]string, error) {
	ret := struct {
		Attributes []struct {
			Key   string
			Value string
		}
	}{}

	resp, err := self.invoke("DescribeTargetGroupAttributes", map[string]string{"TargetGroupArn": id})
	if err != nil {
		return nil, err
	}
	resp.Unmarshal(&ret, "DescribeTargetGroupAttributesResult")

	result := map[string]string{}
	for _, attr := range ret.Attributes {
		result[attr.Key] = attr.Value
	}
	return result, nil
}
