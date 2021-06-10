FROM projects.registry.vmware.com/veba/ce-pcli-base:1.1
ENV TERM linux
ENV PORT 8080

COPY handler.ps1 handler.ps1

CMD ["pwsh","./server.ps1"]
