package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Permission struct {
	Verb string `json:"verb"`
}

type Resource struct {
	Name        string
	Permissions []string
}

type Service struct {
	Name             string
	Resources        []*Resource
	WildcardResource *Resource
}

func LoadService(rbacPath, svcName string) (*Service, error) {
	data := make(map[string][]Permission)
	path := fmt.Sprintf("%s/%s.json", rbacPath, svcName)
	reader, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(reader)
	err = dec.Decode(&data)
	if err != nil {
		return nil, err
	}

	service := &Service{
		Name:      svcName,
		Resources: make([]*Resource, 0, len(data)),
	}

	for resName, perms := range data {
		res := &Resource{
			Name:        resName,
			Permissions: make([]string, 0, len(perms)),
		}

		for _, perm := range perms {
			if perm.Verb == "*" {
				continue
			}
			res.Permissions = append(res.Permissions, perm.Verb)
		}

		if resName == "*" {
			service.WildcardResource = res
		} else {
			service.Resources = append(service.Resources, res)
		}
	}

	return service, nil
}
