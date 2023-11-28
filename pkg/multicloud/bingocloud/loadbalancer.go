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
	"sort"
	"strconv"
	"strings"
	"time"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/utils"

	api "yunion.io/x/cloudmux/pkg/apis/compute"
	"yunion.io/x/cloudmux/pkg/cloudprovider"
	"yunion.io/x/cloudmux/pkg/multicloud"
)

type SElb struct {
	BingoTags
	multicloud.SVirtualResourceBase
	region *SRegion

	Type                string
	LoadBalancerName    string
	DisplayName         string
	LoadBalancerArn     string
	LoadBalancerVersion string
	LoadBalancerId      string
	OwnerId             string
	AvailabilityZones   []struct {
		SubnetId string
		ZoneName string
	}
	CreatedTime          time.Time
	DNSName              string
	VipId                string
	SecurityGroups       []string
	State                State
	InstanceId           string
	IpAddressType        string
	VpcId                string
	SubnetId             string
	Nodes                string
	NodesCount           int
	ReplaceUnhealthyNode bool
	Description          string
}

type State struct {
	Code string
}

func (self *SElb) GetId() string {
	return self.LoadBalancerId
}

func (self *SElb) GetName() string {
	return self.DisplayName
}

func (self *SElb) GetGlobalId() string {
	return self.GetId()
}

func (self *SElb) GetStatus() string {
	switch self.State.Code {
	case "provisioning":
		return api.LB_STATUS_INIT
	case "active":
		return api.LB_STATUS_ENABLED
	case "failed":
		return api.LB_STATUS_START_FAILED
	default:
		return api.LB_STATUS_UNKNOWN
	}
}

func (self *SElb) Refresh() error {
	lb, err := self.region.GetLoadBalancer(self.GetId())
	if err != nil {
		return err
	}
	return jsonutils.Update(self, lb)
}

func (self *SElb) GetSysTags() map[string]string {
	data := map[string]string{}
	data["loadbalance_type"] = self.Type
	attrs, err := self.region.GetElbAttributes(self.GetId())
	if err != nil {
		return data
	}

	for k, v := range attrs {
		data[k] = v
	}
	return data
}

func (self *SElb) GetAddress() string {
	return self.DNSName
}

func (self *SElb) GetAddressType() string {
	switch self.IpAddressType {
	case "internal":
		return api.LB_ADDR_TYPE_INTRANET
	case "internet-facing":
		return api.LB_ADDR_TYPE_INTERNET
	default:
		return api.LB_ADDR_TYPE_INTRANET
	}
}

func (self *SElb) GetNetworkType() string {
	return api.LB_NETWORK_TYPE_VPC
}

func (self *SElb) GetNetworkIds() []string {
	var ret []string
	for i := range self.AvailabilityZones {
		ret = append(ret, self.AvailabilityZones[i].SubnetId)
	}
	return ret
}

func (self *SElb) GetVpcId() string {
	return self.VpcId
}

func (self *SElb) GetZoneId() string {
	var zones []string
	for i := range self.AvailabilityZones {
		zones = append(zones, self.AvailabilityZones[i].ZoneName)
	}

	sort.Strings(zones)
	if len(zones) > 0 {
		z, err := self.region.GetIZoneById(zones[0])
		if err != nil {
			log.Infof("getZoneById %s %s", zones[0], err)
			return ""
		}
		return z.GetGlobalId()
	}

	return ""
}

func (self *SElb) GetZone1Id() string {
	return ""
}

func (self *SElb) GetLoadbalancerSpec() string {
	return self.Type
}

func (self *SElb) GetChargeType() string {
	return api.LB_CHARGE_TYPE_BY_TRAFFIC
}

func (self *SElb) GetEgressMbps() int {
	return 0
}

func (self *SElb) Delete(ctx context.Context) error {
	return self.region.DeleteElb(self.GetId())
}

func (self *SElb) Start() error {
	return nil
}

func (self *SElb) Stop() error {
	return cloudprovider.ErrNotSupported
}

func (self *SElb) GetILoadBalancerListeners() ([]cloudprovider.ICloudLoadbalancerListener, error) {
	var ret []cloudprovider.ICloudLoadbalancerListener
	part, err := self.region.GetElbListeners(self.LoadBalancerId)
	if err != nil {
		return nil, err
	}
	for i := range part {
		part[i].lb = self
		ret = append(ret, &part[i])
	}
	return ret, nil
}

