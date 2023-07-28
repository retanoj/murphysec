package conan

import (
	"encoding/json"
	"github.com/murphysecurity/murphysec/errors"
	"github.com/murphysecurity/murphysec/model"
	"os"
	"strings"
)

type _ConanInfo interface {
	ReadFromFile(path string) error
	Tree() (*model.Dependency, error)
}

// **** for conan version 1 ****
type _ConanInfoJsonItem struct {
	RequiredBy  []string `json:"required_by"`
	DisplayName string   `json:"display_name"`
}

type _ConanInfoJsonFile []_ConanInfoJsonItem

func (t *_ConanInfoJsonFile) ReadFromFile(path string) error {
	data, e := os.ReadFile(path)
	if e != nil {
		return errors.WithCause(ErrReadConanJsonFail, e)
	}
	if e := json.Unmarshal(data, &t); e != nil {
		return errors.WithCause(ErrReadConanJsonFail, e)
	}
	return nil
}

func (t *_ConanInfoJsonFile) Tree() (*model.Dependency, error) {
	var rootName string
	for _, it := range *t {
		if len(it.RequiredBy) == 0 && rootName == "" {
			rootName = it.DisplayName
			break
		}
	}
	if rootName == "" {
		return nil, ErrRootNodeNotFound
	}
	var depGraph = map[string][]string{}
	for _, it := range *t {
		for _, requiredBy := range it.RequiredBy {
			depGraph[requiredBy] = append(depGraph[requiredBy], it.DisplayName)
		}
	}
	return _tree(rootName, depGraph, map[string]bool{}), nil
}
// **** for conan version 1 end ****

// **** for conan version 2 ****
type _Conan2InfoJsonItem struct {
	Binary       string
	Context      string
	Label        string
	Dependencies map[string]struct {
		Ref     string
		Run     string
		Libs    string
		Skip    string
		Test    string
		Force   string
		Direct  string
		Build   string
		Headers string
		Visible string
	}
}

type _Conan2InfoJsonFile map[string]_Conan2InfoJsonItem

func (t *_Conan2InfoJsonFile) ReadFromFile(path string) error {
	data, e := os.ReadFile(path)
	if e != nil {
		return errors.WithCause(ErrReadConanJsonFail, e)
	}

	var file struct {
		Graph struct {
			Nodes _Conan2InfoJsonFile
		}
	}
	if e := json.Unmarshal(data, &file); e != nil {
		return errors.WithCause(ErrReadConanJsonFail, e)
	}
	*t = file.Graph.Nodes

	return nil
}

func (t *_Conan2InfoJsonFile) Tree() (*model.Dependency, error) {
	var rootName string
	var depGraph = map[string][]string{}
	for key, it := range *t {
		if key == "0" {
			rootName = it.Label
		}

		if it.Binary == "Skip" {
			continue
		}

		for _, deps := range it.Dependencies {
			if deps.Skip == "True" ||
				deps.Direct == "False" ||
				(deps.Headers == "False" && deps.Libs == "False") {
				continue
			}
			depGraph[it.Label] = append(depGraph[it.Label], deps.Ref)
		}
	}
	return _tree(rootName, depGraph, map[string]bool{}), nil
}

// **** for conan version 2 end ****
func _tree(name string, g map[string][]string, visitedName map[string]bool) *model.Dependency {
	if visitedName[name] {
		return nil
	}
	visitedName[name] = true
	defer delete(visitedName, name)
	r := strings.SplitN(name, "/", 2)
	d := &model.Dependency{
		Name:         r[0],
		Version:      "",
		Dependencies: nil,
	}
	if len(r) > 1 {
		d.Version = r[1]
	}
	for _, it := range g[name] {
		t := _tree(it, g, visitedName)
		if t == nil {
			continue
		}
		d.Dependencies = append(d.Dependencies, *t)
	}
	return d
}
