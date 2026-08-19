#!/usr/bin/env python3
"""Synthesize speech from text via CosyVoice / Qwen-Audio-TTS models (WebSocket or HTTP NRT API).

CosyVoice and Qwen-Audio-TTS models require the DashScope SDK.
Run with --help for usage.

Dependencies:
    pip install dashscope>=1.25.17

Or with venv:
    python3 -m venv .venv && source .venv/bin/activate && pip install dashscope>=1.25.17
"""
from __future__ import annotations

import sys

if sys.version_info < (3, 9):
    print(f"Error: Python 3.9+ required (found {sys.version}).", file=sys.stderr)
    sys.exit(1)

# Check dashscope dependency before other imports
try:
    import dashscope
    from dashscope.audio.tts_v2 import SpeechSynthesizer
    from dashscope.audio.http_tts.http_speech_synthesizer import HttpSpeechSynthesizer
except ImportError:
    print(
        "Error: dashscope SDK not installed.\n\n"
        "Install with:\n"
        "  pip install dashscope>=1.25.17\n\n"
        "Or use venv:\n"
        "  python3 -m venv .venv\n"
        "  source .venv/bin/activate  # Windows: .venv\\Scripts\\activate\n"
        "  pip install dashscope>=1.25.17",
        file=sys.stderr,
    )
    sys.exit(1)

import argparse
import json
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from qianwen_lib import require_api_key, run_update_signal  # noqa: E402

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

WEBSOCKET_URL = "wss://dashscope.aliyuncs.com/api-ws/v1/inference"
DEFAULT_MODEL = "cosyvoice-v3-flash"
DEFAULT_VOICE = "longanyang"

VOICES = {
    "longanyang": "Sunny young man (male)",
    "longanhuan": "Energetic cheerful female",
    "longhuhu_v3": "Innocent lively girl",
}

# System voices (built-in) — only supported by v3 models
SYSTEM_VOICES = set(VOICES.keys())

# Models that require custom voice IDs (no system voice support)
CUSTOM_VOICE_ONLY_MODELS = {
    "cosyvoice-v3.5-flash",
    "cosyvoice-v3.5-plus",
}

# Models that support the HTTP NRT API (HttpSpeechSynthesizer)
NRT_MODELS = {
    "cosyvoice-v3.5-flash",
    "cosyvoice-v3.5-plus",
    "cosyvoice-v3-flash",
    "cosyvoice-v3-plus",
    "qwen-audio-3.0-tts-plus",
    "qwen-audio-3.0-tts-flash",
}

