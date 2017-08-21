slack:
	kubeless function deploy slack --trigger-http --runtime python2.7 --handler bot.handler --from-file bot.py --dependencies requirements.txt
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/slack/"
events:
        kubeless function deploy k8s-events --trigger-topic k8s --from-file botevents.py --handler botevents.handler --runtime python2.7 --dependencies requirements.txt
