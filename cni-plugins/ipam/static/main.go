package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/pkg/errors"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
)

type Net struct {
	Name       string      `json:"name"`
	CNIVersion string      `json:"cniVersion"`
	IPAM       *IPAMConfig `json:"ipam"`

	RuntimeConfig struct {
		IPs []string `json:"ips,omitempty"`
	} `json:"runtimeConfig,omitempty"`

	Args *struct {
		A *IPAMArgs `json:"cni"`
	} `json:"args"`
}

type IPAMConfig struct {
	Name      string
	Type      string         `json:"type"`
	Routes    []*types.Route `json:"routes"`
	Addresses []Address      `json:"addresses,omitempty"`
	DNS       types.DNS      `json:"dns"`
}

type IPAMEnvArgs struct {
	types.CommonArgs
	IP      types.UnmarshallableString `json:"ip,omitempty"`
	GATEWAY types.UnmarshallableString `json:"gateway,omitempty"`
}

type IPAMArgs struct {
	IPs []string `json:"ips"`
}

type Address struct {
	AddressStr string `json:"address"`
	Gateway    net.IP `json:"gateway,omitempty"`
	Address    net.IPNet
	Version    string
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("static"))
}

func cmdAdd(args *skel.CmdArgs) error {
	ipamConf, confVersion, err := LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return errors.Wrap(err, "load ipam config")
	}

	result := &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		DNS:        ipamConf.DNS,
		Routes:     ipamConf.Routes,
	}

	for _, v := range ipamConf.Addresses {
		result.IPs = append(result.IPs, &current.IPConfig{
			Address: v.Address,
			Gateway: v.Gateway,
		},
		)
	}

	return types.PrintResult(result, confVersion)
}

func cmdCheck(args *skel.CmdArgs) error {
	ipamConf, _, err := LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	// Get PrevResult from stdin... store in RawPrevResult
	n, _, err := loadNetConf(args.StdinData)
	if err != nil {
		return err
	}

	if n.RawPrevResult == nil {
		return errors.New("required prevResult missing")
	}

	if err := version.ParsePrevResult(n); err != nil {
		return err
	}

	result, err := current.NewResultFromResult(n.PrevResult)
	if err != nil {
		return err
	}

	for _, rangeset := range ipamConf.Addresses {
		for _, ips := range result.IPs {
			// Ensure values are what we expect
			if rangeset.Address.IP.Equal(ips.Address.IP) {
				if rangeset.Gateway == nil {
					break
				} else if rangeset.Gateway.Equal(ips.Gateway) {
					break
				}
				return fmt.Errorf("static: Failed to match addr %v on interface %v", ips.Address.IP, args.IfName)
			}
		}
	}

	return nil
}

func cmdDel(_ *skel.CmdArgs) error {
	return nil
}

func LoadIPAMConfig(bytes []byte, envArgs string) (*IPAMConfig, string, error) {
	n := Net{}
	if err := json.Unmarshal(bytes, &n); err != nil {
		return nil, "", err
	}
	if n.IPAM == nil {
		return nil, "", fmt.Errorf("IPAM config missing 'ipam' key")
	}

	for i, ele := range n.IPAM.Addresses {
		if ele.Address.IP == nil {
			ip, addr, err := net.ParseCIDR(ele.AddressStr)
			if err != nil {
				return nil, "", fmt.Errorf("parse CIDR:%s error", ele.AddressStr)
			}
			n.IPAM.Addresses[i].Address = *addr
			n.IPAM.Addresses[i].Address.IP = ip
		}
	}

	n.IPAM.Name = n.Name

	return n.IPAM, n.CNIVersion, nil
}

func loadNetConf(bytes []byte) (*types.NetConf, string, error) {
	n := &types.NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, "", fmt.Errorf("failed to load netconf: %v", err)
	}
	return n, n.CNIVersion, nil
}