# Models that support the instruction parameter
INSTRUCTION_MODELS = {
    "cosyvoice-v3.5-flash",
    "cosyvoice-v3.5-plus",
    "cosyvoice-v3-flash",
    "qwen-audio-3.0-tts-plus",
    "qwen-audio-3.0-tts-flash",
}


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    run_update_signal(caller=__file__)

    parser = argparse.ArgumentParser(
        description="CosyVoice / Qwen-Audio-TTS via DashScope SDK",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=f"""\
models:
  cosyvoice-v3-flash     (default) High quality, fast, supports system voices
  cosyvoice-v3-plus      Highest quality, supports system voices
  cosyvoice-v3.5-flash   High-performance, instruction control, 11 langs
                         ⚠️  Custom voices only (no system voice support)
  cosyvoice-v3.5-plus    Ultra-expressive, instruction control, 11 langs
                         ⚠️  Custom voices only (no system voice support)
  qwen-audio-3.0-tts-plus   High-quality professional scenarios
  qwen-audio-3.0-tts-flash  Low-latency real-time interaction

system voices (cosyvoice-v3-flash / v3-plus only):
  longanyang           (default) Sunny young man
  longanhuan           Energetic cheerful female
  longhuhu_v3          Innocent lively girl

voices (qwen-audio-3.0-tts-plus):
  longanlingxin        Warm and caring female
  longanlufeng         Bright and cheerful male

voices (qwen-audio-3.0-tts-flash):
  longanhuan_v3.6      Energetic cheerful female
  longjielidou_v3.6    Innocent boy
  loongeva_v3.6        Intelligent female (English)
  loongjohn            Steady male (English)

note:
  cosyvoice-v3.5 models do NOT support system voices. You must provide a
  custom voice ID created via Voice Cloning or Voice Design on the platform.

examples:
  # Basic synthesis (v3, default)
  python {Path(__file__).name} --text "Hello, world!"

  # Chinese with specific voice (v3)
  python {Path(__file__).name} --text "你好世界" --voice longanhuan

  # v3.5 with custom voice ID
  python {Path(__file__).name} --text "Hello" --model cosyvoice-v3.5-flash --voice <your-custom-voice-id>

  # With instruction control (v3.5 + custom voice)
  python {Path(__file__).name} --text "欢迎光临" --model cosyvoice-v3.5-flash --voice <id> --instruction "用热情洋溢的声音"

  # Qwen-Audio-TTS
  python {Path(__file__).name} --text "你好" --model qwen-audio-3.0-tts-plus --voice longanlingxin

  # Save to specific file
  python {Path(__file__).name} --text "Hello" --output hello.mp3
""",
    )
    parser.add_argument("--text", "-t", required=True, help="Text to synthesize")
    parser.add_argument("--model", "-m", default=None, help=f"Model (default: {DEFAULT_MODEL})")
    parser.add_argument("--voice", "-v", default=DEFAULT_VOICE, help=f"Voice (default: {DEFAULT_VOICE})")
    parser.add_argument("--output", "-o", type=Path, default=Path("output/qianwen-audio-tts/cosyvoice.mp3"), help="Output file (default: output/qianwen-audio-tts/cosyvoice.mp3)")
    parser.add_argument("--format", "-f", default="mp3", choices=["mp3", "wav", "pcm"], help="Audio format (default: mp3)")
    parser.add_argument("--instruction", type=str, default=None, help="Free-style instruction for speech control (v3.5 and Qwen-Audio-TTS models)")
    parser.add_argument("--language-hints", type=str, default=None, help="Target language hint (e.g. zh, en)")
    args = parser.parse_args()

    # Model priority: CLI > default
    model = args.model or DEFAULT_MODEL

    # Validate: v3.5 models do not support system voices
    if model in CUSTOM_VOICE_ONLY_MODELS and args.voice in SYSTEM_VOICES:
        print(
            f'ERROR: Model "{model}" does not support system voices.\n'
            "You must provide a custom voice ID created via Voice Cloning or Voice Design.\n"
            "Steps: Visit https://platform.qianwenai.com/docs/developer-guides/speech/voice-cloning → Upload 10-20s audio → Get voice ID\n"
            f"Then: python3 {Path(__file__).name} -t \"text\" -m {model} -v <your-custom-voice-id>",
            file=sys.stderr,
        )
        sys.exit(1)

    # Setup
    api_key = require_api_key(script_file=__file__, domain="CosyVoice TTS")
    dashscope.api_key = api_key
    dashscope.base_websocket_api_url = WEBSOCKET_URL

    # Validate instruction parameter
    if args.instruction and model not in INSTRUCTION_MODELS:
        print(f"Warning: --instruction is not supported by model '{model}'. "
              f"Supported models: {', '.join(sorted(INSTRUCTION_MODELS))}",
              file=sys.stderr)

    # Choose API path: HTTP NRT for supported models, WebSocket for others
    if model in NRT_MODELS:
        # Use HTTP NRT API (HttpSpeechSynthesizer)
        print(f"Synthesizing (HTTP NRT): model={model}, voice={args.voice}", file=sys.stderr)
        kwargs: dict = {
            "model": model,
            "text": args.text,
            "voice": args.voice,
            "format": args.format,
            "sample_rate": 24000,
            "stream": False,
            "api_key": api_key,
        }
        if args.instruction and model in INSTRUCTION_MODELS:
            kwargs["instruction"] = args.instruction
        if args.language_hints:
            kwargs["language_hints"] = [args.language_hints]

        try:
            result = HttpSpeechSynthesizer.call(**kwargs)
        except Exception as e:
            print(f"Error: {e}", file=sys.stderr)
            sys.exit(1)

        audio_url = getattr(result, "audio_url", None)
        if not audio_url:
            print(f"Error: No audio URL in response: {result}", file=sys.stderr)
            sys.exit(1)

        # Download audio from URL
        import urllib.request
        output = args.output
        if output.suffix.lower() not in {".mp3", ".wav", ".pcm"}:
            output = output.with_suffix(f".{args.format}")
        output.parent.mkdir(parents=True, exist_ok=True)

        try:
            urllib.request.urlretrieve(audio_url, str(output))
        except Exception as e:
            print(f"Warning: Could not download audio: {e}", file=sys.stderr)
            print(f"Audio URL (manual download): {audio_url}", file=sys.stderr)
            print(json.dumps({"audio_url": audio_url}))
            return

        print(f"Audio saved to {output}", file=sys.stderr)
        print(json.dumps({"audio_file": str(output), "audio_url": audio_url,
                          "size_bytes": output.stat().st_size}))
    else:
        # Use WebSocket API (SpeechSynthesizer) for legacy or unsupported models
        print(f"Synthesizing (WebSocket): model={model}, voice={args.voice}", file=sys.stderr)
        try:
            synthesizer = SpeechSynthesizer(model=model, voice=args.voice)
            audio_data = synthesizer.call(args.text)
        except Exception as e:
            print(f"Error: {e}", file=sys.stderr)
            sys.exit(1)

        if not audio_data:
            print("Error: No audio data returned.", file=sys.stderr)
            sys.exit(1)

        # Save
        output = args.output
        if output.suffix.lower() not in {".mp3", ".wav", ".pcm"}:
            output = output.with_suffix(f".{args.format}")
        output.parent.mkdir(parents=True, exist_ok=True)
        output.write_bytes(audio_data)

        print(f"Audio saved to {output}", file=sys.stderr)
        print(json.dumps({"audio_file": str(output), "size_bytes": len(audio_data)}))


if __name__ == "__main__":
    main()
