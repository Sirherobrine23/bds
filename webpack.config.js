import { fileURLToPath } from "node:url";

/** @type {import("webpack").Configuration} */
export default {
  mode: "production",
  entry: {
    index: [
      fileURLToPath(new URL("web_src/js/global.ts", import.meta.url)),
    ],
  },
  output: {
    path: fileURLToPath(new URL("web_src/js", import.meta.url))
  },
  module: {
    rules: [
      {
        test: /\.ts$/i,
        exclude: /node_modules/,
        use: [
          {
            loader: 'esbuild-loader',
            options: {
              loader: 'ts',
              target: 'es2020',
            },
          },
        ],
      },
    ]
  }
}