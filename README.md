# Telegram Боты для Управления Анкетами

Этот проект представляет собой систему из двух Telegram-ботов:

- **Клиентский бот:** Позволяет пользователям создавать, просматривать и удалять свои анкеты. При регистрации бот проводит диалог, в котором запрашивает такие данные, как имя, возраст, рост, вес, инвентарь, фото, ранг и команда. При просмотре анкеты пользователем выводятся только публичные поля (без внутреннего ID и TG-username). А так же отмечаться или не отмечатся в событиях, и при отметке получать валюту.

- **Админский бот:** Позволяет администратору управлять анкетами пользователей. Для доступа к функциям админского бота требуется аутентификация по паролю (пароль задается в коде). После успешной аутентификации администратор может:
  - Просмотреть список всех анкет (выводятся только уникальный ID анкеты и TG-username того, кто отправил анкету).
  - Просмотреть подробную информацию выбранной анкеты (включая внутренний ID, TG-username, фото и прочие поля).
  - Редактировать отдельные поля анкеты.
  - Удалять анкеты по их уникальному ID.
  - Добавлять валюту полтзователю по уникальному ID.
  - Создавать собития с названием и назначением валюты за отмеки на немю.


При успешном создании анкеты клиентским ботом полная информация (с внутренним ID и TG-username) автоматически пересылается админскому боту.

> **Важно:**  
> Для работы проекта используются:
> - [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — чисто-Go драйвер для SQLite,  
> - [github.com/go-telegram-bot-api/telegram-bot-api/v5](https://pkg.go.dev/github.com/go-telegram-bot-api/telegram-bot-api/v5) — библиотека для работы с Telegram API.

---

## Функциональность

### Клиентский бот

- **Диалог регистрации анкеты:**  
  При выполнении команд `/start` или `/createprofile` бот начинает диалог, последовательно запрашивая:
  - **Имя**
  - **Возраст** (целое число)
  - **Рост** (дробное число, например, 175.5)
  - **Вес** (дробное число, например, 70.2)
  - **Инвентарь** (текстовое описание)
  - **Фото** (file_id или URL изображения)
  - **Ранг**
  - **Команда**

- **Команды, доступные для пользователя:**
  - `/start` — выводит справочное меню с описанием всех доступных команд и запускает регистрацию, если анкета отсутствует.
  - `/createprofile` — инициирует создание новой анкеты.
  - `/profile` — позволяет просмотреть созданную анкету (без внутренних данных: ID и TG-username).
  - `/deleteprofile` — удаляет анкету пользователя.
  - Дополнительные команды для изменения отдельных полей, например, `/setname`, `/setage` и т.д.
  - `/help` — выводит список всех доступных команд клиентского бота.
  - `/attend <ID>` — отметиться на активном событии, получив валюту, указанную в этом событии.
  - `/unattend <ID>` — отменить участие в активном событии.



### Админский бот

- **Аутентификация:**  
  Перед использованием функционала админского бота требуется аутентификация. Бот принимает команду:
/auth <пароль>
При правильном вводе пароля (например, `SuperSecret123`, заданного в коде) администратор станет аутентифицированным, и будут доступны остальные команды.

- **Команды, доступные администратору:**
- `/help` — выводит список всех команд админского бота.
- `/auth <пароль>` — аутентификация администратора.
- `/allprofiles` — выводит список всех анкет; для каждой анкеты показывается только уникальный ID и TG-username пользователя.
- `/viewprofile <ID>` — просмотр подробной информации анкеты с уникальным ID. В профиле отображены все поля, включая фото (если оно задано).
- `/editprofile <ID> <поле> <значение>` — редактирование выбранного поля анкеты. Допустимые поля: `name`, `age`, `height`, `weight`, `inventory`, `photo`, `rank`, `team`.
- `/deleteprofilebyid <ID>` — удаление анкеты по уникальному ID.

---

---
- **Создание события:**
- 
`/createevent <название|валюта|количество>` — создание нового события для начисления валюты. Например:

`/createevent Сбор пиастров|piastres|100`

Название события: Указывается как текст.

Валюта: Тип валюты (piastres или oblomki).

Количество: Целое число, указывающее, сколько валюты будет начислено за участие.

После создания событие сохраняется в базе данных и уведомление рассылается всем зарегистрированным пользователям.

- **Управление валютой:**
- `/addcurrency <ID> <тип валюты> <количество>` — добавление валюты в профиль пользователя. Например:

- `/addcurrency 1 piastres 50`

ID: Уникальный идентификатор профиля пользователя.

Тип валюты: piastres или oblomki.

Количество: Целое число, которое будет добавлено.

## Установка

1. **Клонирование репозитория:**

   git clone https://github.com/yourusername/telegram-bot.git
   cd telegram-bot
Установка Go (если еще не установлен): Скачайте и установите последнюю версию Go.

Установка зависимостей: Находясь в корневой папке проекта, выполните:


go mod tidy
Настройка конфигурации:

Токены ботов:

Токен клиентского бота можно передать через переменную окружения BOT_TOKEN или задать в коде main.go.

Токен админского бота прописан в main.go (например, 8051322387:AAG4pnS8hch0JHBWgVS1qLt12JQCjd_JyB0).

Админский пароль: Пароль для аутентификации админского бота задан в файле handlers/admin_bot.go (например, SuperSecret123). При необходимости измените его.

Chat ID для администратора: В функциях, где админский бот отправляет уведомления (например, SendProfileToSecondBot), замените значение 123456789 на фактический chat_id администратора, с которого начат диалог с ботом.

База данных: База данных хранится в файле bot.db (SQLite). При внесении изменений в схему таблицы может потребоваться удалить существующий файл bot.db или выполнить миграцию.

Запуск ботов:


go run main.go

После запуска одновременно начнут работать два update-цикла: один для клиентского бота, второй — для админского бота.
