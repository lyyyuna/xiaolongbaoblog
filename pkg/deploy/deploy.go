package deploy

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/lyyyuna/xiaolongbaoblog/pkg/config"
)

type Deploy struct {
	Repos []Repo
}

type Repo struct {
	Origin string
	Remote string
	Branch string
}

func NewDeploy(conf *config.Config) *Deploy {
	repos := make([]Repo, 0)

	for i, d := range conf.Deploys {
		if d.Type != "git" {
			log.Printf("%v is not git, not supported", d.Repository)
		}

		repos = append(repos, Repo{
			Remote: d.Repository,
			Branch: d.Branch,
			Origin: fmt.Sprintf("remote_%v", i),
		})
	}
	return &Deploy{
		Repos: repos,
	}
}

func (d *Deploy) DeployToServer(path string) {
	d.init(path)
	d.remoteAdd(path)
	d.add(path)
	d.commit(path)
	d.pushWithForce(path)
}

func (d *Deploy) init(path string) {
	cmd := exec.Command("git", "init", ".")
	cmd.Dir = path

	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))

	if err != nil {
		log.Fatalln("fail to git init")
	}
}

func (d *Deploy) remoteAdd(path string) {
	for _, r := range d.Repos {
		cmd := exec.Command("git", "remote", "add", r.Origin, r.Remote)
		cmd.Dir = path

		output, err := cmd.CombinedOutput()
		fmt.Println(string(output))

		if err != nil {
			log.Fatalln("fail to git remote add " + r.Remote)
		}
	}
}

func (d *Deploy) add(path string) {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = path

	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))

	if err != nil {
		log.Fatalln("fail to git add")
	}
}

func (d *Deploy) commit(path string) {
	cmd := exec.Command("git", "commit", "-m", "site updated: "+time.Now().String())
	cmd.Dir = path

	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))

	if err != nil {
		log.Fatalln("fail to git commit")
	}
}

func (d *Deploy) pushWithForce(path string) {
	for _, r := range d.Repos {
		cmd := exec.Command("git", "push", r.Origin, r.Branch, "-f")
		cmd.Dir = path

		output, err := cmd.CombinedOutput()
		fmt.Println(string(output))

		if err != nil {
			log.Fatalln("fail to git push to " + r.Remote)
		}
	}
}
