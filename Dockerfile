FROM alpine
COPY kube-apiserver-audit-exporter \
	/usr/bin/kube-apiserver-audit-exporter
ENTRYPOINT ["/usr/bin/kube-apiserver-audit-exporter"]
