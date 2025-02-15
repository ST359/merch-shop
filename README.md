# Avito trainee backend assignment 2025

- [English](#Running)
  - [Running](#Running)
  - [Problems and Solutions](#problems-and-solutions)
- [Русский](#Запуск)
  - [Запуск](#Запуск)
  - [Проблемы и Решения](#проблемы-и-решения)

## Running
To run the service, Docker and Docker Compose are required.
```sh
git clone https://github.com/ST359/avito-trainee-backend-winter-2025
cd avito-trainee-backend-winter-2025
docker-compose up --build
```

## Issues and Solutions
The questions mainly concerned the use of various libraries and frameworks. During the process, I would naturally follow the accepted standards in the company, if any, regarding solutions of this level.
- As a query builder for the database, it was decided to use [Squirrel](https://github.com/Masterminds/squirrel). This library allows for convenient query construction while avoiding potential SQL injections.
- The framework used is [Gin](https://github.com/gin-gonic/gin), due to its speed (both the framework itself and the development process), built-in logger, and recovery features.
- To ensure the most accurate compliance with the provided API schema and to speed up development, code generation was utilized with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen).

## Запуск
Для запуска требуется Docker и Docker-compose
```sh
git clone https://github.com/ST359/avito-trainee-backend-winter-2025
cd avito-trainee-backend-winter-2025
docker-compose up --build
```
## Проблемы и решения
Вопросы касались преимущественно использования различных библиотек, фреймворков - в процессе работы, само собой, я бы следовал принятым в компании стандартам, если таковые имеются касательно решений такого уровня
- В качестве билдера запросов к базе данных было решено использовать [Squirrel](https://github.com/Masterminds/squirrel), эта библиотека позволяет удобно строить запросы, избегая при этом потенциальных SQL-инъекций
- Фреймворк - [Gin](https://github.com/gin-gonic/gin), в виду скорости(как самого фреймворка, так и разработки на нем), встроенного логгера и рекавери
- Для наиболее точного соответствия предоставленной схеме API, а также ускорения разработки, была использована кодогенерация [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)
