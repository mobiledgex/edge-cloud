FROM python:2.7 
WORKDIR /tmp
COPY /setup-env/monitor/requirements.txt /tmp
RUN pip install -r requirements.txt

RUN apt-get update -y && \
    apt-get install -y python-pip python-dev

COPY /setup-env/monitor/edgectl /usr/local/bin

COPY /setup-env/monitor/monitor_web.py /usr/local/bin
COPY /setup-env/monitor/monitor.css /root
COPY /tls/out/mex-ca.crt /root/tls/mex-ca.crt
COPY /tls/out/mex-client.crt /root/tls/mex-client.crt
COPY /tls/out/mex-client.key /root/tls/mex-client.key
ENTRYPOINT [ "monitor_web.py" ]
