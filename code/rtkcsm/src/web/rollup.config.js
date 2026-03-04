import resolve from "@rollup/plugin-node-resolve"
import commonjs from "@rollup/plugin-commonjs"
import json from "@rollup/plugin-json"
import nodePolyfills from "rollup-plugin-polyfill-node"

export default {
    input: "dist/index.js",
    output: {
        format: "umd",
        sourcemap: true,
        file: "../static/bundle.js",
        inlineDynamicImports: true
    },
    context: "window", 
    plugins: [
        nodePolyfills(),
        resolve({
            browser: true
        }),
        commonjs(),
        json(),
    ],
}