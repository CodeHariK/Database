import { fetchAllBTree } from './collection';
import { shapes, dia } from '@joint/core';

class Ui {
  graph: dia.Graph
  paper: dia.Paper

  constructor() {
    this.graph = new dia.Graph({}, { cellNamespace: shapes });
    this.paper = new dia.Paper({
      el: document.getElementById('paper'),
      model: this.graph,
      width: 900,
      height: 900,
      background: { color: '#F5F5F5' },
      gridSize: 10,
      interactive: true,
      cellViewNamespace: shapes,
      async: true,
      sorting: dia.Paper.sorting.APPROX,
    });
  }
}

export const ui = new Ui()

fetchAllBTree()
