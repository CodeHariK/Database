import { fetchAllBTree } from './collection';
import { shapes, dia } from '@joint/core';
import { BTree, BTreeNode, TreeDef } from './tree';

class Ui {
  graph: dia.Graph
  paper: dia.Paper

  NODECOLOR = new Map<number, string>()
  SELECTEDNODE = new Map<BTreeNode, boolean>()

  BOXWIDTH = 200
  BOXHEIGHT = 200

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
      gridSize: 10,
      drawGrid: true,
      // width: 600,
      // height: 100,
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
