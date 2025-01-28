import htmx, { config as htmxConfig } from "htmx.org"

declare let __webpack_public_path__: string;
declare global {
  type Writable<T> = { -readonly [K in keyof T]: T[K] };
  interface Window {
    htmx: Omit<htmx, 'config'> & {
      config?: Writable<typeof htmxConfig>,
    },
    _globalHandlerErrors: Array<ErrorEvent & PromiseRejectionEvent> & {
      _inited: boolean,
      push: (e: ErrorEvent & PromiseRejectionEvent) => void | number,
    },
    __webpack_public_path__: string;
    turnstile: any,
    codeEditors: any[],
    updateCloneStates: () => void,
  }
}
