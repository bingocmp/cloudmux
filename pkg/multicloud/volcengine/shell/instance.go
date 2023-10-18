// Copyright 2023 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shell

import (
	"context"
	"fmt"

	"yunion.io/x/cloudmux/pkg/cloudprovider"
	"yunion.io/x/cloudmux/pkg/multicloud/volcengine"
	"yunion.io/x/pkg/util/shellutils"
)

func init() {
	type InstanceListOptions struct {
		Id        []string `help:"IDs of instances to show"`
		Zone      string   `help:"Zone ID"`
		Limit     int      `help:"page size"`
		NextToken string   `help:"next token"`
	}
	shellutils.R(&InstanceListOptions{}, "instance-list", "List intances", func(cli *volcengine.SRegion, args *InstanceListOptions) error {
		instances, _, e := cli.GetInstances(args.Zone, args.Id, args.Limit, args.NextToken)
		if e != nil {
			return e
		}
		printList(instances, 0, 0, 0, nil)
		return nil
	})

	type InstanceDiskOperationOptions struct {
		ID   string `help:"instance ID"`
		DISK string `help:"disk ID"`
	}

	shellutils.R(&InstanceDiskOperationOptions{}, "instance-attach-disk", "Attach a disk to instance", func(cli *volcengine.SRegion, args *InstanceDiskOperationOptions) error {
		err := cli.AttachDisk(args.ID, args.DISK)
		if err != nil {
			return err
		}
		return nil
	})

	shellutils.R(&InstanceDiskOperationOptions{}, "instance-detach-disk", "Detach a disk to instance", func(cli *volcengine.SRegion, args *InstanceDiskOperationOptions) error {
		err := cli.DetachDisk(args.ID, args.DISK)
		if err != nil {
			return err
		}
		return nil
	})

	type InstanceOperationOptions struct {
		ID string `help:"instance ID"`
	}

	shellutils.R(&InstanceOperationOptions{}, "instance-start", "Start a instance", func(cli *volcengine.SRegion, args *InstanceOperationOptions) error {
		err := cli.StartVM(args.ID)
		if err != nil {
			return err
		}
		return nil
	})

	shellutils.R(&InstanceOperationOptions{}, "instance-delete", "Delete a instance", func(cli *volcengine.SRegion, args *InstanceOperationOptions) error {
		err := cli.DeleteVM(args.ID)
		if err != nil {
			return err
		}
		return nil
	})

	type InstanceStopOptions struct {
		ID           string `help:"instance ID"`
		Force        bool   `help:"Force stop instance"`
		StopCharging bool   `help:"Stop Charging"`
	}

	shellutils.R(&InstanceStopOptions{}, "instance-stop", "Stop a instance", func(cli *volcengine.SRegion, args *InstanceStopOptions) error {
		err := cli.StopVM(args.ID, args.Force, args.StopCharging)
		if err != nil {
			return err
		}
		return nil
	})

	type InstanceDeployOptions struct {
		ID            string `help:"instance ID"`
		Name          string `help:"new instance name"`
		Hostname      string `help:"new hostname"`
		Keypair       string `help:"Keypair Name"`
		DeleteKeypair bool   `help:"Remove SSH keypair"`
		Password      string `help:"new password"`
		Description   string `help:"new instances description"`
	}

	shellutils.R(&InstanceDeployOptions{}, "instance-deploy", "Deploy keypair/password to a stopped virtual server", func(cli *volcengine.SRegion, args *InstanceDeployOptions) error {
		err := cli.DeployVM(args.ID, args.Name, args.Password, args.Keypair, args.DeleteKeypair, args.Description)
		if err != nil {
			return err
		}
		return nil
	})

	type InstanceChangeConfigOptions struct {
		ID           string `help:"instance ID"`
		InstanceType string `help:"instance type"`
	}

	shellutils.R(&InstanceChangeConfigOptions{}, "instance-change-config", "Deploy keypair/password to a stopped virtual server", func(cli *volcengine.SRegion, args *InstanceChangeConfigOptions) error {
		err := cli.ChangeConfig(args.ID, args.InstanceType)
		if err != nil {
			return err
		}
		return nil
	})

	type InstanceRebuildRootOptions struct {
		ID       string `help:"instance ID"`
		Image    string `help:"Image ID"`
		Password string `help:"pasword"`
		Keypair  string `help:"keypair name"`
		UserData string `help:"user data"`
	}

	shellutils.R(&InstanceRebuildRootOptions{}, "instance-rebuild-root", "Reinstall virtual server system image", func(cli *volcengine.SRegion, args *InstanceRebuildRootOptions) error {
		_, err := cli.ReplaceSystemDisk(context.Background(), args.ID, args.Image, args.Password, args.Keypair, args.UserData)
		if err != nil {
			return err
		}
		// volcengine does not return diskID
		fmt.Printf("Instance rebuild root done")
		return nil
	})

	type InstanceUpdatePasswordOptions struct {
		ID     string `help:"Instance ID"`
		PASSWD string `help:"new password"`
	}
	shellutils.R(&InstanceUpdatePasswordOptions{}, "instance-update-password", "Update instance password", func(cli *volcengine.SRegion, args *InstanceUpdatePasswordOptions) error {
		err := cli.UpdateInstancePassword(args.ID, args.PASSWD)
		return err
	})

	type InstanceSaveImageOptions struct {
		ID         string `help:"Instance ID"`
		IMAGE_NAME string `help:"Image name"`
		Notes      string `hlep:"Image desc"`
	}
	shellutils.R(&InstanceSaveImageOptions{}, "instance-save-image", "Save instance to image", func(cli *volcengine.SRegion, args *InstanceSaveImageOptions) error {
		opts := cloudprovider.SaveImageOptions{
			Name:  args.IMAGE_NAME,
			Notes: args.Notes,
		}
		image, err := cli.SaveImage(args.ID, &opts)
		if err != nil {
			return err
		}
		printObject(image)
		return nil
	})
}
