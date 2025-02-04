# Приложение для выбора игроков из пула в группы

1. Приложение редставляет собой web сервис с API, предоставляющий доступ к пулу игроков и реализующий операции добавления и удаления игроков, просмотра групп игроков и статистики групп.
    1. Для добавления игрока в пул используйте запрос POST /users со структурой в теле запроса:
    ```json
    {
    "name":"John",
    "skill":64.7,
    "latency":35.7
    }
    ```
   2. Для удаления игрока используйте запрос DELETE /user/{name} где name - имя удаляемого игрока
   3. Для просмотра созданных групп используйте запрос GET /groups - если группы ещё не сформированы, вернётся пустой массив.
   4. Для просмотра статистики группы используйте запрос GET /groups/{number} где number - номер группы, для которой запрашивается статистика. Для несуществующей группы вернётся ответ с нулевыми метриками.
   5. Для сброса групп и перемещения всех игроков обратно в очередь ожидания для формирования новых групп используйте запрос GET /groups/reset

2. Конгфигурация приложения задаётся с помощью переменных окружения:

| Переменная | Описание | Значение по умолчанию |
|---|---|---|
| HOST | Хост-адрес веб приложения | 127.0.0.1 |
| PORT | TCP порт, слушаемый приложением | 8080 |
| MAX_GROUP_SIZE | Размер группы игроков | 3 |
| STORE_IN_DB | Сохранять ли игроков в базе данных | false |
| BUFFER_SIZE | Размер буферизированных каналов для добавления и удаления игроков в БД | 16 |
| DB_HOST | Адрес базы данных | localhost |
| DB_USER | Ипя пользователя для подключения к БД | test_user |
| DB_PASSWORD | Пароль пользователя для подключения к БД | postgres |
| DB_SSL_MODE | Значение параметра ssl_mode в строке подклбчения к БД | disable |

В базе данных используется имя для базы данных по умолчанию - `gamers`, имя таблицы по умолчанию - `gamers`.

1. Приложение состоит из нескольких пакетов, каждый из которых расположен в одноименной папке:
    - main - пакет, содержащий остновную функцию main, с которой начинается работа программы. Также в этом пакете реализован API с использование пакета gorilla/mux.
    - gamer - этот пакет содержит описание структуры данных для игрока, а также отвечает за хранение данных игроков в памяти в виде ассоциативного массива.
    - db - в этом пакете реализовано подключение к базе данных, запись и удаление игроков. Для подключения используется драйвер для работы с БД PostgreSQL.
    - groups - здесь реализована логика управления группами, здесь лежит основной алгоритм распределения игроков.
    - service - здесь реализована логика, связывающая все остальные пакеты.

В качестве основного хранилища игроков используется map в обёртке для потокобезопасной работы с ней (Gamerspool в пакете gamer). 
Для очереди игроков используется вспомогательная структура памяти в виде map (Структура GamersGroups в пакете groups).

