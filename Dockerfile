FROM scratch

COPY istio-state-metrics /
VOLUME /tmp

ENTRYPOINT ["/istio-state-metrics", "--port=9090"]

EXPOSE 9090