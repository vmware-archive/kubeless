# Scheduling the trigger of a function 

Kubeless has its own CronJobTrigger which uses [Kubernetes CronJob](https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs/) to trigger your function in a given schedule. On this page, we're going to cover how to use it, and some basic features.

## Creating a new CronJobTrigger

You can create a new cron trigger using `kubeless-cli`. In this section, we're going to show you how to create a simple function that logs `Hello world!` every 1 minute. 

For this example you're going to need the following tools:

* [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
* [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [Kubeless CLI](/docs/quick-start/)

After installing all the requirements, you can proceed to the step-by-step guide:

### Step 1: Create a new Minikube cluster

In this step, you're going to create a new Minikube cluster called `kubeless`, where you're going to deploy the function triggers. You can run the following command on your shell:

```shell
minikube start -p kubeless
```

**IMPORTANT:** If you have already created any Minikube cluster called `kubeless` you should delete it first, with `minikube delete -p kubeless`

### Step 2: Install Kubeless on your cluster

Now that you have a Minikube cluster running, you can run the following command to install the latest version of Kubeless:

```shell
RELEASE=$(curl -s https://api.github.com/repos/kubeless/kubeless/releases/latest | grep tag_name | cut -d '"' -f 4) && \
  kubectl create ns kubeless && \
  kubectl create -f https://github.com/kubeless/kubeless/releases/download/$RELEASE/kubeless-$RELEASE.yaml
```

### Step 3: Deploy a test function

For this CronJob test, we're going to use a simple function that just logs a "Hello world!" message. Since this isn't a tutorial explaining how to deploy a function you can just run the following command:

```shell
 kubectl apply -f https://gist.githubusercontent.com/odelucca/1f3a71b7ff05f31d492dc5bfd3f3afba/raw/5237991f018f99a697e937a85e60e57dd8ac1a1c/function.yaml
```

### Step 4: Create a new CronJob trigger

To create a CronJob trigger with `kubeless-cli` you can run the following command:

```shell
kubeless trigger cronjob create \
  cron-test-hello-world \
  --function cron-test-hello-world \
  --schedule "*/1 * * * *"
```

About the provided arguments:

* **The first argument** must be the trigger name you want to use
* **--function** should be the name of the function you want to trigger with that cron
* **--schedule** the cron pattern to trigger your function

### Step 5: Take a look on your function logs

Now, wait 1 or 2 minutes and take a look at your function logs with this command:

```shell
kubeless function logs cron-test-hello-world
```

You should see some `Hello world!` logs, showing that our CronJob is working as expected.

## Advanced concepts

In this section, we're going to cover some advanced concepts regarding the CronJob trigger. Each item in this section will cover a given feature that you can use on your triggers.

### Passing payload data to the function

While triggering a function you could pass also a payload data to it. Those will be available on `event.data` (like any other request data). You can do so with the following command:

```shell
kubeless trigger cronjob (create or update) --payload <stringified JSON>
```

If you're not willing to provide a stringified JSON to the `--payload` argument, you can use `--payload-from-file` instead and pass a file path. You can provide files on the following extensions:

* `.json`
* `.yaml`

**IMPORTANT:** Your payload must be an object, so you cannot provide a JSON array to it, but you can add a key on your object that can contain a list of items instead.
