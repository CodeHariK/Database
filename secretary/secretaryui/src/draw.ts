import { dia, elementTools, shapes } from '@joint/core';
import { ui } from './main';
import { NODECOLOR, randomColor } from './utils';
import { displayNode } from './collection';

export type BPlusTreeNode = {
    nodeID: number;
    keys: string[];
    value: string[];
    children: BPlusTreeNode[];
};

const BOXWIDTH = 200
const BOXHEIGHT = 200

let MouseCurrentNode: BPlusTreeNode | null = null

let newBox = (x: number, y: number, node: BPlusTreeNode) => {

    let nodeColor = NODECOLOR.get(node.nodeID)
    if (!nodeColor) {
        NODECOLOR.set(node.nodeID, randomColor())
    }
    nodeColor = NODECOLOR.get(node.nodeID)

    const element = new shapes.standard.HeaderedRectangle();
    element.resize(BOXWIDTH, BOXHEIGHT);
    element.position(x, y);
    element.attr('root/tabindex', 12);
    element.attr('root/title', 'shapes.standard.HeaderedRectangle');
    element.attr('header/fill', '#000000');
    element.attr('header/fillOpacity', 0.1);
    element.attr('headerText/text', node.nodeID);
    element.attr('body/fill', nodeColor);
    element.attr('body/fillOpacity', 0.5);
    element.attr('bodyText/text', "Keys" + node.keys.map((a) => { return "\n" + a }).toString() + "\n\nValues" + node.value.map((a) => { return "\n" + a }).toString());
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

        MouseCurrentNode = node

        if (ev == elementView) {
            displayNode(MouseCurrentNode)
        }
    });

    ui.paper.on('element:mouseleave', function (ev) {
        ev.hideTools();

        MouseCurrentNode = null
    });

    return element
}

// export function createBPlusTreeFromJSON(treeData: BPlusTreeNode, x = 400, y = 50, xSpacing = BOXWIDTH * 1.4, ySpacing = BOXHEIGHT * 1.5) {
//     if (!treeData) return null;

//     const node = newBox(x, y, treeData);

//     if (treeData.children && treeData.children.length > 0) {
//         const numChildren = treeData.children.length;
//         const startX = x - ((numChildren - 1) * xSpacing) / 2;

//         treeData.children.forEach((child, index) => {
//             const childX = startX + index * xSpacing;
//             const childY = y + ySpacing;
//             const childNode = createBPlusTreeFromJSON(child, childX, childY, xSpacing / 1.5, ySpacing);

//             if (childNode) {
//                 const link = new shapes.standard.Link();
//                 link.source(node);
//                 link.target(childNode);

//                 // link.appendLabel({
//                 //     attrs: {
//                 //         text: {
//                 //             text: 'to the'
//                 //         }
//                 //     }
//                 // });
//                 link.router('orthogonal');
//                 link.connector('straight', { cornerType: 'line' });

//                 ui.graph.addCell(link);
//             }
//         });
//     }
//     return node;
// }


export function createBPlusTreeFromJSON(
    treeData: BPlusTreeNode,
    order: number,
    x = 400,
    y = 50
) {
    if (!treeData) return null;

    const height = getTreeHeight(treeData);
    const maxSpacing = Math.pow(order, height - 1) * BOXWIDTH / 3;

    return createTreeRecursive(treeData, order, height, 1, x, y, maxSpacing);
}

function createTreeRecursive(
    treeData: BPlusTreeNode,
    order: number,
    height: number,
    level: number,
    x: number,
    y: number,
    maxSpacing: number
) {
    if (!treeData) return null;

    const node = newBox(x, y, treeData);

    if (treeData.children && treeData.children.length > 0) {
        const numChildren = treeData.children.length;
        const spacing = maxSpacing / Math.pow(order, level - 1);
        const startX = x - ((numChildren - 1) * spacing) / 2;

        treeData.children.forEach((child, index) => {
            const childX = startX + index * spacing;
            const childY = y + BOXHEIGHT * 1.5;
            const childNode = createTreeRecursive(child, order, height, level + 1, childX, childY, maxSpacing);

            if (childNode) {
                const link = new shapes.standard.Link();
                link.source(node);
                link.target(childNode);
                link.router('orthogonal');
                link.connector('straight', { cornerType: 'line' });

                ui.graph.addCell(link);
            }
        });
    }

    return node;
}

// Helper function to determine tree height
function getTreeHeight(node: BPlusTreeNode): number {
    if (!node.children || node.children.length === 0) return 1;
    return 1 + Math.max(...node.children.map(getTreeHeight));
}


document.addEventListener('DOMContentLoaded', () => {

    ui.paper.scale(1);
    ui.paper.setInteractivity({
        elementMove: true,
        linkMove: true
    });
    let panning = false;
    let lastPosition = { x: 0, y: 0 };

    document.addEventListener('mousedown', (e) => {
        if (MouseCurrentNode == null) {
            panning = true;
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
            nodeID: node.nodeID,
            keys: node.key,
            value: node.value,
            children: node.children.map(convertNode),
        };
    }

    return convertNode(json.root);
}
