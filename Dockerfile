FROM alpine:3.2
ADD galeb-statsd-logstash /bin/galeb-statsd-logstash
ENTRYPOINT ["/bin/galeb-statsd-logstash"]
