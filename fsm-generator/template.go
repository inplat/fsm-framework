package fsm_generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
)

//go:embed templates/*
var templatesContent embed.FS

type TemplateModel struct {
	Model *Model
	State *State
}

type TemplateBuilder struct {
	Templates *template.Template
}

func NewTemplateBuilder() (*TemplateBuilder, error) {
	var err error

	tb := &TemplateBuilder{}

	tb.Templates, err = template.New("fsm_templates").Funcs(template.FuncMap{
		"snake":           strcase.ToSnake,
		"screaming_snake": strcase.ToScreamingSnake,
		"camel":           func(name string) string { return strcase.ToCamel(strcase.ToSnake(name)) },
		"lower":           strings.ToLower,
		"upper":           strings.ToUpper,
	}).ParseFS(templatesContent, templatesDir)
	if err != nil {
		return nil, err
	}

	return tb, nil
}

// GenerateModel генерирует файлы модели по указанному пути
func (t *TemplateBuilder) GenerateModel(model *Model) error {
	tm := &TemplateModel{
		Model: model,
		State: nil,
	}

	// filepath путь до папки с fsm моделью
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	fp := filepath.Join(pwd, outputFsmDir, strcase.ToSnake(model.Name))

	err = os.Mkdir(fp, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// модель
	modelFp := fp + "/" + model.Name + "_model.fsm.go"

	err = t.GenerateFromTemplate("model.fsm.go.tpl", modelFp, tm)
	if err != nil {
		return err
	}

	// интерфейс сервиса
	svcFp := fp + "/service.fsm.go"

	err = t.GenerateFromTemplate("service.fsm.go.tpl", svcFp, tm)
	if err != nil {
		return err
	}

	for _, state := range model.States {
		tm.State = state

		// состояние
		stateFp := fp + "/" + strcase.ToSnake(state.Name) + ".fsm.go"

		err = t.GenerateFromTemplate("state.fsm.go.tpl", stateFp, tm)
		if err != nil {
			return err
		}

		// обработчик состояния
		handlerFp := fp + "/" + strcase.ToSnake(state.Name) + ".handler.fsm.go"
		if _, err = os.Stat(handlerFp); os.IsNotExist(err) {
			err = t.GenerateFromTemplate("state.handler.fsm.go.tpl", handlerFp, tm)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *TemplateBuilder) GenerateFromTemplate(template string, path string, tm *TemplateModel) error {
	// open file
	f, err := os.OpenFile(filepath.Clean(path), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755) //nolint: gosec
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	// exec template
	err = t.Templates.ExecuteTemplate(buf, template, tm)
	if err != nil {
		return err
	}

	// apply go format
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	// write in file
	_, err = f.Write(formatted)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}

func (m *TemplateModel) MinRetryDelayFormatted() string {
	return m.formatDur(m.State.MinRetryDelay)
}

func (m *TemplateModel) CancellationTTLFormatted() string {
	return m.formatDur(m.State.CancellationTTL)
}

func (m *TemplateModel) formatDur(d time.Duration) string {
	return fmt.Sprintf("%d * time.Second // %s", int(d.Seconds()), d.String())
}
