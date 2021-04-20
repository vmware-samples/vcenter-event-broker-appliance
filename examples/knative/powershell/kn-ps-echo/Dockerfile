FROM photon:3.0
ENV TERM linux
ENV PORT 8080

# Set terminal. If we don't do this, weird readline things happen.
RUN echo "/usr/bin/pwsh" >> /etc/shells && \
    echo "/bin/pwsh" >> /etc/shells && \
    tdnf install -y powershell-7.0.3-2.ph3 unzip && \
    pwsh -c "Set-PSRepository -Name PSGallery -InstallationPolicy Trusted" && \
    find / -name "net45" | xargs rm -rf && \
    tdnf erase -y unzip && \
    tdnf clean all
RUN pwsh  -Command 'Install-Module ThreadJob -Force -Confirm:$false'
RUN pwsh -Command 'Install-Module -Name CloudEvents.Sdk'

COPY server.ps1 ./
COPY handler.ps1 handler.ps1

CMD ["pwsh","./server.ps1"]
