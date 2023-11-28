package bingocloud

import (
	"fmt"
	"strconv"
	"time"

	api "yunion.io/x/cloudmux/pkg/apis/compute"

	"yunion.io/x/jsonutils"

	"yunion.io/x/cloudmux/pkg/cloudprovider"
)

type SSnapshot struct {
	region       *SRegion
	SnapshotId   string
	SnapshotName string
	BackupId     string
	VolumeId     string
	Status       SnapshotStatusType
	StartTime    string
	Progress     string
	OwnerId      string
	VolumeSize   string
	IsBackup     bool
	IsRoot       bool
	IsHead       bool
	FileType     string
	FileSize     string
	Description  string
	StorageId    string
}

func (self *SSnapshot) GetId() string {
	return self.SnapshotId
}

func (self *SSnapshot) GetName() string {
	return self.SnapshotName
}

func (self *SSnapshot) GetGlobalId() string {
	return self.SnapshotId
}

func (self *SSnapshot) GetCreatedAt() time.Time {
	ct, _ := time.Parse("2006-01-02T15:04:05.000Z", self.StartTime)
	return ct
}

func (self *SSnapshot) GetDescription() string {
	return self.Description
}

func (self *SSnapshot) GetStatus() string {
	if self.Status == SnapshotStatusAccomplished {
		return api.SNAPSHOT_READY
	} else if self.Status == SnapshotStatusProgress {
		return api.SNAPSHOT_CREATING
	} else {
		return api.SNAPSHOT_FAILED
	}
}

func (self *SSnapshot) Refresh() error {
	newSnapshot, err := self.region.getSnapshots(self.SnapshotId, "")
	if err != nil {
		return err
	}
	if len(newSnapshot) == 1 {
		newSnapshot[0].region = self.region
		return jsonutils.Update(&self, newSnapshot[0])
	}
	return cloudprovider.ErrNotFound
}

func (self *SSnapshot) IsEmulated() bool {
	return false
}

func (self *SSnapshot) GetSysTags() map[string]string {
	return nil
}

func (self *SSnapshot) GetTags() (map[string]string, error) {
	return nil, nil
}

func (self *SSnapshot) SetTags(tags map[string]string, replace bool) error {
	return nil
}

func (self *SSnapshot) GetProjectId() string {
	return ""
}

func (self *SSnapshot) GetSizeMb() int32 {
	size, _ := strconv.Atoi(self.VolumeSize)
	return int32(size) * 1024
}

func (self *SSnapshot) GetDiskId() string {
	return self.VolumeId
}

func (self *SSnapshot) GetDiskType() string {
	if self.IsRoot {
		return api.DISK_TYPE_SYS
	}
	switch self.FileType {
	case "system":
		return api.DISK_TYPE_SYS
	case "data":
		return api.DISK_TYPE_DATA
	default:
		return api.DISK_TYPE_DATA
	}
}

func (self *SSnapshot) Delete() error {
	return self.region.deleteSnapshot(self.SnapshotId)
}

func (self *SRegion) createSnapshot(volumeId, storageId, name string, desc string, rootInfo map[string]string) (string, error) {
	params := map[string]string{}
	params["VolumeId"] = volumeId
	params["SnapshotName"] = name
	params["Description"] = desc

	if storageId != "" {
		params["StorageId"] = storageId
	}
	if rootInfo != nil {
		params["IsRoot"] = "true"
		params["RootInfo"] = jsonutils.Marshal(rootInfo).String()
	}

	resp, err := self.invoke("CreateSnapshot", params)
	if err != nil {
		return "", err
	}
	newId := ""
	err = resp.Unmarshal(&newId, "snapshotId")

	return newId, err
}

func (self *SRegion) getSnapshots(id, name string) ([]SSnapshot, error) {
	params := map[string]string{}
	if id != "" {
		params["SnapshotId.1"] = id
	}
	if name != "" {
		params["Filter.1.Name"] = name
	}
	params[fmt.Sprintf("Filter.%d.Name", 1)] = "owner-id"
	params[fmt.Sprintf("Filter.%d.Value.1", 1)] = self.client.user

	resp, err := self.invoke("DescribeSnapshots", params)
	if err != nil {
		return nil, err
	}

	var ret []SSnapshot
	_ = resp.Unmarshal(&ret, "snapshotSet")

	return ret, err
}

func (self *SRegion) deleteSnapshot(id string) error {
	params := map[string]string{}
	params["SnapshotId"] = id
	_, err := self.invoke("DeleteSnapshot", params)
	return err
}
