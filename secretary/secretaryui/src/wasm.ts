import { resultDiv } from "./dom";
import { ui } from "./main";

let WASM = "secretary.wasm"

declare const Go: any;

export interface Func {
    add: (a: number, b: number) => number
    allTree: () => string
    set: (v: string) => null
}

interface Window {
    lib: Func;
}

declare let window: Window;

const go = new Go();
let mod: WebAssembly.Module, inst: WebAssembly.Instance;
WebAssembly.instantiateStreaming(fetch(WASM), go.importObject).then(async (result) => {
    mod = result.module;
    inst = result.instance;

    go.run(inst); // Start Go runtime

    ui.func = window.lib

}).catch((err) => {
    console.error(err);
});

export function SetupWASM() {
    document.getElementById("add-btn")?.addEventListener("click", async () => {
        add()
    });
}

async function add() {

    if (!inst) {
        console.error("WebAssembly instance not loaded yet!");
        return;
    }

    if (typeof ui.func.add === "function") {
        resultDiv.innerText = "Result: " + ui.func.add(10, 20);
    } else {
        console.error("ui.func.add is undefined!");
    }
}

async function allTree() {
    if (!inst) {
        console.error("WebAssembly instance not loaded yet!");
        return;
    }

    if (typeof ui.func.allTree === "function") {
        resultDiv.innerText = "Result: " + ui.func.allTree();
    } else {
        console.error("ui.func.allTree is undefined!");
    }
}

async function set() {
    if (!inst) {
        console.error("WebAssembly instance not loaded yet!");
        return;
    }

    if (typeof ui.func.set === "function") {
        resultDiv.innerText = "Result: " + ui.func.set("hello");
    } else {
        console.error("ui.func.set is undefined!");
    }
}
