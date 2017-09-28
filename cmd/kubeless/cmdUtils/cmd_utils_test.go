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
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateDeploymentInputs(t *testing.T) {
	var cmd = &cobra.Command{}
	cmd.Flags().StringP("runtime", "", "", "")
	cmd.Flags().StringP("handler", "", "", "")
	cmd.Flags().StringP("from-file", "", "", "")
	cmd.Flags().StringP("runtime-image", "", "test", "")
	// it should not throw an error if just the runtime is specified
	err := ValidateDeploymentInputs(cmd.Flags())
	if err != nil {
		t.Fatalf("Should not throw an error if the runtime image is not empty")
	}
	// it should not throw an error if the other commands are properly specified
	cmd = &cobra.Command{}
	cmd.Flags().StringP("runtime", "", "python2.7", "")
	cmd.Flags().StringP("handler", "", "test.foo", "")
	cmd.Flags().StringP("from-file", "", "test", "")
	cmd.Flags().StringP("runtime-image", "", "", "")
	err = ValidateDeploymentInputs(cmd.Flags())
	if err != nil {
		t.Fatalf("Should not throw an error if the correct parameters are specified")
	}
	// it should throw an error if the runtime is not supported
	cmd = &cobra.Command{}
	cmd.Flags().StringP("runtime", "", "python2.6", "")
	cmd.Flags().StringP("handler", "", "test.foo", "")
	cmd.Flags().StringP("from-file", "", "test", "")
	cmd.Flags().StringP("runtime-image", "", "", "")
	err = ValidateDeploymentInputs(cmd.Flags())
	if err == nil {
		t.Fatalf("Should throw an error if the runtime is not supported")
	}
	// it should throw an error if the options are incompatible
	cmd = &cobra.Command{}
	cmd.Flags().StringP("runtime", "", "python2.6", "")
	cmd.Flags().StringP("handler", "", "test.foo", "")
	cmd.Flags().StringP("from-file", "", "test", "")
	cmd.Flags().StringP("runtime-image", "", "test", "")
	err = ValidateDeploymentInputs(cmd.Flags())
	if err == nil {
		t.Fatalf("Should throw an error if all the options are specified")
	}
}
