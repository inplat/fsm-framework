package caller

import (
	"runtime"
	"strings"
)

// Unknown заглушка, когда название вызывающей функции определить не удалось
const Unknown = "unknown"

// ClearFunction сокращает длинную строку пакета
// из github.com/foo/backend/bar.git/internal/app/fsm/state.(*processPipeline).updateProgressTx
// в state.processPipeline.updateProgressTx
func ClearFunction(packageFunc string) string {
	var li int
	if li = strings.LastIndex(packageFunc, "/"); li == -1 {
		return packageFunc
	}

	var bStructName []byte

	for i := li + 1; i < len(packageFunc); i++ {
		b := packageFunc[i]
		if checkRune(b) {
			bStructName = append(bStructName, b)
		}
	}

	return string(bStructName)
}

func checkRune(b byte) bool {
	return b == '.' || b == '-' || b == '_' || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

// CurrentFuncName имя текущей функции/метода в формате bla.com/blah/package.Func
func CurrentFuncName() string {
	return GetFrame(1).Function
}

// CurrentFuncNameClear имя текущей функции/метода в формате package.Func
func CurrentFuncNameClear() string {
	// Skip GetCurrentFunctionName
	return ClearFunction(GetFrame(1).Function)
}

// FuncName имя вызывающей функции/метода в формате bla.com/blah/package.Func
func FuncName() string {
	return GetFrame(2).Function
}

// FuncNameClear имя вызывающей функции/метода в формате package.Func
func FuncNameClear() string {
	return ClearFunction(GetFrame(2).Function)
}

// GetFrame возвращает нужный фрейм из стека вызовов
func GetFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: Unknown}

	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])

		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame

			frameCandidate, more = frames.Next()

			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
