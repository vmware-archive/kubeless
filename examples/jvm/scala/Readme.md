# Scala on runtime JVM

!! WIP the jar-file is to large for the storage backend, you have to pass a url to the jar file.

`sbt assembly` Build the jar with all deps

`kubeless function deploy testscala --runtime jvm1.8 --from-file target/scala-2.12/scala-test.jar --handler de_inoio_Handler.fooBar` The package name use `_` instead of `.` for the path.
