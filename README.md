# Hex Loader — загрузчик HEX в Arduino (Go + Fyne)

Изменение для Workflow

GUI-приложение для прошивки Arduino через `arduino-cli`.

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)
![Fyne](https://img.shields.io/badge/Fyne-v2.8-blue)
![CI/CD](https://img.shields.io/badge/CI%2FCD-GitHub%20Actions-success)

---

## 📖 Оглавление

- [О проекте](#about)
- [Стек технологий](#stack)
- [Структура проекта](#structure)
- [Этапы разработки](#stages)
- [CI/CD Pipeline](#cicd)
- [Скачать бинарники](#download)
- [Локальная сборка](#local-build)
- [Требования для работы](#requirements)
- [Как пользоваться](#usage)

---

<a name="about"></a>
## 🎯 О проекте

**Hex Loader** — это кроссплатформенное GUI-приложение на Go, которое позволяет прошивать Arduino-платы готовыми HEX-файлами через утилиту `arduino-cli`.

Приложение автоматически:

- Определяет подключённые платы через `arduino-cli board list`
- Позволяет выбрать HEX-файл с фильтром `.hex`
- Запоминает последний выбранный файл через `Preferences`
- Показывает статус прошивки в реальном времени
- Проверяет наличие прав на последовательные порты (группа `dialout`)

---

<a name="stack"></a>
## 🛠️ Стек технологий

| Технология | Назначение |
|---|---|
| **Go 1.23** | Язык программирования |
| **Fyne v2.8** | Кроссплатформенный GUI-фреймворк |
| **arduino-cli** | Утилита для прошивки Arduino |
| **CGO + GLFW** | Нативные библиотеки для окна (компилируются из C) |
| **Docker** | Локальная сборка и тесты |
| **GitHub Actions** | CI/CD pipeline |
| **MSYS2 + MinGW-w64** | GCC для CGO на Windows |

---

<a name="structure"></a>
## 📁 Структура проекта

```
hex-loader/
├── .github/
│   └── workflows/
│       └── ci.yml              # GitHub Actions: CI/CD
├── Dockerfile.test             # Образ для локальной сборки
├── main.go                     # Весь код приложения
├── go.mod                      # Модуль Go
├── go.sum                      # Контрольные суммы зависимостей
├── icon.png                    # Иконка приложения
├── README.md                   # Этот файл
└── .gitignore
```

---

<a name="stages"></a>
## 🚀 Этапы разработки

### Этап 1. Создание структуры проекта

Создана базовая структура с файлами `go.mod`, `main.go`, `Dockerfile.test`, `.github/workflows/ci.yml` и `.gitignore`.

```bash
mkdir -p hex-loader/.github/workflows
cd hex-loader
```

### Этап 2. Написание GUI-кода

В `main.go` реализовано:

- **Кастомная тема** (`myTheme`) — увеличенный размер текста, убранные тени
- **Кастомная кнопка с обводкой** (`buttonWithBorder`) — через `CreateRenderer()`
- **Проверка `arduino-cli`** — если утилиты нет, кнопки блокируются
- **Работа с портами** — через `exec.Command("arduino-cli", "board", "list")`
- **Асинхронная прошивка** — через `goroutine` + `fyne.Do` для обновления UI

### Этап 3. Локальная сборка в Docker

Собран образ `hex-loader-test` с GUI-зависимостями:

```bash
docker build -f Dockerfile.test -t hex-loader-test .
```

Сгенерирован `go.sum`:

```bash
MSYS_NO_PATHCONV=1 docker run --rm -u "$(id -u):$(id -g)" \
  -e HOME=/tmp -e GOPATH=/tmp/go -e GOCACHE=/tmp/go-cache \
  -v "$(pwd)":/app -v ~/.go-docker-cache:/tmp/go \
  -w /app hex-loader-test go mod tidy
```

### Этап 4. Форматирование и smoke-check

```bash
MSYS_NO_PATHCONV=1 docker run --rm -u "$(id -u):$(id -g)" \
  -e HOME=/tmp -e GOPATH=/tmp/go -e GOCACHE=/tmp/go-cache \
  -v "$(pwd)":/app -v ~/.go-docker-cache:/tmp/go \
  -w /app hex-loader-test sh -c "gofmt -w . && gofmt -l ."

MSYS_NO_PATHCONV=1 docker run --rm -u "$(id -u):$(id -g)" \
  -e HOME=/tmp -e GOPATH=/tmp/go -e GOCACHE=/tmp/go-cache \
  -e CGO_ENABLED=1 \
  -v "$(pwd)":/app -v ~/.go-docker-cache:/tmp/go \
  -w /app hex-loader-test go build -o /tmp/hex-loader .
```

### Этап 5. Публикация на GitHub

```bash
git init
git add .
git commit -m "Initial commit: Hex Loader (Go+Fyne) with CI/CD"
git branch -M main
git remote add origin https://github.com/VellinHellin/hex-loader.git
git push -u origin main
```

### Этап 6. Первый запуск CI

Workflow `Hex Loader CI/CD` запустился автоматически после push в `main`. Job `test` прошел успешно за **2m 58s**.

### Этап 7. Создание релиза

```bash
git tag v0.1.0
git push origin v0.1.0
```

Job `release` развернулся в матрицу из 3 задач и за **8m 4s** собрал 3 бинарника параллельно.

---

<a name="cicd"></a>
## 🔄 CI/CD Pipeline

```
push в main ──► job test
                 ├─ gofmt -l .
                 ├─ go vet ./...
                 └─ go build

push тега v* ──► job test ──► job release (matrix)
                                ├─ ubuntu-latest   → hex-loader-linux-x64
                                ├─ macos-14        → hex-loader-macos-arm64
                                └─ windows-latest  → hex-loader-windows-x64.exe
                                       │
                                       ▼
                                GitHub Release v0.1.0
```

**Ключевые особенности:**

- **Матричная сборка** — 3 ОС параллельно, каждая со своим нативным компилятором
- **MSYS2 на Windows** — для CGO нужен GCC, путь определяется динамически
- **`-H windowsgui`** — скрывает консольное окно в Windows-бинарнике
- **`-ldflags "-X main.version=..."`** — версия подставляется из git-тега
- **`softprops/action-gh-release@v2`** — публикует артефакты в Release

---

<a name="download"></a>
## 📦 Скачать бинарники

👉 [**Releases**](https://github.com/VellinHellin/hex-loader/releases)

**Текущая версия: v0.1.0**

| ОС | Файл | Размер |
|---|---|---|
| 🐧 Linux x64 | `hex-loader-linux-x64` | 22.6 MB |
| 🍎 macOS ARM64 | `hex-loader-macos-arm64` | 21.9 MB |
| 🪟 Windows x64 | `hex-loader-windows-x64.exe` | 22.9 MB |

---

<a name="local-build"></a>
## 🧪 Локальная сборка

Требования: Docker Desktop, Git Bash.

```bash
docker build -f Dockerfile.test -t hex-loader-test .

MSYS_NO_PATHCONV=1 docker run --rm -u "$(id -u):$(id -g)" \
  -e HOME=/tmp -e GOPATH=/tmp/go -e GOCACHE=/tmp/go-cache \
  -e CGO_ENABLED=1 \
  -v "$(pwd)":/app -v ~/.go-docker-cache:/tmp/go \
  -w /app hex-loader-test go build -o hex-loader .
```

---

<a name="requirements"></a>
## ⚠️ Требования для работы

1. **`arduino-cli` в PATH** — без него кнопки заблокированы
2. **Права на порты (Linux):**

   ```bash
   sudo usermod -a -G dialout $USER
   ```

   После этого выйдите из системы и зайдите заново

---

<a name="usage"></a>
## 🔧 Как пользоваться

1. **📂 Выбрать HEX** — диалог с фильтром `.hex`. Путь сохраняется в настройках
2. **🔌 Выбрать порт** — вызов `arduino-cli board list` покажет список плат
3. **⬆️ Загрузить** — начнётся прошивка. Статус отображается внизу окна

---

## 📋 Что было изучено

- ✅ Создание GUI на Go с помощью **Fyne**
- ✅ Кастомные темы и виджеты (через `CreateRenderer`)
- ✅ Работа с `os/exec` для вызова внешних утилит
- ✅ CGO и компиляция C-библиотек (GLFW) внутри Go
- ✅ Матричные сборки в **GitHub Actions** (3 ОС параллельно)
- ✅ Настройка **MSYS2 + MinGW-w64** для CGO на Windows
- ✅ Динамический поиск `gcc.exe` на Windows runner
- ✅ Автоматическая публикация в **GitHub Releases**
- ✅ Семантическое версионирование (`v0.1.0`)
- ✅ Подстановка версии из тега через `-ldflags "-X main.version=..."`

---

## 👤 Автор

**VellinHellin** — [GitHub](https://github.com/VellinHellin)
