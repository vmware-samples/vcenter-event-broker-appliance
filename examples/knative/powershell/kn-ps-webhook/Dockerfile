FROM ghcr.io/vmware-samples/vcenter-event-broker-appliance/ce-ps-base:1.5

LABEL maintainer="vCenter Event Broker Appliance Community"
LABEL org.opencontainers.image.source="https://github.com/vmware-samples/vcenter-event-broker-appliance"

COPY server.ps1 server.ps1
COPY handler.ps1 handler.ps1

CMD ["pwsh","./server.ps1"]
