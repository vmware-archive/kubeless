slack:
	kubeless function create slack --trigger-http --runtime python27 --handler bot.handler --from-file bot.py --dependencies requirements.txt
	echo "curl localhost:8080/api/v1/proxy/namespaces/default/services/slack/"
