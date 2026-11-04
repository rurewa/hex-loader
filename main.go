package main

import (
	"bytes"
	"context"
	"debug/buildinfo"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"hex-loader/internal/assets"
)

// version устанавливается при сборке через -ldflags "-X main.version=..."
var version = "dev"

// sizeVersionText — кастомный размер для строки с версией.
const sizeVersionText fyne.ThemeSizeName = "versionText"

// Таймауты для вызовов arduino-cli.
const (
	timeoutQuick  = 15 * time.Second  // board list, core list
	timeoutUpload = 120 * time.Second // upload (большие прошивки на ESP32)
)

// BoardInfo хранит информацию о подключенной плате
type BoardInfo struct {
	Port string
	FQBN string
}

// boardEntry — один элемент в ответе `arduino-cli board list --json`.
// Используется и для {"detected_ports": [...]}, и для {"boards": [...]},
// и для старого формата-массива [...] — поля совпадают.
type boardEntry struct {
	MatchingBoards []struct {
		Name string `json:"name"`
		FQBN string `json:"fqbn"`
	} `json:"matching_boards"`
	Port struct {
		Address string `json:"address"`
	} `json:"port"`
}

// myTheme — кастомная тема: увеличенный текст, убранная тень,
// адаптивные цвета из системной темы.
type myTheme struct {
	fyne.Theme
}

func (t myTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 16
	case theme.SizeNamePadding:
		return 14
	case theme.SizeNameHeadingText:
		return 22
	case sizeVersionText:
		return 12
	default:
		return t.Theme.Size(name)
	}
}

func (t myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameShadow:
		return color.Transparent
	default:
		return t.Theme.Color(name, variant)
	}
}

// buttonWithBorder — кастомная кнопка с обводкой
type buttonWithBorder struct {
	widget.Button
}

func newButtonWithBorder(label string, onTap func()) *buttonWithBorder {
	btn := &buttonWithBorder{}
	btn.ExtendBaseWidget(btn)
	btn.Text = label
	btn.OnTapped = onTap
	return btn
}

func (b *buttonWithBorder) CreateRenderer() fyne.WidgetRenderer {
	return &buttonRenderer{btn: b}
}

type buttonRenderer struct {
	btn       *buttonWithBorder
	label     *canvas.Text
	bg        *canvas.Rectangle
	border    *canvas.Rectangle
	objects   []fyne.CanvasObject
	container *fyne.Container
}

func (r *buttonRenderer) init() {
	if r.container != nil {
		return
	}
	r.label = canvas.NewText(r.btn.Text, theme.Color(theme.ColorNameForeground))
	r.label.Alignment = fyne.TextAlignCenter
	r.label.TextSize = theme.TextSize()

	r.bg = canvas.NewRectangle(theme.Color(theme.ColorNameButton))
	r.bg.CornerRadius = 8

	r.border = canvas.NewRectangle(color.Transparent)
	r.border.StrokeWidth = 2
	r.border.StrokeColor = theme.Color(theme.ColorNameForeground)
	r.border.CornerRadius = 8

	r.container = container.NewStack(r.bg, r.border, r.label)
	r.objects = []fyne.CanvasObject{r.container}
}

func (r *buttonRenderer) Layout(size fyne.Size) {
	r.init()
	r.container.Resize(size)
}

func (r *buttonRenderer) MinSize() fyne.Size {
	r.init()
	minSize := r.container.MinSize()
	return fyne.NewSize(minSize.Width+40, minSize.Height+16)
}

func (r *buttonRenderer) Refresh() {
	r.init()
	r.label.Text = r.btn.Text
	r.label.Color = theme.Color(theme.ColorNameForeground)
	r.label.TextSize = theme.TextSize()

	if r.btn.Disabled() {
		r.bg.FillColor = theme.Color(theme.ColorNameDisabled)
		r.border.StrokeColor = theme.Color(theme.ColorNameDisabled)
	} else {
		r.bg.FillColor = theme.Color(theme.ColorNameButton)
		r.border.StrokeColor = theme.Color(theme.ColorNameForeground)
	}
	r.bg.CornerRadius = 8
	r.border.CornerRadius = 8

	r.label.Refresh()
	r.bg.Refresh()
	r.border.Refresh()
	canvas.Refresh(r.btn)
}

