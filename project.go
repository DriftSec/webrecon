package main

import (
	"errors"
	"net"
	"strings"
	"sync"
	"webrecon/core"
)

type DNStoIPMap map[string][]string

type Project struct {
	Name             string
	Scope            core.Scope
	RootDoms         []string
	DNSMap           DNStoIPMap
	DataDir          string
	Targets          []string
	ResultsPath      string
	ReconVars        core.VarMap
	ReconCallbacks   core.CallBacks
	FlyoverVars      core.VarMap
	FlyoverCallbacks core.CallBacks
	MaxThreads       int
}

func validateProject(p *Project) error {
	if p.Name == "" {
		return errors.New("project requires name")
	}
	if p.DataDir == "" {
		return errors.New("no data directory specified")
	}
	if len(p.Scope.Ranges) <= 0 {
		return errors.New("no scope specified (Scope.Ranges)")
	}
	if !strings.HasSuffix(p.DataDir, "/") {
		p.DataDir = p.DataDir + "/"
	}
	return nil
}

func NewProject() (Project, error) {
	var p Project
	p.DNSMap = make(DNStoIPMap)
	return p, nil
}

func (p *Project) StartRecon() error {
	err := validateProject(p)
	if err != nil {
		core.Panic(err)
	}

	p.mapHostnames()

	recontasks := core.NewCmdRunner()
	recontasks.CallBacks = p.ReconCallbacks

	recontasks.VarMap = p.ReconVars

	recontasks.MaxThreads = p.MaxThreads

	err = recontasks.Run(c.Recon.TargetID)
	if err != nil {
		core.Eprint(err)
	}

	// start flyover
	fr := core.NewCmdRunner()
	fr.CallBacks = p.FlyoverCallbacks

	fr.VarMap = p.FlyoverVars
	fr.MaxThreads = p.MaxThreads

	err = fr.RunWait(c.Recon.Flyover)
	if err != nil {
		core.Eprint(err)
	}
	return nil
}

func (p *Project) mapHostnames() {
	var wgDoms = new(sync.WaitGroup)
	var wgDomCnt int
	const maxResolves = 5
	var mutex = &sync.Mutex{}
	ips := p.Scope.GetInScopeIPs()
	wgDomCnt = 0

	for _, ip := range ips {
		if wgDomCnt >= maxResolves {
			wgDoms.Wait()
		}
		wgDoms.Add(1)
		wgDomCnt++
		go func(ip string, p *Project) {
			defer wgDoms.Done()
			core.Dprint("resolving:", ip)
			hosts, err := net.LookupAddr(ip)
			if err == nil {
				for _, dom := range hosts {
					dom = strings.TrimSuffix(dom, ".")
					mutex.Lock()
					p.DNSMap[dom] = append(p.DNSMap[dom], ip)
					mutex.Unlock()
				}
			}
			wgDomCnt--
		}(ip, p)
	}
	wgDoms.Wait()

}
