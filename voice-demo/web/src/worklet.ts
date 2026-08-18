// AudioWorklet 处理器集合。
// 1) pcm-recorder：麦克风 Float32 -> 24kHz 单声道 PCM16，分块发给主线程。
// 2) pcm-player：常驻输出节点，用环形缓冲做主线程推来的 PCM 无缝播放。
// 注：Worklet 运行在独立线程，不能 import 其它模块。

declare class AudioWorkletProcessor {
  readonly port: MessagePort;
  constructor();
  process(
    inputs: Float32Array[][],
    outputs: Float32Array[][],
    parameters: Record<string, Float32Array>,
  ): boolean;
}

declare function registerProcessor(
  name: string,
  processorCtor: new () => AudioWorkletProcessor,
): void;

class PCMRecorder extends AudioWorkletProcessor {
  private buffer: Float32Array;
  private offset = 0;

  constructor() {
    super();
    this.buffer = new Float32Array(2048);
  }

  process(inputs: Float32Array[][]): boolean {
    const channel = inputs[0] && inputs[0][0];
    if (!channel) {
      return true;
    }

    for (let i = 0; i < channel.length; i += 1) {
      this.buffer[this.offset] = channel[i];
      this.offset += 1;
      if (this.offset === this.buffer.length) {
        this.emit(this.buffer);
        this.offset = 0;
      }
    }
    return true;
  }

  private emit(samples: Float32Array): void {
    const pcm = new Int16Array(samples.length);
    for (let i = 0; i < samples.length; i += 1) {
      const value = Math.max(-1, Math.min(1, samples[i]));
      pcm[i] = value < 0 ? value * 0x8000 : value * 0x7fff;
    }
    // pcm.buffer 每次都是新建的 ArrayBuffer，可直接转移。
    this.port.postMessage(pcm.buffer, [pcm.buffer]);
  }
}

class PCMPlayer extends AudioWorkletProcessor {
  private buffer: Float32Array;
  private size = 0;
  private read = 0;
  private write = 0;

  constructor() {
    super();
    this.buffer = new Float32Array(24000 * 60); // 60 秒环形缓冲，避免长回复灌太快丢尾
    this.port.onmessage = (event: MessageEvent<ArrayBuffer | { type: string }>) => {
      const data = event.data;
      if (typeof data === "object" && data && (data as { type: string }).type === "clear") {
        this.clear();
        return;
      }
      if (data instanceof ArrayBuffer) {
        this.pushPCM16(new Int16Array(data));
      }
    };
  }

  process(_inputs: Float32Array[][], outputs: Float32Array[][]): boolean {
    const channel = outputs[0] && outputs[0][0];
    if (!channel) {
      return true;
    }

    const hadAudio = this.size > 0;
    for (let i = 0; i < channel.length; i += 1) {
      if (this.size > 0) {
        channel[i] = this.buffer[this.read];
        this.read = (this.read + 1) % this.buffer.length;
        this.size -= 1;
      } else {
        channel[i] = 0;
      }
    }

    // 只在"有声音 -> 播完"的瞬间通知主线程，避免刷屏。
    if (hadAudio && this.size === 0) {
      this.port.postMessage({ type: "drained" });
    }
    return true;
  }

  private pushPCM16(pcm: Int16Array): void {
    for (let i = 0; i < pcm.length; i += 1) {
      if (this.size >= this.buffer.length) {
        // 正好满：覆盖最旧样本、保留最新，保证不丢尾。
        this.buffer[this.read] = pcm[i] / 32768;
        this.read = (this.read + 1) % this.buffer.length;
        this.write = this.read;
        continue;
      }
      this.buffer[this.write] = pcm[i] / 32768;
      this.write = (this.write + 1) % this.buffer.length;
      this.size += 1;
    }
  }

  private clear(): void {
    this.buffer.fill(0);
    this.size = 0;
    this.read = 0;
    this.write = 0;
  }
}

registerProcessor("pcm-recorder", PCMRecorder);
registerProcessor("pcm-player", PCMPlayer);
