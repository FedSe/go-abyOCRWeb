package main

import (
	"bytes"
	"fmt"
	abbyyocr "go-abbyyocr/cmd"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (app *application) homePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, "Ошибка парсинга формы")
		return
	}

	uploadedfile, fhandler, err := r.FormFile("file")
	if err != nil {
		app.clientError(w, http.StatusBadRequest, "Ошибка получения файла")
		return
	}
	defer uploadedfile.Close()

	soup := fmt.Sprintf("%d", time.Now().Unix())
	sourceFile, err := downloadPDFFile(soup, uploadedfile, fhandler)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, "Ошибка сохранения файла: "+err.Error())
		return
	}

	ocr := &abbyyocr.OCRService{
		Uploads:   Uploads,
		Results:   Results,
		ABBYYPath: ABBYYPath,
		Source:    sourceFile,
		Soup:      soup,
	}

	result, err := ocr.ProcessFile(r)
	if err != nil {
		app.serverError(w, fmt.Errorf("ошибка обработки файла: %w", err))
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(result.FilePath)))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", result.Size))

	http.ServeFile(w, r, result.FilePath)
}

func (app *application) homeGet(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/pdf.page.tmpl",
		"./ui/html/base.layout.tmpl",
		"./ui/html/footer.partial.tmpl",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = ts.Execute(w, nil)
	if err != nil {
		app.serverError(w, err)
	}
}

func downloadPDFFile(soup string, uploadedfile multipart.File, fhandler *multipart.FileHeader) (string, error) {
	ext := filepath.Ext(fhandler.Filename)
	if ext != ".pdf" {
		return "", fmt.Errorf("файл не является PDF-документом")
	}

	// Проверка первых 4 байт на сигнатуру PDF (%PDF)
	buf := make([]byte, 4)
	_, err := uploadedfile.Read(buf)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения файла: %v", err)
	}
	uploadedfile.Seek(0, 0)
	if !bytes.HasPrefix(buf, []byte("%PDF")) {
		return "", fmt.Errorf("файл не является PDF-документом")
	}

	_fname := fmt.Sprintf("%s-%s", soup, fhandler.Filename)
	outPath := filepath.Join(Uploads, _fname)
	os.MkdirAll(Uploads, os.ModePerm)

	outFile, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании файла")
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, uploadedfile)
	if err != nil {
		return "", fmt.Errorf("ошибка при записи файла")
	}

	return _fname, nil
}
