FROM ubuntu
LABEL tanisobe <tanisobe@example.com>
RUN apt-get update \
    && apt-get install -y \
    snmp snmpd \
    --no-install-recommends
COPY snmpd.conf /etc/snmp/snmpd.conf
ENTRYPOINT /usr/sbin/snmpd -f -L
EXPOSE 161/udp