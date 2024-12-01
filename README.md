# PathFinder

PathFinder is a tool designed to visualize call graphs of source code. It parses source files to extract function declarations and calls, and then generates a visual representation of the call graph. This tool is particularly useful for understanding the structure and flow of the applications, especially in complex codebases. Additionally, the main reason I developed this application/extension is to assist me during my secure code review process. For detailed usage scenarios, please refer to the 'Use-Cases' section.

## Support Languages

Below, I’ve added a list of supported languages and those planned for future support. Even for the languages I consider supported, there might still be numerous bugs or edge cases. In such cases, I can say: "PRs are welcome!"

- [x] Golang
- [ ] Python - *It would be nice to do it :)*
- [ ] Java - *It would be nice to do it :)*

## Features

- **Function Declaration and Call Detection**: Parses source files to identify function declarations and calls.
- **Endpoint Detection**: Identifies and visualizes HTTP endpoints in web applications. This part might not work perfectly, but I believe we can improve it over time to get closer to perfection.
- **JSON Output**: Converts the call graph data into JSON format for further processing or visualization. Under the `tools/` folder, you can find the analyzers. Each language will have its own analyzer, but their outputs should follow the same format. This way, we can integrate analyzers for different languages with minimal effort. For an example output, please refer to the *Analyzer(s)* section.
- **Interactive Visualization**: Provides an interactive web-based visualization of the call graph using D3.js. 

## User Interface Features

- **Node Interaction**: By clicking on any node, the flow leading to this node from the outermost level can be highlighted.
- **Node Deletion**: Right-click on a node to delete the related flow and clicked node.
- **Zoom and Pan**: Use the mouse wheel to zoom in and out of the graph. Click and drag to pan around the visualization, allowing you to explore different parts of the call graph.
- **Search Functionality**: A search bar is available to quickly locate specific functions within the graph. Enter the function name to highlight and focus on the corresponding node.
- **Show Only Endpoints**: By clicking the 'Show Only Endpoints' button, nodes identified as endpoints in web applications can be highlighted.
- **Jump to Code Location**: By hovering over any node and performing a "Command + Left Click" you can see where that node (i.e., function) is called within the codebase. Additionally, you can click on the listed locations to navigate directly to their positions in the editor.
- **Filter Nodes**: By entering a number in the **Filter Nodes** box, you can remove functions that are called more than the specified number of times from the graph. To explain its necessity, consider the Go programming language: the `Println()` function is called by many functions, which can clutter the graph with unnecessary and unhelpful visuals. By filtering out commonly called functions, you can clean up the graph and focus on more relevant details.
- **Tooltip Information**: Hover over nodes and edges to see tooltips with additional information, such as function argument(s).
- **Dark Mode**: As you know, Dark & Light modes.

## Analyzer(s)

As mentioned above, we will add an analyzer for each language under the `tools/` folder over time. Currently, only the analyzer for the Go programming language is available. To avoid frequent UI changes and enable quick integrations in the future, all analyzers should produce the same output format. Below is an example of an analyzer output:

Example source code:

```golang
//third.go
import "fmt"

func AdditionalFunction3() {
        fmt.Println("Inside additionalFunction from secondary.go")
        branchX()
}

func branchX() {
        fmt.Println("Inside branchE")
        branchY()
}

func branchY() {
        fmt.Println("Inside branchF")
}
```

```bash
$ ./analyzer ../../example/folder1/third.go|jq
{
  "nodes": [
    {
      "id": "AdditionalFunction3",
      "name": "AdditionalFunction3",
      "args": [],
      "returns": []
    },
    {
      "id": "branchX",
      "name": "branchX",
      "args": [],
      "returns": []
    },
    {
      "id": "branchY",
      "name": "branchY",
      "args": [],
      "returns": []
    },
    {
      "id": "Println",
      "name": "Println",
      "args": [
        "\"Inside additionalFunction from secondary.go\""
      ],
      "returns": []
    }
  ],
  "edges": [
    {
      "source": "AdditionalFunction3",
      "target": "Println",
      "line": 7,
      "file": "../../example/folder1/third.go",
      "endpoint": ""
    },
    {
      "source": "AdditionalFunction3",
      "target": "branchX",
      "line": 8,
      "file": "../../example/folder1/third.go",
      "endpoint": ""
    },
    {
      "source": "branchX",
      "target": "Println",
      "line": 13,
      "file": "../../example/folder1/third.go",
      "endpoint": ""
    },
    {
      "source": "branchX",
      "target": "branchY",
      "line": 14,
      "file": "../../example/folder1/third.go",
      "endpoint": ""
    },
    {
      "source": "branchY",
      "target": "Println",
      "line": 18,
      "file": "../../example/folder1/third.go",
      "endpoint": ""
    }
  ]
}
```

## Real World Use Cases

* You’ve identified a vulnerable function and want to see how accessible it is from external sources. Search for the function and left-click on it. You’ll be able to view the path leading from external sources to this function.
* You can visually see how an exposed endpoint branches out behind it. Click the 'Show Only Endpoints' button to view only the endpoints. Then, explore the flows originating from these endpoints to analyze their paths.
* You can quickly understand the functions of nodes based on their colors. For instance, orange-colored boxes represent functions with at least one argument, while red-colored ones indicate endpoints only.
* If you're curious about where a function is being called from, "Command + Left Click" on the node to see the call locations, and return to the editor for a detailed inspection.
* You can spot favorite functions. :)
* I'm not sure if it's necessary, but you can spot functions defined in the codebase but not used anywhere—floating aimlessly within the graph. :)

## Installation

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/fatihhcelik/pathfinder
   cd pathfinder
   ```

2. **Install Dependencies**:
   Ensure you have Go installed on your system. Then, run:
   ```bash
   cd tools/golang/ && go mod tidy
   ```

   Also, install npm dependencies;
   ```bash
   npm i
   ``` 

3. **Run the extension**:
   - Open project with Vscode and hit F5.
   - Choose your project folder to be analyzed. 
   - Hit "Command + Shift + P" and search for "Analyze Call Graph"
   - Choose one of the options, "Active File" or "All Files in the Project".

## Code Structure

- **Main Logic**: The main logic for parsing and generating the call graph is located in `tools/analyzer.go`.
- **Visualization**: The HTML and JavaScript for the interactive visualization are located in `media/graph.html`.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request with your changes.

