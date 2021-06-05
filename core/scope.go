package core

import (
	"errors"
	"net"
	"strconv"
	"strings"

	netaddr "github.com/xgfone/netaddr"
)

type IPs []string

type Scope struct {
	Ranges   IPs
	Excludes IPs
}

// GetInScopeIPs returns a slice of all IPs that meet the scope criteria.
func (s Scope) GetInScopeIPs() IPs {
	var tmp, tmpx IPs
	for _, i := range s.Ranges {

		li, err := expandNmapRange(i)
		if err != nil {
			continue
		}
		tmp = append(tmp, li...)

	}
	for _, x := range s.Excludes {
		lx, err := expandNmapRange(x)
		if err != nil {
			continue
		}
		tmpx = append(tmpx, lx...)
	}
	return removeExcludedIPs(tmp, tmpx)
}

// IsIPInscope compares an address to lists of in scope and excluded ips (IPv4 or IPv6), and returns true if the address is in scope.
func (s Scope) IsIPInscope(ipaddr string) bool {
	i := s.Ranges
	x := s.Excludes
	result := false
	for _, a := range i {
		if ipaddr == a || rangeContains(ipaddr, a) {
			result = true
			break
		}
	}
	for _, b := range x {
		if ipaddr == b || rangeContains(ipaddr, b) {
			return false
		}
	}
	return result
}

// IsDNSInScope resolves a DNS name and checks it against the current scope, returns bool, and slice of inscope ips that resolved
func (s Scope) IsDNSInScope(name string) (bool, []string) {
	var iplist []string
	ips, err := net.LookupHost(name)
	if err != nil {
		return false, []string{}
	}
	for _, ip := range ips {
		if s.IsIPInscope(ip) {
			iplist = append(iplist, ip)
		}
	}
	return len(iplist) > 0, iplist
}

// rangeContains returns true if an address is in a subnet (IPv4 or IPv6).
func rangeContains(ipaddr string, subnet string) bool {
	net_subnet, err := netaddr.NewIPNetwork(subnet) // subnet
	if err != nil {                                 // may be nmap style  .* or .1-255
		try, err := expandNmapRange(subnet)
		if err != nil {
			return false
		}
		return sliceContains(try, ipaddr)
	}
	net_addr, err := netaddr.NewIPNetwork(ipaddr) // address to check
	if err != nil {
		return false
	}
	return net_subnet.Contains(net_addr)
}

//expandNmapRange expands nmap style IPv4 network ranges to a []string slice of ips.  ranges can end with .* or .1-255, ipv6 just returns the orignial.
func expandNmapRange(iprange string) ([]string, error) {
	var list []string
	if strings.Contains(iprange, ":") {
		return []string{iprange}, nil //ipv6 just send it back
	} else {
		if checkIPAddress(iprange) { // 192.168.56.1
			list := append(list, iprange)
			return list, nil
		}
		if strings.Contains(iprange, "/") { // 192.168.56.0/24
			ip := netaddr.MustNewIPNetwork(iprange)
			ip.Walk(func(ip netaddr.IPAddress) {
				list = append(list, ip.String())
			})
			return list, nil
		}
		if strings.Contains(iprange, "*") { // 192.168.56.*
			for i := 1; i < 256; i++ {
				list = append(list, strings.Replace(iprange, "*", strconv.Itoa(i), 1))
			}
			return list, nil
		}
		if strings.Contains(iprange, "-") { // 192.168.56.3-50
			tmp := strings.Split(iprange, ".")
			if len(tmp) != 4 {
				return []string{}, errors.New("ip parse error")
			}
			a := tmp[0]
			b := tmp[1]
			c := tmp[2]
			r := strings.Split(tmp[3], "-")
			if len(r) != 2 {
				return []string{}, errors.New("ip parse error")
			}
			low, _ := strconv.Atoi(r[0])
			high, _ := strconv.Atoi(r[1])
			for i := low; i <= high; i++ {
				list = append(list, a+"."+b+"."+c+"."+strconv.Itoa(i))
			}
			return list, nil
		}
	}
	return []string{}, errors.New("ip parse error")

}

// removeExcludedIPs compares a slice of ips, with a slice of excluded ips, and returns a slice of only in scope IPs
func removeExcludedIPs(inscope []string, outscope []string) []string {
	var list []string
OUTER:
	for _, curip := range inscope {
		for _, curXip := range outscope {
			if curip == curXip {
				continue OUTER
			}
		}
		list = append(list, curip)
	}
	return list
}

func sliceContains(lst []string, word string) bool {
	for _, v := range lst {
		if v == word {
			return true
		}
	}
	return false
}

// // CheckIPAddress validates an IPv4 or IPv6 address. returns true if valid
func checkIPAddress(ip string) bool {
	return net.ParseIP(ip) != nil
}