func (r *buttonRenderer) Objects() []fyne.CanvasObject {
	r.init()
	return r.objects
}

func (r *buttonRenderer) Destroy() {}

// showCustomInformation — кастомный информационный диалог с кнопкой с обводкой.
// Используем NewCustomWithoutButtons, чтобы Fyne не добавлял невидимую
// dismiss-кнопку под контентом.
func showCustomInformation(title, message string, w fyne.Window) {
	label := widget.NewLabel(message)
	label.Wrapping = fyne.TextWrapWord

	var d *dialog.CustomDialog

	d = dialog.NewCustomWithoutButtons(title, container.NewVBox(
		label,
		container.NewCenter(newButtonWithBorder("OK", func() {
			if d != nil {
				d.Hide()
			}
		})),
	), w)
	d.Resize(fyne.NewSize(500, 250))
	d.Show()
}

// checkArduinoCLI проверяет, доступен ли arduino-cli в PATH
func checkArduinoCLI() error {
	_, err := exec.LookPath("arduino-cli")
	return err
}

// checkUserInGroup проверяет, состоит ли текущий пользователь в указанной группе.
// Группы dialout/uucp существуют только в Linux. На macOS и Windows права
// на последовательные порты работают иначе, поэтому проверка там не имеет смысла
// и не должна показывать ложное предупреждение.
func checkUserInGroup(groupName string) (bool, error) {
	if runtime.GOOS != "linux" {
		return false, nil
	}
	if _, err := exec.LookPath("groups"); err != nil {
		return false, nil
	}

	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}

	cmd := exec.Command("groups", currentUser.Username)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	groups := strings.Fields(string(output))
	for _, g := range groups {
		if g == groupName {
			return true, nil
		}
	}
	return false, nil
}

