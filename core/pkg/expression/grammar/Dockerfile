# https://github.com/antlr/antlr4/blob/master/docker/Dockerfile
FROM eclipse-temurin:17-jre

ARG ANTLR4_VERSION=4.13.1

RUN curl -sSL https://repo1.maven.org/maven2/org/antlr/antlr4/${ANTLR4_VERSION}/antlr4-${ANTLR4_VERSION}-complete.jar \
         -o /usr/local/lib/antlr4-tool.jar

RUN useradd -ms "/bin/bash" --uid=1000 appuser;
USER appuser:appuser

WORKDIR /work

ENTRYPOINT ["java", "-Xmx500M", "-cp", "/usr/local/lib/antlr4-tool.jar", "org.antlr.v4.Tool"]
