# Licensing Backend — API для работы с моделями лицензирования

## Модели базы данных

### Таблица `users` — пользователи

| Поле | Тип | Ограничения | Описание |
| --- | --- | --- | --- |
| `id` | Integer | PK, autoincrement, index | Идентификатор |
| `login` | String(50) | NOT NULL, UNIQUE | Логин пользователя |
| `password` | String(100) | NOT NULL | Пароль |
| `is_moderator` | Boolean | NOT NULL, default=false | Флаг модератора |

### Таблица `licensings` — модели лицензирования

| Поле | Тип | Ограничения | Описание |
| --- | --- | --- | --- |
| `id` | Integer | PK, autoincrement, index | Идентификатор |
| `title` | String(100) | NOT NULL | Название модели |
| `description` | String(255) | NULLABLE | Краткое описание |
| `commission_per_unit` | Float | NOT NULL, default=0 | Комиссия за 1 единицу |
| `min_for_calc` | Integer | NOT NULL, default=0 | Минимальное число для расчёта |
| `image_url` | String(255) | NOT NULL, default='' | URL изображения |
| `video_url` | String(255) | NOT NULL, default='' | URL видео |
| `status` | ENUM `licensing_status` | NOT NULL | `published` / `draft` / `deleted` |
| `creator_id` | Integer | FK → `users.id`, NOT NULL | Автор модели |
| `created_at` | TIMESTAMP(tz) | NOT NULL | Дата создания |
| `published_at` | TIMESTAMP(tz) | NULLABLE | Дата публикации |

### Таблица `likes` — лайки (м-м пользователь-модель)

| Поле | Тип | Ограничения | Описание |
| --- | --- | --- | --- |
| `id` | Integer | PK, autoincrement, index | Идентификатор |
| `user_id` | Integer | FK → `users.id`, NOT NULL | Кто поставил |
| `licensing_id` | Integer | FK → `licensings.id`, NOT NULL | Какой модели |

**Уникальный индекс:** `(user_id, licensing_id)` — один пользователь может лайкнуть модель только один раз.

---

## API методы

Все эндпоинты имеют префикс `/api`. Тег `licensings` — для работы с моделями лицензирования, `users` — для пользователей.

### Пользователи (`/api/users`)

### `POST /api/users/register` — регистрация

**Тело запроса (JSON):**

```json
{
  "login": "testuser1",
  "password": "secret"
}
```

**Ответы:**

| Код | Описание |
| --- | --- |
| `201` | Пользователь создан |
| `400` | Ошибка валидации |
| `500` | Пользователь с таким логином уже существует |

### `POST /api/users/login` — вход

Заглушка. Пока не реализовано.

**Тело запроса (JSON):**

```json
{
  "login": "testuser1",
  "password": "secret"
}
```

**Ответ `200`:**

```json
{
  "message": "заглушка авторизации"
}
```

### `POST /api/users/logout` — выход

Заглушка. Пока не реализовано.

**Тело запроса (JSON):**

```json
{}
```

**Ответ `200`:**

```json
{
  "message": "заглушка деавторизации"
}
```

---

### Модели лицензирования (`/api/licensings`)

### `GET /api/licensings` — каталог моделей

Возвращает список **опубликованных** моделей с количеством лайков и флагом «является ли текущий пользователь создателем».

**Query-параметры:**

| Параметр | Тип | Обяз. | Описание |
| --- | --- | --- | --- |
| `max_commission` | float | Нет | Фильтр: показать модели с комиссией **меньше или равно** указанной |

**Ответ `200`:**

```json
[
  {
    "id": 1,
    "title": "Per User",
    "description": "Модель на одного пользователя.",
    "commission_per_unit": 120,
    "min_for_calc": 1,
    "image_url": "http://localhost:9000/licensing-images/licensing-1-image.jpg",
    "video_url": "http://localhost:9000/licensing-images/licensing-1-video.mp4",
    "status": "published",
    "creator_id": 1,
    "created_at": "2026-09-20T19:28:40Z",
    "published_at": "2026-09-20T19:30:00Z",
    "is_creator": 1,
    "likes_count": 3
  }
]
```

Где `is_creator` — 0/1: совпадает ли текущий пользователь с создателем модели.

**Ответы:**

| Код | Описание |
| --- | --- |
| `200` | Список получен |
| `400` | Ошибка валидации `max_commission` |
| `500` | Внутренняя ошибка |

---

### `GET /api/licensings/draft` — получить черновик

