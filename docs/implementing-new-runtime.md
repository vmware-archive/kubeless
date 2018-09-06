# How to implement a new Kubeless run time 

As of today, Kubeless supports the following languages:

* Ruby
* Node.js
* Python
* PHP
* Golang

Each language interpreter is implemented as an image of a Docker container, dispatched by the Kubeless controller.

All the structures that implement  language support are coded in the file `langruntime.go`, located under the `<..>kuberless/pkg/langruntime` directory of the project tree repository.

If you want to extend it and make another language available it is necessary to change the following components:

## 1. [Optional] Create an Init image for installing deps and compiling

The first step is to create an image for installing function dependencies or compile source code. This step is optional depending on the target language. If the runtime doesn't require compilation and there is already an available image with the necessary packages to install dependencies you can skip this step.

In case a custom init image is required, create a Dockerfile under the folder `docker/runtime/<language>/Dockerfile.<version>.init`. Note that you can skip `<version>` if just one version is supported.

The goal of this image is to have available any tools or files necessary to compile a function or install dependencies. It is not necessary to specify the `CMD` to compile or install dependencies, that will be specified in the Kubeless source code.

In case that the function server needs to be compiled at this step, see the requirements in the next section.

See an example of an init image for [Go](https://github.com/kubeless/kubeless/blob/master/docker/runtime/golang/Dockerfile.init).

## 2. Create a runtime image

The second step is to generate the image that will be used to run the HTTP server that will load and execute functions. Usually  at least a Dockerfile and a source code file written in the target language will be needed. Those files will be placed at:

```
docker/runtime/<language>/Dockerfile.<version>
docker/runtime/<language>/kubeless.<language_extension>
```

Note that you can skip `<version>` if just one version is supported.

The HTTP server should satisfy the following requirements:

 - The file to load can be specified using an environment variable `MOD_NAME`.
 - The function to load can be specified using an environment variable `FUNC_HANDLER`.
 - The server should return `200 - OK` to requests at `/healthz`.
 - Functions should receive two parameters: `event` and `context` and should return the value that will be used as HTTP response. See [the functions standard signature](/docs/runtimes#runtimes-interface) for more information. The information that will be available in `event` parameter will be received as HTTP headers.
 - Requests should be served in parallel.
 - Exceptions in the function should be caught. The server should not exit due to a function error.

See an example of an runtime image for [Python](https://github.com/kubeless/kubeless/blob/master/docker/runtime/python/Dockerfile.2.7).

## 2.1 Additional features

Apart from the requirements above, the runtime should satisfy:

 - The port used to expose the service can be modified using an environment variable `FUNC_PORT`.
 - Functions should run `FUNC_TIMEOUT` as maximum. If, due to language limitations, it is not possible not stop the user function, at least a `408 - Timeout` response should be returned to the HTTP request.
 - Requests should be logged to stdout including date, HTTP method, requested path and status code of the response.
 - The function should expose Prometheus statistics in the path `/metrics`. At least it should expose:
   - Calls per HTTP method
   - Errors per HTTP method
   - Histogram with the execution time per HTTP method

In any case, it is not necessary that the native runtime fulfill the above. Those features can be obtained adding a Go proxy that already implente those features and redirect the request to the new runtime. For adding it is only necessary to add the proxy binary to the image and run it as the `CMD`. See the [ruby example](https://github.com/kubeless/kubeless/blob/master/docker/runtime/ruby/).

## 3. Update the kubeless-config configmap

In this configmap there is a set of images corresponding to the ones described in the previous sections. You need to add the references to these images along with general information about the language that will be added.

There are two entries - one for the runtime image and another for the init container that will be used to install dependencies or compile the function in the build process.

In the example below, a custom image for `go` has been added. You can optionally add `imagePullSecrets` if they are necessary to pull the image from a private Docker registry.

```patch
runtime-images
...
+      {
+        "ID": "go",
+        "compiled": true,
+        "versions": [
+          {
+            "name": "go1.10",
+            "version": "1.10",
+            "runtimeImage": "andresmgot/go:1",
+            "initImage": "andresmgot/go-init:19",
+            "imagePullSecrets": [
+	             {
+                "imageSecret": "Secret"
+              }
+            ]
+          }
+        ],
+        "depName": "Gopkg.toml",
+        "fileNameSuffix": ".go"
+      }
``` 

Restart the controller after updating the configmap. In case that you want to submit the new runtime specify the new images in the file `kubeless.jsonnet` at the root of this repository.

## 4. Add the instructions to install dependencies in the runtime

Each runtime has specific instructions to install its dependencies. These instructions are specified in the method `GetBuildContainer()`. About this method you should know:
 - The runtime is specified as a string in the form of `<name><version>` e.g. `go1.10`
 - The folder with the function and the dependency files is mounted at `installVolume.MountPath`. This is the path in which the function dependencies should be installed. Any change outside this folder will be ignored.

## 5. [Optional] Add the instructions to compile the runtime

In case it is a compiled language you will need to add the required instructions to compile the source code in the method `GetCompilationContainer()`. As in `GetBuildContainer()` the result should be stored in the folder that `installVolume.MountPath` points to. Anything outside that folder will be ignored.

## 6. [Optional] Update the deployment to load requirements for the runtime image

Some languages require to specify an environment variable in order to load dependencies from a certain path. If that is the case, update the function `updateDeployment()` to include the required environment variable:

```patch
func UpdateDeployment(dpm *v1beta1.Deployment, depsPath, runtime string) {
	switch {
...
+	case strings.Contains(runtime, "ruby"):
+		dpm.Spec.Template.Spec.Containers[0].Env = append(
+			dpm.Spec.Template.Spec.Containers[0].Env,
+			v1.EnvVar{
+				Name:  "GEM_HOME",
+				Value: path.Join(depsPath, "ruby/2.4.0"),
+			})
```

This function is called if there are requirements to be injected in your runtime or if it is a custom runtime.

## 6. Add examples

In order to demonstrate the usage of the new runtime it will be necessary to add at least three different examples:

 - GET Example: A simple example in which the function returns a "hello world" string or similar.
 - POST Example: Another example in which the function reads the received input and returns a response.
 - Deps Example: In this example the runtime should load an external library (installed via the build process) and produce a response.

 The examples should be added to the folder `examples/<language_id>/` and should be added as well to the Makefile present in `examples/Makefile`. Note that the target should be `get-<language_id>`, `post-<language_id>` and `get-<language_id>-deps` for three examples above.

## 7. Add tests

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

