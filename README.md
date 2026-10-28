## Hex Loader

[![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Fyne](https://img.shields.io/badge/Fyne-UI-blueviolet)](https://fyne.io/)


![Логотип](/internal/assets/hex-loader.png)

**GUI-загрузчик hex-файлов** на базе **Arduino CLI**, созданный на **Go** с использованием **Fyne** для Linux.

### 🚀 Что умеет Hex Loader?

- Проверка прав пользователя Linux - предупреждает о необходимости группы `dialout` (`uucp`).
- Выбирать **HEX**-файлы через стандартный диалог.
- Сохранение последнего выбранного HEX-файла.
- Обнаруживать подключенные платы (даже если они не распознаны `arduino-cli`).
- Предлагать выбрать `FQBN` вручную для нераспознанных плат (из списка или вводом своего).
- Загружать прошивку на плату с использованием `arduino-cli`.
- Работать без интернета (используя локально установленные ядра и инструменты).
- Работать в **Alt Linux p11** и других Linux-дистрибутивах с установленными зависимостями.

![Img](/assets/images/1.png)
![Img](/assets/images/2.png)
![Img](/assets/images/3.png)
![Img](/assets/images/4.png)
![Img](/assets/images/5.png)
![Img](/assets/images/6.png)

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

**После переноса восстановите права:**
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

### Проверить работу Hex Loader

1. Создать скетч.
```cpp
// big_sketch.ino
// Скетч занимает ~30 КБ в памяти Arduino Uno
// Большой массив данных
const int arraySize = 15000;
unsigned char bigArray[arraySize];

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
- [ ] Расширить по высоте и сделать контур поля выбора FQBN в окне "Выбор FQBN"
- [ ] Обновить инструкцию по назначению прав на `~/.arduino15/packages/`
- [ ] Обновить инструкцию по копированию `icon.png` и `hexloader` в `./hexloader.AppDir/`
- [ ] Испытать бинарник в разных linux
- [x] Сборка в AppImage в Alt 11
- [ ] Сборка в AppImage в Alt 10
- [ ] Упаковка в DEB-пакет (Debian/Ubuntu)
- [ ] Упаковка в RPM-пакет (Fedora/RedHat)
- [ ] Сборка в Степлер на Alt Space
- [ ] Упаковка в RPM-пакет для Alt Linux 10
- [ ] Упаковка в RPM-пакет для Alt Linux 11