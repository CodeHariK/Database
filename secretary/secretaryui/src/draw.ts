import { dia, elementTools, shapes } from '@joint/core';
import { ui } from './main';
import { randomColor } from './utils';
import { displayNode } from './collection';
import { BTreeNode, NodeDef } from './tree';
import { canvasSection, searchInput } from './dom';

let newBox = (x: number, y: number, node: BTreeNode) => {

    let nodeDef = ui.NODEMAP.get(node.nodeID);
    ui.NODEMAP.set(node.nodeID, {
        color: nodeDef ? nodeDef.color : randomColor(),
        selected: false,
        box: null,
        node: node,
    })
    nodeDef = ui.NODEMAP.get(node.nodeID)!;

    const element = new shapes.standard.HeaderedRectangle();
    element.resize(ui.BOXWIDTH, ui.BOXHEIGHT);
    element.position(x, y);
    element.attr('headerText/text', node.nodeID);
    element.attr({
        root: {
            tabindex: 12,
            title: 'shapes.standard.HeaderedRectangle',
        },
        body: {
            fill: nodeDef.color,
            fillOpacity: searchInput.value ? (nodeDef.selected ? 0.5 : 0) : 0.3,
            // rx: 20,
            // ry: 20,
            strokeWidth: 1,
            strokeDasharray: '4,2'
        },
        label: {
            text: 'Hello',
            fill: '#ECF0F1',
            fontSize: 11,
            fontVariant: 'small-caps'
        },
        header: {
            fill: "#000000",
            fillOpacity: 0.1,
            strokeWidth: 1
        },
    });

    element.attr('bodyText/text', "Next " + node.nextID + "\nPrev " + node.prevID +
        "\nKeys" + node.keys.map((a) => { return "\n" + (a == searchInput.value ? " > " : "") + a }).toString() +
        "\n\nValues" + node.value.map((a, i) => { return "\n" + (node.keys[i] == searchInput.value ? " > " : "") + a }).toString());

    element.addTo(ui.graph);

    const boundaryTool = new elementTools.Boundary();
    const removeButton = new elementTools.Remove();
    const btnButton = new elementTools.Button();

    const toolsView = new dia.ToolsView({
        tools: [boundaryTool, removeButton, btnButton]
    });

    const elementView = element.findView(ui.paper);
    elementView.addTools(toolsView);
    elementView.hideTools();

    ui.paper.on('element:mouseenter', function (ev) {
        ev.showTools();

        if (ev == elementView) {
            ui.MouseCurrentNode = node
        }
    });

    ui.paper.on('element:mouseleave', function (ev) {
        ev.hideTools();

        ui.MouseCurrentNode = null
    });

    nodeDef.box = element

    return element
}

function drawLink(nodeDef: NodeDef, childDef: NodeDef) {
    const link = new shapes.standard.Link();
    link.source(nodeDef.box!);
    link.target(childDef.box!);

    link.appendLabel({
        attrs: {
            text: {
                text: nodeDef.node?.nodeID + " " + childDef.node?.nodeID,
                refY: nodeDef.node!.nodeID > childDef.node!.nodeID ? 30 : 0,
            }
        }
    });
    link.router('orthogonal');
    link.connector('straight', { cornerType: 'line' });

    var stroke = "#aaa";
    if (nodeDef.selected && childDef.selected) {
        stroke = "#000";
    }

    link.attr(
        {
            line: {
                // connection: true,
                stroke: stroke,
                strokeWidth: 2,
                strokeDasharray: 4,
                // refY: nodeDef.node!.nodeID > childDef.node!.nodeID ? 30 : 0,
            },
        }
    );

    ui.graph.addCell(link);
}

function createTreeRecursive(
    node: BTreeNode,
    order: number,
    height: number,
    level: number,
    x: number,
    y: number,
    maxSpacing: number
): shapes.standard.HeaderedRectangle | null {
    if (!node) return null;

    const box = newBox(x, y, node);

    let nodeDef = ui.NODEMAP.get(node.nodeID)!;

    if (node.children && node.children.length > 0) {
        const numChildren = node.children.length;
        const spacing = maxSpacing / Math.pow(order, level - 1);
        const startX = x - ((numChildren - 1) * spacing) / 2;

        node.children.forEach((child, index) => {
            const childX = startX + index * spacing;
            const childY = y + ui.BOXHEIGHT * 1.5;
            const childNode = createTreeRecursive(child, order, height, level + 1, childX, childY, maxSpacing);

            if (childNode) {

                let childDef = ui.NODEMAP.get(child.nodeID)!;

                drawLink(nodeDef, childDef);
            }
        });
    }

    return box;
}

export function BuildTreeFromJSON() {
    ui.graph.clear()

    let x = 400, y = 50

    let tree = ui.getTree()
    if (!tree) return null;
    let order = ui.currentTreeDef!.order

    const height = tree.height();
    const maxSpacing = Math.pow(order, height - 1) * ui.BOXWIDTH / 3;

    createTreeRecursive(tree.root, order, height, 1, x, y, maxSpacing);

    ui.NODEMAP.forEach((nodeDef) => {
        if (nodeDef.node?.nextID) {
            let nextDef = ui.NODEMAP.get(nodeDef.node!.nextID)!
            drawLink(nodeDef, nextDef)
        }
        if (nodeDef.node?.prevID) {
            let prevDef = ui.NODEMAP.get(nodeDef.node!.prevID)!
            drawLink(nodeDef, prevDef)
        }
    })
}


document.addEventListener('DOMContentLoaded', () => {

    ui.paper.scale(1);
    ui.paper.setInteractivity({
        elementMove: true,
        linkMove: true
    });
    let panning = false;
    let lastPosition = { x: 0, y: 0 };

    canvasSection().addEventListener('mousedown', (e) => {
        if (ui.MouseCurrentNode == null) {
            panning = true;
        } else {
            displayNode(ui.MouseCurrentNode)
        }
        lastPosition = { x: e.x, y: e.y };
    })
    canvasSection().addEventListener('mouseup', () => {
        panning = false;
    })
    canvasSection().addEventListener('mousemove', (e) => {
        if (panning) {
            const dx = (e.x - lastPosition.x); // Reduce speed
            const dy = (e.y - lastPosition.y); // Reduce speed
            const currentTranslate = ui.paper.translate();
            ui.paper.translate(currentTranslate.tx + dx, currentTranslate.ty + dy);
            lastPosition = { x: e.x, y: e.y };
        }
    })

    canvasSection().addEventListener('wheel', (event) => {
        event.preventDefault();
        const scaleFactor = 1.1;
        let scale = ui.paper.scale().sx;
        scale = event.deltaY < 0 ? scale * scaleFactor : scale / scaleFactor;
        ui.paper.scale(Math.max(0.2, Math.min(2, scale)));
    }, { passive: false });
});

window.onresize = () => {
    ui.paper.setDimensions(canvasSection().clientWidth, canvasSection().clientHeight)
}
