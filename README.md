## Hex Loader

[![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Fyne](https://img.shields.io/badge/Fyne-UI-blueviolet)](https://fyne.io/)
![Логотип](/internal/assets/hex-loader.png)

**GUI-загрузчик hex-файлов** на базе **Arduino CLI**, созданный на **Go** с использованием **Fyne** для **Linux/MacOS/Linux**.

### 🚀 Что умеет Hex Loader?

- Проверка прав пользователя Linux - предупреждает о необходимости группы `dialout` (`uucp`).
- Выбирать **HEX**-файлы через стандартный диалог.
- Сохранение последнего выбранного HEX-файла.
- Обнаруживать подключенные платы (даже если они не распознаны `arduino-cli`).
- Предлагать выбрать `FQBN` вручную для нераспознанных плат (из списка или вводом своего).
- Загружать прошивку на плату с использованием `arduino-cli`.
- Работать без интернета (используя локально установленные ядра и инструменты).
- Работать в **Alt Linux p11** и других Linux-дистрибутивах с установленными зависимостями.

![Img](/screenshots/1.png)
![Img](/screenshots/2.png)
![Img](/screenshots/3.png)
![Img](/screenshots/4.png)
![Img](/screenshots/5.png)
![Img](/screenshots/6.png)

---

### Проверить работу Hex Loader

1. Создать скетч.
```cpp
// big_sketch.ino
// Скетч занимает ~30 КБ в памяти Arduino Uno
// Большой массив данных
const int arraySize = 15000;
unsigned char bigArray[arraySize];git tag -d v0.1.0
git push --delete origin v0.1.0
# (если релиз остался — удалить вручную на GitHub)
git tag v0.1.0
git push origin v0.1.0

void setup() {
  Serial.begin(9600);
  // Заполняем массив данными
  for (int i = 0; i < arraySize; i++) {
    bigArray[i] = i % 256;
  }
  Serial.println("Большой скетч загружен!");
}

void loop() {
  // Ничего не делаем
  delay(1000);
}
```
2. Скомпилировать его в **Arduino IDE**, сохранить HEX-файл (Скетч → Экспортировать скомпилированный бинарный файл).
3. Загрузить полученный HEX-файл в плату с помощью **Hex Loader**.

---

#### Опционально
**Проверка работы `arduino-cli`** (подключённые платы):
```bash
arduino-cli board list
```

---

### 🌍 Arduino CLI не может загрузить индексы в РФ из-за санкций

**Варианты решения проблемы:**

1. **Используйте системный VPN** — самый надёжный способ.
2. **Скачайте и распакуйте вручную** файлы индексов:
   - `https://downloads.arduino.cc/packages/package_index.tar.bz2`
   - `https://downloads.arduino.cc/libraries/library_index.tar.bz2`

   в папку `~/.arduino15`.
3. **Перенесите папку `~/.arduino15/packages/`** с другого компьютера, где уже установлены ядра.

**После переноса восстановите права (не протестировано!):**
```bash
# Восстановить владельца
chown -R $USER:$USER ~/.arduino15/
# Права на файлы и папки
find ~/.arduino15/packages/ -type f -exec chmod 644 {} \;
find ~/.arduino15/packages/ -type d -exec chmod 755 {} \;
# Исполняемые файлы
find ~/.arduino15/packages/ -name "avrdude" -exec chmod +x {} \;
find ~/.arduino15/packages/ -name "*.so*" -exec chmod +x {} \;
find ~/.arduino15/packages/ -path "*/bin/*" -exec chmod +x {} \;
```

---

## 📄 Лицензия

- Этот проект распространяется под лицензией **MIT**. Подробнее см. в файле [LICENSE](LICENSE).
- This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Благодарности

Программа разработана при участии **AI-ассистента** (DeepSeek AI и Qwen3.8-Max) и доработана на основе обратной связи сообщества.

Если у вас есть идеи по улучшению, создавайте **Issue** или отправляйте **Pull Request**!

---

## 📝 Дополнительная информация

- Скрипт тестировался в **ALT Education 11.2** и **Ubuntu 22.04/24.04**, но должен работать в любом дистрибутиве Linux с графическим интерфейсом.
- Для корректной установки плат, перед запуском Hex Loader запустите `arduino-cli` с VPN, чтобы загрузить индексы.

---

### 🛠️ Планы по улучшению

- [x] Сохранение последнего выбранного FQBN
- [x] Автообновление списка портов
- [x] Встроить иконку в окно приложения
- [x] Расширить по высоте и сделать контур поля выбора FQBN в окне "Выбор FQBN"
- [ ] Обновить инструкцию по назначению прав на `~/.arduino15/packages/`
- [x] Испытать бинарник в разных linux
- [ ] Испытать бинарник в разных MacOS
- [ ] Испытать бинарник в разных Windows
- [x] Сборка в AppImage
- [ ] Испытать AppImage в разных linux (включая Alt Linux 10)
- [ ] Упаковка в DEB-пакет (Debian/Ubuntu)
- [ ] Упаковка в RPM-пакет (Fedora/RedHat)
- [ ] Упаковка в RPM-пакет Alt Linux 10
- [ ] Упаковка в RPM-пакет Alt Linux 11
- [ ] Сборка в Степлер на Alt Space