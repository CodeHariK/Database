import { ui } from "./main";

import { BPlusTreeNode, convertJsonToBPlusTree, createBPlusTreeFromJSON } from "./draw";

const url = "http://localhost:8080"
let currentTree
let currentTreeSnapshotIndex = 0
let TreeSnapshots: Array<BPlusTreeNode> = []

const prevTreeBtn = document.getElementById("prev-tree") as HTMLButtonElement;
const nextTreeBtn = document.getElementById("next-tree") as HTMLButtonElement;

const insertBtn = document.getElementById("insert-btn") as HTMLButtonElement;
const deleteBtn = document.getElementById("delete-btn") as HTMLButtonElement;
const searchBtn = document.getElementById("search-btn") as HTMLButtonElement;
const resultDiv = document.getElementById("result") as HTMLDivElement;

export function displayNode(node: BPlusTreeNode) {
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

function showSnapshot() {
    ui.graph.clear()

    createBPlusTreeFromJSON(TreeSnapshots[currentTreeSnapshotIndex], currentTree!.order);
}

async function fetchCurrentTree() {

    const response = await fetch(`${url}/getbtree?table=${currentTree!.collectionName}`);

    if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
    }

    const data = await response.json(); // Assuming data is an array

    const bPlusTree = convertJsonToBPlusTree(data)
    TreeSnapshots.push(bPlusTree)
    currentTreeSnapshotIndex = TreeSnapshots.length - 1

    showSnapshot()
}

export async function fetchAllBTree() {
    try {
        const response = await fetch(`${url}/getallbtree`);

        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
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

                currentTree = tree
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
        const response = await fetch(`${url}/insert?table=${currentTree!.collectionName}`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
        });
        const result = await response.json();
        resultDiv.textContent = JSON.stringify(result, null, 2);

        fetchCurrentTree()
    } catch (error) {
        console.error("Error inserting data:", error);
    }
}

async function deleteData() {
    const id = (document.getElementById("delete-id") as HTMLInputElement).value;
    try {
        const response = await fetch(`${url}/delete?table=/${currentTree!.collectionName}/${id}`, { method: "DELETE" });
        const result = await response.json();
        resultDiv.textContent = JSON.stringify(result, null, 2);
    } catch (error) {
        console.error("Error deleting data:", error);
    }
}

async function searchData() {
    const search = (document.getElementById("search") as HTMLInputElement).value;
    try {
        const response = await fetch(`${url}/search/${currentTree!.collectionName}/${search}`);
        const result = await response.json();
        resultDiv.textContent = JSON.stringify(result, null, 2);
    } catch (error) {
        console.error("Error searching data:", error);
    }
}

document.addEventListener("DOMContentLoaded", () => {

    prevTreeBtn.addEventListener("click", () => {
        if (currentTreeSnapshotIndex != 0) {
            currentTreeSnapshotIndex--;
            showSnapshot()
        }
    });
    nextTreeBtn.addEventListener("click", () => {
        if (currentTreeSnapshotIndex != (TreeSnapshots.length - 1)) {
            currentTreeSnapshotIndex++;
            showSnapshot()
        }
    });

    insertBtn.addEventListener("click", insertData);
    deleteBtn.addEventListener("click", deleteData);
    searchBtn.addEventListener("click", searchData);
});
