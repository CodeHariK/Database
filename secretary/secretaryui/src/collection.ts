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
const queryBtn = document.getElementById("query-btn") as HTMLButtonElement;
const resultDiv = document.getElementById("result") as HTMLDivElement;

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

async function queryData() {
    const query = (document.getElementById("query") as HTMLInputElement).value;
    try {
        const response = await fetch(`/query?table=${currentTree!.collectionName}/id=${query}`);
        const result = await response.json();
        resultDiv.textContent = JSON.stringify(result, null, 2);
    } catch (error) {
        console.error("Error querying data:", error);
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
    queryBtn.addEventListener("click", queryData);
});
