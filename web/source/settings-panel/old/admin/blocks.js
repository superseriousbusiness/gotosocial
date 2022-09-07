/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

"use strict";

const Promise = require("bluebird");
const React = require("react");
const fileDownload = require("js-file-download");

function sortBlocks(blocks) {
	return blocks.sort((a, b) => { // alphabetical sort
		return a.domain.localeCompare(b.domain);
	});
}

function deduplicateBlocks(blocks) {
	let a = new Map();
	blocks.forEach((block) => {
		a.set(block.id, block);
	});
	return Array.from(a.values());
}

module.exports = function Blocks({oauth}) {
	const [blocks, setBlocks] = React.useState([]);
	const [info, setInfo] = React.useState("Fetching blocks");
	const [errorMsg, setError] = React.useState("");
	const [checked, setChecked] = React.useState(new Set());

	React.useEffect(() => {
		Promise.try(() => {
			return oauth.apiRequest("/api/v1/admin/domain_blocks", undefined, undefined, "GET");
		}).then((json) => {
			setInfo("");
			setError("");
			setBlocks(sortBlocks(json));
		}).catch((e) => {
			setError(e.message);
			setInfo("");
		});
	}, []);

	let blockList = blocks.map((block) => {
		function update(e) {
			let newChecked = new Set(checked.values());
			if (e.target.checked) {
				newChecked.add(block.id);
			} else {
				newChecked.delete(block.id);
			}
			setChecked(newChecked);
		}

		return (
			<React.Fragment key={block.id}>
				<div><input type="checkbox" onChange={update} checked={checked.has(block.id)}></input></div>
				<div>{block.domain}</div>
				<div>{(new Date(block.created_at)).toLocaleString()}</div>
			</React.Fragment>
		);
	});

	function clearChecked() {
		setChecked(new Set());
	}

	function undoChecked() {
		let amount = checked.size;
		if(confirm(`Are you sure you want to remove ${amount} block(s)?`)) {
			setInfo("");
			Promise.map(Array.from(checked.values()), (block) => {
				console.log("deleting", block);
				return oauth.apiRequest(`/api/v1/admin/domain_blocks/${block}`, "DELETE");
			}).then((res) => {
				console.log(res);
				setInfo(`Deleted ${amount} blocks: ${res.map((a) => a.domain).join(", ")}`);
			}).catch((e) => {
				setError(e);
			});

			let newBlocks = blocks.filter((block) => {
				if (checked.size > 0 && checked.has(block.id)) {
					checked.delete(block.id);
					return false;
				} else {
					return true;
				}
			});
			setBlocks(newBlocks);
			clearChecked();
		}
	}

	return (
		<section className="blocks">
			<h1>Blocks</h1>
			<div className="error accent">{errorMsg}</div>
			<div>{info}</div>
			<AddBlock oauth={oauth} blocks={blocks} setBlocks={setBlocks} />
			<h3>Blocks:</h3>
			<div style={{display: "grid", gridTemplateColumns: "1fr auto"}}>
				<span onClick={clearChecked} className="accent" style={{alignSelf: "end"}}>uncheck all</span>
				<button onClick={undoChecked}>Unblock selected</button>
			</div>
			<div className="blocklist overflow">
				{blockList}
			</div>
			<BulkBlocking oauth={oauth} blocks={blocks} setBlocks={setBlocks}/>
		</section>
	);
};

