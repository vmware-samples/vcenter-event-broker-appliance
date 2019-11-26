FROM vmware/powerclicore:latest

RUN mkdir -p /home/app
USER root
RUN echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/openfaas/faas/releases/download/0.9.14/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && cp /usr/bin/fwatchdog /root

# Add PSSlack https://github.com/RamblingCookieMonster/PSSlack
RUN pwsh -c "\$ProgressPreference = \"SilentlyContinue\"; Install-Module PSSlack"

WORKDIR /root

USER root

# Populate example here - i.e. "cat", "sha512sum" or "node index.js"
SHELL [ "pwsh", "-command" ]
ENV fprocess="xargs pwsh ./function/script.ps1"
COPY function function
# Set to true to see request in function logs
ENV write_debug="true"

EXPOSE 8080

HEALTHCHECK --interval=3s CMD [ -e /tmp/.lock ] || exit 1
CMD [ "fwatchdog" ]
