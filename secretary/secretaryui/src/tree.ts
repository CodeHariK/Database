import { shapes } from "@joint/core"

export class TreeDef {
    collectionName: string
    order: number

    constructor(collectionName: string, order: number) {
        this.collectionName = collectionName
        this.order = order
    }
}

export class NodeDef {
    color: string = "#ECF0F1"
    selected: boolean = false
    node: BTreeNode | null = null
    box: shapes.standard.HeaderedRectangle | null = null
}


export type BTreeNode = {
    nodeID: number;
    nextID: number;
    prevID: number;
    parentID: number;
    keys: string[];
    value: string[];
    children: BTreeNode[];
    errors: string[];
};

export class BTree {
    name: string;
    root: BTreeNode;

    constructor(name: string, json: any) {
        this.name = name
        this.root = this.convertNode(json);
    }

    convertNode = (node: any): BTreeNode => {
        return {
            nodeID: node.nodeID,
            nextID: node.nextID,
            prevID: node.prevID,
            parentID: node.parentID,
            keys: node.key,
            value: node.value,
            children: node.children ? node.children.map((child: any) => this.convertNode(child)) : [],
            errors: node.errors
        };
    };

    searchLeafNode(key: string): { node: BTreeNode; index: number; found: boolean, path: BTreeNode[] } {
        let node = this.root;
        let path: BTreeNode[] = [node];

        // Traverse internal nodes
        while (node.children.length > 0) {
            let { index, found } = this.searchKey(node, key);
            node = found ? node.children[index + 1] : node.children[index];
            path.push(node);
        }

        // Search within the leaf node
        let { index, found } = this.searchKey(node, key);
        return { node, index, found, path };
    }

    searchKey(node: BTreeNode, key: string): { index: number; found: boolean } {
        for (let i = 0; i < node.keys.length; i++) {
            if (key === node.keys[i]) return { index: i, found: true };
            if (key < node.keys[i]) return { index: i, found: false };
        }
        return { index: node.keys.length, found: false };
    }

    height(node: BTreeNode = this.root): number {
        return node.children.length === 0 ? 1 : 1 + Math.max(...node.children.map(child => this.height(child)));
    }
}
