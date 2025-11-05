import { getAllTree, setupCollectionRequest } from './collection';
import { shapes, dia } from '@joint/core';
import { BTree, BTreeNode, NodeDef, TreeDef } from './tree';
import { canvasSection, modalOverlay, openModalBtn, themeToggle } from './dom';
import { setupDraw } from './draw';
import Split from 'split.js'

Split(['#json-section', "#canvas-section", '#form-section'],
  {
    sizes: [15, 70, 15],
    // gutterSize: 4,
  }
);
(document.querySelector('.container') as HTMLElement).style.visibility = 'visible';

declare const Go: any;

export interface Func {
  getAllTree: () => { data: any, error: string }
  getTree: (collectionName: string) => { data: any, error: string }
  newTree: (collectionName: string, order: number, numLevel: number, baseSize: number, increment: number, compactionBatchSize: number) => { data: any, error: string }
  clearTree: (collectionName: string) => { data: any, error: string }
  getRecord: (collectionName: string, key: string) => { data: any, error: string }
  setRecord: (collectionName: string, key: string, value: string) => { data: any, error: string }
  sortedSetRecord: (collectionName: string, value: number) => { data: any, error: string }
  deleteRecord: (collectionName: string, key: string) => { data: any, error: string }
}

interface Window {
  lib: Func;
}

declare let window: Window;

const go = new Go();

class Ui {
  graph: dia.Graph
  paper: dia.Paper

  NODEMAP = new Map<number, NodeDef>()

  BOXWIDTH = 280
  BOXHEIGHT = 200

  DARK: boolean

  func!: Func
  WASM: boolean = false

  router: "normal" | "orthogonal" = "orthogonal"
  connector: "straight" | "curve" = "straight"

  MouseCurrentNode: BTreeNode | null = null

  url = "http://localhost:8080"

  currentTreeDef: TreeDef | null = null

  currentTreeSnapshotIndex = 0
  TreeSnapshots: Array<BTree> = []

  constructor(dark: boolean) {

    this.DARK = dark

    this.graph = new dia.Graph({}, { cellNamespace: shapes });
    this.paper = new dia.Paper({
      el: document.getElementById('paper'),
      model: this.graph,
      background: { color: dark ? '#111111' : '#F5F5F5' },
      gridSize: 20,
      drawGrid: true,
      width: canvasSection().clientWidth,
      height: canvasSection().clientHeight,
      interactive: true,
      cellViewNamespace: shapes,
      async: true,
      sorting: dia.Paper.sorting.APPROX,
    });
  }

  getTree() {
    return this.TreeSnapshots[this.currentTreeSnapshotIndex]
  }
}

let dark = localStorage.getItem("theme") === "dark"
export let ui = new Ui(dark)

WebAssembly.instantiateStreaming(fetch("secretary.wasm"), go.importObject).then(async (result) => {
  // let mod = result.module;
  let inst = result.instance;

  go.run(inst); // Start Go runtime

  ui.func = window.lib

  console.log(window.lib)

  getAllTree()
}).catch((err) => {
  console.error(err);
});

setupCollectionRequest()
setupDraw()

openModalBtn.addEventListener('click', () => {
  modalOverlay.classList.add('active');
});

modalOverlay.addEventListener('click', (e) => {
  if (e.target === modalOverlay) {
    modalOverlay.classList.remove('active');
  }
});

const handleToggleClick = () => {
  document.documentElement.classList.toggle("dark");
  const dark = document.documentElement.classList.contains("dark");
  localStorage.setItem("theme", dark ? "dark" : "light");

  ui = new Ui(dark)
};
themeToggle.addEventListener("click", handleToggleClick);
