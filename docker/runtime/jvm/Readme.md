# Generic Runtime for JVM base languages

The runtime environment jvm can be used to execute compiled JVM code in kubeless.

The idea is that you can use your own development environment and workflows to create your code. This way you can run your quality assurance and tests before each new release and roll out this result.

Only one jar-file has to be passed to kubeless, there will be no more changes of your code in Kubeless.

lapse of a treaty 
Compile your code and put it with all dependencies into a jar file:
`gradle shadowJar`
Pass this jar file to kubeless and define the handler method. The package path must be underscored instead of a dot (for the class io.ino.Handler):
`kubeless function deploy test --runtime jvm1.8 --from-file build/libs/jvm-test-0.1-all.jar --handler io_ino_Handler.sayHello`

Known issues
The jar must not be larger than 1MB, since the file is stored in etcd. This can be solved by passing a URL to the JAR file:
`kubeless function deploy test --runtime jvm1.8 --from-file https://example.com/build/libs/jvm-test-0.1-all.jar --handler io_ino_Handler.sayHello`