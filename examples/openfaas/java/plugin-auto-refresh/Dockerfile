FROM openjdk:8u212-jdk-alpine as builder

RUN apk --no-cache add curl \
    && addgroup -S app && adduser -S -g app app \
    && mkdir -p /opt/ && cd /opt/ \
    && echo "Downloading gradle.." \
    && curl -sSfL https://services.gradle.org/distributions/gradle-5.5.1-bin.zip -o gradle-bin.zip \
    && unzip gradle-bin.zip -d /opt/ \
    && rm gradle-bin.zip \
    && apk del curl

# Export some environment variables

FROM openfaas/of-watchdog:0.7.6 as watchdog

FROM openjdk:8u212-jdk-alpine as ship

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

RUN addgroup -S app && adduser -S -g app app
USER app

ARG JAR_FILE=target/*.jar
COPY ${JAR_FILE} app.jar

ENV fprocess="java -jar /app.jar"
ENV write_debug=true

ENV mode="http"
ENV upstream_url="http://127.0.0.1:8089/sample-ui/veba/host/updated"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]