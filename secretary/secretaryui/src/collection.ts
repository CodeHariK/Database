import { ui } from "./main";

import { RedrawTree, ZoomToElement } from "./draw";
import { BTree, BTreeNode } from "./tree";
import { deleteBtn, deleteInput, setBtn, nextTreeBtn, prevTreeBtn, resultDiv, runTestBtn, getBtn, getInput, setKey, setValue, treeForm, clearBtn, sortedSetBtn, sortedSetValue, sortedSetSubBtn, sortedSetAddBtn, zoomBtn } from "./dom";

let TestCounter = 0
let Tests = [
    "0000000000000026", "0000000000000020", "0000000000000021",
    "0000000000000018", "0000000000000019", "0000000000000022",
    "0000000000000024", "0000000000000023",
    "0000000000000025", "0000000000000017",
]
async function runTest() {
    if (!ui.getTree()) return

    if (TestCounter < Tests.length) {
        runTestBtn.innerHTML = `Test ${TestCounter + 1}/${Tests.length}`

        let t = Tests[TestCounter]

        await deleteRecord(t)
    }
    TestCounter++
}

export function displayNode(node: BTreeNode) {
    const infoBox = document.getElementById("info-box");
    if (!infoBox) return;

    // Clear previous content
    infoBox.innerHTML = "";

    // Create Node ID section
    const nodeIdDiv = document.createElement("div");
    nodeIdDiv.textContent = `Node ID: ${node.nodeID}`;
    nodeIdDiv.style.fontSize = "15px";
    nodeIdDiv.style.marginBottom = "5px";
    infoBox.appendChild(nodeIdDiv);

    // Create the list container
    const list = document.createElement("ul");
    list.style.listStyle = "none";
    list.style.padding = "0";

    node.keys.forEach((key, index) => {
        const listItem = document.createElement("li");
        listItem.style.marginTop = "5px";
        listItem.style.alignItems = "center";

        // Key text
        const keySpan = document.createElement("span");
        keySpan.textContent = `${key} `;

        // Input for modifying value
        const valueInput = document.createElement("textarea");
        valueInput.value = node.value[index];
        valueInput.oninput = () => {
            node.value[index] = valueInput.value;
            valueInput.style.height = "auto"; // Reset height
            valueInput.style.height = valueInput.scrollHeight + "px"; // Set new height
        };

        const updateButton = document.createElement("button");
        updateButton.textContent = " Update ";
        updateButton.style.backgroundColor = "#c8ffc3";
        updateButton.onclick = async () => {
            await setRecord(key, valueInput.value)
            displayNode(node);
        };

        const deleteButton = document.createElement("button");
        deleteButton.textContent = " Delete ";
        deleteButton.style.backgroundColor = "#ffa1a1";
        deleteButton.onclick = async () => {
            await deleteRecord(key)
            displayNode(node);
        };

        listItem.appendChild(keySpan);
        if (node.value[index]) {
            listItem.appendChild(valueInput);
            listItem.appendChild(updateButton);
            listItem.appendChild(deleteButton);
        }
        list.appendChild(listItem);
    });

    infoBox.appendChild(list);
}

async function getCurrentTree() {

    let res: ResponseObject

    let collectionName = ui.currentTreeDef!.collectionName

    if (ui.WASM) {
        res = toResponseObject("wasm:getTree", ui.func.getTree(collectionName))
    } else {
        res = await makeRequest(
            `${ui.url}/gettree/${collectionName}`,
            undefined,
        )
    }

    appendResult(res);

    if (!res.error) {
        const bTree = new BTree(collectionName, res.data)

        // if (ui.getTree()?.name != bTree.name) {
        ui.TreeSnapshots = []
        ui.currentTreeSnapshotIndex = 0
        ui.NODEMAP = new Map()
        ui.MouseCurrentNode = null
        // }
        ui.TreeSnapshots.push(bTree)
        ui.currentTreeSnapshotIndex = ui.TreeSnapshots.length - 1

        let height = ui.currentTreeDef?.order ?? 4
        ui.BOXHEIGHT = height * 36

        RedrawTree()
    }
}

interface ResponseObject {
    url: RequestInfo | URL;
    data: any;
    logs: any;
    error: string | null;
}

