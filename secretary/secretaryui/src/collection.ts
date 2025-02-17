import { ui } from "./main";

import { RedrawTree } from "./draw";
import { BTree, BTreeNode } from "./tree";
import { deleteBtn, insertBtn, nextTreeBtn, prevTreeBtn, resultDiv, searchBtn, searchInput, treeForm } from "./dom";

export function displayNode(node: BTreeNode) {
    const infoBox = document.getElementById("info-box");
    if (!infoBox) return;

    // Clear previous content
    infoBox.innerHTML = "";

    // Create Node ID section
    const nodeIdDiv = document.createElement("div");
    nodeIdDiv.textContent = `Node ID: ${node.nodeID}`;
    nodeIdDiv.style.fontSize = "17px";
    nodeIdDiv.style.marginBottom = "10px";
    infoBox.appendChild(nodeIdDiv);

    // Create the list container
    const list = document.createElement("ul");
    list.style.listStyle = "none";
    list.style.padding = "0";

    node.keys.forEach((key, index) => {
        const listItem = document.createElement("li");
        listItem.style.marginTop = "15px";
        listItem.style.alignItems = "center";

        // Key text
        const keySpan = document.createElement("span");
        keySpan.textContent = `${key} `;

        // Input for modifying value
        const valueInput = document.createElement("textarea");
        valueInput.value = node.value[index];
        valueInput.style.width = '100%'
        valueInput.style.padding = '7px'
        valueInput.style.overflow = "hidden"; // Hide scrollbar
        valueInput.style.resize = "none"; // Disable manual resize
        valueInput.style.minHeight = "20px"; // Set a minimum height
        valueInput.oninput = () => {
            node.value[index] = valueInput.value;

            valueInput.style.height = "auto"; // Reset height
            valueInput.style.height = valueInput.scrollHeight + "px"; // Set new height
        };

        const updateButton = document.createElement("button");
        updateButton.textContent = " Update ";
        updateButton.style.backgroundColor = "#be7eb1";
        updateButton.style.color = "white";
        updateButton.style.border = "none";
        updateButton.style.cursor = "pointer";
        updateButton.onclick = () => {
            // node.keys.splice(index, 1);
            // node.value.splice(index, 1);
            displayNode(node); // Re-render the updated node
        };

        const deleteButton = document.createElement("button");
        deleteButton.textContent = " Delete ";
        deleteButton.style.backgroundColor = "#f15858";
        deleteButton.style.color = "white";
        deleteButton.style.border = "none";
        deleteButton.style.cursor = "pointer";
        deleteButton.onclick = () => {
            // node.keys.splice(index, 1);
            // node.value.splice(index, 1);
            displayNode(node); // Re-render the updated node
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

    const response = await fetch(`${ui.url}/gettree/${ui.currentTreeDef!.collectionName}`);

    if (!response.ok) {
        throw new Error(`Error fetch current tree : HTTP error! Status: ${response.statusText} ${response.status}`);
    }

    const data = await response.json(); // Assuming data is an array

    const bTree = new BTree(ui.currentTreeDef!.collectionName, data)

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

export async function fetchAllBTree() {
    try {
        const response = await fetch(`${ui.url}/getalltree`);

        if (!response.ok) {
            throw new Error(`Error fetch all btree : HTTP error! Status: ${response.statusText} ${response.status}`);
        }

        const data: any[] = await response.json(); // Assuming data is an array

        const collectionsDiv = document.getElementById("collections")!;
        collectionsDiv.innerHTML = ""; // Clear existing content

        data.forEach((tree, _) => {
            const card = document.createElement("div");
            card.className = "card";
            card.innerHTML = `<p>${JSON.stringify(tree, null, 2)}</p>`;
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

    } catch (error) {
        console.error("Error fetching data:", error);
        document.getElementById("collections")!.textContent = "Error fetching data";
    }
}

async function insertData() {
    const value = (document.getElementById("value") as HTMLInputElement).value;
    const payload = { value };
    try {
        const response = await fetch(`${ui.url}/insert/${ui.currentTreeDef!.collectionName}`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
        });
        if (!response.ok) {
            throw new Error(`Error inserting data : HTTP error! Status: ${response.statusText} ${response.status}`);
        }

        const result = await response.json();
        resultDiv.textContent = JSON.stringify(result, null, 2);

        fetchCurrentTree()
    } catch (error) {
        console.error("Error inserting data:", error);
    }
}

async function deleteData() {
    const id = (document.getElementById("delete-id") as HTMLInputElement).value;
    const response = await fetch(`${ui.url}/delete/${ui.currentTreeDef!.collectionName}/${id}`, { method: "DELETE" });
    if (!response.ok) {
        throw new Error(`Error deleting : HTTP error! Status: ${response.statusText} ${response.status}`);
    }

    const result = await response.json();
    resultDiv.textContent = JSON.stringify(result, null, 2);

    fetchCurrentTree()
}

async function searchRecord() {
    const search = searchInput.value;
    const response = await fetch(`${ui.url}/search/${ui.currentTreeDef!.collectionName}/${search}`);
    if (!response.ok) {
        throw new Error(`Error search : HTTP error! Status: ${response.statusText} ${response.status}`);
    }

    const result = await response.json();
    resultDiv.textContent = JSON.stringify(result, null, 2);

    let searchResult = ui.getTree().searchLeafNode(search)

    ui.NODEMAP.forEach((n) => {
        n.selected = false
    })
    searchResult.path.forEach((e) => {
        ui.NODEMAP.get(e.nodeID)!.selected = true
    })

    RedrawTree()
}

async function newTreeRequest(event: SubmitEvent) {
    event.preventDefault(); // Prevent page reload

    const requestData = {
        CollectionName: (document.getElementById("collectionName") as HTMLInputElement).value,
        Order: Number((document.getElementById("order") as HTMLInputElement).value),
        BatchNumLevel: Number((document.getElementById("batchNumLevel") as HTMLInputElement).value),
        BatchBaseSize: Number((document.getElementById("batchBaseSize") as HTMLInputElement).value),
        BatchIncrement: Number((document.getElementById("batchIncrement") as HTMLInputElement).value),
        BatchLength: Number((document.getElementById("batchLength") as HTMLInputElement).value),
    };

    try {
        const response = await fetch(`${ui.url}/newtree`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(requestData),
        });

        const result = await response.json();
        resultDiv.textContent = `Success: ${JSON.stringify(result)}`;
        resultDiv.style.color = "green";

        fetchAllBTree()
    } catch (error) {
        resultDiv.textContent = `Error: ${error}`;
        resultDiv.style.color = "red";
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

        insertBtn.addEventListener("click", insertData);
        deleteBtn.addEventListener("click", deleteData);
        searchBtn.addEventListener("click", searchRecord);

        treeForm.addEventListener("submit", newTreeRequest)
    });
}
