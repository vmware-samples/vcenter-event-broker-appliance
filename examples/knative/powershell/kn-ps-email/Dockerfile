FROM ghcr.io/vmware-samples/vcenter-event-broker-appliance/ce-ps-base:1.5

LABEL maintainer="vCenter Event Broker Appliance Community"
LABEL org.opencontainers.image.source="https://github.com/vmware-samples/vcenter-event-broker-appliance"

ARG MAILKIT_VERSION="3.1.0"

RUN pwsh -c "Install-Module -Name Send-MailKitMessage -RequiredVersion ${MAILKIT_VERSION}"

COPY handler.ps1 handler.ps1

CMD ["pwsh","./server.ps1"]
