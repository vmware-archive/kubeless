# http trigger improvements

Though there is no standard on what http/https triggers of FAAS platform should support, most hosted FAAS solution like AWS Lambda, google cloud functions, IBM cloud functions, Azure functions etc provide common functionality

- http/https endpoint for the function: a fully qualified URL is automatically generated and assigned to an HTTP triggered Cloud Function which can retrieved through respective cli or consoles
- ability to customize the endpoint url by specifying route
- a way to authenticate and authorize the function invoker through url
- a way to restrict http method GET/POST/DELETE etc used to invoke function
- stage and version functions

Kubeless already supports `--trigger-http`. It does seem reasonable to expect similar functionality with kubeless for http triggers. This proposal would like to articulate current gaps and suggest changes.

## Challenges

- In typical FAAS platforms you have [API gateway](https://martinfowler.com/articles/serverless.html) or router (for e.g AWS API gateway) which receives the requests and calls the relevant FaaS function. Some times also perform authentication, input validation, response code mapping, etc. Control path (rest endpoint to create/update/delete functions) and data path (invoking function and getting response) are either combined into one enity or sepearated. Kubeless as native Kubernetes solution intelligently leverages kubernetes constructs and offloads control path (Kubernetes API server through CRD) and data path (through services). While it helps in many aspects, it also means we are constrained by kubernetes constructs. For e.g authenticating the function caller.

- Leveraging kubernetes service as data path to call the function means we need to deal with various service types of Kubernetes for various scenarios. For e.g, a function deployed with Kubeless, if its only caller is microservice running in-cluster then perhaps service of `clusterIP` is needed. If you expect out of cluster callers but does not care about L7 then service of type NodePort is enough. For cases where you want L7, tls etc then Kubeless already leverages Kubernetes Ingress. Also there is PR to support headless service which make sense for baremetal deployment. While Kubeless should be flexible to allow differnet use-cases, it will be challenging to generate a http endpoint for the function.

- On managed FAAS platform, since function user/developer is completly taken out of the infrastrcutre ops there is clear seperation of concerns. Kubeless leveraging K8S constructs, conscious effort must be put not to splill infrastructure or k8s concpets on to the function user. In other words be mindful of two personas using kubeless: function user and cluster/kubeless deployment operator.

## Gaps

While its debatable whats the desirable kubeless view of http-triggers, here are current gaps from the point of view of this proposal.

* tight coupling of function routing with kubernetes ingress.So what if some one does not want ingress (i.e want to use node port, headless service etc)
  * `kubeless ingress`: function user explicitly dealing with ingress
  * though optional `--hostname` flag for `kubeless ingress`, why shoud function user be aware of ingress object hosts?
  * function user specifying the tls flags
- kubeless client creates ingress objects
- as function user how do i know what http/https endpoint for my function. alternatively how controller can generate consitently the URL for the function. 
- as a cluster operator how do i tell which ingress controller to use, or how do i customize my ingress

## proposed changes

- [formal specification](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/#validation) of the function spec that is devoid of any k8s/infrastcture constucts for writable fields
- function user just do the route management i.e) express desired path for the function. Function spec to carry the function user intent ie. path to function mapping.
- rename `kubeless ingress` to `kubeless route`
- move the ingress object creation to kubeless controller.
- kubeless controller that has ability to provision service (as cluster ip, node port, load balancer or headless) backing the function as desired by the cluster operator  
- introduce configmap for the controller that cluster oprator can use to configure controller. for e.g details like which ingress controller to use
- controllers ability consistently generate http endpoint for the function irrespective the service type backing the function and whether kubernetes ingress is used or not
- clean up the current kubeless flags related to ingress

## what to do we achieve?

- small step toward some of the comman functionality of http triggers in other FAAS platforms
- clean separation of concerns of function user and cluster/kubeless operator
- extensibility of controller (where applicable) with configmap

## tracking issues

- [#417](https://github.com/kubeless/kubeless/issues/417) kubeless list option should give info on http/https endpoint
- [#476](https://github.com/kubeless/kubeless/issues/476) move ingress object creation to ingress controller
- [#474](https://github.com/kubeless/kubeless/issues/474) support flexible service types for the service backing functions
- [#475](https://github.com/kubeless/kubeless/issues/475) introduce configmap for kubeless-controller
- [#478](https://github.com/kubeless/kubeless/issues/478) rename `ingress` command to `route` 


      
