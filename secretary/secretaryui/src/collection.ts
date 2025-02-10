import { convertJsonToBPlusTree, createBPlusTreeFromJSON } from "./draw";

async function handleCardClick(item: any) {

    const response = await fetch("http://localhost:8080/getbtree?key=" + item.collectionName);

    if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
    }

    const data = await response.json(); // Assuming data is an array

    const bPlusTree = convertJsonToBPlusTree(data)

    createBPlusTreeFromJSON(bPlusTree);
}

export async function fetchAllBTree() {
    try {
        const response = await fetch("http://localhost:8080/getallbtree?key=users");

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
            card.onclick = () => handleCardClick(item);

            collectionsDiv.appendChild(card);
        });

    } catch (error) {
        console.error("Error fetching data:", error);
        document.getElementById("collections")!.textContent = "Error fetching data";
    }
}
