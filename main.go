package main

import (
	"webrecon/core"
)

var c core.Config

func main() {

	// var c core.Config
	err := c.Init("./config.yaml")
	if err != nil {
		core.Panic("Failed parse config!!!")
	}
	core.Debug = c.General.Debug
	core.Errors = c.General.Errors

	p, err := NewProject()
	if err != nil {
		core.Eprint(err)
	}
	p.Name = "test"
	p.DataDir = "/tmp/data/" + p.Name
	p.Scope = core.Scope{
		Ranges:   []string{"192.168.56.*"},
		Excludes: []string{"192.168.56.11-250"},
	}
	p.RootDoms = []string{"test.com", "admin.test.com"}
	p.MaxThreads = 5
	p.ReconVars = core.VarMap{
		"OutFile":      p.genOutfile,
		"RootDomsCSV":  p.genRootDomsCSV,
		"RootDomsFile": p.genRootDomsFile,
		"IPFile":       p.exampleVarFunc,
	}
	p.ReconCallbacks = core.CallBacks{
		"domains": p.domainsCallback,
	}

	p.FlyoverVars = core.VarMap{
		"OutDir":     p.genOutputDir,
		"IPFile":     p.genIPFile,
		"DomsFile":   p.genDomsFile,
		"DomsIPFile": p.genAllFile,
	}
	p.FlyoverCallbacks = core.CallBacks{
		"aq": p.aqCallback,
	}

	p.StartRecon()

}
