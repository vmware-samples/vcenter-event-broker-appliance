FROM projects.registry.vmware.com/veba/ce-ps-base:1.1
ENV TERM linux
ENV PORT 8080

ARG MAILKIT_VERSION="3.1.0"

RUN pwsh -c "Install-Module -Name Send-MailKitMessage -RequiredVersion ${MAILKIT_VERSION}"

COPY handler.ps1 handler.ps1

CMD ["pwsh","./server.ps1"]
