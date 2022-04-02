package main

import (
	"log"

	fsm_generator "github.com/inplat/fsm-framework.git/fsm-generator"
)

// Генератор кода fsm моделей
func main() {
	// todo: парсинг параметров, help, пути до моделей и генерированных файлов
	err := fsm_generator.Generate()
	if err != nil {
		log.Fatal(err)
	}
}
