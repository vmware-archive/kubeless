# Create a custom image with a python function
FROM kubeless/python@sha256:565bebecb08d9a7b804c588105677a3572f10ff2032cef7727975061a653fb98
ENV FUNC_HANDLER=foo \
    MOD_NAME=helloget
ADD helloget.py /
RUN mkdir -p /kubeless/
RUN chown 1000:1000 /kubeless
ENTRYPOINT [ "bash", "-c", "cp /helloget.py /kubeless/ && python2.7 /kubeless.py"]
