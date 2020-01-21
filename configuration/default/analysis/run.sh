podname=productpage-v1-8554d58bff-d7j8d
ns=default

istioctl proxy-config listener $podname -n $ns -o json > lds.json

istioctl proxy-config router $podname -n $ns -o json > rds.json

istioctl proxy-config cluster $podname -n $ns -o json > cds.json

istioctl proxy-config endpoint $podname -n $ns -o json > eds.json
