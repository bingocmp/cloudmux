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
	"testing"
)

var bingoCloudClient, _ = NewBingoCloudClient(&BingoCloudConfig{
	endpoint:  "http://10.203.136.11",
	accessKey: "98DDB4D334328A0989C8",
	secretKey: "W0NGMTIyRUYyNERGQzQxRDA0NTcwNkUxQkZCOTY1",
	debug:     false,
})

var managerBingoCloudClient, _ = NewBingoCloudClient(&BingoCloudConfig{
	endpoint:  "http://10.203.136.11",
	accessKey: "373ABE5836BB9E82FCFB",
	secretKey: "WzkzNzk0Mzg2ODQxODQ5REJEN0U4QTg1NTY5NzA2",
	debug:     false,
})

func TestNewBingoCloudClient(t *testing.T) {
	bingoCloudClient = managerBingoCloudClient
	bingoCloudClient.managerClient = managerBingoCloudClient
	accounts, _ := bingoCloudClient.GetSubAccounts()
	for _, account := range accounts {
		t.Log(account)
	}
	regions := bingoCloudClient.GetIRegions()
	bingoCloudClient.debug = true
	vpcs, _ := regions[0].GetIVpcs()
	for _, vpc := range vpcs {
		tags, _ := vpc.GetTags()
		t.Log(vpc.GetId(), vpc.GetName(), tags)
	}
	//regs := bingoCloudClient.GetIRegions()
	//lbcs, err := regs[0].GetILoadBalancerCertificates()
	//t.Log(lbcs, err)
	//elb, err := regions[0].GetLoadBalancer("applb-FDBCC55C")
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//groups, err := regions[0].GetElbBackendGroups("")
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//for _, group := range groups {
	//	attrs, err := elb.region.GetElbBackendGroupAttributes(group.TargetGroupId)
	//	if err != nil {
	//		t.Error(err)
	//		return
	//	}
	//	t.Log(attrs)
	//}
	//lbl, err := elb.GetILoadBalancerListenerById("lbl-8BC21426")
	//t.Log(lbl.GetId(), err)

	//access, err := bingoCloudClient.listAccessKeys("CloudTest0711")
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//ak, sk := access.decryptKeys(bingoCloudClient.managerClient.secretKey)
	//t.Log(ak, sk)
}
