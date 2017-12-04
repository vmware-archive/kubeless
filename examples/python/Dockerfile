# Create a custom image with a python function
FROM kubeless/python@sha256:ba948a6783b93d75037b7b1806a3925d441401ae6fba18282f712a1b1a786899
ENV FUNC_HANDLER=foo \
    MOD_NAME=helloget
ADD helloget.py /
ENTRYPOINT [ "bash", "-c", "mv /helloget.py /kubeless/ && python /kubeless.py"]
