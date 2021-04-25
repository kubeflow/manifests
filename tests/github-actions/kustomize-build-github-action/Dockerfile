FROM python:3.8-alpine

RUN apk add --update --no-cache git bash wget
RUN wget -O /usr/bin/kustomize "https://github.com/kubernetes-sigs/kustomize/releases/download/v3.2.0/kustomize_3.2.0_linux_amd64"
RUN chmod +x /usr/bin/kustomize
RUN mkdir -p /src
COPY . /src
RUN pip install -r /src/requirements.txt
ENTRYPOINT ["/src/run_tests.py"]
