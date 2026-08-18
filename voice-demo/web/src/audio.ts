// 麦克风采集与 PCM16 播放封装。

// MicrophoneStream 把麦克风音频切成 PCM16 块并 base64，通过回调交给上层发送。
export class MicrophoneStream {
  private context: AudioContext | null = null;
  private stream: MediaStream | null = null;
  private source: MediaStreamAudioSourceNode | null = null;
  private worklet: AudioWorkletNode | null = null;

  constructor(private readonly onChunk: (base64: string) => void) {}

  async start(): Promise<void> {
    await this.stop();

    this.stream = await navigator.mediaDevices.getUserMedia({
      audio: { channelCount: 1, sampleRate: 24000 },
    });
    this.context = new AudioContext({ sampleRate: 24000 });
    await this.context.audioWorklet.addModule("/assets/worklet.js");

    this.source = this.context.createMediaStreamSource(this.stream);
    this.worklet = new AudioWorkletNode(this.context, "pcm-recorder");
    this.worklet.port.onmessage = (event: MessageEvent<ArrayBuffer>) => {
      this.onChunk(bytesToBase64(new Uint8Array(event.data)));
    };

    this.source.connect(this.worklet);
    // 连到输出以驱动 worklet 运行；process 不动 outputs，输出为零，无回声。
    this.worklet.connect(this.context.destination);
  }

  async stop(): Promise<void> {
    this.worklet?.disconnect();
    this.source?.disconnect();
    this.stream?.getTracks().forEach((track) => track.stop());
    if (this.context) {
      await this.context.close().catch(() => undefined);
    }
    this.worklet = null;
    this.source = null;
    this.stream = null;
    this.context = null;
  }
}

// PcmPlayer 把 AI 返回的 PCM16 推给常驻 playback worklet 做无缝播放。
export class PcmPlayer {
  private context: AudioContext | null = null;
  private node: AudioWorkletNode | null = null;

  async start(): Promise<void> {
    await this.stop();
    this.context = new AudioContext({ sampleRate: 24000 });
    await this.context.audioWorklet.addModule("/assets/worklet.js");
    await this.context.resume();
    this.node = new AudioWorkletNode(this.context, "pcm-player");
    this.node.connect(this.context.destination);
  }

  append(deltaBase64: string): void {
    if (!this.node) {
      return;
    }
    const pcm = base64ToInt16(deltaBase64);
    if (pcm.length === 0) {
      return;
    }
    const bytes = pcm.buffer.slice(pcm.byteOffset, pcm.byteOffset + pcm.byteLength);
    this.node.port.postMessage(bytes, [bytes]);
  }

  clear(): void {
    this.node?.port.postMessage({ type: "clear" });
  }

  async stop(): Promise<void> {
    this.node?.disconnect();
    this.node?.port.close();
    this.node = null;
    if (this.context) {
      await this.context.close().catch(() => undefined);
    }
    this.context = null;
  }
}

function bytesToBase64(bytes: Uint8Array): string {
  let binary = "";
  for (let i = 0; i < bytes.length; i += 1) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function base64ToBytes(base64: string): Uint8Array {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

function base64ToInt16(base64: string): Int16Array {
  const bytes = base64ToBytes(base64);
  const count = Math.floor(bytes.length / 2);
  const samples = new Int16Array(count);
  for (let i = 0; i < count; i += 1) {
    samples[i] = bytes[i * 2] | (bytes[i * 2 + 1] << 8);
  }
  return samples;
}
