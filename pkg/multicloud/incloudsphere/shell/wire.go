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

	"yunion.io/x/cloudmux/pkg/multicloud/incloudsphere"
)

func init() {
	type WireListOptions struct {
	}
	shellutils.R(&WireListOptions{}, "wire-list", "list wires", func(cli *incloudsphere.SRegion, args *WireListOptions) error {
		wires, err := cli.GetWires()
		if err != nil {
			return err
		}
		printList(wires, 0, 0, 0, []string{})
		return nil
	})

	type WireIdOptions struct {
		ID string
	}

	shellutils.R(&WireIdOptions{}, "wire-show", "show wire", func(cli *incloudsphere.SRegion, args *WireIdOptions) error {
		ret, err := cli.GetWire(args.ID)
		if err != nil {
			return err
		}
		printObject(ret)
		return nil
	})
}
