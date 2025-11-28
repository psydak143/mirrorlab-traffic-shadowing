# ===== build =====
FROM maven:3.9.9-eclipse-temurin-21 AS build
WORKDIR /build

COPY pom.xml .
RUN mvn -q -e -DskipTests dependency:go-offline

COPY src ./src
RUN mvn -q -DskipTests package

# ===== runtime =====
FROM eclipse-temurin:21-jre-alpine
ENV JAVA_TOOL_OPTIONS="-XX:+UseContainerSupport -XX:MaxRAMPercentage=75.0"
WORKDIR /app

COPY --from=build /build/target/demo-0.0.1-SNAPSHOT.jar /app/app.jar

EXPOSE 4001
ENTRYPOINT ["java","-jar","/app/app.jar"]
