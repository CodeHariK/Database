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

async function fetchCurrentTree() {
    await makeRequest(
        `${ui.url}/gettree/${ui.currentTreeDef!.collectionName}`,
        undefined,
        (data) => {
            const bTree = new BTree(ui.currentTreeDef!.collectionName, data)

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
    )
}

export async function fetchAllBTree() {
    await makeRequest(
        `${ui.url}/getalltree`,
        undefined,
        (data: any[], error: boolean) => {

            console.log(data)

            if (error) {
                ui.WASM = true

                // ui.func.allTree()

                console.log(ui.func)

                console.log("---> ", ui.func.allTree())

                return
            }

            const collectionsDiv = document.getElementById("collections")!;
            collectionsDiv.innerHTML = "";

            data.forEach((tree, _) => {
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
                    fetchCurrentTree()
                };

                collectionsDiv.appendChild(card);
            });
        }
    )
}

async function setRecord(key: string | null, value: string | null) {
    if (!ui.getTree()) return

    key = key ?? setKey.value;
    value = value ?? setValue.value;
    const payload = { key, value };
    await makeRequest(
        `${ui.url}/set/${ui.currentTreeDef!.collectionName}`,
        {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
        },
        () => {
            fetchCurrentTree()
        }
    )
}

let SortedSetValue = 1
async function sortedSetRecord(add: number | null) {
    if (!ui.getTree()) return

    let value = add != null ? SortedSetValue + add : Number(sortedSetValue.value);
    if (value <= 0) { return }
    SortedSetValue = value
    const payload = { value };
    sortedSetValue.value = String(value)
    await makeRequest(
        `${ui.url}/sortedset/${ui.currentTreeDef!.collectionName}`,
        {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
        },
        () => {
            fetchCurrentTree()
        }
    )
}

async function deleteRecord(id: string | null) {
    if (!ui.getTree()) return

    await makeRequest(
        `${ui.url}/delete/${ui.currentTreeDef!.collectionName}/${id ?? deleteInput.value}`,
        { method: "DELETE" },
        () => {
            fetchCurrentTree()
        }
    )
}

async function getRecord(getId: string | null) {
    if (!ui.getTree()) return

    getId = getId ?? getInput.value

    ui.NODEMAP.forEach((n) => {
        n.selected = false
    })
    if (getId) {
        await makeRequest(
            `${ui.url}/get/${ui.currentTreeDef!.collectionName}/${getId}`,
            undefined, () => {
                let result = ui.getTree().searchLeafNode(getId)
                result.path.forEach((e) => {
                    ui.NODEMAP.get(e.nodeID)!.selected = true
                })

            }
        )
    }
    RedrawTree()
}

async function clearTree() {
    if (!ui.getTree()) return

    await makeRequest(
        `${ui.url}/clear/${ui.currentTreeDef!.collectionName}`,
        { method: "DELETE", },
        () => {
            fetchCurrentTree()
        }
    )
}

async function newTreeRequest(event: SubmitEvent) {
    event.preventDefault(); // Prevent page reload

    const requestData = {
        CollectionName: (document.getElementById("collectionName") as HTMLInputElement).value,
        Order: Number((document.getElementById("order") as HTMLInputElement).value),
        NumLevel: Number((document.getElementById("NumLevel") as HTMLInputElement).value),
        BaseSize: Number((document.getElementById("BaseSize") as HTMLInputElement).value),
        Increment: Number((document.getElementById("Increment") as HTMLInputElement).value),
    };

    await makeRequest(
        `${ui.url}/newtree`,
        {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(requestData),
        },
        () => {
            fetchAllBTree()
        }
    )
}

async function makeRequest(
    url: RequestInfo | URL,
    parameters: RequestInit | undefined,
    after: (result: any, error: boolean) => void
) {
    let response;
    let result;

    try {
        response = await fetch(url, parameters);
        result = await response.json();
    } catch {
        result = { data: null, logs: "Invalid JSON response" };
    }

    const hasError = !response?.ok || response.status !== 200;
    appendResult(url, result.data, result.logs, hasError);
    after(result.data, hasError);
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

function appendResult(url: RequestInfo | URL, data: any, logs: string, error: boolean) {
    const uniqueId = `json-content-${Date.now()}`;
    const formattedData = JSON.stringify(data, null, 2);
    const formattedLogs = logs ? `<pre style="color: gray;">${logs}</pre>` : "";

    const newResultHTML = `
        <div style="background-color:${error ? "#ffb9b9" : "#bdffbd"}; 
                    color: black; border:1px solid; padding:5px; margin-bottom: 10px;">
            <div onclick="window.toggleVisibility('${uniqueId}')"
                style="cursor: pointer; font-weight: bold; padding: 4px;">
                ${url.toString().split(ui.url)[1]}
            </div>
            <div id="${uniqueId}" style="display: none; padding: 4px;">
                <pre>${formattedData}</pre>
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
