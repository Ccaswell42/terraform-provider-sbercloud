// Copyright 2022 Huawei Technologies Co.,Ltd.
//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package region

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	regionsFileEnv            = "HUAWEICLOUD_SDK_REGIONS_FILE"
	defaultRegionsFileDirName = ".huaweicloud"
	defaultRegionsFileName    = "regions.yaml"
)

var (
	profileOnce  sync.Once
	profileCache *ProfileCache
)

type ProfileCache struct {
	value map[string]*Region
}

func getProfileCache() *ProfileCache {
	profileOnce.Do(func() {
		profileCache = &ProfileCache{value: resolveProfile()}
	})

	return profileCache
}

func getRegionsFilePath() string {
	if path := os.Getenv(regionsFileEnv); path != "" {
		return path
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(dir, defaultRegionsFileDirName, defaultRegionsFileName)
}

func isPathExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func resolveProfile() map[string]*Region {
	result := make(map[string]*Region)

	path := getRegionsFilePath()
	if !isPathExist(path) {
		return result
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("failed to read file: '%s'\n%s", path, err.Error()))
	}

	var tempMap map[string][]map[string]string
	err = yaml.Unmarshal(bytes, &tempMap)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve file: '%s'\n%s", path, err.Error()))
	}

	for service, list := range tempMap {
		for _, m := range list {
			id := m["id"]
			endpoint := m["endpoint"]
			if id == "" || endpoint == "" {
				continue
			}
			result[strings.ToUpper(service)+id] = NewRegion(id, endpoint)
		}
	}
	return result
}

type ProfileProvider struct {
	serviceName string
}

func NewProfileProvider(serviceName string) *ProfileProvider {

	return &ProfileProvider{
		serviceName: strings.ToUpper(serviceName),
	}
}

func (p *ProfileProvider) GetRegion(regionId string) *Region {
	return getProfileCache().value[p.serviceName+regionId]
}
