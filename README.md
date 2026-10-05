# Проектирование систем и продуктовая веб-разработка 2026 (Разработка Интернет Приложений)

### Тема: Расчёт стоимости лицензирования ПО в зависимости от модели. Услуги - модели лицензирования (per_user, per_core, subscription и тд), заявка - расчёт общей стоимости лицензий для заданной конфигурации системы и количества пользователей

## API

Запуск: `docker compose up -d`, `go run ./cmd/migrate`, `go run ./cmd/licensing-calc`. Все методы под `http://localhost:8080/api`.

- При ошибке тело пустое, код `400` или `404`. Незаполненные поля приходят как `null`.
- Текущий создатель пока задан на бэкенде константой (id = 1).

Тело ответа «модель лицензирования»:

```json
{
  "id": 16,
  "title": "Per User",
  "description": "Модель лицензирования на одного пользователя",
  "image_url": "http://localhost:9000/licensing-images/licensing-16-image.jpg",
  "video_url": "http://localhost:9000/licensing-images/licensing-16-video.mp4",
  "commission_per_unit": 120,
  "min_for_calc": 5,
  "likes_count": 1
}
```

### Модели лицензирования

| Метод | URL | Тело/параметры | Ответ |
| --- | --- | --- | --- |
| GET | `/licensings` | `?max_commission=N`: если > 0, комиссия от N ₽ и ниже | 200, массив опубликованных моделей; у каждой дополнительно `is_creator`: 1, если её создал текущий создатель, иначе 0 |
| GET | `/licensings/feed` | `?id=N&next=true`: следующая после id; без параметров — первая; с `id=N` — конкретная | 200, модель; 404 |
| GET | `/licensings/draft` | — | 200, черновик текущего создателя; 404, если нет |
| POST | `/licensings` | form-data: `title`, `image` (jpeg/png/webp), `video` (webm/mp4) | 201, модель; 400, если черновик уже есть или данные некорректны |
| PUT | `/licensings/:id/publish` | JSON, все обязательны: `description`, `commission_per_unit`, `min_for_calc` | 200, модель; 404, если не черновик текущего создателя |
| DELETE | `/licensings/:id` | — (логическое удаление опубликованной модели текущего создателя) | 200 без тела; 404, если модель чужая или не опубликована |
| POST | `/licensings/:id/like` | JSON `{"like": 1}` ставит, `{"like": 0}` снимает | 200, модель; 404, если не опубликована |

### Создатели

| Метод | URL | Тело/параметры | Ответ |
| --- | --- | --- | --- |
| POST | `/users/register` | JSON `{login, password}` | 201, `{id, login, is_moderator}`; 400, если логин занят |
| POST | `/users/login` | — (заглушка) | 200 без тела |
| POST | `/users/logout` | — (заглушка) | 200 без тела |

## Таблицы БД

### licensings — модели лицензирования

| Поле | Тип | Описание |
| --- | --- | --- |
| `id` | bigint | первичный ключ |
| `title` | varchar(100) | название |
| `description` | varchar(500) | краткое описание, заполняется при публикации |
| `status` | varchar(15) | черновик, опубликован или удалён |
| `image_url` | varchar(255) | ссылка на фото в MinIO |
| `video_url` | varchar(255) | ссылка на видео в MinIO |
| `commission_per_unit` | integer | комиссия за 1 единицу, ₽ |
| `min_for_calc` | integer | минимальное число для расчёта |
| `created_at` | timestamptz | дата создания |
| `creator_id` | bigint | создатель, внешний ключ на `users` |
| `published_at` | timestamptz | дата публикации |

### users — пользователи

| Поле | Тип | Описание |
| --- | --- | --- |
| `id` | bigint | первичный ключ |
| `login` | varchar(25) | логин, уникальный |
| `password` | varchar(100) | пароль |
| `is_moderator` | boolean | признак модератора |

### likes — лайки

| Поле | Тип | Описание |
| --- | --- | --- |
| `id` | bigint | первичный ключ |
| `user_id` | bigint | кто поставил лайк, внешний ключ на `users` |
| `licensing_id` | bigint | какой модели, внешний ключ на `licensings` |