// resolveVersion возвращает версию для отображения: из -ldflags, либо из build info,
// либо "dev" для локальной сборки.
func resolveVersion() string {
	if version != "" && version != "dev" {
		return version
	}
	if info, err := buildinfo.ReadFile(os.Args[0]); err == nil {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "dev"
}

// showPermissionWarning показывает предупреждение о правах.
// NewCustomWithoutButtons — чтобы не было невидимой dismiss-кнопки.
func showPermissionWarning(w fyne.Window) {
	var d *dialog.CustomDialog

	label := widget.NewLabel(
		"⚠️ У вас нет прав для работы с последовательными портами.\n\n" +
			"Чтобы Hex Loader мог определять и прошивать платы,\n" +
			"добавьте пользователя в группу dialout:\n\n" +
			"  sudo usermod -a -G dialout $USER\n\n" +
			"После этого выйдите из системы и зайдите заново.\n\n" +
			"Вы всё равно можете использовать программу,\n" +
			"но порты могут не определяться.",
	)

	btnOk := newButtonWithBorder("Понятно", func() {
		if d != nil {
			d.Hide()
		}
	})

	content := container.NewVBox(label, btnOk)
	d = dialog.NewCustomWithoutButtons("Внимание", content, w)
	d.Resize(fyne.NewSize(500, 300))
	d.Show()
}

// uploadWithProgress выполняет загрузку с отображением статуса.
// Использует context, чтобы загрузку можно было прервать при закрытии окна.
func uploadWithProgress(ctx context.Context, hexPath, portPath, fqbn string, statusLabel *widget.Label) error {
	// DoAndWait гарантирует, что UI обновится до блокирующего вызова cmd.Run().
	fyne.DoAndWait(func() {
		statusLabel.SetText("⏳ Подождите, идёт загрузка...")
		statusLabel.Refresh()
	})

	uploadCtx, cancel := context.WithTimeout(ctx, timeoutUpload)
	defer cancel()

	cmd := exec.CommandContext(uploadCtx, "arduino-cli",
		"upload",
		"-p", portPath,
		"--fqbn", fqbn,
		"--input-file", hexPath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		if uploadCtx.Err() == context.DeadlineExceeded {
			fyne.Do(func() {
				statusLabel.SetText("❌ Таймаут загрузки")
				statusLabel.Refresh()
			})
			return fmt.Errorf("таймаут загрузки (%s)", timeoutUpload)
		}
		if uploadCtx.Err() == context.Canceled {
			fyne.Do(func() {
				statusLabel.SetText("❌ Загрузка отменена")
				statusLabel.Refresh()
			})
			return fmt.Errorf("загрузка отменена")
		}
		fyne.Do(func() {
			statusLabel.SetText("❌ Ошибка загрузки")
			statusLabel.Refresh()
		})
		return fmt.Errorf("ошибка загрузки: %v: %s", err, stderr.String())
	}

	fyne.Do(func() {
		statusLabel.SetText("✅ Загрузка успешно завершена!")
		statusLabel.Refresh()
	})
	return nil
}

func main() {
	a := app.NewWithID("com.example.hexloader")
	a.Settings().SetTheme(&myTheme{theme.DefaultTheme()})

	w := a.NewWindow("Загрузчик HEX в Arduino")
	w.Resize(fyne.NewSize(700, 540))

	var hexPath string
	var portPath string
	var fqbn string

	// appCtx — отменяется при закрытии окна.
	// От него наследуются контексты для board list и upload,
	// поэтому закрытие окна прерывает и текущую загрузку, и discovery.
	appCtx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()

	var uploadMu sync.Mutex
	isUploading := false

	prefs := a.Preferences()

	lastHexPath := prefs.String("lastHexPath")
	if lastHexPath != "" {
		hexPath = lastHexPath
	}

	hexLabel := widget.NewLabel("Файл не выбран")
	if hexPath != "" {
		hexLabel.SetText(hexPath)
	}

	portLabel := widget.NewLabel("Порт не выбран")
	statusLabel := widget.NewLabel("Готов к работе")

	// --- ИКОНКА В ИНТЕРФЕЙСЕ ---
	iconResource := fyne.NewStaticResource("hex-loader.png", assets.IconPNG)
	iconImage := canvas.NewImageFromResource(iconResource)
	iconImage.FillMode = canvas.ImageFillContain
	iconImage.SetMinSize(fyne.NewSize(64, 64))
	iconImage.Resize(fyne.NewSize(64, 64))
	iconImage.Refresh()

	// Заголовок и версия — через widget.RichText с Alignment: Center.
	title := widget.NewRichText(&widget.TextSegment{
		Text: "Загрузчик HEX файлов в Arduino",
		Style: widget.RichTextStyle{
			SizeName:  theme.SizeNameHeadingText,
			TextStyle: fyne.TextStyle{Bold: true},
			Alignment: fyne.TextAlignCenter,
		},
	})
	title.Wrapping = fyne.TextWrapWord

	versionLabel := widget.NewRichText(&widget.TextSegment{
		Text: "Версия: " + resolveVersion(),
		Style: widget.RichTextStyle{
			SizeName:  sizeVersionText,
			Alignment: fyne.TextAlignCenter,
		},
	})
	versionLabel.Wrapping = fyne.TextWrapWord

	header := container.NewVBox(
		container.NewCenter(iconImage),
		title,
		versionLabel,
	)

	btnSelectHex := newButtonWithBorder("📂 Выбрать HEX", func() {
		fileFilter := storage.NewExtensionFileFilter([]string{".hex"})
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()
			hexPath = reader.URI().Path()
			hexLabel.SetText(hexPath)
			statusLabel.SetText("HEX файл выбран")

			prefs.SetString("lastHexPath", hexPath)
		}, w)
		fileDialog.SetFilter(fileFilter)
		fileDialog.Show()
	})

	// showPortChooser открывает диалог выбора платы из готового списка.
	showPortChooser := func(boards []BoardInfo) {
		if len(boards) == 1 {
			selectBoard(boards[0], &portPath, &fqbn, portLabel, statusLabel, w)
			return
		}

		items := make([]string, len(boards))
		for i, b := range boards {
			label := b.Port
			if b.FQBN != "" && strings.Count(b.FQBN, ":") >= 2 {
				label += " (" + b.FQBN + ")"
			} else {
				label += " (тип не определён)"
			}
			items[i] = label
		}
		selected := widget.NewSelect(items, func(s string) {
			port := strings.Split(s, " ")[0]
			for _, b := range boards {
				if b.Port == port {
					selectBoard(b, &portPath, &fqbn, portLabel, statusLabel, w)
					break
				}
			}
		})
		dialog.ShowCustom("Выберите плату", "OK", selected, w)
	}

	btnSelectPort := newButtonWithBorder("🔌 Выбрать порт", func() {
		boards, err := getAvailableBoards(appCtx)
		if err != nil {
			dialog.ShowError(fmt.Errorf("не удалось получить список плат: %v", err), w)
			return
		}
		if len(boards) == 0 {
			showCustomInformation("Нет плат", "Подключенные Arduino платы не найдены.", w)
			return
		}
		showPortChooser(boards)
	})

	// btnRefreshPorts — перепроверяет список портов.
	// Если текущий порт всё ещё подключён — оставляет его.
	// Если исчез — сбрасывает выбор или предлагает выбрать заново.
	btnRefreshPorts := newButtonWithBorder("🔄 Обновить порты", func() {
		boards, err := getAvailableBoards(appCtx)
		if err != nil {
			dialog.ShowError(fmt.Errorf("не удалось получить список плат: %v", err), w)
			return
		}

		if len(boards) == 0 {
			portPath = ""
			fqbn = ""
			portLabel.SetText("Порт не выбран")
			statusLabel.SetText("Платы не найдены")
			return
		}

		// Проверяем, есть ли текущий выбранный порт в списке.
		currentPortPresent := false
		if portPath != "" {
			for _, b := range boards {
				if b.Port == portPath {
					currentPortPresent = true
					break
				}
			}
		}

		if currentPortPresent {
			statusLabel.SetText("Порты обновлены. Порт " + portPath + " на месте.")
			return
		}

		// Текущий порт исчез или ещё не выбран — предлагаем выбрать.
		if len(boards) == 1 {
			selectBoard(boards[0], &portPath, &fqbn, portLabel, statusLabel, w)
			statusLabel.SetText("Порты обновлены. Выбран " + boards[0].Port)
			return
		}
		showPortChooser(boards)
	})

	btnUpload := newButtonWithBorder("⬆️ Загрузить", func() {
		uploadMu.Lock()
		if isUploading {
			uploadMu.Unlock()
			showCustomInformation("Подождите", "Загрузка уже выполняется.", w)
			return
		}
		uploadMu.Unlock()

		if hexPath == "" {
			showCustomInformation("Ошибка", "Сначала выберите HEX файл.", w)
			return
		}
		if portPath == "" {
			showCustomInformation("Ошибка", "Сначала выберите порт.", w)
			return
		}

		// Проверяем, что HEX-файл всё ещё существует (мог быть удалён после выбора).
		if _, err := os.Stat(hexPath); err != nil {
			showCustomInformation("Ошибка",
				"HEX-файл не найден по пути:\n"+hexPath+"\n\nВозможно, он был перемещён или удалён. Выберите файл заново.", w)
			return
		}

		if fqbn == "" {
			showFQBNInputDialog(w, &fqbn, statusLabel)
			if fqbn == "" {
				return
			}
		}

		// Валидация FQBN: хотя бы два двоеточия (vendor:arch:board).
		if strings.Count(fqbn, ":") < 2 {
			showCustomInformation("Ошибка",
				"Некорректный FQBN: \""+fqbn+"\".\n\nОжидается формат vendor:architecture:board, например arduino:avr:uno.", w)
			return
		}

		// Проверяем, установлено ли ядро.
		coreName := strings.Split(fqbn, ":")[0] + ":" + strings.Split(fqbn, ":")[1]
		coreCtx, cancelCore := context.WithTimeout(appCtx, timeoutQuick)
		defer cancelCore()
		cmd := exec.CommandContext(coreCtx, "arduino-cli", "core", "list")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			if coreCtx.Err() == context.DeadlineExceeded {
				showCustomInformation("Ошибка", "Команда `arduino-cli core list` не ответила за "+timeoutQuick.String()+".", w)
				return
			}
			showCustomInformation("Ошибка", "Не удалось получить список ядер: "+err.Error(), w)
			return
		}
		if !strings.Contains(out.String(), coreName) {
			showCustomInformation("Ошибка",
				"Ядро "+coreName+" не установлено.\n\nУстановите его через `arduino-cli core install "+coreName+"` или скопируйте папку packages.", w)
			return
		}

		uploadMu.Lock()
		isUploading = true
		uploadMu.Unlock()

		go func() {
			defer func() {
				uploadMu.Lock()
				isUploading = false
				uploadMu.Unlock()
			}()

			// Отдельный контекст загрузки, производный от appCtx.
			// Закрытие окна отменит appCtx → отменится и uploadCtx.
			uploadCtx, cancelUpload := context.WithCancel(appCtx)
			defer cancelUpload()

			err := uploadWithProgress(uploadCtx, hexPath, portPath, fqbn, statusLabel)
			if err != nil {
				if err.Error() == "загрузка отменена" {
					return
				}
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("ошибка загрузки: %v", err), w)
				})
			}
		}()
	})

	if err := checkArduinoCLI(); err != nil {
		btnSelectHex.Disable()
		btnSelectPort.Disable()
		btnRefreshPorts.Disable()
		btnUpload.Disable()
		statusLabel.SetText("❌ arduino-cli не найден")
	}

	// Проверка прав на порты — только в Linux. На macOS/Windows групп dialout/uucp нет.
	if runtime.GOOS == "linux" {
		inDialout, _ := checkUserInGroup("dialout")
		inUucp, _ := checkUserInGroup("uucp")
		if !inDialout && !inUucp {
			go func() {
				time.Sleep(500 * time.Millisecond)
				showPermissionWarning(w)
			}()
			statusLabel.SetText("⚠️ Нет прав на доступ к портам (нужна группа dialout)")
		}
	}

	// Закрытие окна: отменяем appCtx (а с ним — текущую загрузку и discovery).
	w.SetCloseIntercept(func() {
		cancelApp()
		w.Close()
	})

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		container.NewHBox(btnSelectHex, hexLabel),
		container.NewHBox(btnSelectPort, btnRefreshPorts, portLabel),
		widget.NewSeparator(),
		btnUpload,
		statusLabel,
	)

	w.SetContent(content)
	w.ShowAndRun()
}

