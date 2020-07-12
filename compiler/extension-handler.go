// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compiler

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"

	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	ext_plugin "github.com/googleapis/gnostic/extensions"
	yaml "gopkg.in/yaml.v3"
)

// ExtensionHandler describes a binary that is called by the compiler to handle specification extensions.
type ExtensionHandler struct {
	Name string
}

// HandleExtension calls a binary extension handler.
func HandleExtension(context *Context, in *yaml.Node, extensionName string) (bool, *any.Any, error) {
	handled := false
	var errFromPlugin error
	var outFromPlugin *any.Any

	if context != nil && context.ExtensionHandlers != nil && len(*(context.ExtensionHandlers)) != 0 {
		for _, customAnyProtoGenerator := range *(context.ExtensionHandlers) {
			outFromPlugin, errFromPlugin = customAnyProtoGenerator.handle(in, extensionName)
			if outFromPlugin == nil {
				continue
			} else {
				handled = true
				break
			}
		}
	}
	log.Printf("%s %t %+v", extensionName, handled, outFromPlugin)
	return handled, outFromPlugin, errFromPlugin
}

func (extensionHandlers *ExtensionHandler) handle(in *yaml.Node, extensionName string) (*any.Any, error) {
	if extensionHandlers.Name != "" {
		binary, _ := yaml.Marshal(in)

		request := &ext_plugin.ExtensionHandlerRequest{}

		version := &ext_plugin.Version{}
		version.Major = 0
		version.Minor = 1
		version.Patch = 0
		request.CompilerVersion = version

		request.Wrapper = &ext_plugin.Wrapper{}

		request.Wrapper.Version = "v2"
		request.Wrapper.Yaml = string(binary)
		request.Wrapper.ExtensionName = extensionName

		requestBytes, _ := proto.Marshal(request)
		cmd := exec.Command(extensionHandlers.Name)
		log.Printf("calling %s", extensionHandlers.Name)
		log.Printf("with request %+v", request)
		cmd.Stdin = bytes.NewReader(requestBytes)
		output, err := cmd.Output()
		log.Printf("output %+v", output)
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
			return nil, err
		}
		response := &ext_plugin.ExtensionHandlerResponse{}
		err = proto.Unmarshal(output, response)
		if err != nil {
			fmt.Printf("Error: %+v\n", err)
			fmt.Printf("%s\n", string(output))
			return nil, err
		}
		log.Printf("response %+v", response)
		if !response.Handled {
			return nil, nil
		}
		if len(response.Error) != 0 {
			message := fmt.Sprintf("Errors when parsing: %+v for field %s by vendor extension handler %s. Details %+v", in, extensionName, extensionHandlers.Name, strings.Join(response.Error, ","))
			return nil, errors.New(message)
		}
		return response.Value, nil
	}
	return nil, nil
}
