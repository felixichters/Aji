import {
  type ServerMessage,
  decodeServerMessage,
  encodeJoin,
  encodePlace,
} from "./protocol";

export type MessageHandler = (msg: ServerMessage) => void;

export interface ConnectionOptions {
  url: string;
  onMessage: MessageHandler;
  onOpen?: () => void;
  onClose?: (code: number, reason: string) => void;
  onError?: (err: Event) => void;
}

export class Connection {
  private ws: WebSocket | null = null;
  private opts: ConnectionOptions;

  constructor(opts: ConnectionOptions) {
    this.opts = opts;
  }

  connect(): void {
    this.ws = new WebSocket(this.opts.url);
    this.ws.onopen = () => this.opts.onOpen?.();
    this.ws.onclose = (ev) => this.opts.onClose?.(ev.code, ev.reason);
    this.ws.onerror = (ev) => this.opts.onError?.(ev);
    this.ws.onmessage = (ev) => {
      const msg = decodeServerMessage(ev.data as string);
      this.opts.onMessage(msg);
    };
  }

  join(playerId: string): void {
    this.ws?.send(encodeJoin(playerId));
  }

  place(x: number, y: number): void {
    this.ws?.send(encodePlace(x, y));
  }

  close(): void {
    this.ws?.close();
    this.ws = null;
  }
}
