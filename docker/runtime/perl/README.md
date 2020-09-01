
# Introduction
`kubeless` framework supports deploying functions written in `python`, `golang`, `nodejs`, `ruby`, `php`, `java` and `dotnet`. Now, the support is added to deploy `perl` functions without much effort.

# Architecture
* [Dancer](http://perldancer.org) is a simple but powerful web application framework designed for `perl`. `Dancer` is used to implement the `perl` runtime support in `kubeless`.
* A Docker Image is already created with `Dancer`, `perl 5.26` (compiled with Thread support) and other necessary dependencies. The main entry file i.e. `kubeless.pl` is packaged inside the docker image and is started when the image is deployed and started.
* The main entry file `kubeless.pl` is a `dancer` perl application that starts and waits for the incoming HTTP requests.
* When `kubeless` deploys the docker image, few key pieces of information are provided in the following environrment variables.

```perl
MOD_NAME      => 'The name of the module to be loaded.'
FUNC_HANDLER  => 'The name of the function to be invoked.'
FUNC_PORT     => 'Port on which the web application should listen.'
FUC_TIMEOUT   => 'Timeout (in seconds) in which the function should complete the execution.'
```

*  The `kubeless.pl` file parses the incoming requests, executes the function (in a separate thread) and returns the result. The function must complete it's execution in a fixed timeout. (Default is 10 seconds). If the function is not done executing, it will
be killed and an error is returned to the client.

## Default Route End-points
* Following are few default routes defined by `perl` framework in `kubelss`
* `/healthz` : GET Method is implemented for /healthz route. This route can be used to check if the application is up and running or not. A successful GET request will return 'OK' string in the response.
* `/metrics` : This route is useful to capture the event monitoring. Currently, it's not implemented and will return 'Not implemented' string in the response.
* All other routes that match `/*` path are routed to the function handler. `GET`, `POST`, `PATCH`, `DELETE` methods are implemented for all these routes.

## Structure of a perl module
Following is a template for a simple perl module.
```perl
# helloget.pm
# The file name must end with .pm suffix
# Each function will be invoked with two arguments.
# First one is a reference to an event object.
# Second one is a reference to a function context object.
use strict;
sub hello_world {
   $event_ref = shift;
   $func_ref = shift;
   # The POST data in the incoming HTTP request is available
   # in data attribute of event object.
   $data = $event_ref->{data};
   # If the client POSTS the JSON data, then the $data
   # will be a perl has variable with proper key / value / pairs.
   # Ex: JSON data: {'id' : 123, 'foo' : 'bar'}
   # data{id} = 123, data{foo} = 'bar'
   return "Result Value"
}

# The last line of the perl module MUST be 1
1
```

## Building Perl Docker image
```bash
 cd docker/runtime/perl
 make # This will build docker image
```
* By default, kubeless/perl:5.26 is the tag created for the images built.
* You can use the following commands to push to your own repository.
```bash
  docker tag kubeless/perl:5.26 <username>/perl-kubeless:5.26
  docker push <username>/perl-kubeless:5.26
```

## Examples
```bash
 cd examples/perl
 make # Make file to run all tests.
 #check all .pm files
```

### Updating the kubeless config map.
```
 kubectl edit -n kubeless configmap kubeless-config
 # Add the following entry
 {
  "ID": "perl",
  "compiled": false,
  "versions": [
    {
      "name": "perl526”,
      "version": "5.26”,
      "runtimeImage": "stanguturi/perl-kubeless:5.26”,
      "initImage": "perl:5.26-threaded"
    }
  ],
  "depName": "",
  "fileNameSuffix": ".pm",
 }
  # load the controller manager.
 kubectl delete pods -n kubeless -l kubeless=controller
```

## Deploying the perl module
* Write a proper perl module as mentioned above.
* Currently only perl5.26 is the supported perl runtime.
* Use the following commandi in `kubeless` to deply a function.
```
kubeless deploy function hello_world \
         --from-file helloget.pm
         --handler helloget.hello
         --runtime perl5.26
```

## Tips for writing a proper perl Module
* The name of the perl module file must end with .pm extension.
* The last line of the perl module must be 1
```perl
1
```
* Make sure that there are no compilation errors in the perl module.
* Make sure that the perl module runs with perl 5.26.
* Add `use strict` at the beginning of the perl module.
* The `perl` package installed in the `container image` is built with `threading support`. If the perl module has any dependies with `perl threads`, it should just work.
* Make sure that the function finishes the execution in the specific timeout. The default timeout is 10 seconds.
* Make sure that a proper scalar string is returned from the perl function.

**Happy Perl function Deploying**
