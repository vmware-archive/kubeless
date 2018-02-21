# How to implement a new Kubeless run time 

As of today, Kubeless supports the following languages:

* Ruby
* Node.js
* Python
* .NET

Each language interpreter is implemented as an image of a Docker container, dispatched by the Kubeless controller.

All the structures that implement  language support are coded in the file `langruntime.go`, located under the `<..>kuberless/pkg/langruntime` directory of the project tree repository.

If you want to extend it and make another language available it is necessary to change the following components:

## 1. Update the kubeless-config configmap

### Custom runtime images
In this configmap there is a set of runtime images. You need to update this set with an entry pointing to the repository of the Docker container image for the runtime of the new language.

Usually there are three entries - one for HTTP triggers, another for event based functions and another one for the Init container that will be used to install dependencies in the build process. If your new runtime implementation will support only HTTP triggers, then create only two entries as follows:

In the example below, a custom image for `dotnetcore` has been added. You can optionally add `imagePullSecrets` if they are necessary to pull the image from a private Docker registry.

```patch
runtime-images
...
+        - ID: "dotnetcore"
+		  versions:
+		  - name: "dotnetcore2"
+		    version: "2.0"
+			httpImage: "mydocker/kubeless-netcore:latest"
+           initImage: "mydocker/kubeless-netcore-build:latest"
+           imagePullSecrets:
+           - imageSecret: "Secret"
+         depName: "requirements.xml"
+         fileNameSuffix: ".cs"
``` 

### Custom Provisioning container Images
In the configmap there is a default provisioning container image. This image is used for init-container inside the function pods definition. One can specify a custom provisioning container by updating `provision-container-image` in the configmap

```patch
provision-container-image
...
+ Image: "mycustom/provision/image:latest"
+ ImageSecret: "Secret"
```

Restart the controller after updating the configmap.

## 2. Add the build instructions to include dependencies in the runtime

Each runtime has specific instructions to install its dependencies. These instructions are specified in the method `GetBuildContainer()`. About this method you should know:
 - The folder with the function and the dependency files is mounted at `depsVolume.MountPath`
 - Dependencies should be installed in the folder `runtimeVolume.MountPath`

## 3. Update the deployment to load requirements for the runtime image

Some languages require to specify an environment variable in order to load dependencies from a certain path. If that is the case, update the function `updateDeployment()` to include the required environment variable:

```patch
func UpdateDeployment(dpm *v1beta1.Deployment, depsPath, runtime string) {
	switch {
...
+	case strings.Contains(runtime, "ruby"):
+		dpm.Spec.Template.Spec.Containers[0].Env = append(dpm.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{
+			Name:  "GEM_HOME",
+			Value: path.Join(depsPath, "ruby/2.4.0"),
+		})
```

This function is called if there are requirements to be injected in your runtime or if it is a custom runtime.

## 4. Add examples

In order to demonstrate the usage of the new runtime it will be necessary to add at least three different examples:
 - GET Example: A simple example in which the function returns a "hello world" string or similar.
 - POST Example: Another example in which the function reads the received input and returns a response.
 - Deps Example: In this example the runtime should load an external library (installed via the build process) and produce a response.

 The examples should be added to the folder `examples/<language_id>/` and should be added as well to the Makefile present in `examples/Makefile`. Note that the target should be `get-<language_id>`, `post-<language_id>` and `get-<language_id>-deps` for three examples above.

## 5. Add tests

For each new runtime, there should be integration tests that deploys the three examples above and check that the function is successfully deployed and that the output of the function is the expected one. For doing so add the counterpart `get-<language_id>-verify`, `post-<language_id>-verify` and `get-<language_id>-deps-verify` in the `examples/Makefile` and enable the execution of these tests in the script `test/integration-tests.bats`:

```patch
...
+@test "Test function: get-dotnetcore" {
+  test_kubeless_function get-dotnetcore
+}
...
+@test "Test function: post-dotnetcore" {
+  test_kubeless_function post-dotnetcore
+}
...
+@test "Test function: get-dotnetcore-deps" {
+  test_kubeless_function get-dotnetcore-deps
+}
```

Unit test are encouraged but not required. If there is a specific behaviour that needs testing (and it is not covered by the integration tests) you can add tests at `pkg/langruntime/langruntime_test.go`.

## Conclusion

Once you have followed the above steps send PR to the kubeless project, the CI system will pick your changes and test them. Once the PR is accepted by one of the maintainers the changes will be merged and included in the next release of Kubeless. 

Stay tuned for future documentations on these additional steps! :-)

