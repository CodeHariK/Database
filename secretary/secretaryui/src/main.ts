import * as d3 from "d3";

export interface BPlusTreeNode {
  keys: number[];
  children?: BPlusTreeNode[];
}

export function renderBPlusTree(treeData: BPlusTreeNode, order: number) {
  const width = window.innerWidth;
  const height = window.innerHeight;
  const svg = d3.select("#tree-container")
    .attr("width", width)
    .attr("height", height)
    .call(d3.zoom().on("zoom", (event) => {
      g.attr("transform", event.transform);
    }))
    .append("g");

  svg.selectAll("*").remove(); // Clear previous drawings

  const treeLayout = d3.tree<BPlusTreeNode>().size([width - 200, height - 200]);
  const root = d3.hierarchy(treeData);
  treeLayout(root);

  const g = svg.append("g");

  const lineGenerator = d3.line<d3.HierarchyPointLink<BPlusTreeNode>>()
    .x(d => d.x)
    .y(d => d.y)
    .curve(d3.curveBasis);

  const links = g.selectAll(".link")
    .data(root.links())
    .enter()
    .append("path")
    .attr("class", "link")
    .attr("fill", "none")
    .attr("stroke", "#007acc")
    .attr("stroke-width", 3)
    .attr("d", d => lineGenerator([d.source, d.target]));

  const nodes = g.selectAll(".node")
    .data(root.descendants())
    .enter()
    .append("g")
    .attr("class", "node")
    .attr("transform", d => `translate(${d.x}, ${d.y})`);

  nodes.append("rect")
    .attr("width", d => d.data.keys.length * 40 + 20)
    .attr("height", 50)
    .attr("x", d => -((d.data.keys.length * 40 + 20) / 2))
    .attr("y", -25)
    .attr("rx", 10)
    .attr("ry", 10)
    .attr("fill", "lightblue")
    .attr("stroke", "black");

  nodes.selectAll(".key-box")
    .data(d => d.data.keys.map((key, i) => ({ key, index: i, total: d.data.keys.length })))
    .enter()
    .append("rect")
    .attr("class", "key-box")
    .attr("width", 30)
    .attr("height", 30)
    .attr("x", d => -((d.total * 40 + 20) / 2) + d.index * 40 + 10)
    .attr("y", -15)
    .attr("rx", 5)
    .attr("ry", 5)
    .attr("fill", "white")
    .attr("stroke", "black");

  nodes.selectAll(".key-text")
    .data(d => d.data.keys.map((key, i) => ({ key, index: i, total: d.data.keys.length })))
    .enter()
    .append("text")
    .attr("class", "key-text")
    .attr("x", d => -((d.total * 40 + 20) / 2) + d.index * 40 + 25)
    .attr("y", 0)
    .attr("text-anchor", "middle")
    .attr("alignment-baseline", "middle")
    .text(d => d.key);

  const dragHandler = d3.drag<SVGGElement, d3.HierarchyNode<BPlusTreeNode>>()
    .on("start", function (event) {
      d3.select(this).raise().classed("active", true);
    })
    .on("drag", function (event, d) {
      d3.select(this).attr("transform", `translate(${event.x}, ${event.y})`);
      d.x = event.x;
      d.y = event.y;
      links.attr("d", d => lineGenerator([d.source, d.target]));
    })
    .on("end", function () {
      d3.select(this).classed("active", false);
    });

  dragHandler(nodes);
}

const sampleTree: BPlusTreeNode = {
  keys: [20, 40],
  children: [
    { keys: [5, 10, 15], children: [{ keys: [2, 3, 4] }, { keys: [6, 7, 9] }, { keys: [11, 12, 14] }] },
    { keys: [25, 30, 35], children: [{ keys: [21, 22, 24] }, { keys: [26, 27, 29] }, { keys: [31, 32, 34] }] },
    { keys: [45, 50], children: [{ keys: [41, 42, 44] }, { keys: [46, 47, 49] }] }
  ]
};

renderBPlusTree(sampleTree, 10)