func (self *SElb) GetILoadBalancerBackendGroups() ([]cloudprovider.ICloudLoadbalancerBackendGroup, error) {
	var ret []cloudprovider.ICloudLoadbalancerBackendGroup
	listeners, err := self.GetILoadBalancerListeners()
	if err != nil {
		return nil, err
	}
	for _, listener := range listeners {
		part, err := self.region.GetElbBackendGroups(listener.GetBackendGroupId())
		if err != nil {
			return nil, errors.Wrapf(err, "GetElbBackendGroups")
		}
		for i := range part {
			part[i].region = self.region
			ret = append(ret, &part[i])
		}
	}
	return ret, nil
}

func (self *SElb) CreateILoadBalancerBackendGroup(group *cloudprovider.SLoadbalancerBackendGroup) (cloudprovider.ICloudLoadbalancerBackendGroup, error) {
	backendGroup, err := self.region.CreateElbBackendGroup(group)
	if err != nil {
		return nil, errors.Wrap(err, "CreateElbBackendGroup")
	}

	backendGroup.region = self.region

	return backendGroup, nil
}

func (self *SElb) GetILoadBalancerBackendGroupById(groupId string) (cloudprovider.ICloudLoadbalancerBackendGroup, error) {
	lbbg, err := self.region.GetElbBackendGroup(groupId)
	if err != nil {
		return nil, err
	}
	lbbg.region = self.region
	return lbbg, nil
}

func (self *SElb) CreateILoadBalancerListener(ctx context.Context, listener *cloudprovider.SLoadbalancerListenerCreateOptions) (cloudprovider.ICloudLoadbalancerListener, error) {
	ret, err := self.region.CreateElbListener(self.LoadBalancerId, listener)
	if err != nil {
		return nil, errors.Wrap(err, "CreateElbListener")
	}

	ret.lb = self
	return ret, nil
}