function toResponseObject(url: RequestInfo | URL, response: { data: any, error: string }): ResponseObject {
    let data: any
    let logs: any

    if (!response.error) {
        try {
            let jsondata = JSON.parse(response.data);
            data = jsondata.data
            logs = jsondata.logs
        } catch (error) {
        }
    }

    return {
        url: url,
        data: data,
        logs: logs,
        error: response.error,
    }
}

export async function getAllTree() {
    let res = await makeRequest(
        `${ui.url}/getalltree`,
        undefined,
    )

    if (res.error != "") {
        ui.WASM = true

        res = toResponseObject("wasm:getalltree", ui.func.getAllTree())
    }

    appendResult(res);

    const collectionsDiv = document.getElementById("collections")!;
    collectionsDiv.innerHTML = "";

    (res.data as any[]).forEach((tree, _) => {
        const card = document.createElement("div");
        card.className = "card";
        card.innerHTML = `<pre>${JSON.stringify(tree, null, 2)}</pre>`;
        card.onclick = () => {
            Array.from(document.getElementsByClassName("highlight")).
                forEach((e) => {
                    e.classList.remove("highlight")
                })
            card.classList.add("highlight");

            ui.currentTreeDef = tree
            getCurrentTree()
        };

        collectionsDiv.appendChild(card);
    });
}

async function setRecord(key: string | null, value: string | null) {
    if (!ui.getTree()) {
        return
    }

    key = key ?? setKey.value;
    value = value ?? setValue.value;
    const payload = { key, value };

    let res: ResponseObject

    let collectionName = ui.currentTreeDef!.collectionName

    if (ui.WASM) {
        res = toResponseObject("wasm:setRecord", ui.func.setRecord(collectionName, key, value))
    } else {
        res = await makeRequest(
            `${ui.url}/set/${collectionName}`,
            {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            },
        )
    }

    appendResult(res);

    getCurrentTree()
}

let SortedSetValue = 1
async function sortedSetRecord(add: number | null) {
    if (!ui.getTree()) return

    let value = add != null ? SortedSetValue + add : Number(sortedSetValue.value);
    if (value <= 0) { return }
    SortedSetValue = value
    sortedSetValue.value = String(value)

    let res: ResponseObject
    let collectionName = ui.currentTreeDef!.collectionName

    if (ui.WASM) {
        res = toResponseObject("wasm:sortedSetRecord", ui.func.sortedSetRecord(collectionName, value))
    } else {
        res = await makeRequest(
            `${ui.url}/sortedset/${collectionName}/${value}`,
            {
                method: "POST",
                headers: { "Content-Type": "application/json" },
            },
        )
    }

    appendResult(res);

    getCurrentTree()
}

async function deleteRecord(id: string | null) {
    if (!ui.getTree()) return

    let key = id ?? deleteInput.value

    let res: ResponseObject
    let collectionName = ui.currentTreeDef!.collectionName

    if (ui.WASM) {
        res = toResponseObject("wasm:deleteRecord", ui.func.deleteRecord(collectionName, key))
    } else {
        res = await makeRequest(
            `${ui.url}/delete/${collectionName}/${key}`,
            { method: "DELETE" },
        )
    }

    appendResult(res);

    getCurrentTree()
}

async function getRecord(key: string | null) {
    if (!ui.getTree()) return

    key = key ?? getInput.value

    ui.NODEMAP.forEach((n) => {
        n.selected = false
    })

    if (key) {

        let res: ResponseObject
        let collectionName = ui.currentTreeDef!.collectionName

        if (ui.WASM) {
            res = toResponseObject("wasm:getRecord", ui.func.getRecord(collectionName, key))
        } else {
            res = await makeRequest(
                `${ui.url}/get/${collectionName}/${key}`,
                undefined,
            )
        }

        appendResult(res);

        let result = ui.getTree().searchLeafNode(key)
        result.path.forEach((e) => {
            ui.NODEMAP.get(e.nodeID)!.selected = true
        })

    }
    RedrawTree()
}

async function clearTree() {
    if (!ui.getTree()) return

    let res: ResponseObject
    let collectionName = ui.currentTreeDef!.collectionName

    if (ui.WASM) {
        res = toResponseObject("wasm:clearTree", ui.func.clearTree(collectionName))
    } else {
        res = await makeRequest(
            `${ui.url}/clear/${collectionName}`,
            { method: "DELETE", },
        )
    }

    appendResult(res);

    getCurrentTree()
}

