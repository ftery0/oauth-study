# app1 Notebook backend — Spring Boot + MariaDB
FROM maven:3.9-eclipse-temurin-21 AS builder
WORKDIR /src
COPY pom.xml ./
RUN mvn -B -q dependency:go-offline
COPY src ./src
RUN mvn -B -q -DskipTests package

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=builder /src/target/*.jar /app/app.jar
EXPOSE 8011
ENV PORT=8011
ENTRYPOINT ["java","-jar","/app/app.jar"]
