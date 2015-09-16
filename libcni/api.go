// Copyright 2015 CoreOS, Inc.
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

package libcni

import (
	"strings"

	"github.com/appc/cni/pkg/invoke"
	"github.com/appc/cni/pkg/types"
)

type RuntimeConf struct {
	ContainerID string
	NetNS       string
	IfName      string
	Args        [][2]string
}

type NetworkConfig struct {
	Network *types.NetConf
	Bytes   []byte
}

type CNI interface {
	AddNetwork(net *NetworkConfig, rt *RuntimeConf) (*types.Result, error)
	DelNetwork(net *NetworkConfig, rt *RuntimeConf) error
}

type CNIConfig struct {
	Path []string
}

func (c *CNIConfig) AddNetwork(net *NetworkConfig, rt *RuntimeConf) (*types.Result, error) {
	return c.execPlugin("ADD", net, rt)
}

func (c *CNIConfig) DelNetwork(net *NetworkConfig, rt *RuntimeConf) error {
	_, err := c.execPlugin("DEL", net, rt)
	return err
}

// =====

func (c *CNIConfig) execPlugin(action string, conf *NetworkConfig, rt *RuntimeConf) (*types.Result, error) {
	pluginPath := invoke.FindInPath(conf.Network.Type, c.Path)

	args := &invoke.Args{
		Command:     action,
		ContainerID: rt.ContainerID,
		NetNS:       rt.NetNS,
		PluginArgs:  rt.Args,
		IfName:      rt.IfName,
		Path:        strings.Join(c.Path, ":"),
	}
	return invoke.ExecPlugin(pluginPath, conf.Bytes, args)
}
