FROM python:2.7-slim

RUN pip install boto3 prometheus_client

ADD kubeless.py .

EXPOSE 8080

CMD ["python", "/kubeless.py"]
