# Examples

This directory contains basic examples for kubeless.

Specifically it contains examples that we can test quickly using the `Makefile`. Some of these examples are run during our integration tests.

Check the [Makefile](Makefile)

Then run some of the examples like so:

```
make post-python
```

Or a different runtime:

```
make post-dotnetcore
```

Or a PubSub example:

```
make pubsub-python
```

# Looking for more function examples?

You can find more examples at [https://github.com/kubeless/functions](https://github.com/kubeless/functions)
