const vscode = require('vscode');
const {
	exec
} = require('child_process');
const path = require('path');
const fs = require('fs');

const getGoFiles = (dirPath) => {
	let results = [];
	const list = fs.readdirSync(dirPath);

	list.forEach(file => {
		const filePath = path.join(dirPath, file);
		const stat = fs.statSync(filePath);

		if (stat && stat.isDirectory()) {
			results = results.concat(getGoFiles(filePath));
		} else if (file.endsWith('.go')) {
			results.push(filePath);
		}
	});

	return results;
};


function activate(context) {
	let disposable = vscode.commands.registerCommand('callgraph.analyze', async () => {
		const options = ['Active File', 'All Files in Project'];
		const selectedOption = await vscode.window.showQuickPick(options, {
			placeHolder: 'Select an analysis option'
		});

		if (!selectedOption) {
			return;
		}

		const editor = vscode.window.activeTextEditor;
		let filePath = editor.document.uri.fsPath;

		const workspaceFolders = vscode.workspace.workspaceFolders;
		let rootPath = workspaceFolders && workspaceFolders.length > 0 ? workspaceFolders[0].uri.fsPath : '';

		//vscode.window.showInformationMessage(`File Path: ${filePath}`);
		//vscode.window.showInformationMessage(`Root Path: ${rootPath}`);

		if (selectedOption === 'Active File') {
			try {
				const analysisData = await analyzeGoFile(filePath, selectedOption);
				createVisualization(context, analysisData);
			} catch (err) {
				vscode.window.showErrorMessage(`Analysis failed: ${err.message}`);
			}
		} else if (selectedOption === 'All Files in Project') {
			const dirPath = rootPath;
			//vscode.window.showInformationMessage(`Root Path: ${dirPath}`);

			try {
				const files = getGoFiles(dirPath);
				if (files.length === 0) {
					vscode.window.showWarningMessage('No Go files found in the directory.');
				} else {
					//vscode.window.showInformationMessage(`Found Go files: ${files.join(', ')}`);

					const analysisData = await analyzeGoFile(files, selectedOption);
					createVisualization(context, analysisData);
				}
			} catch (err) {
				vscode.window.showErrorMessage(`Error reading directory: ${err.message}`);
			}
		}
	});

	context.subscriptions.push(disposable);
}

async function analyzeGoFile(filePaths, option) {
	return new Promise((resolve, reject) => {
		const analyzerPath = path.join(__dirname, 'tools/golang', 'analyzer');
		// TODO: Is it necessary to build the analyzer every time?
		exec(`cd ${path.join(__dirname, 'tools/golang')} && go build -o analyzer analyzer.go`, (buildErr) => {
			if (buildErr) {
				reject(buildErr);
				return;
			}

			if (option === 'All Files in Project') {
				let command = `${analyzerPath} ${filePaths.join(' ')}`;
				exec(command, (error, stdout, stderr) => {
					if (error) {
						reject(error);
						return;
					}
					try {
						const data = JSON.parse(stdout);
						//vscode.window.showInformationMessage(`Analyzer Output: ${stdout}`);
						resolve(data);
					} catch (err) {
						reject(new Error('Failed to parse analyzer output'));
					}
				});
			} else if (option === 'Active File') {
				let command = `${analyzerPath} ${filePaths}`;
				exec(command, (error, stdout, stderr) => {
					if (error) {
						reject(error);
						return;
					}
					try {
						const data = JSON.parse(stdout);
						//vscode.window.showInformationMessage(`Analyzer Output: ${stdout}`);
						resolve(data);
					} catch (err) {
						reject(new Error('Failed to parse analyzer output'));
					}
				});
			} else {
				reject(new Error('Invalid option selected.'));
			}
		});
	});
}

function createVisualization(context, analysisData) {
	const panel = vscode.window.createWebviewPanel(
		'goCallGraph',
		'Go Call Graph',
		vscode.ViewColumn.One, {
			enableScripts: true
		}
	);

	panel.webview.onDidReceiveMessage(async message => {
		if (message.command === 'openFile') {
			//console.log("Received file path:", message);
			const absoluteFilePath = vscode.Uri.file(path.resolve(__dirname, message.file));
			vscode.workspace.openTextDocument(absoluteFilePath).then(doc => {
				vscode.window.showTextDocument(doc, {
					selection: new vscode.Range(message.line - 1, 0, message.line - 1, 0)
				});
			});
		}
	});

	const htmlPath = vscode.Uri.file(path.join(__dirname, 'media', 'graph.html'));
	fs.readFile(htmlPath.fsPath, 'utf8', (err, htmlContent) => {
		if (err) {
			console.error("Error reading HTML file:", err);
			return;
		}

		const updatedHtmlContent = htmlContent.replace("{{data}}", JSON.stringify(analysisData));
		panel.webview.html = updatedHtmlContent;
	});
}

module.exports = {
	activate,
	deactivate: () => {}
};