import resolve from "@rollup/plugin-node-resolve"
import commonjs from "@rollup/plugin-commonjs"
import json from "@rollup/plugin-json"
import terser from "@rollup/plugin-terser"

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
        resolve({
            browser: true
        }),
        commonjs(),
        json(),
        terser()
    ],
}