func (self *SElb) GetILoadBalancerListenerById(listenerId string) (cloudprovider.ICloudLoadbalancerListener, error) {
	lis, err := self.region.GetElbListeners(self.LoadBalancerId)
	if err != nil {
		return nil, err
	}
	for i := range lis {
		if lis[i].ListenerId == listenerId {
			lis[i].lb = self
			return &lis[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound
}

func (self *SElb) GetIEIP() (cloudprovider.ICloudEIP, error) {
	return nil, nil
}

func (self *SRegion) DeleteElb(id string) error {
	elb, err := self.GetILoadBalancerById(id)
	if err != nil {
		return err
	}
	elbgs, err := elb.GetILoadBalancerBackendGroups()
	if err != nil {
		return err
	}
	for _, elbg := range elbgs {
		elbg.Delete(context.Background())
	}
	params := map[string]string{"LoadBalancerArn": id}
	_, err = self.invoke("DeleteLoadBalancer", params)
	return err
}

func (self *SRegion) GetElbBackendGroups(targetGroupId string) ([]SElbBackendGroup, error) {
	params := map[string]string{}
	if len(targetGroupId) > 0 {
		params["TargetGroupArns.member.1"] = targetGroupId
	}
	params["ownerId"] = self.client.user
	resp, err := self.invoke("DescribeTargetGroups", params)
	if err != nil {
		return nil, errors.Wrapf(err, "DescribeTargetGroups")
	}

	var rets []SElbBackendGroup
	resp.Unmarshal(&rets, "DescribeTargetGroupsResult", "TargetGroups")
	return rets, nil
}

func (self *SRegion) GetElbBackendGroup(id string) (*SElbBackendGroup, error) {
	groups, err := self.GetElbBackendGroups(id)
	if err != nil {
		return nil, err
	}
	for i := range groups {
		if groups[i].TargetGroupId == id {
			groups[i].region = self
			return &groups[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func ToOnecloudHealthCode(s string) string {
	ret := []string{}

	segs := strings.Split(s, ",")
	for _, seg := range segs {
		codes := strings.Split(seg, "-")
		for _, code := range codes {
			c, _ := strconv.Atoi(code)
			if c >= 400 && !utils.IsInStringArray(api.LB_HEALTH_CHECK_HTTP_CODE_4xx, ret) {
				ret = append(ret, api.LB_HEALTH_CHECK_HTTP_CODE_4xx)
			} else if c >= 300 && !utils.IsInStringArray(api.LB_HEALTH_CHECK_HTTP_CODE_3xx, ret) {
				ret = append(ret, api.LB_HEALTH_CHECK_HTTP_CODE_3xx)
			} else if c >= 200 && !utils.IsInStringArray(api.LB_HEALTH_CHECK_HTTP_CODE_2xx, ret) {
				ret = append(ret, api.LB_HEALTH_CHECK_HTTP_CODE_2xx)
			}
		}

		if len(codes) == 2 {
			min, _ := strconv.Atoi(codes[0])
			max, _ := strconv.Atoi(codes[1])

			if min >= 200 && max >= 400 {
				if !utils.IsInStringArray(api.LB_HEALTH_CHECK_HTTP_CODE_3xx, ret) {
					ret = append(ret, api.LB_HEALTH_CHECK_HTTP_CODE_3xx)
				}
			}
		}
	}

	return strings.Join(ret, ",")
}

func (self *SRegion) CreateElbBackendGroup(opts *cloudprovider.SLoadbalancerBackendGroup) (*SElbBackendGroup, error) {
	params := map[string]string{
		"Protocol":           strings.ToUpper(opts.Protocol),
		"Name":               opts.Name,
		"Port":               strconv.Itoa(opts.ListenPort),
		"TargetType":         opts.TargetType,
		"VpcId":              opts.VpcId,
		"HealthCheckEnabled": fmt.Sprintf("%v", opts.HealthCheckEnabled),
	}
	if opts.HealthCheckEnabled {
		params["HealthCheckIntervalSeconds"] = strconv.Itoa(opts.HealthCheckIntervalSeconds)
		params["HealthCheckPath"] = opts.HealthCheckPath
		params["HealthCheckProtocol"] = strings.ToUpper(opts.HealthCheckProtocol)
		params["HealthCheckTimeoutSeconds"] = strconv.Itoa(opts.HealthCheckTimeoutSeconds)
		params["HealthyThresholdCount"] = strconv.Itoa(opts.HealthyThresholdCount)
		params["UnhealthyThresholdCount"] = strconv.Itoa(opts.UnhealthyThresholdCount)
	}
	var ret struct {
		TargetGroups []SElbBackendGroup
	}
	resp, err := self.invoke("CreateTargetGroup", params)
	if err != nil {
		return nil, err
	}
	resp.Unmarshal(&ret, "CreateTargetGroupResult")
	for i := range ret.TargetGroups {
		return &ret.TargetGroups[i], nil
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, "after created")
}

func (self *SRegion) GetLoadBalancer(id string) (*SElb, error) {
	part, err := self.GetLoadbalancers(id)
	if err != nil {
		return nil, errors.Wrapf(err, "GetLoadbalancers")
	}
	for i := range part {
		if part[i].GetGlobalId() == id {
			part[i].region = self
			return &part[i], nil
		}
	}
	return nil, errors.Wrap(cloudprovider.ErrNotFound, id)
}

func (self *SRegion) GetLoadbalancers(id string) ([]SElb, error) {
	params := map[string]string{"LoadBalancerVersion": "v2"}
	if len(id) > 0 {
		params["LoadBalancerArns.member.1"] = id
	}
	params["ownerId"] = self.client.user

	resp, err := self.invoke("DescribeLoadBalancers", params)
	if err != nil {
		return nil, errors.Wrapf(err, "DescribeLoadBalancers")
	}

	var ret struct {
		LoadBalancers []SElb
	}
	err = resp.Unmarshal(&ret, "DescribeLoadBalancersResult")
	return ret.LoadBalancers, err
}

func (self *SRegion) CreateLoadbalancer(opts *cloudprovider.SLoadbalancerCreateOptions) (*SElb, error) {
	params := map[string]string{
		"Name":  opts.Name,
		"VpcId": opts.VpcId,
	}

	for i, net := range opts.NetworkIds {
		params[fmt.Sprintf("Subnets.member.%d", i+1)] = net
	}

	var ret struct {
		LoadBalancers []SElb
	}

	resp, err := self.invoke("CreateLoadBalancer", params)
	if err != nil {
		return nil, errors.Wrapf(err, "CreateLoadBalancer")
	}
	resp.Unmarshal(&ret, "CreateLoadBalancerResult")
	for i := range ret.LoadBalancers {
		ret.LoadBalancers[i].region = self
		return &ret.LoadBalancers[i], nil
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, "after created")
}

func (self *SRegion) GetElbAttributes(id string) (map[string]string, error) {
	ret := struct {
		Attributes []struct {
			Key   string
			Value string
		}
	}{}
	params := map[string]string{"LoadBalancerArn": id}
	resp, err := self.invoke("DescribeLoadBalancerAttributes", params)
	if err != nil {
		return nil, err
	}

	resp.Unmarshal(&ret)
	result := map[string]string{}
	for _, attr := range ret.Attributes {
		result[attr.Key] = attr.Value
	}
	return result, nil
}
