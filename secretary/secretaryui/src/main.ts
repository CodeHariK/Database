import { shapes, dia, elementTools } from '@joint/core';

type BPlusTreeNode = {
  keys: number[];
  children: BPlusTreeNode[];
};

let newBox = (paper: dia.Paper, graph: dia.Graph, x: number, y: number, keys: number[]) => {

  const bevelSize = 15;

  const element = new shapes.standard.Polygon();
  element.position(x, y);
  element.resize(100, 40);
  element.attr({
    body: {
      fill: 'white',
      stroke: '#C94A46',
      strokeWidth: 2,
      // rx: 10, ry: 10
      refPoints: `0,0 150,0 150,${30 - bevelSize} 135,30 ${bevelSize},30 0,${30 - bevelSize}`
    },
    label: {
      text: keys.toString(),
      fill: 'black',
      fontSize: 14
    }
  });
  element.addTo(graph);

  const boundaryTool = new elementTools.Boundary();
  const removeButton = new elementTools.Remove();

  const toolsView = new dia.ToolsView({
    tools: [boundaryTool, removeButton]
  });

  const elementView = element.findView(paper);
  elementView.addTools(toolsView);
  elementView.hideTools();

  paper.on('element:mouseenter', function (elementView) {
    elementView.showTools();
  });

  paper.on('element:mouseleave', function (elementView) {
    elementView.hideTools();
  });

  return element
}


function createBPlusTreeFromJSON(paper: dia.Paper, graph: dia.Graph, treeData: BPlusTreeNode, x = 400, y = 50, xSpacing = 200, ySpacing = 150) {
  if (!treeData) return null;

  const node = newBox(paper, graph, x, y, treeData.keys);

  if (treeData.children && treeData.children.length > 0) {
    const numChildren = treeData.children.length;
    const startX = x - ((numChildren - 1) * xSpacing) / 2;

    treeData.children.forEach((child, index) => {
      const childX = startX + index * xSpacing;
      const childY = y + ySpacing;
      const childNode = createBPlusTreeFromJSON(paper, graph, child, childX, childY, xSpacing / 1.5, ySpacing);

      if (childNode) {
        const link = new shapes.standard.Link();
        link.source(node);
        link.target(childNode);

        link.appendLabel({
          attrs: {
            text: {
              text: 'to the'
            }
          }
        });
        link.router('orthogonal');
        link.connector('straight', { cornerType: 'line' });

        graph.addCell(link);
      }
    });
  }
  return node;
}


document.addEventListener('DOMContentLoaded', () => {
  const namespace = shapes;

  const graph = new dia.Graph({}, { cellNamespace: namespace });

  const paper = new dia.Paper({
    el: document.getElementById('paper'),
    model: graph,
    width: 900,
    height: 900,
    background: { color: '#F5F5F5' },
    gridSize: 10,
    interactive: true,
    cellViewNamespace: namespace,
    async: true,
    sorting: dia.Paper.sorting.APPROX,
  });

  // Enable panning and zooming
  paper.scale(1);
  paper.setInteractivity({
    elementMove: true,
    linkMove: true
  });
  let panning = false;
  let lastPosition = { x: 0, y: 0 };

  paper.on('blank:pointerdown', function (_, x, y) {
    panning = true;
    lastPosition = { x, y };
  });

  paper.on('blank:pointermove', function (_, x, y) {
    if (panning) {
      const dx = (x - lastPosition.x); // Reduce speed
      const dy = (y - lastPosition.y); // Reduce speed
      const currentTranslate = paper.translate();
      paper.translate(currentTranslate.tx + dx, currentTranslate.ty + dy);
      lastPosition = { x, y };
    }
  });

  paper.on('blank:pointerup', function () {
    panning = false;
  });

  document.addEventListener('wheel', (event) => {
    event.preventDefault();
    const scaleFactor = 1.1;
    let scale = paper.scale().sx;
    scale = event.deltaY < 0 ? scale * scaleFactor : scale / scaleFactor;
    paper.scale(Math.max(0.2, Math.min(2, scale)));
  }, { passive: false });

  const sampleTree: BPlusTreeNode = {
    keys: [10, 20],
    children: [
      { keys: [5, 8], children: [] },
      { keys: [12, 18], children: [] },
      {
        keys: [22, 25], children: [
          { keys: [5, 8], children: [] },
          { keys: [5, 8], children: [] },
        ]
      }
    ]
  };

  createBPlusTreeFromJSON(paper, graph, sampleTree);
});
