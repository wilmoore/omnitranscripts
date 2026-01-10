//go:build cgo

package lib

/*
#cgo CFLAGS: -I../third_party/whisper.cpp/include -I../third_party/whisper.cpp/ggml/include
#cgo LDFLAGS: -L../third_party/whisper.cpp/build/src -L../third_party/whisper.cpp/build/ggml/src -L../third_party/whisper.cpp/build/ggml/src/ggml-blas -L../third_party/whisper.cpp/build/ggml/src/ggml-metal
#cgo LDFLAGS: -Wl,-rpath,../third_party/whisper.cpp/build/src -Wl,-rpath,../third_party/whisper.cpp/build/ggml/src -Wl,-rpath,../third_party/whisper.cpp/build/ggml/src/ggml-blas -Wl,-rpath,../third_party/whisper.cpp/build/ggml/src/ggml-metal
#cgo LDFLAGS: -lwhisper -lggml -lggml-base -lggml-cpu -lm -lstdc++
#cgo darwin LDFLAGS: -lggml-metal -lggml-blas
#cgo darwin LDFLAGS: -framework Accelerate -framework Metal -framework Foundation -framework CoreGraphics
#include <whisper.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// WhisperContext wraps the C whisper context
type WhisperContext struct {
	ctx *C.struct_whisper_context
}

// InitWhisper initializes whisper with a model file
func InitWhisper(modelPath string) (*WhisperContext, error) {
	cPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cPath))

	ctx := C.whisper_init_from_file_with_params(cPath, C.whisper_context_default_params())
	if ctx == nil {
		return nil, fmt.Errorf("failed to initialize whisper with model: %s", modelPath)
	}

	return &WhisperContext{ctx: ctx}, nil
}

// Free releases the whisper context
func (w *WhisperContext) Free() {
	if w.ctx != nil {
		C.whisper_free(w.ctx)
		w.ctx = nil
	}
}

// TranscribeAudio transcribes the given audio samples
func (w *WhisperContext) TranscribeAudio(samples []float32) ([]TranscriptSegment, error) {
	if w.ctx == nil {
		return nil, fmt.Errorf("whisper context is nil")
	}

	// Get default parameters
	params := C.whisper_full_default_params(C.WHISPER_SAMPLING_GREEDY)
	params.print_realtime = C.bool(false)
	params.print_progress = C.bool(false)
	params.print_timestamps = C.bool(false)
	params.print_special = C.bool(false)
	params.translate = C.bool(false)
	params.language = C.CString("en")
	defer C.free(unsafe.Pointer(params.language))

	// Run the full pipeline
	if C.whisper_full(w.ctx, params, (*C.float)(&samples[0]), C.int(len(samples))) != 0 {
		return nil, fmt.Errorf("whisper_full failed")
	}

	// Extract segments
	nSegments := int(C.whisper_full_n_segments(w.ctx))
	segments := make([]TranscriptSegment, nSegments)

	for i := 0; i < nSegments; i++ {
		startTime := int64(C.whisper_full_get_segment_t0(w.ctx, C.int(i))) * 10 // Convert to milliseconds
		endTime := int64(C.whisper_full_get_segment_t1(w.ctx, C.int(i))) * 10   // Convert to milliseconds
		text := C.GoString(C.whisper_full_get_segment_text(w.ctx, C.int(i)))

		segments[i] = TranscriptSegment{
			Text:      text,
			StartTime: startTime,
			EndTime:   endTime,
		}
	}

	return segments, nil
}

// IsWhisperAvailable checks if whisper.cpp is available
func IsWhisperAvailable() bool {
	// This is a simple check - we could make it more sophisticated
	return true // Since we built it, it should be available
}