import { shapes, dia, elementTools } from '@joint/core';
import { ui } from './main';

type BPlusTreeNode = {
    keys: number[];
    value: string[];
    children: BPlusTreeNode[];
};

let newBox = (paper: dia.Paper, graph: dia.Graph, x: number, y: number, node: BPlusTreeNode) => {

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
            text: node.keys.toString() + " " + node.value.toString(),
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

export function createBPlusTreeFromJSON(treeData: BPlusTreeNode, x = 400, y = 50, xSpacing = 200, ySpacing = 150) {
    if (!treeData) return null;

    const node = newBox(ui.paper, ui.graph, x, y, treeData);

    if (treeData.children && treeData.children.length > 0) {
        const numChildren = treeData.children.length;
        const startX = x - ((numChildren - 1) * xSpacing) / 2;

        treeData.children.forEach((child, index) => {
            const childX = startX + index * xSpacing;
            const childY = y + ySpacing;
            const childNode = createBPlusTreeFromJSON(child, childX, childY, xSpacing / 1.5, ySpacing);

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

                ui.graph.addCell(link);
            }
        });
    }
    return node;
}

document.addEventListener('DOMContentLoaded', () => {

    // Enable panning and zooming
    ui.paper.scale(1);
    ui.paper.setInteractivity({
        elementMove: true,
        linkMove: true
    });
    let panning = false;
    let lastPosition = { x: 0, y: 0 };

    ui.paper.on('blank:pointerdown', function (_, x, y) {
        panning = true;
        lastPosition = { x, y };
    });

    ui.paper.on('blank:pointermove', function (_, x, y) {
        if (panning) {
            const dx = (x - lastPosition.x); // Reduce speed
            const dy = (y - lastPosition.y); // Reduce speed
            const currentTranslate = ui.paper.translate();
            ui.paper.translate(currentTranslate.tx + dx, currentTranslate.ty + dy);
            lastPosition = { x, y };
        }
    });

    ui.paper.on('blank:pointerup', function () {
        panning = false;
    });

    document.addEventListener('wheel', (event) => {
        event.preventDefault();
        const scaleFactor = 1.1;
        let scale = ui.paper.scale().sx;
        scale = event.deltaY < 0 ? scale * scaleFactor : scale / scaleFactor;
        ui.paper.scale(Math.max(0.2, Math.min(2, scale)));
    }, { passive: false });
});


export function convertJsonToBPlusTree(json: any): BPlusTreeNode {

    function convertNode(node: any): BPlusTreeNode {
        return {
            keys: node.key.map((k: string) => k.charCodeAt(0)), // Convert string keys to ASCII codes
            value: node.value,
            children: node.children.map(convertNode),
        };
    }

    return convertNode(json.root);
}