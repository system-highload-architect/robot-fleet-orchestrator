# Модель данных

## Таблица `robots`
| Поле          | Тип          | Описание |
|---------------|--------------|----------|
| id            | UUID         | PK       |
| name          | VARCHAR(100) |         |
| type          | VARCHAR(50)  | delivery, drone, loader |
| status        | VARCHAR(20)  | idle, busy, offline, charging |
| location_x    | FLOAT        |         |
| location_y    | FLOAT        |         |
| battery       | FLOAT        | 0–100   |
| capacity      | FLOAT        | грузоподъёмность |
| updated_at    | TIMESTAMP    |         |

## Таблица `tasks`
| Поле          | Тип          | Описание |
|---------------|--------------|----------|
| id            | UUID         | PK       |
| type          | VARCHAR(50)  | delivery, move, charge |
| priority      | INT          | 1–5      |
| status        | VARCHAR(20)  | pending, assigned, completed, failed |
| assigned_to   | UUID         | FK robots.id |
| location_from | VARCHAR(200) |         |
| location_to   | VARCHAR(200) |         |
| created_at    | TIMESTAMP    |         |
| updated_at    | TIMESTAMP    |         |

## Таблица `locations`
| Поле          | Тип          | Описание |
|---------------|--------------|----------|
| id            | UUID         | PK       |
| name          | VARCHAR(100) |         |
| type          | VARCHAR(50)  | warehouse, city, drone_zone |
| min_x         | FLOAT        | границы зоны |
| max_x         | FLOAT        |         |
| min_y         | FLOAT        |         |
| max_y         | FLOAT        |         |
| allowed_types | VARCHAR(255) | допустимые типы роботов через запятую |