Возвращает **единственный** черновик текущего пользователя. ID не указывается.

**Ответ `200`:**

```json
{
  "id": 6,
  "title": "Trial",
  "description": "Пробная модель.",
  "commission_per_unit": 0,
  "min_for_calc": 1,
  "image_url": "",
  "video_url": "",
  "status": "draft",
  "creator_id": 1,
  "created_at": "2026-09-20T19:28:40Z",
  "published_at": null
}
```

**Ответы:**

| Код | Описание |
| --- | --- |
| `200` | `{"id": 6, "...": "..."}` |
| `404` | Черновик не найден |

---

### `POST /api/licensings` — создать черновик

Создаёт новый черновик. Загружает изображение и видео в MinIO. Один пользователь — один черновик.

**Тело запроса (`multipart/form-data`):**

| Поле | Тип | Обяз. | Ограничения |
| --- | --- | --- | --- |
| `title` | string | Да | — |
| `description` | string | Нет | — |
| `commission_per_unit` | float | Нет | `≥ 0` |
| `min_for_calc` | int | Нет | `≥ 0` |
| `image` | file | Да | `content_type` начинается с `image/` |
| `video` | file | Да | `content_type` начинается с `video/` |

**Ответ `201`:**

```json
{
  "id": 8,
  "title": "Новая модель",
  "description": "Описание",
  "commission_per_unit": 250,
  "min_for_calc": 3,
  "image_url": "http://localhost:9000/licensing-images/licensing-8-image.jpg",
  "video_url": "http://localhost:9000/licensing-images/licensing-8-video.mp4",
  "status": "draft",
  "creator_id": 1,
  "created_at": "2026-09-20T19:28:40Z",
  "published_at": null
}
```

**Ответы:**

| Код | Описание |
| --- | --- |
| `201` | Черновик создан |
| `400` | Не указано название / недопустимый тип файла |
| `409` | У черновика уже существует запись |
| `500` | Ошибка сохранения в MinIO |

**Что происходит:**

1. Проверка `content_type` файлов.
2. Поиск существующего черновика пользователя (если есть — 409).
3. Создание записи в БД.
4. Загрузка файлов в MinIO (`licensing-images`).
5. Запись `image_url` и `video_url` в БД.

---

### `PUT /api/licensings/{id}/publish` — публикация

Меняет статус черновика на `published` и проставляет `published_at`. Только **создатель**.

**Path-параметры:**

| Параметр | Тип | Описание |
| --- | --- | --- |
| `id` | int | ID черновика |

**Тело запроса (JSON):**

```json
{}
```

**Ответ `200`:**

```json
{
  "id": 8,
  "title": "Новая модель",
  "status": "published",
  "published_at": "2026-09-20T19:30:00Z",
  "is_liked": 0,
  "likes_count": 0
}
```

**Ответы:**

| Код | Описание |
| --- | --- |
| `200` | Модель опубликована |
| `400` | Невалидный ID |
| `404` | Черновик не найден / чужой / уже не `draft` |

---

### `DELETE /api/licensings/{id}` — логическое удаление

Меняет `status` на `deleted`. Удалить может **только создатель**.

**Path-параметры:**

| Параметр | Тип | Описание |
| --- | --- | --- |
| `id` | int | ID модели |

**Ответы:**

| Код | Описание |
| --- | --- |
| `200` | Удалено (логически) |
| `400` | Невалидный ID |
| `404` | Модель не найдена / чужая / уже удалена |

---

### Лента (`/api/licensing/feed`)

### `GET /api/licensing/feed` — лента (одна модель)

Возвращает **одну** опубликованную модель — для отображения в ленте.

**Query-параметры:**

| Параметр | Тип | Обяз. | Описание |
| --- | --- | --- | --- |
| `id` | int | Нет | Текущая модель (для поиска следующей) |
| `next` | bool | Нет | Если `true` — вернуть следующую после `id` |

**Логика выбора:**

- Без параметров — первая опубликованная.
- С `id=N` — конкретная по ID.
- С `id=N&next=true` — следующая после N.

**Ответ `200`:**

```json
{
  "id": 1,
  "title": "Per User",
  "description": "Модель на одного пользователя.",
  "commission_per_unit": 120,
  "min_for_calc": 1,
  "image_url": "http://localhost:9000/licensing-images/licensing-1-image.jpg",
  "video_url": "http://localhost:9000/licensing-images/licensing-1-video.mp4",
  "status": "published",
  "creator_id": 1,
  "created_at": "2026-09-20T19:28:40Z",
  "published_at": "2026-09-20T19:30:00Z",
  "is_liked": 1,
  "likes_count": 3
}
```

