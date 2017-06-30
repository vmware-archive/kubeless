# Kubeless Build Pipeline Proposal
 
Author(s): gus@
Date: 2017/04/10
Type: Design Doc
Status: Draft
 
## OBJECTIVE 
Design proposal for Kubeless to generate and use docker images.  This reduces the start time of the Kubeless run-time pods, decreases the per-function overhead, and allows us to build on standard Deployment objects to gain staged rollout and rollback functionality.

## BACKGROUND
Kubeless accepts simple function definitions from the user, combines these with a per-language “runtime” and runs the result as pods in Kubernetes.  Currently the function is provided in a ConfigMap and the “combination” (build step) with the runtime is performed in init-containers, which leads to several limitations:

* Where the “build” is expensive or slow, this delays execution of the new pod
Init results are not cached or shared between pod instances or invocations, leading to greater per-function overhead
* No way to “roll back” to previous versions of a function or dependency (when the function definition and runtime pulls in external dependencies)
* When the function is updated, there is no built-in mechanism that triggers the runtime to update with the new definition.
* Any syntax errors or build failures are associated with each runtime pod in which they occur, and not with the Function.  This makes viewing logs and finding debugging information surprising.

## OVERVIEW
This proposal seeks to bring Kubeless even closer to “standard” Kubernetes infrastructure, by moving away from ConfigMaps to regular docker images, and treating the Kubeless Function essentially as input into an internal docker build/push pipeline.

The proposal is to move the per-language runtime to a per-language “build time”, and do as much work as possible up-front, capturing the result in a docker image.  The docker image is pushed to a new private docker registry that is now included alongside the Kubeless controller.  The Kubeless controller now drives a per-function state machine that watches for new or changed Function objects and steps through the build process:

* Notice new Function definition.
* Trigger a Kubernetes “build” Job with a docker image determined by the language runtime, and providing it with a pointer to the Function definition.
* The build Job generates a new OCI image capturing the “compiled” and ready-to-run version of the Function.  Exactly what this means varies greatly for each language.
* Build results (positive or negative) are reported as an event on the Function object.
* The compiled OCI image is pushed to a local private docker registry provided by the kubeless framework.
* The Function Deployment is updated with the new image version.
* The Deployment rolls out the new version following its usual healthcheck, minimum availability, etc strategies.
* The new Function is now deployed.

## DETAILED DESIGN
### Impact on Kubeless deployment

This proposal requires a new docker registry component run as part of Kubeless.  The images are assumed to not be security sensitive (see Security section below), but should be configured to only be accessible within the cluster, and could use per-namespace pull secrets if desired.

This proposal also requires a small “upload proxy” running in the Kubeless namespace.  This proxy accepts an OCI image, validates the uploader, and writes the result to the private docker registry under an image name determined by the sender namespace.  This is to provide per-namespace write isolation and may be removed in future if an alternative method of write isolation can be performed directly by the docker registry.
### Impact on Kubeless end-user (developer) experience
Minimal change to user experience.  The same Function object is created, but now the “status” of the Function can be polled for rollout and error feedback.  In particular, syntax errors, missing dependencies, and other build failures are now reported directly on the Function object.

A new or updated Function will take slightly longer to become live (additional image push/pull step), but repeated runtime instances of the same function (eg: for scaling or fault recovery) will be much faster.  The disk and cpu overhead of additional invocations of the same function will also be greatly reduced, allowing greater scalability within the same physical hardware footprint.
### Impact on Kubeless per-language runtime
Large changes here.  The per-language runtimes need to be recreated and now effectively become per-language “build times”.  Each of these is given the Function text and needs to output an OCI image containing the “compiled” form of the function, including any dependencies.  Ideally any work that is needed in order to execute the function should be done during this step in order to make the later pod faster to (re)start.  Ideally any syntax checking should be done during this step in order to provide prompt and accurate feedback to the user before the resulting pod is instantiated.

Kubeless will (as part of this proposal) provide two tools to help runtime developers:
* A tool to fetch the content of the Function object given the (namespace, name, version) tuple
* A tool to upload the resulting OCI image into the private docker registry, via the Kubeless docker upload proxy

The build step is essentially a linear script that needs to exit 0 on success.  If unsuccessful, the pod stderr will be reported to the user to aid in debugging, so should include file/line error information, etc.  The build will be executed as a Kubernetes Job, and will be automatically retried multiple (but unspecified) times on failure.  Intermittent failures, such as problems downloading externally hosted dependencies should result in the build exiting with failure, and the surrounding logic will handle backoff/retry.

Docker-in-docker is tricky on Kubernetes, since it requires a privileged container.  The recommendation is to build the OCI image directly as filesystem layers on disk, since this is simply a matter of constructing a tarball and some json metadata and requires no special privileges.  The OCI image will be converted to native docker format by the Kubeless upload proxy.

Where possible the build step should optimise for reusing output layers as much as possible and minimise the impact of the final per-function layer.  Consider the case where 1000 different functions are using the same runtime language, and ideally the final layer of each image would only include the unique results of the function itself.  In particular languages like go probably want to dynamically link resulting binaries so they can share common system libraries, unlike the existing “minimal image” common practice of statically linking all binaries.

### Kubeless Function schema
The user-provided portions of the spec section do not change.  Additional “read only” attributes are added to track the resulting automatically created Job and Deployment objects.  The status section is updated to include a Status field giving an overall indication of the position in the build/deploy state machine (eg: Status=Building, Failed, Deployed, etc).  These new fields are all “owned” by the Kubeless controller.

### Kubeless controller
Changes to overall control flow.  The controller now has an overall control loop that watches for new Function resources in any namespace and spawns a goroutine for each Function found.  The goroutine progresses the state of a single (namespace, name) Function through the build/deploy pipeline (see outline in Overview section above).  This mostly consists of spawning an appropriate build Job and Deployment update at the correct times, watching for their completion, and updating the Function object with current status.  Note in particular that the Job definition needs to include the Function version, so a restarted controller can determine whether the correct build job has been run.

## SECURITY
Where possible, all Kubernetes resources should be created in the same namespace as the originating Function object.  In particular, the build Job and final Deployment should not have permissions unavailable to the originating namespace.

The build result needs to be pushed to the private docker registry, but we don’t want to allow other namespaces to be able to overwrite built images.  This document proposes writing to the docker image via a trusted proxy running in the Kubeless namespace, which authenticates the build job via a bearer token and then writes the image to a known unique image name (derived from the original Function name and namespace).

Ideally, the resulting image would also be only able to be read by the originating namespace, since the user might have unwisely embedded sensitive secrets in their Function.  This is not practical with Kubernetes at this point, however, since images are cached in the host-level docker daemon and can be reused by any pod on the same host.  It is out of scope to address this problem in Kubeless at this time, and thus we should ensure that the Kubeless documentation makes it clear that secrets should be stored via the usual Kubernetes Secret mechanisms and not embedded in Function text.

Required RBAC rules
* Kubeless controller needs to read, update and append Events to Function objects in all namespaces
* Kubeless controller needs to start, update, and delete Jobs and Deployments in all namespaces

## POTENTIAL FUTURE WORK
Ideally the build step would be “magically” provided with the Function contents on disk, and the result would be found on disk and “magically” processed.  The need to provide tools for these two steps that are invoked directly by the build step is a workaround for not having a good “sequential job” primitive in Kubernetes.  This will potentially change as this area matures.

