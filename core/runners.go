package core

import (
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

type CmdRunner struct {
	CallBacks  CallBacks // CallBacks is a map[string]CbFunc, used to set callbacks for runners
	VarMap     VarMap    // VarMap is a map[string]String, used to set replacement variables for runners.
	MaxThreads int       // MaxThreads sets the max number of concurrent threads for this CmdRunner
	RunningQ   Queue     // RunningQ is a map[int]Cmd of currently running Cmds
	WaitingQ   Queue     // WaitingQ is a map[int]Cmd of Cmds currently in the wait Queue
}

type Runners []Cmd
type Cmd struct {
	Name       string `yaml:"name"`
	CmdLine    string `yaml:"cmdline"`
	CallBack   string `yaml:"callback"`
	Status     string
	Output     string
	OutputFile string
	QID        int
}

type Queue map[int]Cmd

type CbFunc func(c Cmd) error
type CallBacks map[string]CbFunc

type VarFunc func(c *Cmd) string
type VarMap map[string]VarFunc

var wg sync.WaitGroup

// NewCmdRunner initializes a the module, and returns a CmdRunner
func NewCmdRunner() CmdRunner {
	var ret CmdRunner
	ret.CallBacks = make(CallBacks)
	ret.VarMap = make(VarMap)
	ret.RunningQ = make(Queue)
	ret.WaitingQ = make(Queue)
	return ret
}

func (c *CmdRunner) validateRunner(r Runners) error {
	for _, i := range r {
		if i.CallBack != "none" {
			if _, ok := c.CallBacks[i.CallBack]; !ok {
				return errors.New(i.CallBack + " is not in CmdRunner.CallBacks")
			}
		}
		re := regexp.MustCompile("{{ .(.*?) }}")
		matches := re.FindAllStringSubmatch(i.CmdLine, -1)
		// fmt.Println(matches)
		for _, m := range matches {
			if _, ok := c.VarMap[m[1]]; !ok {
				return errors.New(m[1] + " is not in CmdRunner.VarMap")
			}
		}
	}
	return nil
}

// Run runs commands with threads, and waits for them all to finish.  each thread will call its callback define in CallBacks upon completion.
func (c *CmdRunner) Run(r Runners) error {
	err := c.validateRunner(r)
	if err != nil {
		return err
	}
	for _, i := range r {
		c.runOrQueue(i)
	}
	wg.Wait()
	return nil
}

// RunWait runs commands one by one waiting for each to finish.
func (c *CmdRunner) RunWait(r Runners) error {
	err := c.validateRunner(r)
	if err != nil {
		return err
	}
	for _, i := range r {
		c.parseVars(&i)
		run := exec.Command("bash", "-c", i.CmdLine)
		out, err := run.CombinedOutput()
		if err != nil {
			i.Output = string(out)
			i.Status = "error"
		} else {
			i.Output = string(out)
			i.Status = "success"
		}
		if i.CallBack != "none" {
			err := c.CallBacks[i.CallBack](i)
			if err != nil {
				i.Output = "callback failed"
				i.Status = "error"
			}
		}

	}
	return nil
}

func (c *CmdRunner) parseVars(cmd *Cmd) {
	for k := range c.VarMap {
		cmd.CmdLine = strings.ReplaceAll(cmd.CmdLine, "{{ ."+k+" }}", c.VarMap[k](cmd))
	}
}

func (c *CmdRunner) startRunner(cmd Cmd) {
	c.parseVars(&cmd)
	run := exec.Command("bash", "-c", cmd.CmdLine)
	out, err := run.CombinedOutput()
	if err != nil {
		cmd.Output = string(out)
		cmd.Status = "error"
	} else {
		cmd.Output = string(out)
		cmd.Status = "success"
	}
	if cmd.CallBack != "none" {
		err := c.CallBacks[cmd.CallBack](cmd)
		if err != nil {
			cmd.Output = "callback failed"
			cmd.Status = "error"
		}
	}
	delete(c.RunningQ, cmd.QID)
	c.doNextRunner()
	wg.Done()

}

func (c *CmdRunner) GetStatus() (Queue, Queue) {
	return c.RunningQ, c.WaitingQ
}

// GetNextRunner gets the QID of the next runner in the wait queue, handy for "next command" status lines
func (c *CmdRunner) GetNextRunner() Cmd {
	var curlow int
	for runid := range c.WaitingQ {
		if curlow == 0 {
			curlow = runid
		} else {
			if runid < curlow {
				curlow = runid
			}
		}
	}
	return c.WaitingQ[curlow]
}

func (c *CmdRunner) doNextRunner() {
	if len(c.RunningQ) >= c.MaxThreads {
		// Dprint("concurrency still maxed")
		return
	}
	if len(c.WaitingQ) < 1 {
		// Dprint("nothing in wait queue")
		return
	}

	n := c.GetNextRunner()

	delete(c.WaitingQ, n.QID)
	nid := newQueueID(&c.RunningQ)
	c.RunningQ[nid] = n
	n.QID = nid
	n.Status = "running"
	Dprint("starting next:", nid)
	wg.Add(1)
	go c.startRunner(n)
}

func (c *CmdRunner) runOrQueue(cmd Cmd) {
	if len(c.RunningQ) < c.MaxThreads {
		nid := newQueueID(&c.RunningQ)
		Dprint("Starting Thread:", nid)
		c.RunningQ[nid] = cmd
		cmd.QID = nid
		cmd.Status = "running"
		wg.Add(1)
		go c.startRunner(cmd)
	} else {
		nid := newQueueID(&c.WaitingQ)
		cmd.QID = nid
		cmd.Status = "queued"
		c.WaitingQ[nid] = cmd
		Dprint("Queueing Thread:", nid)
	}
}

func newQueueID(daQ *Queue) int {
	var ids []int
	for key := range *daQ {
		ids = append(ids, key)
	}
	if len(ids) == 0 {
		ids = append(ids, 1)
	}
	var total int
	for j := 1; j < total; j++ {
		if ids[0] < ids[j] {
			ids[0] = ids[j]
		}

	}
	return ids[0] + 1
}
