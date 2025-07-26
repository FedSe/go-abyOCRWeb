package abbyyocr

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type OCRService struct {
	Uploads   string
	Results   string
	Source    string
	ABBYYPath string
	Soup      string
}

type ProcessResult struct {
	FilePath string
	Size     int64
}

func (s *OCRService) ProcessFile(r *http.Request) (*ProcessResult, error) {
	language := r.FormValue("lang")
	recognizerLng := "Russian"
	cmdlng := recognizerLng
	if language != "ru" {
		recognizerLng = "Russian,English"
		cmdlng = "Russian English"
	}

	outFormat, optionsFile, err := s.generateOptions(r, recognizerLng)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации опций: %w", err)
	}

	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить рабочую директорию: %w", err)
	}

	upfilePath := filepath.Join(workDir, s.Uploads, s.Source)
	outFilePath := filepath.Join(workDir, s.Results, s.Source+outFormat)
	repFile := filepath.Join(workDir, s.Soup+"rep.xml")

	args := []string{
		upfilePath,
		"/lang", cmdlng,
		"/optionsFile", filepath.Join(workDir, optionsFile),
		"/out", outFilePath,
		"/report", repFile,
	}

	cmd := exec.Command(strings.Trim(s.ABBYYPath, "\""), args...)
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output

	fmt.Print("Распознавание...")
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ошибка ABBYY: %v, вывод: %s", err, output.String())
	}
	fmt.Println(" успешно.")

	// Анализ отчёта
	sizef := xmlMessageConstructor(repFile)
	for _, val := range sizef {
		fmt.Println(val)
	}

	file, err := os.Open(outFilePath)
	if err != nil {
		return nil, fmt.Errorf("файл не создан: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить информацию о файле: %w", err)
	}

	filesToDelete := []string{repFile, optionsFile}
	for _, file := range filesToDelete {
		if err := os.Remove(file); err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("Предупреждение: не удалось удалить %s: %v\n", file, err)
			}
		} else {
			fmt.Printf("Файл %s успешно удален\n", file)
		}
	}

	return &ProcessResult{
		FilePath: outFilePath,
		Size:     fileInfo.Size(),
	}, nil
}

func (s *OCRService) generateOptions(r *http.Request, lang string) (outFormat, optionsFile string, err error) {
	workmode := r.FormValue("mode")
	if workmode == "ocr" {
		outFormat = ".pdf"
		pictureTextMode := "BPEM_TextOnImage"
		if r.FormValue("oc") != "toi" {
			pictureTextMode = "BPEM_ImageOnText"
		}
		optionsFile, err = xmlOptionsConstructor(s.Soup, pictureTextMode, "exportMode=\"", "<pdfOptions", lang)
	} else {
		outFormat = ".docx"
		mapping := map[string]string{
			"a": "RTFEOA_KeepLines",
			"b": "RTFEOA_KeepPages",
			"c": "RTFEOA_HighlightErrorsBackground",
			"d": "RTFEOA_KeepTextAndBackgroundColor",
			"f": "RTFEOA_KeepRunningTitles",
			"g": "RTFEOA_KeepPictures",
			"h": "RTFEOA_RemoveSoftHyphens",
		}
		var xmlOptionsPDF []string
		for _, char := range r.Form["co"] {
			if word, exists := mapping[char]; exists {
				xmlOptionsPDF = append(xmlOptionsPDF, word)
			}
		}
		xmlOptionsPDF = append(xmlOptionsPDF, "RTFEOA_ForceFixedPageSize")
		optionsFile, err = xmlOptionsConstructor(s.Soup, strings.Join(xmlOptionsPDF, ","), "options=\"", "<rtfOptions", "Russian")
	}
	return outFormat, optionsFile, err
}

func xmlOptionsConstructor(soup, pdf_options, distOption, typeExport, language string) (string, error) {
	filePath := "OptionsFileTemplate.xml"
	optionsFile := "OptionsFile" + soup + ".xml"

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения файла")
	}

	text := string(content)

	rtfOptionsIndex := strings.Index(text, typeExport)
	subText := text[rtfOptionsIndex:]
	optionsIndex := strings.Index(subText, distOption)
	absoluteOptionsIndex := rtfOptionsIndex + optionsIndex
	if absoluteOptionsIndex < 2200 || optionsIndex < 0 {
		return "", fmt.Errorf("файл конфигурации не корректен")
	}

	insertPosition := absoluteOptionsIndex + len(distOption)
	newXml := text[:insertPosition] + pdf_options + text[insertPosition:]
	newXml = strings.ReplaceAll(newXml, "ages=\"", "ages=\""+language)

	err = os.WriteFile(optionsFile, []byte(newXml), 0644)
	if err != nil {
		return "", fmt.Errorf("ошибка записи в файл")
	}
	return optionsFile, nil
}

func xmlMessageConstructor(filePath string) []string {
	var messages []string
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	text := string(content)

	rtfRecognizeIndex := strings.Index(text, "class=\"TaskAutomation.RecognizeOperation")
	subText := text[rtfRecognizeIndex:]

	durationIndex_start := strings.Index(subText, "duration=\"")
	subText = subText[durationIndex_start+len("duration=\""):]

	durationIndex_end := strings.Index(subText, "\">")
	messages = append(messages, "Длительность распознавания: "+subText[:durationIndex_end])

	mess_start := strings.Index(subText, "severity=\"")
	if mess_start > 0 {
		for mess_start > 0 {
			newMes := "* Сообщение "
			typeMes := subText[mess_start+len("severity=\"") : mess_start+len("severity=\"")+3]

			switch typeMes {
			case "inf":
				newMes += "(информационное): "
			case "war":
				newMes += "(предупреждение): "
			default:
				newMes += "(важное): "
			}

			subText = subText[strings.Index(subText, "text=\"")+len("text=\""):]
			newMes += subText[:strings.Index(subText, "\"")]
			subText = subText[strings.Index(subText, "page=\"")+len("page=\""):]
			newMes += " Страница " + subText[:strings.Index(subText, "\"")]

			messages = append(messages, newMes)
			mess_start = strings.Index(subText, "severity=\"")
		}
	}

	return messages
}
