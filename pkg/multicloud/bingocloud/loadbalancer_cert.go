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
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin/pkg/log"
	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/errors"

	api "yunion.io/x/cloudmux/pkg/apis/compute"
	"yunion.io/x/cloudmux/pkg/cloudprovider"
	"yunion.io/x/cloudmux/pkg/multicloud"
)

type SElbCertificate struct {
	multicloud.SResourceBase
	BingoTags
	region *SRegion
	cert   *x509.Certificate

	Path                  string
	ServerCertificateName string
	ServerCertificateId   string
	UploadDate            time.Time
	OwnerId               string
	Expiration            time.Time
	NotAfter              time.Time
	NotBefore             time.Time

	CertificateBody  string
	CertificateChain string
}

func (self *SElbCertificate) GetId() string {
	return self.ServerCertificateId
}

func (self *SElbCertificate) GetName() string {
	return self.ServerCertificateName
}

func (self *SElbCertificate) GetGlobalId() string {
	return self.GetId()
}

func (self *SElbCertificate) GetStatus() string {
	return api.LB_STATUS_ENABLED
}

func (self *SElbCertificate) Refresh() error {
	icert, err := self.region.GetILoadBalancerCertificateById(self.GetId())
	if err != nil {
		return err
	}

	err = jsonutils.Update(self, icert)
	if err != nil {
		return err
	}

	return nil
}

func (self *SElbCertificate) IsEmulated() bool {
	return false
}

func (self *SElbCertificate) GetProjectId() string {
	return ""
}

func (self *SElbCertificate) Sync(name, privateKey, publickKey string) error {
	return cloudprovider.ErrNotSupported
}

func (self *SElbCertificate) Delete() error {
	return self.region.deleteElbCertificate(self.GetName())
}

func (self *SElbCertificate) GetCommonName() string {
	cert, err := self.ParsePublicKey()
	if err != nil {
		return ""
	}

	return cert.Issuer.CommonName
}

func (self *SElbCertificate) GetSubjectAlternativeNames() string {
	cert, err := self.ParsePublicKey()
	if err != nil {
		return ""
	}
	names, err := getOtherSANsFromX509Extensions(cert.Extensions)
	if err != nil {
		return ""
	}
	return strings.Join(names, ",")
}

func (self *SElbCertificate) GetFingerprint() string {
	publicKey := self.GetPublickKey()
	if len(publicKey) == 0 {
		return ""
	}

	_fp := sha1.Sum([]byte(publicKey))
	fp := fmt.Sprintf("sha1:% x", _fp)
	return strings.Replace(fp, " ", ":", -1)
}

func (self *SElbCertificate) GetExpireTime() time.Time {
	return self.Expiration
}

func (self *SElbCertificate) GetPublickKey() string {
	if self.CertificateBody == "" {
		params := map[string]string{}
		params["ServerCertificateName"] = self.ServerCertificateName
		resp, err := self.region.invoke("GetServerCertificate", params)
		if err != nil {
			log.Errorf("GetServerCertificate %v", err)
			return ""
		}
		err = resp.Unmarshal(&self.CertificateBody, "GetServerCertificateResult", "ServerCertificate", "CertificateBody")
		if err != nil {
			log.Errorf("Unmarshal serverCertificate %v", err)
			return ""
		}
	}
	return self.CertificateBody
}

func (self *SElbCertificate) GetPrivateKey() string {
	return ""
}

func (self *SElbCertificate) ParsePublicKey() (*x509.Certificate, error) {
	if self.cert != nil {
		return self.cert, nil
	}

	publicKey := self.GetPublickKey()
	if len(publicKey) == 0 {
		return nil, fmt.Errorf("SElbCertificate ParsePublicKey public key is empty")
	}

	block, _ := pem.Decode([]byte(publicKey))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "ParseCertificate")
	}

	self.cert = cert
	return cert, nil
}

func (self *SRegion) deleteElbCertificate(certName string) error {
	params := map[string]string{}
	params["ServerCertificateName"] = certName
	_, err := self.invoke("DeleteServerCertificate", params)
	return err
}