Где `is_liked` — 0/1: поставил ли **текущий пользователь** лайк этой модели.

**Ответы:**

| Код | Описание |
| --- | --- |
| `200` | Модель получена |
| `400` | Ошибка валидации |
| `404` | Модель не найдена / конец ленты |

---

### Лайки (`/api/licensings/{id}/like`)

### `POST /api/licensings/{id}/like` — поставить/убрать лайк

**Path-параметры:**

| Параметр | Тип | Описание |
| --- | --- | --- |
| `id` | int | ID модели |

**Тело запроса (JSON):**

```json
{
  "like": 1
}
```

Где `like`:

- `1` — поставить лайк
- `0` — убрать лайк

**Ответ `200`:**

```json
{
  "id": 1,
  "title": "Per User",
  "is_liked": 1,
  "likes_count": 4
}
```

**Ответы:**

| Код | Описание |
| --- | --- |
| `200` | Лайк поставлен/убран |
| `400` | Невалидный ID / `like` не 0 и не 1 |
| `404` | Модель не найдена |

---

## Сводная таблица эндпоинтов

| № | Метод | URL | Content-Type | Body / Query | Ответ |
| --- | --- | --- | --- | --- | --- |
| 1 | GET | `/api/licensings` | — | `?max_commission=` | `200` массив |
| 2 | GET | `/api/licensing/feed` | — | `?id=&next=` | `200` объект |
| 3 | GET | `/api/licensings/draft` | — | — | `200` объект / `404` |
| 4 | POST | `/api/licensings` | `multipart/form-data` | title, description, commission_per_unit, min_for_calc, image, video | `201` объект |
| 5 | PUT | `/api/licensings/{id}/publish` | `application/json` | `{}` | `200` объект |
| 6 | DELETE | `/api/licensings/{id}` | — | — | `200` пусто |
| 7 | POST | `/api/licensings/{id}/like` | `application/json` | `{"like": 1}` | `200` объект |
| 8 | POST | `/api/users/register` | `application/json` | `{"login": "...", "password": "..."}` | `201` объект |
| 9 | POST | `/api/users/login` | `application/json` | `{"login": "...", "password": "..."}` | `200` заглушка |
| 10 | POST | `/api/users/logout` | `application/json` | `{}` | `200` заглушка |

---

## Статусы модели лицензирования

| Статус | Каталог | Лента | Редактирование | Удаление |
| --- | --- | --- | --- | --- |
| `draft` | ❌ | ❌ | ✅ (создатель) | ✅ |
| `published` | ✅ | ✅ | ❌ | ✅ (создатель) |
| `deleted` | ❌ | ❌ | ❌ | ❌ |

**Правила переходов:**

- `draft` → `published` (через `PUT /publish`)
- `draft` → `deleted` (через `DELETE`)
- `published` → `deleted` (через `DELETE`)
- `deleted` → ничего (обратный переход запрещён)

---

## Хранилище MinIO

**Бакет:** `licensing-images`

**Правила именования файлов:**

- Изображение: `licensing-{id}-image.{ext}`
- Видео: `licensing-{id}-video.{ext}`

**Допустимые Content-Type:**

- Изображения: `image/jpeg`, `image/png`, `image/webp`
- Видео: `video/mp4`, `video/webm`

**Формат URL:**

```
http://localhost:9000/licensing-images/{filename}
```

**Требование:** бакет должен быть публичным.

---

## Текущий пользователь

Пока не реализована авторизация (до ЛР4), текущий пользователь зафиксирован через singleton:

```go
const currentUserID uint = 1
```

Это значит:

- Все операции идут от имени пользователя с `id=1`
- `is_creator` = 1, если `creator_id` модели = 1
- Удалить/опубликовать можно только модели с `creator_id=1`

---

## Стек

| Компонент | Технология |
| --- | --- |
| Язык | Go 1.20+ |
| Веб-фреймворк | Gin |
| ORM | GORM v2 |
| БД | PostgreSQL 18 |
| Хранилище | MinIO |
| Конфиг | Viper + dotenv |
| Логи | logrus |

---

## Запуск

```bash
docker-compose up -d
go run cmd/migrate/main.go
go run cmd/licensing-calc/main.go
```

**Порты:**

- Приложение: `8080`
- PostgreSQL: `5433`
- Adminer: `8081`
- MinIO API: `9000`
- MinIO Console: `9001`