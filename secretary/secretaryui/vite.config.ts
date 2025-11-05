import { defineConfig } from "vite";
import path from "path";
import fs from "fs";

export default defineConfig({
    base: '/Database',
    server: {
        port: 3000,
        host: '0.0.0.0',
    },
    plugins: [

        {
            name: 'copy-wasm_exec.js',
            writeBundle() {
                fs.copyFileSync('wasm_exec.js', 'dist/wasm_exec.js');
            }
        },
        {
            name: 'copy-secretary.wasm',
            writeBundle() {
                fs.copyFileSync('secretary.wasm', 'dist/secretary.wasm');
            }
        },
        {
            name: 'copy-404',
            writeBundle() {
                fs.copyFileSync('404.html', 'dist/404.html');
            }
        },
        {
            name: 'copy-logo',
            writeBundle() {
                fs.copyFileSync('logo.png', 'dist/logo.png');
            }
        }
    ],
    build: {
        outDir: "./dist",
        emptyOutDir: true,
        rollupOptions: {
            input: path.resolve(__dirname, "index.html"), // Ensure Vite finds the entry file
        },
    },
});
