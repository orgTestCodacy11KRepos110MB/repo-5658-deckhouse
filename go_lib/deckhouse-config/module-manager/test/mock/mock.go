/*
Copyright 2022 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mock

import (
	"github.com/flant/addon-operator/pkg/module_manager"
	"github.com/flant/addon-operator/pkg/utils"
	"github.com/flant/addon-operator/pkg/values/validation"
	"k8s.io/utils/pointer"
)

var (
	EnabledByBundle  = pointer.Bool(true)
	EnabledByModule  = pointer.Bool(true)
	EnabledByScript  = pointer.Bool(true)
	DisabledByBundle = pointer.Bool(false)
	DisabledByModule = pointer.Bool(false)
	DisabledByScript = pointer.Bool(false)
)

// NewModuleManager returns mocked ModuleManager to test hooks
// without running values validations.
func NewModuleManager(modules ...ModuleMock) *ModuleManagerMock {
	// Index input list of modules.
	modulesMap := map[string]*module_manager.Module{}
	enabledModules := map[string]struct{}{}
	for _, mod := range modules {
		modulesMap[mod.module.Name] = mod.module
		if mod.enabled == nil || *mod.enabled {
			enabledModules[mod.module.Name] = struct{}{}
		}
	}

	return &ModuleManagerMock{
		modules:        modulesMap,
		enabledModules: enabledModules,
	}
}

type ModuleManagerMock struct {
	module_manager.ModuleManager
	modules        map[string]*module_manager.Module
	enabledModules map[string]struct{}
}

func (m *ModuleManagerMock) IsModuleEnabled(name string) bool {
	_, has := m.enabledModules[name]
	return has
}

func (m *ModuleManagerMock) GetModule(name string) *module_manager.Module {
	mod, has := m.modules[name]
	if has {
		return mod
	}
	return nil
}

func (m *ModuleManagerMock) GetModuleNames() []string {
	names := make([]string, 0)
	for modName := range m.modules {
		names = append(names, modName)
	}
	return names
}

func (m *ModuleManagerMock) GetValuesValidator() *validation.ValuesValidator {
	return nil
}

type ModuleMock struct {
	module  *module_manager.Module
	enabled *bool
}

func NewModule(name string, enabledByBundle *bool, enabledByScript *bool) ModuleMock {
	return ModuleMock{
		module: &module_manager.Module{
			Name: name,
			CommonStaticConfig: &utils.ModuleConfig{
				IsEnabled: enabledByBundle,
			},
			StaticConfig: &utils.ModuleConfig{
				IsEnabled: nil,
			},
			State: &module_manager.ModuleState{},
		},
		enabled: enabledByScript,
	}
}