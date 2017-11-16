# Task List. Simple API app

This is an example for deploying a simple serverless API.

It uses MongoDB as database so first we need to deploy it:

```console
kubectl create -f https://raw.githubusercontent.com/bitnami/bitnami-docker-mongodb/3.4.10-r0/kubernetes.yml
```

Now we need to create a zip file containing all the API code. Generate it executing:

```console
zip -r api.zip api server.js
```

Wait for the database to be deployed (check the status with `kubectl get pods`). Once it is ready deploy the API functions:

```console
kubeless function deploy api-list \
                        --from-file api.zip \
                        --handler server.list \
                        --runtime nodejs6 \
                        --dependencies package.json \
                        --trigger-http
kubeless function deploy api-add \
                        --from-file api.zip \
                        --handler server.add \
                        --runtime nodejs6 \
                        --dependencies package.json \
                        --trigger-http
kubeless function deploy api-update \
                        --from-file api.zip \
                        --handler server.update \
                        --runtime nodejs6 \
                        --dependencies package.json \
                        --trigger-http
kubeless function deploy api-delete \
                        --from-file api.zip \
                        --handler server.delete \
                        --runtime nodejs6 \
                        --dependencies package.json \
                        --trigger-http
```

Once the functions are finally deployed you can invoke them:

```console
$ kubeless function call api-add --data '{"name": "Create an API"}'
{"__v":0,"name":"Create an API","_id":"5a0c78018cef7f00017f39ca","status":["pending"],"Created_date":"2017-11-15T17:23:13.718Z"}
$ kubeless function call api-list
[{"_id":"5a0c78ae8cef7f00017f39cb","name":"Create an API","__v":0,"status":["pending"],"Created_date":"2017-11-15T17:26:06.453Z"}]
$ kubeless function call api-update --data '{"name": "Create an API", "status": ["completed"]}'
{"_id":"5a0c71429bf23b0001cea0d2","name":"Create an API","__v":0,"status":["completed"],"Created_date":"2017-11-15T16:54:26.865Z"}
$ kubeless function call api-delete --data '{"name": "Create an API"}'
{"message":"Task successfully deleted"}
```