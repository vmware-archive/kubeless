# Java on runtime JVM

`gradle shadowJar` Build the jar with all deps

`kubeless function deploy test --runtime jvm1.8 --from-file build/libs/jvm-test-0.1-all.jar --handler io_ino_Handler.sayHello` The package name use `_` instead of `.` for the path.
