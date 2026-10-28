## CI/CD

### Управление релизами с помощью тэгов

Получить локальные теги:
```shell
git tag
```
Получить удалённые теги на (remote)
```shell
git ls-remote --tags origin
```
Удалить локально
```shell
git tag -d v0.0.1
git tag -d v0.1.0
git tag -d v0.1.1
# и т.д. — перечислите все, что не нужны
```
Удалить на GitHub
```shell
git push --delete origin v0.0.1
git push --delete origin v0.1.0
git push --delete origin v0.1.1
```
Новый релиз

Подготовим изменения
```shell
git add .
```
Добавим коммит
```shell
git commit -m "feat: adaptive theme support"
```
Запушить
```shell
git push
```
Создайте тег
```shell
git tag v0.1.0
```
```shell
git push origin v0.1.0
```

- PATCH (0.1.0 → 0.1.1): исправили баг, ничего нового.
- MINOR (0.1.1 → 0.2.0): добавили функцию (например, прогресс-бар, сохранение FQBN).
- MAJOR (0.2.0 → 1.0.0): сломали совместимость (например, изменили формат конфига) или вышли в стабильную версию.

В целом для перевыпуска достаточно этих команд:
```shell
git tag -d v0.1.0
git push --delete origin v0.1.0
# (если релиз остался — удалить вручную на GitHub)
git tag v0.1.0
git push origin v0.1.0
```