import { convertJsonToBPlusTree, createBPlusTreeFromJSON } from "./draw";
import { ui } from "./main";

const url = "http://localhost:8080"
let currentCollection = ""

const insertBtn = document.getElementById("insert-btn") as HTMLButtonElement;
const deleteBtn = document.getElementById("delete-btn") as HTMLButtonElement;
const queryBtn = document.getElementById("query-btn") as HTMLButtonElement;
const resultDiv = document.getElementById("result") as HTMLDivElement;

async function handleCardClick(item: any) {

    currentCollection = item.collectionName

    const response = await fetch(`${url}/getbtree?table=${currentCollection}`);

    if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
    }

    const data = await response.json(); // Assuming data is an array

    const bPlusTree = convertJsonToBPlusTree(data)

    ui.graph.clear()

    createBPlusTreeFromJSON(bPlusTree, item.order);
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

        data.forEach((item, _) => {
            const card = document.createElement("div");
            card.className = "card";
            card.innerHTML = `<p>${JSON.stringify(item, null, 2)}</p>`;
            card.onclick = () => {
                Array.from(document.getElementsByClassName("highlight")).
                    forEach((e) => {
                        e.classList.remove("highlight")
                    })
                card.classList.add("highlight");
                handleCardClick(item)
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
        const response = await fetch(`${url}/api/insert?table=${currentCollection}`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
        });
        const result = await response.json();
        resultDiv.textContent = JSON.stringify(result, null, 2);
    } catch (error) {
        console.error("Error inserting data:", error);
    }
}

async function deleteData() {
    const id = (document.getElementById("delete-id") as HTMLInputElement).value;
    try {
        const response = await fetch(`${url}/api/delete?table=/${currentCollection}/${id}`, { method: "DELETE" });
        const result = await response.json();
        resultDiv.textContent = JSON.stringify(result, null, 2);
    } catch (error) {
        console.error("Error deleting data:", error);
    }
}

async function queryData() {
    const query = (document.getElementById("query") as HTMLInputElement).value;
    try {
        const response = await fetch(`/api/query?table=${currentCollection}/id=${query}`);
        const result = await response.json();
        resultDiv.textContent = JSON.stringify(result, null, 2);
    } catch (error) {
        console.error("Error querying data:", error);
    }
}

document.addEventListener("DOMContentLoaded", () => {
    insertBtn.addEventListener("click", insertData);
    deleteBtn.addEventListener("click", deleteData);
    queryBtn.addEventListener("click", queryData);
});
