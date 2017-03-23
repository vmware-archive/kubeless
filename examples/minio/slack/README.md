# Send messages on objects upload to a bucket

This function is triggered when an object is added to a bucket

```
kubeless function deploy slack --trigger-topic s3 --from-file bot.py --handler bot.handler --runtime python2.7 --dependencies requirements.txt
```
