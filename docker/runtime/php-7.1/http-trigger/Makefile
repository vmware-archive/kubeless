

build:
	docker build -t paolomainardi/kubeless-php:latest .

push: build
	docker push paolomainardi/kubeless-php:latest

php-cs-fixer:
	docker run -it --rm -v $${PWD}:/wd -w /wd shouldbee/php-cs-fixer fix Controller.php || true
