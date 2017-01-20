// Copyright 2016 CNI authors
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

package libcni_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/containernetworking/cni/libcni"
	"github.com/containernetworking/cni/pkg/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Loading configuration from disk", func() {
	Describe("LoadConf", func() {
		var (
			configDir    string
			pluginConfig []byte
		)

		BeforeEach(func() {
			var err error
			configDir, err = ioutil.TempDir("", "plugin-conf")
			Expect(err).NotTo(HaveOccurred())

			pluginConfig = []byte(`{ "name": "some-plugin", "some-key": "some-value" }`)
			Expect(ioutil.WriteFile(filepath.Join(configDir, "50-whatever.conf"), pluginConfig, 0600)).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(configDir)).To(Succeed())
		})

		It("finds the network config file for the plugin of the given type", func() {
			netConfig, err := libcni.LoadConf(configDir, "some-plugin")
			Expect(err).NotTo(HaveOccurred())
			Expect(netConfig).To(Equal(&libcni.NetworkConfig{
				Network: &types.NetConf{Name: "some-plugin"},
				Bytes:   pluginConfig,
			}))
		})

		Context("when the config directory does not exist", func() {
			BeforeEach(func() {
				Expect(os.RemoveAll(configDir)).To(Succeed())
			})

			It("returns a useful error", func() {
				_, err := libcni.LoadConf(configDir, "some-plugin")
				Expect(err).To(MatchError("no net configurations found"))
			})
		})

		Context("when the config file is .json extension instead of .conf", func() {
			BeforeEach(func() {
				Expect(os.Remove(configDir + "/50-whatever.conf")).To(Succeed())
				pluginConfig = []byte(`{ "name": "some-plugin", "some-key": "some-value" }`)
				Expect(ioutil.WriteFile(filepath.Join(configDir, "50-whatever.json"), pluginConfig, 0600)).To(Succeed())
			})
			It("finds the network config file for the plugin of the given type", func() {
				netConfig, err := libcni.LoadConf(configDir, "some-plugin")
				Expect(err).NotTo(HaveOccurred())
				Expect(netConfig).To(Equal(&libcni.NetworkConfig{
					Network: &types.NetConf{Name: "some-plugin"},
					Bytes:   pluginConfig,
				}))
			})
		})

		Context("when there is no config for the desired plugin", func() {
			It("returns a useful error", func() {
				_, err := libcni.LoadConf(configDir, "some-other-plugin")
				Expect(err).To(MatchError(ContainSubstring(`no net configuration with name "some-other-plugin" in`)))
			})
		})

		Context("when a config file is malformed", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(configDir, "00-bad.conf"), []byte(`{`), 0600)).To(Succeed())
			})

			It("returns a useful error", func() {
				_, err := libcni.LoadConf(configDir, "some-plugin")
				Expect(err).To(MatchError(`error parsing configuration: unexpected end of JSON input`))
			})
		})

		Context("when the config is in a nested subdir", func() {
			BeforeEach(func() {
				subdir := filepath.Join(configDir, "subdir1", "subdir2")
				Expect(os.MkdirAll(subdir, 0700)).To(Succeed())

				pluginConfig = []byte(`{ "name": "deep", "some-key": "some-value" }`)
				Expect(ioutil.WriteFile(filepath.Join(subdir, "90-deep.conf"), pluginConfig, 0600)).To(Succeed())
			})

			It("will not find the config", func() {
				_, err := libcni.LoadConf(configDir, "deep")
				Expect(err).To(MatchError(HavePrefix("no net configuration with name")))
			})
		})
	})

	Describe("ConfFromFile", func() {
		Context("when the file cannot be opened", func() {
			It("returns a useful error", func() {
				_, err := libcni.ConfFromFile("/tmp/nope/not-here")
				Expect(err).To(MatchError(HavePrefix(`error reading /tmp/nope/not-here: open /tmp/nope/not-here`)))
			})
		})
	})

	Describe("LoadConfList", func() {
		var (
			configDir  string
			configList []byte
		)

		BeforeEach(func() {
			var err error
			configDir, err = ioutil.TempDir("", "plugin-conf")
			Expect(err).NotTo(HaveOccurred())

			configList = []byte(`{
  "name": "some-list",
  "cniVersion": "0.2.0",
  "plugins": [
    {
      "type": "host-local",
      "subnet": "10.0.0.1/24"
    },
    {
      "type": "bridge",
      "mtu": 1400
    },
    {
      "type": "port-forwarding",
      "ports": {"20.0.0.1:8080": "80"}
    }
  ]
}`)
			Expect(ioutil.WriteFile(filepath.Join(configDir, "50-whatever.conflist"), configList, 0600)).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(configDir)).To(Succeed())
		})

		It("finds the network config file for the plugin of the given type", func() {
			netConfigList, err := libcni.LoadConfList(configDir, "some-list")
			Expect(err).NotTo(HaveOccurred())
			Expect(netConfigList).To(Equal(&libcni.NetworkConfigList{
				Name:       "some-list",
				CNIVersion: "0.2.0",
				Plugins: []*libcni.NetworkConfig{
					{
						Network: &types.NetConf{Type: "host-local"},
						Bytes:   []byte(`{"subnet":"10.0.0.1/24","type":"host-local"}`),
					},
					{
						Network: &types.NetConf{Type: "bridge"},
						Bytes:   []byte(`{"mtu":1400,"type":"bridge"}`),
					},
					{
						Network: &types.NetConf{Type: "port-forwarding"},
						Bytes:   []byte(`{"ports":{"20.0.0.1:8080":"80"},"type":"port-forwarding"}`),
					},
				},
				Bytes: configList,
			}))
		})

		Context("when the config directory does not exist", func() {
			BeforeEach(func() {
				Expect(os.RemoveAll(configDir)).To(Succeed())
			})

			It("returns a useful error", func() {
				_, err := libcni.LoadConfList(configDir, "some-plugin")
				Expect(err).To(MatchError("no net configuration lists found"))
			})
		})

		Context("when there is no config for the desired plugin list", func() {
			It("returns a useful error", func() {
				_, err := libcni.LoadConfList(configDir, "some-other-plugin")
				Expect(err).To(MatchError(ContainSubstring(`no net configuration list with name "some-other-plugin" in`)))
			})
		})

		Context("when a config file is malformed", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(configDir, "00-bad.conflist"), []byte(`{`), 0600)).To(Succeed())
			})

			It("returns a useful error", func() {
				_, err := libcni.LoadConfList(configDir, "some-plugin")
				Expect(err).To(MatchError(`error parsing configuration list: unexpected end of JSON input`))
			})
		})

		Context("when the config is in a nested subdir", func() {
			BeforeEach(func() {
				subdir := filepath.Join(configDir, "subdir1", "subdir2")
				Expect(os.MkdirAll(subdir, 0700)).To(Succeed())

				configList = []byte(`{
  "name": "deep",
  "cniVersion": "0.2.0",
  "plugins": [
    {
      "type": "host-local",
      "subnet": "10.0.0.1/24"
    },
  ]
}`)
				Expect(ioutil.WriteFile(filepath.Join(subdir, "90-deep.conflist"), configList, 0600)).To(Succeed())
			})

			It("will not find the config", func() {
				_, err := libcni.LoadConfList(configDir, "deep")
				Expect(err).To(MatchError(HavePrefix("no net configuration list with name")))
			})
		})
	})

	Describe("ConfListFromFile", func() {
		Context("when the file cannot be opened", func() {
			It("returns a useful error", func() {
				_, err := libcni.ConfListFromFile("/tmp/nope/not-here")
				Expect(err).To(MatchError(HavePrefix(`error reading /tmp/nope/not-here: open /tmp/nope/not-here`)))
			})
		})
	})

	Describe("InjectConf", func() {
		var testNetConfig *libcni.NetworkConfig

		BeforeEach(func() {
			testNetConfig = &libcni.NetworkConfig{Network: &types.NetConf{Name: "some-plugin"},
				Bytes: []byte(`{ "name": "some-plugin" }`)}
		})

		Context("when function parameters are incorrect", func() {
			It("returns unmarshal error", func() {
				conf := &libcni.NetworkConfig{Network: &types.NetConf{Name: "some-plugin"},
					Bytes: []byte(`{ cc cc cc}`)}

				_, err := libcni.InjectConf(conf, "", nil)
				Expect(err).To(MatchError(HavePrefix(`unmarshal existing network bytes`)))
			})

			It("returns key  error", func() {
				_, err := libcni.InjectConf(testNetConfig, "", nil)
				Expect(err).To(MatchError(HavePrefix(`key value can not be empty`)))
			})

			It("returns newValue  error", func() {
				_, err := libcni.InjectConf(testNetConfig, "test", nil)
				Expect(err).To(MatchError(HavePrefix(`newValue must be specified`)))
			})
		})

		Context("when new string value added", func() {
			It("adds the new key & value to the config", func() {
				newPluginConfig := []byte(`{"name":"some-plugin","test":"test"}`)

				resultConfig, err := libcni.InjectConf(testNetConfig, "test", "test")
				Expect(err).NotTo(HaveOccurred())
				Expect(resultConfig).To(Equal(&libcni.NetworkConfig{
					Network: &types.NetConf{Name: "some-plugin"},
					Bytes:   newPluginConfig,
				}))
			})

			It("adds the new value for exiting key", func() {
				newPluginConfig := []byte(`{"name":"some-plugin","test":"changedValue"}`)

				resultConfig, err := libcni.InjectConf(testNetConfig, "test", "test")
				Expect(err).NotTo(HaveOccurred())

				resultConfig, err = libcni.InjectConf(resultConfig, "test", "changedValue")
				Expect(err).NotTo(HaveOccurred())

				Expect(resultConfig).To(Equal(&libcni.NetworkConfig{
					Network: &types.NetConf{Name: "some-plugin"},
					Bytes:   newPluginConfig,
				}))
			})

			It("adds existing key & value", func() {
				newPluginConfig := []byte(`{"name":"some-plugin","test":"test"}`)

				resultConfig, err := libcni.InjectConf(testNetConfig, "test", "test")
				Expect(err).NotTo(HaveOccurred())

				resultConfig, err = libcni.InjectConf(resultConfig, "test", "test")
				Expect(err).NotTo(HaveOccurred())

				Expect(resultConfig).To(Equal(&libcni.NetworkConfig{
					Network: &types.NetConf{Name: "some-plugin"},
					Bytes:   newPluginConfig,
				}))
			})

			It("adds sub-fields of NetworkConfig.Network to the config", func() {

				expectedPluginConfig := []byte(`{"dns":{"domain":"local","nameservers":["server1","server2"]},"name":"some-plugin","type":"bridge"}`)
				servers := []string{"server1", "server2"}
				newDNS := &types.DNS{Nameservers: servers, Domain: "local"}

				// inject DNS
				resultConfig, err := libcni.InjectConf(testNetConfig, "dns", newDNS)
				Expect(err).NotTo(HaveOccurred())

				// inject type
				resultConfig, err = libcni.InjectConf(resultConfig, "type", "bridge")
				Expect(err).NotTo(HaveOccurred())

				Expect(resultConfig).To(Equal(&libcni.NetworkConfig{
					Network: &types.NetConf{Name: "some-plugin", Type: "bridge", DNS: types.DNS{Nameservers: servers, Domain: "local"}},
					Bytes:   expectedPluginConfig,
				}))
			})
		})
	})
})