func (self *SRegion) listLoadBalancerCertificates() ([]SElbCertificate, error) {
	var params = map[string]string{
		"MaxItems": "3000",
	}

	resp, err := self.invoke("ListServerCertificates", params)
	if err != nil {
		return nil, errors.Wrap(err, "ListServerCertificates")
	}

	var ret struct {
		ServerCertificateMetadataList []SElbCertificate
	}
	err = resp.Unmarshal(&ret, "ListServerCertificatesResult")
	return ret.ServerCertificateMetadataList, err
}

func (self *SRegion) GetILoadBalancerCertificateById(certId string) (cloudprovider.ICloudLoadbalancerCertificate, error) {
	certs, err := self.GetILoadBalancerCertificates()
	if err != nil {
		return nil, errors.Wrap(err, "GetILoadBalancerCertificates")
	}

	for i := range certs {
		if certs[i].GetId() == certId {
			return certs[i], nil
		}
	}

	return nil, errors.Wrap(cloudprovider.ErrNotFound, "GetILoadBalancerCertificateById")
}

func (self *SRegion) CreateILoadBalancerCertificate(cert *cloudprovider.SLoadbalancerCertificate) (cloudprovider.ICloudLoadbalancerCertificate, error) {
	var params = map[string]string{
		"ServerCertificateName": cert.Name,
		"CertificateBody":       cert.Certificate,
		"PrivateKey":            cert.PrivateKey,
	}
	resp, err := self.invoke("UploadServerCertificate", params)
	if err != nil {
		return nil, errors.Wrap(err, "region.CreateILoadBalancerCertificate.UploadServerCertificate")
	}

	serverCertificateId := ""
	resp.Unmarshal(&serverCertificateId, "UploadServerCertificateResult", "ServerCertificateMetadata", "ServerCertificateId")

	return self.GetILoadBalancerCertificateById(serverCertificateId)
}

func (self *SRegion) GetILoadBalancerCertificates() ([]cloudprovider.ICloudLoadbalancerCertificate, error) {
	certs, err := self.listLoadBalancerCertificates()
	if err != nil {
		return nil, errors.Wrap(err, "listLoadBalancerCertificates")
	}

	icerts := make([]cloudprovider.ICloudLoadbalancerCertificate, 0)
	for i := range certs {
		if certs[i].OwnerId == self.client.user {
			certs[i].region = self
			icerts = append(icerts, &certs[i])
		}
	}

	return icerts, nil
}

func forEachSAN(extension []byte, callback func(tag int, data []byte) error) error {
	var seq asn1.RawValue
	rest, err := asn1.Unmarshal(extension, &seq)
	if err != nil {
		return err
	} else if len(rest) != 0 {
		return fmt.Errorf("x509: trailing data after X.509 extension")
	}
	if !seq.IsCompound || seq.Tag != 16 || seq.Class != 0 {
		return asn1.StructuralError{Msg: "bad SAN sequence"}
	}

	rest = seq.Bytes
	for len(rest) > 0 {
		var v asn1.RawValue
		rest, err = asn1.Unmarshal(rest, &v)
		if err != nil {
			return err
		}

		if err := callback(v.Tag, v.FullBytes); err != nil {
			return err
		}
	}

	return nil
}

func getOtherSANsFromX509Extensions(exts []pkix.Extension) ([]string, error) {
	var ret []string
	type otherName struct {
		TypeID asn1.ObjectIdentifier
		Value  asn1.RawValue
	}
	for _, ext := range exts {
		if !ext.Id.Equal([]int{2, 5, 29, 17}) {
			continue
		}
		err := forEachSAN(ext.Value, func(tag int, data []byte) error {
			if tag != 0 {
				return nil
			}

			var other otherName
			_, err := asn1.UnmarshalWithParams(data, &other, "tag:0")
			if err != nil {
				return fmt.Errorf("could not parse requested other SAN: %v", err)
			}
			var s string
			_, err = asn1.Unmarshal(other.Value.Bytes, &s)
			if err != nil {
				return err
			}
			ret = append(ret, s)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}