function BulkBlocking({oauth, blocks, setBlocks}) {
	const [bulk, setBulk] = React.useState("");
	const [blockMap, setBlockMap] = React.useState(new Map());
	const [output, setOutput] = React.useState();

	React.useEffect(() => {
		let newBlockMap = new Map();
		blocks.forEach((block) => {
			newBlockMap.set(block.domain, block);
		});
		setBlockMap(newBlockMap);
	}, [blocks]);

	const fileRef = React.useRef();

	function error(e) {
		setOutput(<div className="error accent">{e}</div>);
		throw e;
	}

	function fileUpload() {
		let reader = new FileReader();
		reader.addEventListener("load", (e) => {
			try {
				// TODO: use validatem?
				let json = JSON.parse(e.target.result);
				json.forEach((block) => {
					console.log("block:", block);
				});
			} catch(e) {
				error(e.message);
			}
		});
		reader.readAsText(fileRef.current.files[0]);
	}

	React.useEffect(() => {
		if (fileRef && fileRef.current) {
			fileRef.current.addEventListener("change", fileUpload);
		}
		return function cleanup() {
			fileRef.current.removeEventListener("change", fileUpload);
		};
	});

	function textImport() {
		Promise.try(() => {
			if (bulk[0] == "[") {
				// assume it's json
				return JSON.parse(bulk);
			} else {
				return bulk.split("\n").map((val) => {
					return {
						domain: val.trim()
					};
				});
			}
		}).then((domains) => {
			console.log(domains);
			let before = domains.length;
			setOutput(`Importing ${before} domain(s)`);
			domains = domains.filter(({domain}) => {
				return (domain != "" && !blockMap.has(domain));
			});
			setOutput(<span>{output}<br/>{`Deduplicated ${before - domains.length}/${before} with existing blocks, adding ${domains.length} block(s)`}</span>);
			if (domains.length > 0) {
				let data = new FormData();
				data.append("domains", new Blob([JSON.stringify(domains)], {type: "application/json"}), "import.json");
				return oauth.apiRequest("/api/v1/admin/domain_blocks?import=true", "POST", data, "form");
			}
		}).then((json) => {
			console.log("bulk import result:", json);
			setBlocks(sortBlocks(deduplicateBlocks([...json, ...blocks])));
		}).catch((e) => {
			error(e.message);
		});
	}

	function textExport() {
		setBulk(blocks.reduce((str, val) => {
			if (typeof str == "object") {
				return str.domain;
			} else {
				return str + "\n" + val.domain;
			}
		}));
	}

	function jsonExport() {
		Promise.try(() => {
			return oauth.apiRequest("/api/v1/admin/domain_blocks?export=true", "GET");
		}).then((json) => {
			fileDownload(JSON.stringify(json), "block-export.json");
		}).catch((e) => {
			error(e);
		});
	}

	function textAreaUpdate(e) {
		setBulk(e.target.value);
	}

	return (
		<React.Fragment>
			<h3>Bulk import/export</h3>
			<label htmlFor="bulk">Domains, one per line:</label>
			<textarea value={bulk} rows={20} onChange={textAreaUpdate}></textarea>
			<div className="controls">
				<button onClick={textImport}>Import All From Field</button>
				<button onClick={textExport}>Export To Field</button>
				<label className="button" htmlFor="upload">Upload .json</label>
				<button onClick={jsonExport}>Download .json</button>
			</div>
			{output}
			<input type="file" id="upload" className="hidden" ref={fileRef}></input>
		</React.Fragment>
	);
}

function AddBlock({oauth, blocks, setBlocks}) {
	const [domain, setDomain] = React.useState("");
	const [type, setType] = React.useState("suspend");
	const [obfuscated, setObfuscated] = React.useState(false);
	const [privateDescription, setPrivateDescription] = React.useState("");
	const [publicDescription, setPublicDescription] = React.useState("");

	function addBlock() {
		console.log(`${type}ing`, domain);
		Promise.try(() => {
			return oauth.apiRequest("/api/v1/admin/domain_blocks", "POST", {
				domain: domain,
				obfuscate: obfuscated,
				private_comment: privateDescription,
				public_comment: publicDescription
			}, "json");
		}).then((json) => {
			setDomain("");
			setPrivateDescription("");
			setPublicDescription("");
			setBlocks([json, ...blocks]);
		});
	}

	function onDomainChange(e) {
		setDomain(e.target.value);
	}

	function onTypeChange(e) {
		setType(e.target.value);
	}

	function onKeyDown(e) {
		if (e.key == "Enter") {
			addBlock();
		}
	}

	return (
		<React.Fragment>
			<h3>Add Block:</h3>
			<div className="addblock">
				<input id="domain" placeholder="instance" onChange={onDomainChange} value={domain} onKeyDown={onKeyDown} />
				<select value={type} onChange={onTypeChange}>
					<option id="suspend">Suspend</option>
					<option id="silence">Silence</option>
				</select>
				<button onClick={addBlock}>Add</button>
				<div>
					<label htmlFor="private">Private description:</label><br/>
					<textarea id="private" value={privateDescription} onChange={(e) => setPrivateDescription(e.target.value)}></textarea>
				</div>
				<div>
					<label htmlFor="public">Public description:</label><br/>
					<textarea id="public" value={publicDescription} onChange={(e) => setPublicDescription(e.target.value)}></textarea>
				</div>
				<div className="single">
					<label htmlFor="obfuscate">Obfuscate:</label>
					<input id="obfuscate" type="checkbox" value={obfuscated} onChange={(e) => setObfuscated(e.target.checked)}/>
				</div>
			</div>
		</React.Fragment>
	);
}

// function Blocklist() {
// 	return (
// 		<section className="blocklists">
// 			<h1>Blocklists</h1>
// 		</section>
// 	);
// }