func selectBoard(board BoardInfo, portPath *string, fqbn *string, portLabel *widget.Label, statusLabel *widget.Label, w fyne.Window) {
	*portPath = board.Port
	portLabel.SetText(board.Port)
	if board.FQBN != "" && strings.Count(board.FQBN, ":") >= 2 {
		*fqbn = board.FQBN
		portLabel.SetText(board.Port + " (" + board.FQBN + ")")
		statusLabel.SetText("Порт выбран: " + board.Port + " (" + board.FQBN + ")")
	} else {
		statusLabel.SetText("Тип платы не определён. Выберите FQBN.")
		showFQBNInputDialog(w, fqbn, statusLabel)
		if *fqbn != "" {
			portLabel.SetText(board.Port + " (" + *fqbn + ")")
			statusLabel.SetText("Порт выбран: " + board.Port + " (" + *fqbn + ")")
		} else {
			portLabel.SetText(board.Port + " (тип не выбран)")
			statusLabel.SetText("FQBN не выбран")
		}
	}
}

func showFQBNInputDialog(w fyne.Window, fqbn *string, statusLabel *widget.Label) {
	commonFQBNs := []string{
		"arduino:avr:uno",
		"arduino:avr:nano",
		"arduino:avr:mega",
		"arduino:avr:leonardo",
		"esp32:esp32:esp32",
		"esp8266:esp8266:generic",
	}

	// dialogObj объявлен до entry, чтобы OnSubmitted мог его закрыть.
	var dialogObj *dialog.CustomDialog

	selectWidget := widget.NewSelect(commonFQBNs, func(s string) {
		*fqbn = s
		statusLabel.SetText("FQBN выбран: " + s)
	})

	border := canvas.NewRectangle(color.Transparent)
	border.StrokeWidth = 2
	border.StrokeColor = theme.Color(theme.ColorNameForeground)
	border.CornerRadius = 8

	selectWithBorder := container.NewStack(border, selectWidget)
	selectWrapper := container.NewHBox(selectWithBorder)
	selectWrapper.Resize(fyne.NewSize(600, 80))

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Или введите свой FQBN вручную")
	entry.OnSubmitted = func(s string) {
		if s != "" {
			*fqbn = s
			statusLabel.SetText("FQBN введён: " + s)
			if dialogObj != nil {
				dialogObj.Hide()
			}
		}
	}

	okButton := newButtonWithBorder("✅ OK", func() {
		if *fqbn == "" && entry.Text != "" {
			*fqbn = entry.Text
			statusLabel.SetText("FQBN выбран: " + *fqbn)
		}
		if *fqbn == "" && len(commonFQBNs) > 0 {
			*fqbn = commonFQBNs[0]
			statusLabel.SetText("FQBN выбран: " + *fqbn)
		}
		if dialogObj != nil {
			dialogObj.Hide()
		}
	})

	content := container.NewVBox(
		widget.NewLabel("Выберите или введите FQBN для вашей платы:"),
		selectWrapper,
		widget.NewLabel("или"),
		entry,
		container.NewCenter(okButton),
	)

	// NewCustomWithoutButtons — чтобы Fyne не добавлял невидимую dismiss-кнопку
	// под контентом. Закрытие — только через нашу кнопку «OK» или Enter.
	dialogObj = dialog.NewCustomWithoutButtons("Выбор FQBN", content, w)
	dialogObj.Resize(fyne.NewSize(600, 340))
	dialogObj.Show()
}

