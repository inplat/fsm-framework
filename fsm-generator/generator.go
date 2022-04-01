package fsm_generator

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	modelsDir    = "internal/fsm/models"
	previewsDir  = "internal/fsm/previews"
	outputFsmDir = "internal/app/fsm"

	templatesDir = "templates/*.fsm.go.tpl"
)

func Generate() error {
	tb, err := NewTemplateBuilder()
	if err != nil {
		return err
	}

	models, err := loadModels(modelsDir)
	if err != nil {
		return err
	}

	for _, model := range models {
		err = model.ValidateModel()
		if err != nil {
			return fmt.Errorf("%s model validation err: %w", model.Name, err)
		}

		model.PrintStates()

		err = model.MakePreview()
		if err != nil {
			return err
		}

		err = tb.GenerateModel(model)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadModels(modelsPath string) ([]*Model, error) {
	dirEntities, err := os.ReadDir(modelsPath)
	if err != nil {
		return nil, err
	}

	models := make([]*Model, 0, len(dirEntities))

	for _, entity := range dirEntities {
		info, err := entity.Info()
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			continue
		}

		if ext := filepath.Ext(info.Name()); ext != ".yaml" {
			continue
		}

		file, err := os.Open(filepath.Clean(filepath.Join(modelsPath, info.Name())))
		if err != nil {
			return nil, err
		}

		model, err := ParseModel(file)
		if err != nil {
			return nil, err
		}

		model.Name = info.Name()[:len(info.Name())-len(".yaml")]

		models = append(models, model)
	}

	return models, nil
}
