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
	"strconv"
	"time"

	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"

	"yunion.io/x/cloudmux/pkg/cloudprovider"
)

var hostMetricsMap = map[string]cloudprovider.TMetricType{
	"CPUUtilization": cloudprovider.HOST_METRIC_TYPE_CPU_USAGE,
	"MemeryUsage":    cloudprovider.HOST_METRIC_TYPE_MEM_USAGE,
	"NetworkIn":      cloudprovider.HOST_METRIC_TYPE_NET_BPS_RX,
	"NetworkOut":     cloudprovider.HOST_METRIC_TYPE_NET_BPS_TX,
	"DiskReadBytes":  cloudprovider.HOST_METRIC_TYPE_DISK_IO_READ_BPS,
	"DiskWriteBytes": cloudprovider.HOST_METRIC_TYPE_DISK_IO_WRITE_BPS,
	"DiskReadOps":    cloudprovider.HOST_METRIC_TYPE_DISK_IO_READ_IOPS,
	"DiskWriteOps":   cloudprovider.HOST_METRIC_TYPE_DISK_IO_WRITE_IOPS,
}

var vmMetricsMap = map[string]cloudprovider.TMetricType{
	"CPUUtilization": cloudprovider.VM_METRIC_TYPE_CPU_USAGE,
	"MemeryUsage":    cloudprovider.VM_METRIC_TYPE_MEM_USAGE,
	"NetworkIn":      cloudprovider.VM_METRIC_TYPE_NET_BPS_RX,
	"NetworkOut":     cloudprovider.VM_METRIC_TYPE_NET_BPS_TX,
	"DiskUsage":      cloudprovider.VM_METRIC_TYPE_DISK_USAGE,
	"DiskReadBytes":  cloudprovider.VM_METRIC_TYPE_DISK_IO_READ_BPS,
	"DiskWriteBytes": cloudprovider.VM_METRIC_TYPE_DISK_IO_WRITE_BPS,
	"DiskReadOps":    cloudprovider.VM_METRIC_TYPE_DISK_IO_READ_IOPS,
	"DiskWriteOps":   cloudprovider.VM_METRIC_TYPE_DISK_IO_WRITE_IOPS,
}

var lbMetricsMap = map[string]cloudprovider.TMetricType{
	"UnHealthyHostCount": cloudprovider.LB_METRIC_TYPE_UNHEALTHY_SERVER_COUNT,
	"RequestCount":       cloudprovider.LB_METRIC_TYPE_MAX_CONNECTION,
}

var metricsMap = map[cloudprovider.TResourceType]map[string]cloudprovider.TMetricType{
	cloudprovider.METRIC_RESOURCE_TYPE_SERVER: vmMetricsMap,
	cloudprovider.METRIC_RESOURCE_TYPE_HOST:   hostMetricsMap,
	cloudprovider.METRIC_RESOURCE_TYPE_LB:     lbMetricsMap,
}

var resourceTypeMap = map[cloudprovider.TResourceType]string{
	cloudprovider.METRIC_RESOURCE_TYPE_SERVER: "Instance",
	cloudprovider.METRIC_RESOURCE_TYPE_HOST:   "Host",
	cloudprovider.METRIC_RESOURCE_TYPE_LB:     "LoadBalancer",
}

type MetricOutput struct {
	NextToken  string
	Datapoints []DatapointMember
}

type DatapointMember struct {
	ResourceId  string
	MetricName  string
	Average     float64
	Maximum     float64
	Minimum     float64
	SampleCount float64
	Sum         float64
	Timestamp   string
}

func (self DatapointMember) GetValue() float64 {
	return self.Average + self.Maximum + self.Minimum + self.Sum
}

type Datapoints struct {
	Member []DatapointMember
}

func (self *SBingoCloudClient) GetResourceStatistics(resourceType string, since time.Time, until time.Time) (*MetricOutput, error) {
	params := map[string]string{}
	params["ResourceType"] = resourceType
	params["StartTime"] = strconv.FormatInt(since.UTC().Unix(), 10)
	params["EndTime"] = strconv.FormatInt(until.UTC().Unix(), 10)

	ret := &MetricOutput{}

	for {
		tmp := &MetricOutput{}
		resp, err := self.invoke("GetResourceStatistics", params)
		if err != nil {
			return nil, errors.Wrap(err, "GetResourceStatistics err")
		}
		err = resp.Unmarshal(tmp, "GetResourceStatisticsResult")
		if err != nil {
			return nil, errors.Wrap(err, "GetResourceStatistics err")
		}

		ret.Datapoints = append(ret.Datapoints, tmp.Datapoints...)

		if tmp.NextToken == "" {
			break
		}
		params["NextToken"] = tmp.NextToken
	}

	return ret, nil
}

func (self *SBingoCloudClient) GetMetrics(opts *cloudprovider.MetricListOptions) ([]cloudprovider.MetricValues, error) {
	data, err := self.GetResourceStatistics(resourceTypeMap[opts.ResourceType], opts.StartTime, opts.EndTime)
	if err != nil {
		log.Errorf("GetResourceStatistics error: %v", err)
		return nil, err
	}

	var ret []*cloudprovider.MetricValues

	resources := map[string]map[cloudprovider.TMetricType]*cloudprovider.MetricValues{}

	if data != nil && data.Datapoints != nil {
		for _, member := range data.Datapoints {
			metricType := metricsMap[opts.ResourceType][member.MetricName]
			if metricType == "" {
				continue
			}
			var metric = &cloudprovider.MetricValues{
				Id:         member.ResourceId,
				Values:     []cloudprovider.MetricValue{},
				MetricType: metricType,
			}
			if res, exist := resources[member.ResourceId]; exist {
				if mc, exist := res[metricType]; exist {
					metric = mc
				} else {
					resources[member.ResourceId][metricType] = metric
					ret = append(ret, metric)
				}
			} else {
				resources[member.ResourceId] = map[cloudprovider.TMetricType]*cloudprovider.MetricValues{metricType: metric}
				ret = append(ret, metric)
			}
			metricValue := cloudprovider.MetricValue{}

			toInt64, _ := strconv.ParseInt(member.Timestamp, 10, 64)
			metricValue.Timestamp = time.Unix(toInt64, 0).UTC()
			metricValue.Value = member.Average
			metric.Values = append(metric.Values, metricValue)
		}
	}

	result := make([]cloudprovider.MetricValues, len(ret))
	for i, p := range ret {
		result[i] = *p
	}

	return result, nil
}
