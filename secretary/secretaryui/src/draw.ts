import { dia, elementTools, shapes } from '@joint/core';
import { ui } from './main';
import { randomColor } from './utils';
import { displayNode } from './collection';
import { BTreeNode, NodeDef } from './tree';
import { canvasSection, getInput } from './dom';

let MapToNodeDef = (x: number, y: number, node: BTreeNode) => {

    let nodeDef = ui.NODEMAP.get(node.nodeID);
    if (!nodeDef) {
        ui.NODEMAP.set(node.nodeID, {
            color: randomColor(ui.DARK),
            selected: false,
            box: null,
            node: node,
        })
    }
    nodeDef = ui.NODEMAP.get(node.nodeID)!;

    const element = new shapes.standard.HeaderedRectangle();
    element.resize(ui.BOXWIDTH, ui.BOXHEIGHT);
    element.position(x, y);
    element.attr({
        root: {
            tabindex: 12,
            title: 'shapes.standard.HeaderedRectangle',
        },
        bodyText: {
            fill: ui.DARK ? "#fff" : "#000",
            fontSize: 16,
            text: "\n" + node.prevID + "<Prev " + "Parent^" + node.parentID + " Next >" + node.nextID + "\n" +
                node.keys.map((a, i) => {
                    return "\n" + (a == getInput.value ? " > " : "") + a + "  " + (node.value[i] ? node.value[i].substring(0, 12) : "*")
                }).join("") +
                "\n\n" + (node.error ? (node.error + "\n") : "")
        },
        body: {
            fill: node.error ? "#ff6565" : (!getInput.value || nodeDef.selected ? nodeDef.color : (ui.DARK ? "#333" : "#fff")),
            fillOpacity: 1,
            // rx: 20,
            // ry: 20,
            strokeWidth: 1,
            strokeDasharray: '4,2',
        },
        headerText: {
            fill: ui.DARK ? "#fff" : "#000",
            fontSize: 20,
            fontVariant: 'small-caps',
            text: "" + node.nodeID,
        },
        header: {
            fill: node.error ? "#ff6565" : (!getInput.value || nodeDef.selected ? nodeDef.color : (ui.DARK ? "#333" : "#fff")),
            fillOpacity: 1,
            strokeWidth: 1
        },
    });


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

    return nodeDef
}

function drawLink(nodeDef: NodeDef, childDef: NodeDef) {

    const link = new shapes.standard.Link();

    if (!nodeDef?.box || (nodeDef?.box && !ui.graph.getCell(nodeDef.box.id))) {
        // console.log("nodeDef.box is not in the graph!");
        return
    }
    if (!childDef?.box || (childDef?.box && !ui.graph.getCell(childDef.box.id))) {
        // console.log("childDef.box is not in the graph!");
        return
    }

    link.source(nodeDef.box!);
    link.target(childDef.box!);

    link.appendLabel({
        attrs: {
            text: {
                text: nodeDef.node?.nodeID + " " + childDef.node?.nodeID,
                refY: nodeDef.node!.nodeID > childDef.node!.nodeID ? -10 : 10,
            }
        }
    });

    link.router(ui.router);
    link.connector(ui.connector, { cornerType: 'line' });

    var stroke = (nodeDef.selected && childDef.selected) ? (ui.DARK ? "#fff" : "#000") : (ui.DARK ? "#555" : "#bbb");

    link.attr(
        {
            line: {
                // connection: true,
                stroke: stroke,
                strokeWidth: 2,
                strokeDasharray: (nodeDef.selected && childDef.selected) ? 0 : 4,
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
): NodeDef | null {
    if (!node) return null;

    const nodeDef = MapToNodeDef(x, y, node);

    if (node.children && node.children.length > 0) {
        const numChildren = node.children.length;
        const spacing = maxSpacing / Math.pow(order, level - 1);
        const startX = x - ((numChildren - 1) * spacing) / 2;

        node.children.forEach((child, index) => {
            const childX = startX + index * spacing;
            const childY = y + ui.BOXHEIGHT * 1.5;

            const childNodeDef = createTreeRecursive(child, order, height, level + 1, childX, childY, maxSpacing);

            if (childNodeDef) {
                let childDef = ui.NODEMAP.get(child.nodeID)!;

                drawLink(nodeDef, childDef);
            }
        });
    }

    return nodeDef;
}

export function RedrawTree() {
    ui.graph.clear()

    let x = 400, y = 50

    let tree = ui.getTree()
    if (!tree) return null;
    let order = ui.currentTreeDef!.order

    const height = tree.height();
    const maxSpacing = Math.pow(order, height - 1) * ui.BOXWIDTH / 3;

    createTreeRecursive(tree.root, order, height, 1, x, y, maxSpacing);

    ui.NODEMAP.forEach((nodeDef) => {
        if ((nodeDef.node?.nextID ?? 0) > 0) {
            let nextDef = ui.NODEMAP.get(nodeDef.node!.nextID)!
            drawLink(nodeDef, nextDef)
        }
        if ((nodeDef.node?.prevID ?? 0) > 0) {
            let prevDef = ui.NODEMAP.get(nodeDef.node!.prevID)!
            drawLink(nodeDef, prevDef)
        }
    })
}

export function setupDraw() {
    document.addEventListener('DOMContentLoaded', () => {

        ui.paper.scale(1);
        ui.paper.setInteractivity({
            elementMove: true,
            linkMove: true
        });
        let panning = false;
        let lastPosition = { x: 0, y: 0 };

        document.addEventListener('mousedown', (e) => {
            if (ui.MouseCurrentNode == null) {
                panning = true;
            } else {
                displayNode(ui.MouseCurrentNode)
            }
            lastPosition = { x: e.x, y: e.y };
        })
        document.addEventListener('mouseup', () => {
            panning = false;
        })
        document.addEventListener('mousemove', (e) => {
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
}
