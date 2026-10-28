# Локальная сборка Hex-Loader

## Системно (на хосте)

### 📦 Alt Linux p11

**Установка системных зависимостей** (делается один раз):
```shell
epmi --auto \
    gcc-c++ \
    golang \
    rpm-build-golang \
    libXcursor-devel \
    libX11-devel \
    libGL-devel \
    libXrender-devel \
    libXfixes-devel \
    libXi-devel \
    libXinerama-devel \
    libXrandr-devel \
    libwayland-egl-devel \
    libXxf86vm-devel \
    libxkbcommon-devel \
    libwayland-cursor-devel \
    arduino-cli
```
Создание Go-проекта
```bash
mkdir hex-loader && cd hex-loader
micro main.go   # вставьте код
go mod init hex-loader
go get fyne.io/fyne/v2@latest
go install fyne.io/tools/cmd/fyne@latest   # опционально
go mod tidy
go run main.go   # проверка
```
Сборка Go-проекта с оптимизацией размера бинарника:
```bash
go build -ldflags="-s -w" -trimpath -o hex-loader main.go
```
- `-s` — удаляет отладочную информацию.
- `-w` — удаляет DWARF-таблицы.
- `-trimpath` — убирает пути к исходникам из бинарника.

**Запуск:**
```bash
./hex-loader
```
**Проверка зависимостей:**
```bash
ldd hex-loader
```

## Distrobox

- [Установить Distrobox в Alt Linux 11](https://gitflic.ru/project/rurewa/altlinux/blob?file=content/Distrobox.md&branch=master)

Зайти в контейнер с Альтом
```shell
distrobox enter altbox11
```
Установить необходимые библиотеки
```shell
epmi --auto \
    gcc-c++ \
    golang \
    rpm-build-golang \
    libXcursor-devel \
    libX11-devel \
    libGL-devel \
    libXrender-devel \
    libXfixes-devel \
    libXi-devel \
    libXinerama-devel \
    libXrandr-devel \
    libwayland-egl-devel \
    libXxf86vm-devel \
    libxkbcommon-devel \
    libwayland-cursor-devel \
    arduino-cli
```
Собрать
```shell
go build -buildvcs=false -o hex-loader .
```
или с оптимизацией бинарника
```shell
go build -ldflags="-s -w" -trimpath -o hex-loader main.go
```
Опционально дать права на dialout
```shell
sudo usermod -a -G dialout $USER
```
Запустить
```shell
./hex-loader
```

## Сборка и запуск приложения Hex-Loader в Docker

(не самая удачная идея, т.к. запущенное графическое приложение имеет крайне ограниченный функционал)

Проверялось в **Альт Линукс 11 (X11)**

Создать временный контейнер и войти в него:
```shell
docker run --rm -it   -e DISPLAY=$DISPLAY   -v /tmp/.X11-unix:/tmp/.X11-unix:ro   -v $HOME/.Xauthority:/root/.Xauthority:ro   -v "$PWD":/app   -w /app   hex-loader-builder   bash
```
собрать приложение:
```shell
go build -buildvcs=false -o hex-loader .
```
запустить приложение:
```shell
./hex-loader
```
открыть доступ к X-серверу:
```shell
xhost +local:docker
```
закрыть доступ к X-серверу:
```shell
xhost -local:docker
```

***

### Создание Appimage

1. Создаём отдельный каталог для AppImage-пакета.
```shell
mkdir AppImage
```
переходим в него
```shell
cd AppImage
```
2. Создаём каталог `./hexloader.AppDir/`.
```shell
mkdir ./hexloader.AppDir/
```
3. Копируем в `./hexloader.AppDir/` бинарник и иконку.
```shell
cp ../hexloader ./hexloader.AppDir/   # если бинарник в корне проекта
cp ../icon.png ./hexloader.AppDir/icon.png
```
переходим `./hexloader.AppDir/`
```shell
cd ./hexloader.AppDir/
```
4. создаём файл `hexloader.desktop`.
```shell
micro hexloader.desktop
```
Содержимое `.desktop` файла
```.desktop
[Desktop Entry]
Type=Application
Name=Hex Loader
Comment=GUI загрузчик HEX файлов для Arduino
Exec=hexloader
Icon=icon
Categories=Development;Electronics;
Terminal=false
```
> **Важно: имя иконки в .desktop должно совпадать с именем файла иконки (без расширения)**
Правильная структура `./hexloader.AppDir/`
```
./hexloader.AppDir/
├── hexloader          # Ваш бинарник (исполняемый)
├── hexloader.desktop  # Файл .desktop
├── icon.png      # Иконка (имя совпадает с Icon в .desktop)
└── AppRun -> hexloader # Ссылка на бинарник
```
выходим из каталога `./hexloader.AppDir/`.
```shell
cd ..
```
5. выполняем загрузку `appimagetool`.
```shell
wget -c "https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage"
```
делаем `appimagetool` исполняемым
```shell
chmod +x appimagetool-x86_64.AppImage
```
выполняем сборку AppImage-пакета
```shell
ARCH=x86_64 ./appimagetool-x86_64.AppImage ./hexloader.AppDir/
```
запускаем получившийся AppImage-пакет
```shell
./Hex_Loader-x86_64.AppImage
```