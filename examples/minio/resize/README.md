# A Python function to automatically create thumbnails

This example assumes that you have Minio running in your Kubernetes cluster. See this [example](../README.md)

## Goal

We want images that get uploaded to a Minio bucket (aka S3 bucket) to automatically trigger a thumbnail creation.
Each thumbnail should then be stored into another bucket

## Create the function

```
kubeless function deploy thumb1 --trigger-topic s3 --runtime python2.7 --handler resize.thumbnail --from-file resize.py --dependencies requirements.txt
```

This function is written in Python, the modules required to make it work are `kubernetes`, `minio` and `Pillow`

Once it is running, drop images in the `foobar` bucket and watch the thumbnail appear automatically in the `thumb` bucket.

## Details of the function

Looking into the `resize.py` file you will see how the thubmail is created:

```
size=(120,120)
img = Image.open(tf.name)
img.thumbnail(size)
img.save(tf_thumb.name, "JPEG")
```