// getAvailableBoards возвращает список плат через JSON-парсинг `arduino-cli board list --json`.
// Поддерживает несколько форматов ответа:
//   - {"detected_ports": [...]}  — современный формат (v0.19+)
//   - {"boards": [...]}          — старый формат
//   - [...]                       — ещё более старый формат (массив)
func getAvailableBoards(ctx context.Context) ([]BoardInfo, error) {
	listCtx, cancel := context.WithTimeout(ctx, timeoutQuick)
	defer cancel()

	cmd := exec.CommandContext(listCtx, "arduino-cli",
		"board", "list",
		"--json",
		"--discovery-timeout", "5s",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if listCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("команда `arduino-cli board list` не ответила за %s", timeoutQuick)
		}
		return nil, fmt.Errorf("ошибка запуска arduino-cli: %v: %s", err, stderr.String())
	}

	raw := bytes.TrimSpace(out.Bytes())
	if len(raw) == 0 {
		return nil, nil
	}

	var entries []boardEntry

	switch raw[0] {
	case '[':
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, fmt.Errorf("ошибка парсинга JSON (массив): %v\nСырой вывод: %s", err, raw)
		}
	case '{':
		var wrapper struct {
			DetectedPorts []boardEntry `json:"detected_ports"`
			Boards        []boardEntry `json:"boards"`
		}
		if err := json.Unmarshal(raw, &wrapper); err != nil {
			return nil, fmt.Errorf("ошибка парсинга JSON (объект): %v\nСырой вывод: %s", err, raw)
		}
		if len(wrapper.DetectedPorts) > 0 {
			entries = wrapper.DetectedPorts
		} else {
			entries = wrapper.Boards
		}
	default:
		return nil, fmt.Errorf("неожиданный формат ответа arduino-cli: %s", raw)
	}

	var boards []BoardInfo
	for _, b := range entries {
		if b.Port.Address == "" {
			continue
		}
		var fqbn string
		if len(b.MatchingBoards) > 0 {
			fqbn = b.MatchingBoards[0].FQBN
		}
		boards = append(boards, BoardInfo{
			Port: b.Port.Address,
			FQBN: fqbn,
		})
	}
	return boards, nil
}