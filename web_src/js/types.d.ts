declare module '*.svg' {
  const value: string;
  export default value;
}

declare module '*.css' {
  const value: string;
  export default value;
}

declare let __webpack_public_path__: string;

declare module 'htmx.org/dist/htmx.esm.js' {
  const value = await import('htmx.org');
  export default value;
}

declare module 'uint8-to-base64' {
  export function encode(arrayBuffer: Uint8Array): string;
  export function decode(base64str: string): Uint8Array;
}

declare module 'swagger-ui-dist/swagger-ui-es-bundle.js' {
  const value = await import('swagger-ui-dist');
  export default value.SwaggerUIBundle;
}

type Writable<T> = { -readonly [K in keyof T]: T[K] };

interface Window {
  htmx: Omit<typeof import('htmx.org/dist/htmx.esm.js').default, 'config'> & {
    config?: Writable<typeof import('htmx.org').default.config>,
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
