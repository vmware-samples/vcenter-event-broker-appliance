FROM projects.registry.vmware.com/veba/ce-ps-base:1.0
ENV TERM linux
ENV PORT 8080

COPY handler.ps1 handler.ps1

CMD ["pwsh","./server.ps1"]
