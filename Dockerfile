FROM alpine:3.14

COPY config/serverconfig.yaml /etc/deliverable-receiver/serverconfig.yaml
COPY deliverable-receiver.linux64 /usr/local/bin/deliverable-receiver
CMD /usr/local/bin/deliverable-receiver --config /etc/deliverable-receiver/serverconfig.yaml
