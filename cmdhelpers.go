package main

import (
	"fmt"
	"strings"
	"sync"
	"webrecon/core"

	"github.com/google/uuid"
)

// ------------------------------- Variable Generators -------------------------------------------
// -------------------------------- vars for initial recon
func (p *Project) genOutfile(c *core.Cmd) string {
	c.OutputFile = `'` + p.DataDir + `/` + c.Name + `-` + uuid.NewString() + `'`
	return c.OutputFile
}

func (p *Project) genRootDomsCSV(c *core.Cmd) string {
	return `'` + strings.Join(p.RootDoms, ",") + `'`
}

func (p *Project) genRootDomsFile(c *core.Cmd) string {
	fname := p.DataDir + `/RootDoms-` + uuid.NewString()
	core.WriteSliceToFile(p.RootDoms, fname)
	return `'` + fname + `'`
}

// ------------------------------- Vars for flyover
func (p *Project) exampleVarFunc(c *core.Cmd) string {
	return "test" + p.Name + c.Name
}

func (p *Project) genAllFile(c *core.Cmd) string {
	fname := p.DataDir + `/all-targets-` + uuid.NewString()
	a := p.Scope.GetInScopeIPs()
	for key := range p.DNSMap {
		a = append(a, key)
	}
	core.WriteSliceToFile(p.RootDoms, fname)
	p.ResultsPath = fname
	return `'` + fname + `'`
}

func (p *Project) genIPFile(c *core.Cmd) string {
	fname := p.DataDir + `/ip-targets-` + uuid.NewString()
	core.WriteSliceToFile(p.Scope.GetInScopeIPs(), fname)
	p.ResultsPath = fname
	return `'` + fname + `'`
}

func (p *Project) genDomsFile(c *core.Cmd) string {
	fname := p.DataDir + `/domian-targets-` + uuid.NewString()
	var d []string
	for key := range p.DNSMap {
		d = append(d, key)
	}
	core.WriteSliceToFile(d, fname)
	p.ResultsPath = fname
	return `'` + fname + `'`
}

func (p *Project) genOutputDir(c *core.Cmd) string {
	return p.DataDir + `/` + `aquatone/`
}

//----------------------------------- runner callbacks -------------------------------------------
func (p *Project) exampleCallback(c core.Cmd) error {
	fmt.Println(p.Name, c.CmdLine)
	return nil
}

func (p *Project) aqCallback(c core.Cmd) error {
	fmt.Println(c.Output)
	return nil
}

func (p *Project) domainsCallback(c core.Cmd) error {
	var wgDoms = new(sync.WaitGroup)
	var wgDomCnt int
	const maxResolves = 5
	var mutex = &sync.Mutex{}

	doms, err := core.ReadLines(strings.ReplaceAll(c.OutputFile, "'", ""))
	if err != nil {
		core.Eprint(err)
		return err
	}
	doms = core.UniqueSlice(doms)
	wgDomCnt = 0
	for _, dom := range doms {

		if wgDomCnt >= maxResolves {

			wgDoms.Wait()
		}
		wgDoms.Add(1)
		wgDomCnt++
		// go func routine for parrallel resolve in ParseDomains
		go func(dom string, p *Project) {
			defer wgDoms.Done()
			core.Dprint("resolving:", dom)
			ok, ips := p.Scope.IsDNSInScope(dom)
			if ok {
				mutex.Lock()
				p.Targets = append(p.Targets, dom)
				p.DNSMap[dom] = append(p.DNSMap[dom], ips...)
				mutex.Unlock()
			}
			wgDomCnt--
		}(dom, p)
	}
	wgDoms.Wait()
	return nil
}
