#!/bin/bash

# OmniTranscripts - Whisper Model Download Script
# Downloads pre-trained whisper models for offline transcription

set -e

# Colors for output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
MODELS_DIR="models"
BASE_URL="https://huggingface.co/ggerganov/whisper.cpp/resolve/main"

# Model configurations (model_name:file_size_mb:description)
declare -A MODELS=(
    ["ggml-tiny.en.bin"]="37:English-only, fastest (32x realtime)"
    ["ggml-base.en.bin"]="148:English-only, balanced speed/accuracy (16x realtime)"
    ["ggml-small.en.bin"]="483:English-only, better accuracy (6x realtime)"
    ["ggml-medium.en.bin"]="1533:English-only, high accuracy (2x realtime)"
    ["ggml-large-v3.bin"]="3094:Multilingual, best accuracy (1x realtime)"
)

# Default model if no argument provided
DEFAULT_MODEL="ggml-base.en.bin"

print_usage() {
    echo -e "${BLUE}Usage: $0 [model_name|all]${NC}"
    echo ""
    echo "Available models:"
    for model in "${!MODELS[@]}"; do
        IFS=':' read -r size desc <<< "${MODELS[$model]}"
        echo "  ${model} (${size}MB) - ${desc}"
    done
    echo ""
    echo "Examples:"
    echo "  $0                    # Download default model (${DEFAULT_MODEL})"
    echo "  $0 ggml-tiny.en.bin   # Download specific model"
    echo "  $0 all                # Download all models"
}

download_model() {
    local model="$1"
    local url="${BASE_URL}/${model}"
    local output_path="${MODELS_DIR}/${model}"

    if [[ ! "${MODELS[$model]}" ]]; then
        echo -e "${RED}Error: Unknown model '${model}'${NC}"
        return 1
    fi

    IFS=':' read -r size desc <<< "${MODELS[$model]}"

    if [[ -f "$output_path" ]]; then
        echo -e "${YELLOW}Model already exists: ${model}${NC}"
        return 0
    fi

    echo -e "${BLUE}Downloading ${model} (${size}MB)...${NC}"
    echo "Description: ${desc}"

    # Create models directory if it doesn't exist
    mkdir -p "$MODELS_DIR"

    # Download with progress bar
    if command -v curl >/dev/null 2>&1; then
        curl -L --progress-bar "$url" -o "$output_path"
    elif command -v wget >/dev/null 2>&1; then
        wget --progress=bar:force:noscroll "$url" -O "$output_path"
    else
        echo -e "${RED}Error: Neither curl nor wget found. Please install one of them.${NC}"
        return 1
    fi

    # Verify download
    if [[ -f "$output_path" && -s "$output_path" ]]; then
        local file_size=$(du -h "$output_path" | cut -f1)
        echo -e "${GREEN}✓ Downloaded: ${model} (${file_size})${NC}"
    else
        echo -e "${RED}✗ Download failed: ${model}${NC}"
        rm -f "$output_path"
        return 1
    fi
}

create_readme() {
    local readme_path="${MODELS_DIR}/README.md"

    cat > "$readme_path" << 'EOF'
# Whisper Models

This directory contains pre-trained Whisper models for offline speech transcription.

## Available Models

| Model | Size | Language | Speed | Accuracy | Use Case |
|-------|------|----------|-------|----------|----------|
| ggml-tiny.en.bin | ~37MB | English | 32x realtime | Basic | Quick transcription, real-time |
| ggml-base.en.bin | ~148MB | English | 16x realtime | Good | **Recommended for most use cases** |
| ggml-small.en.bin | ~483MB | English | 6x realtime | Better | High-quality English transcription |
| ggml-medium.en.bin | ~1.5GB | English | 2x realtime | High | Professional English transcription |
| ggml-large-v3.bin | ~3GB | Multilingual | 1x realtime | Best | Best quality, supports 99+ languages |

## Model Selection Guide

- **For most users**: Use `ggml-base.en.bin` (default)
- **For real-time applications**: Use `ggml-tiny.en.bin`
- **For highest quality**: Use `ggml-large-v3.bin`
- **For non-English content**: Use `ggml-large-v3.bin`

## Configuration

Set the model path in your `.env` file:

```bash
WHISPER_MODEL_PATH=models/ggml-base.en.bin
```

## Usage

Models are automatically used by the transcription system when available.
The application will fall back to cloud services if models are not present.

## Download

Use the provided script to download models:

```bash
# Download default model
./scripts/download-models.sh

# Download specific model
./scripts/download-models.sh ggml-tiny.en.bin

# Download all models
./scripts/download-models.sh all
```

## License

Models are distributed under the MIT license by OpenAI.
See: https://github.com/openai/whisper
EOF

    echo -e "${GREEN}✓ Created: ${readme_path}${NC}"
}

main() {
    echo -e "${BLUE}OmniTranscripts - Whisper Model Downloader${NC}"
    echo "=============================================="

    local model="${1:-$DEFAULT_MODEL}"

    if [[ "$1" == "--help" || "$1" == "-h" ]]; then
        print_usage
        exit 0
    fi

    # Create models directory
    mkdir -p "$MODELS_DIR"

    if [[ "$model" == "all" ]]; then
        echo -e "${BLUE}Downloading all models...${NC}"
        local success_count=0
        local total_count=${#MODELS[@]}

        for model_name in "${!MODELS[@]}"; do
            if download_model "$model_name"; then
                ((success_count++))
            fi
            echo ""
        done

        echo -e "${BLUE}Download Summary:${NC}"
        echo "Successfully downloaded: ${success_count}/${total_count} models"

    else
        download_model "$model"
    fi

    # Create README
    create_readme

    echo ""
    echo -e "${GREEN}Model download complete!${NC}"
    echo -e "${YELLOW}Configure WHISPER_MODEL_PATH in your .env file to use offline transcription${NC}"

    # Show status
    echo ""
    echo -e "${BLUE}Downloaded models:${NC}"
    for file in "${MODELS_DIR}"/*.bin; do
        if [[ -f "$file" ]]; then
            local size=$(du -h "$file" | cut -f1)
            local name=$(basename "$file")
            echo "  ✓ ${name} (${size})"
        fi
    done
}

# Run main function
main "$@"