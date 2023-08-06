FROM openjdk:8-alpine AS builder

WORKDIR /opt
COPY ./spring-petclinic-data-jdbc .
RUN ./gradlew build

FROM openjdk:8-jre-alpine
RUN adduser runner -D -u 1000 \
    && mkdir -p /var/log/spring \
    && chown -R runner.runner /var/log/spring
USER runner
WORKDIR /home/runner
COPY --from=builder /opt/build/libs/*.jar ./petclinic.jar
ENTRYPOINT ["/usr/bin/java"]
CMD ["-jar", "petclinic.jar"]