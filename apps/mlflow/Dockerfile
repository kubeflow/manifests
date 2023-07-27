FROM python:3.10
RUN pip3 install --upgrade pip && \
 pip3 install mlflow --upgrade && \
 pip3 install awscli --upgrade && \
 pip3 install boto3 --upgrade
RUN pip install --upgrade mysql-connector
RUN pip install --upgrade mysqlclient
ENV PORT 5000
ENV MLFLOW_S3_ENDPOINT_URL http://minio-service.kubeflow.svc.cluster.local:9000
ENV AWS_BUCKET mlflow
ENV AWS_ACCESS_KEY_ID minio
ENV AWS_SECRET_ACCESS_KEY minio123
ENV FILE_DIR /opt//mlflow/backendfile/
RUN mkdir -p /opt/mlflow
COPY run.sh /opt/mlflow
RUN chmod -R 777 /opt/mlflow/
ENTRYPOINT ["/opt/mlflow/run.sh"]