async function newTreeRequest(event: SubmitEvent) {
    event.preventDefault(); // Prevent page reload

    let collectionName = (document.getElementById("collectionName") as HTMLInputElement).value
    let order = Number((document.getElementById("order") as HTMLInputElement).value)
    let numLevel = Number((document.getElementById("NumLevel") as HTMLInputElement).value)
    let baseSize = Number((document.getElementById("BaseSize") as HTMLInputElement).value)
    let increment = Number((document.getElementById("Increment") as HTMLInputElement).value)
    let compactionBatchSize = 0

    const requestData = {
        CollectionName: collectionName,
        Order: order,
        NumLevel: numLevel,
        BaseSize: baseSize,
        Increment: increment,
        compactionBatchSize: compactionBatchSize,
    };

    let res: ResponseObject

    if (ui.WASM) {
        res = toResponseObject("wasm:newTree", ui.func.newTree(collectionName, order, numLevel, baseSize, increment, compactionBatchSize))
    } else {
        res = await makeRequest(
            `${ui.url}/newtree`,
            {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(requestData),
            },
        )
    }

    appendResult(res);

    getAllTree()
}

async function makeRequest(
    url: RequestInfo | URL,
    parameters: RequestInit | undefined,
): Promise<ResponseObject> {
    let response;
    let result;
    let error: string = ""

    try {
        response = await fetch(url, parameters);

        if (!response.ok) {
            const errorText = await response.text(); // Read body only once
            throw new Error(`Error: ${response.status} - ${errorText}`);
        }

        result = await response.json();

    } catch (err) {
        return {
            url,
            data: null,
            logs: null,
            error: err instanceof Error ? err.message : "Unknown error",
        };
    }

    return {
        url: url,
        data: result.data,
        logs: result.logs,
        error: error,
    }
}

declare global {
    interface Window {
        toggleVisibility: (id: string) => void;
    }
}
window.toggleVisibility = function (id: string) {
    const content = document.getElementById(id) as HTMLElement;
    content.style.display = content.style.display === "none" ? "block" : "none";
};

function appendResult(res: ResponseObject) {
    const uniqueId = `json-content-${Date.now()}`;
    const formattedData = JSON.stringify(res.data, null, 2);
    const formattedLogs = res.logs ? `<pre style="color: gray;">${res.logs}</pre>` : "";

    const newResultHTML = `
        <div style="background-color:${res.error ? "#ffb9b9" : "#bdffbd"}; color: black; border:1px solid; padding:5px; margin-bottom: 10px;">
            <div onclick="window.toggleVisibility('${uniqueId}')"
                style="cursor: pointer; font-weight: bold; padding: 4px;">
                ${res.url.toString().includes("wasm") ? res.url : res.url.toString().split(ui.url)[1]}
            </div>
            <div id="${uniqueId}" style="display: none; padding: 4px;">
                <pre>${res.error ? res.error : formattedData}</pre>
                ${formattedLogs}
            </div>
        </div>
    `;

    resultDiv.innerHTML = newResultHTML + resultDiv.innerHTML;
}

function Zoom() {
    let tree = ui.getTree()
    let nodeDef = ui.NODEMAP.get(tree.root.nodeID)
    if (nodeDef?.box) {
        ZoomToElement(nodeDef?.box)
    }
}

export function setupCollectionRequest() {
    document.addEventListener("DOMContentLoaded", () => {

        prevTreeBtn.addEventListener("click", () => {
            if (ui.currentTreeSnapshotIndex != 0) {
                ui.currentTreeSnapshotIndex--;
                RedrawTree()
            }
        });
        nextTreeBtn.addEventListener("click", () => {
            if (ui.currentTreeSnapshotIndex != (ui.TreeSnapshots.length - 1)) {
                ui.currentTreeSnapshotIndex++;
                RedrawTree()
            }
        });

        runTestBtn.addEventListener("click", runTest);
        setBtn.addEventListener("click", () => setRecord(null, null));
        sortedSetBtn.addEventListener("click", () => sortedSetRecord(null));
        sortedSetAddBtn.addEventListener("click", () => sortedSetRecord(1));
        sortedSetSubBtn.addEventListener("click", () => sortedSetRecord(-1));
        deleteBtn.addEventListener("click", () => deleteRecord(null));
        getBtn.addEventListener("click", () => getRecord(null));
        clearBtn.addEventListener("click", () => clearTree());
        zoomBtn.addEventListener("click", () => Zoom());

        treeForm.addEventListener("submit", newTreeRequest)
    });
}
