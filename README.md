# Geo Objects to .mbtiles file converter


[![forthebadge](https://forthebadge.com/images/badges/made-with-go.svg)](https://forthebadge.com)
[![forthebadge](http://forthebadge.com/images/badges/built-with-love.svg)](http://forthebadge.com)


[![Go Report Card](https://goreportcard.com/badge/github.com/vieux/docker-volume-sshfs)](https://goreportcard.com/report/github.com/OkDenAl/mbtiles_converter)

Утилита для конвертации гео объектов хранящихся в `Postgres`
в векторный формат `.mbtiles`. Подробнее про [формат](https://github.com/mapbox/mbtiles-spec/blob/master/1.3/spec.md).

Используемые технологии:
- PostgreSQL (в качестве хранилища исходных данных)
- SQLite (для построения выходного файла .mbtiles)
- Docker (для запуска сервиса)
- pgx (драйвер для работы с PostgreSQL)
- sqlite3 (драйвер для работы с SQLite)
- golang/mock, testify (для тестирования)

Сервис был написан с `Clean Architecture`, что позволяет легко расширять функционал сервиса и тестировать его.

## Getting Started
1) Скопируйте себе данный репозиторий GitHub

    ```$ git clone https://github.com/OkDenAl/mbtiles_converter```


2) Перейдите в папку `./config` и настройте конфигурацию утилиты
   с помощью файла `config.yml`.


3) Для запуска используйте
    ```$ make start```



Выходной файл будет лежать в директории `./mbtiles`