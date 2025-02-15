import { fetchAllBTree, setupCollectionRequest } from './collection';
import { shapes, dia } from '@joint/core';
import { BTree, BTreeNode, NodeDef, TreeDef } from './tree';
import { canvasSection } from './dom';
import { setupDraw } from './draw';

class Ui {
  graph: dia.Graph
  paper: dia.Paper

  NODEMAP = new Map<number, NodeDef>()

  BOXWIDTH = 280
  BOXHEIGHT = 200

  router: "normal" | "orthogonal" = "orthogonal"
  connector: "straight" | "curve" = "straight"

  MouseCurrentNode: BTreeNode | null = null

  url = "http://localhost:8080"

  currentTreeDef: TreeDef | null = null

  currentTreeSnapshotIndex = 0
  TreeSnapshots: Array<BTree> = []

  constructor() {
    this.graph = new dia.Graph({}, { cellNamespace: shapes });
    this.paper = new dia.Paper({
      el: document.getElementById('paper'),
      model: this.graph,
      background: { color: '#F5F5F5' },
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

export const ui = new Ui()

fetchAllBTree()
setupCollectionRequest()
setupDraw()
