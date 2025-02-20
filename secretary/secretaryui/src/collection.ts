import { ui } from "./main";

import { RedrawTree } from "./draw";
import { BTree, BTreeNode } from "./tree";
import { deleteBtn, deleteInput, setBtn, nextTreeBtn, prevTreeBtn, resultDiv, runTestBtn, getBtn, getInput, setInput, treeForm } from "./dom";

let TestCounter = 0
let Tests = [
    "0000000000000026", "0000000000000020", "0000000000000021",
    "0000000000000018", "0000000000000019", "0000000000000022",
    "0000000000000024", "0000000000000023",
    "0000000000000025", //"0000000000000017",
]
function runTest() {
    if (!ui.getTree()) return

    if (TestCounter < Tests.length) {
        runTestBtn.innerHTML = `Test ${TestCounter + 1}/${Tests.length}`

        let t = Tests[TestCounter]

        deleteRecord(t)
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
        updateButton.onclick = () => {
            displayNode(node);
        };

        const deleteButton = document.createElement("button");
        deleteButton.textContent = " Delete ";
        deleteButton.style.backgroundColor = "#ffa1a1";
        deleteButton.onclick = () => {
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

function fetchCurrentTree() {
    makeRequest(
        `${ui.url}/gettree/${ui.currentTreeDef!.collectionName}`,
        undefined,
        (data) => {
            const bTree = new BTree(ui.currentTreeDef!.collectionName, data)

            console.log(data)

            if (ui.getTree()?.name != bTree.name) {
                ui.TreeSnapshots = []
                ui.currentTreeSnapshotIndex = 0
                ui.NODEMAP = new Map()
                ui.MouseCurrentNode = null
            }
            ui.TreeSnapshots.push(bTree)
            ui.currentTreeSnapshotIndex = ui.TreeSnapshots.length - 1

            let height = ui.currentTreeDef?.order ?? 4
            ui.BOXHEIGHT = height * 36

            RedrawTree()
        }
    )
}

export function fetchAllBTree() {
    makeRequest(
        `${ui.url}/getalltree`,
        undefined,
        (data: any[]) => {
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

function setRecord(value: string | null) {
    if (!ui.getTree()) return

    value = value ?? setInput.value;
    const payload = { value };
    makeRequest(
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

function deleteRecord(id: string | null) {
    if (!ui.getTree()) return

    makeRequest(
        `${ui.url}/delete/${ui.currentTreeDef!.collectionName}/${id ?? deleteInput.value}`,
        { method: "DELETE" },
        () => {
            fetchCurrentTree()
        }
    )
}

async function getRecord(getId: string | null) {
    if (!ui.getTree()) return

    ui.NODEMAP.forEach((n) => {
        n.selected = false
    })
    await makeRequest(
        `${ui.url}/get/${ui.currentTreeDef!.collectionName}/${getId ?? getInput.value}`,
        undefined, () => {
            let result = ui.getTree().searchLeafNode(getId ?? getInput.value)
            result.path.forEach((e) => {
                ui.NODEMAP.get(e.nodeID)!.selected = true
            })

        }
    )
    RedrawTree()
}

function newTreeRequest(event: SubmitEvent) {
    event.preventDefault(); // Prevent page reload

    const requestData = {
        CollectionName: (document.getElementById("collectionName") as HTMLInputElement).value,
        Order: Number((document.getElementById("order") as HTMLInputElement).value),
        BatchNumLevel: Number((document.getElementById("batchNumLevel") as HTMLInputElement).value),
        BatchBaseSize: Number((document.getElementById("batchBaseSize") as HTMLInputElement).value),
        BatchIncrement: Number((document.getElementById("batchIncrement") as HTMLInputElement).value),
        BatchLength: Number((document.getElementById("batchLength") as HTMLInputElement).value),
    };

    makeRequest(
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

async function makeRequest(url: RequestInfo | URL, parameters: RequestInit | undefined, after: (result: any, error: boolean) => void) {
    const response = await fetch(url, parameters);

    let result = await response.json();
    appendResult(url, result, !response.ok || response.status != 200)
    after(result, !response.ok || response.status != 200)
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

function appendResult(url: RequestInfo | URL, result: any, error: boolean) {
    const uniqueId = `json-content-${Date.now()}`;
    const formattedJSON = JSON.stringify(result, null, 2);
    const newResultHTML = `
        <div style="background-color:${error ? "#ffb9b9" : "#bdffbd"}; color: black; border:1px solid; padding:5px; margin-bottom: 10px;">
            <div onclick="window.toggleVisibility('${uniqueId}')" 
                style="cursor: pointer; font-weight: bold; padding: 4px;">
                ${url.toString().split(ui.url)[1]}
            </div>
            <div id="${uniqueId}" style="display: none; padding: 4px;">
                <pre>${formattedJSON}</pre>
            </div>
        </div>
    `;
    resultDiv.innerHTML = newResultHTML + resultDiv.innerHTML;
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
        setBtn.addEventListener("click", () => setRecord(null));
        deleteBtn.addEventListener("click", () => deleteRecord(null));
        getBtn.addEventListener("click", () => getRecord(null));

        treeForm.addEventListener("submit", newTreeRequest)
    });
}
