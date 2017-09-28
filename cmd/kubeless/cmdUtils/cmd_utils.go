/*
Copyright (c) 2016-2017 Bitnami

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

package cmdUtils

import (
	"fmt"

	"github.com/kubeless/kubeless/pkg/utils"
	"github.com/spf13/pflag"
)

// ValidateDeploymentInputs validates that the combination of flags are valid for a deployment
func ValidateDeploymentInputs(flags *pflag.FlagSet) error {
	// Check if the runtime is supported
	runtime, _ := flags.GetString("runtime")
	if len(runtime) != 0 {
		_, _, _, _, err := utils.GetRuntimeProperties(runtime, "")
		if err != nil {
			return err
		}
	}
	// Check if runtime image is given the rest of flags are compatible
	if ri, _ := flags.GetString("runtime-image"); len(ri) != 0 {
		// TODO: Discomment if we don't want to support having custom images + kubeless build process
		// unsupportedOption := []string{"runtime", "handler", "from-file", "dependencies"}
		// for i := range unsupportedOption {
		// 	if o, _ := flags.GetString(unsupportedOption[i]); len(o) != 0 {
		// 		return fmt.Errorf("It is not possible to specify --%s when using a custom runtime image", unsupportedOption[i])
		// 	}
		// }
	} else {
		mandatoryFlags := []string{"runtime", "handler", "from-file"}
		for i := range mandatoryFlags {
			if f, _ := flags.GetString(mandatoryFlags[i]); len(f) == 0 {
				return fmt.Errorf("It is necessary to specify --%s deploying a function", mandatoryFlags[i])
			}
		}
		// Check that the handler has the correct format
		handler, err := flags.GetString("handler")
		if err != nil {
			return err
		}
		_, _, err = utils.SplitHandler(handler)
		if err != nil {
			return err
		}
	}
	return nil
}
