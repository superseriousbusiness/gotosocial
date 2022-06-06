require=(function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({11:[function(require,module,exports){
"use strict";

function _slicedToArray(arr, i) { return _arrayWithHoles(arr) || _iterableToArrayLimit(arr, i) || _unsupportedIterableToArray(arr, i) || _nonIterableRest(); }

function _nonIterableRest() { throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method."); }

function _unsupportedIterableToArray(o, minLen) { if (!o) return; if (typeof o === "string") return _arrayLikeToArray(o, minLen); var n = Object.prototype.toString.call(o).slice(8, -1); if (n === "Object" && o.constructor) n = o.constructor.name; if (n === "Map" || n === "Set") return Array.from(o); if (n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n)) return _arrayLikeToArray(o, minLen); }

function _arrayLikeToArray(arr, len) { if (len == null || len > arr.length) len = arr.length; for (var i = 0, arr2 = new Array(len); i < len; i++) { arr2[i] = arr[i]; } return arr2; }

function _iterableToArrayLimit(arr, i) { var _i = arr == null ? null : typeof Symbol !== "undefined" && arr[Symbol.iterator] || arr["@@iterator"]; if (_i == null) return; var _arr = []; var _n = true; var _d = false; var _s, _e; try { for (_i = _i.call(arr); !(_n = (_s = _i.next()).done); _n = true) { _arr.push(_s.value); if (i && _arr.length === i) break; } } catch (err) { _d = true; _e = err; } finally { try { if (!_n && _i["return"] != null) _i["return"](); } finally { if (_d) throw _e; } } return _arr; }

function _arrayWithHoles(arr) { if (Array.isArray(arr)) return arr; }

var Promise = require("bluebird");

var React = require("react");

var ReactDom = require("react-dom");

var oauthLib = require("./oauth.js");

var Auth = require("./auth");

var Settings = require("./settings");

var Blocks = require("./blocks");

require("./style.css");

function App() {
  var _React$useState = React.useState(),
      _React$useState2 = _slicedToArray(_React$useState, 2),
      oauth = _React$useState2[0],
      setOauth = _React$useState2[1];

  var _React$useState3 = React.useState(false),
      _React$useState4 = _slicedToArray(_React$useState3, 2),
      hasAuth = _React$useState4[0],
      setAuth = _React$useState4[1];

  var _React$useState5 = React.useState(localStorage.getItem("oauth")),
      _React$useState6 = _slicedToArray(_React$useState5, 2),
      oauthState = _React$useState6[0],
      setOauthState = _React$useState6[1];

  React.useEffect(function () {
    var state = localStorage.getItem("oauth");

    if (state != undefined) {
      state = JSON.parse(state);
      var restoredOauth = oauthLib(state.config, state);
      Promise["try"](function () {
        return restoredOauth.callback();
      }).then(function () {
        setAuth(true);
      });
      setOauth(restoredOauth);
    }
  }, []);

  if (!hasAuth && oauth && oauth.isAuthorized()) {
    setAuth(true);
  }

  if (oauth && oauth.isAuthorized()) {
    return /*#__PURE__*/React.createElement(AdminPanel, {
      oauth: oauth
    });
  } else if (oauthState != undefined) {
    return "processing oauth...";
  } else {
    return /*#__PURE__*/React.createElement(Auth, {
      setOauth: setOauth
    });
  }
}

function AdminPanel(_ref) {
  var oauth = _ref.oauth;

  /* 
  	Features: (issue #78)
  	- [ ] Instance information updating
  		  GET /api/v1/instance PATCH /api/v1/instance
  	- [ ] Domain block creation, viewing, and deletion
  		  GET /api/v1/admin/domain_blocks
  		  POST /api/v1/admin/domain_blocks
  		  GET /api/v1/admin/domain_blocks/DOMAIN_BLOCK_ID, DELETE /api/v1/admin/domain_blocks/DOMAIN_BLOCK_ID
  	- [ ] Blocklist import/export
  		  GET /api/v1/admin/domain_blocks?export=true
  		  POST json file as form field domains to /api/v1/admin/domain_blocks
  */
  return /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement(Logout, {
    oauth: oauth
  }), /*#__PURE__*/React.createElement(Settings, {
    oauth: oauth
  }), /*#__PURE__*/React.createElement(Blocks, {
    oauth: oauth
  }));
}

function Logout(_ref2) {
  var oauth = _ref2.oauth;
  return /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("button", {
    onClick: oauth.logout
  }, "Logout"));
}

ReactDom.render( /*#__PURE__*/React.createElement(App, null), document.getElementById("root"));

},{"./auth":9,"./blocks":10,"./oauth.js":12,"./settings":13,"./style.css":14,"bluebird":15,"react":23,"react-dom":20}],14:[function(require,module,exports){
require("../../node_modules/icssify/global-css-loader.js"); module.exports = {};
},{"../../node_modules/icssify/global-css-loader.js":5}],13:[function(require,module,exports){
"use strict";

function _extends() { _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; }; return _extends.apply(this, arguments); }

function _toConsumableArray(arr) { return _arrayWithoutHoles(arr) || _iterableToArray(arr) || _unsupportedIterableToArray(arr) || _nonIterableSpread(); }

function _nonIterableSpread() { throw new TypeError("Invalid attempt to spread non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method."); }

function _iterableToArray(iter) { if (typeof Symbol !== "undefined" && iter[Symbol.iterator] != null || iter["@@iterator"] != null) return Array.from(iter); }

function _arrayWithoutHoles(arr) { if (Array.isArray(arr)) return _arrayLikeToArray(arr); }

function _typeof(obj) { "@babel/helpers - typeof"; if (typeof Symbol === "function" && typeof Symbol.iterator === "symbol") { _typeof = function _typeof(obj) { return typeof obj; }; } else { _typeof = function _typeof(obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; }; } return _typeof(obj); }

function _slicedToArray(arr, i) { return _arrayWithHoles(arr) || _iterableToArrayLimit(arr, i) || _unsupportedIterableToArray(arr, i) || _nonIterableRest(); }

function _nonIterableRest() { throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method."); }

function _unsupportedIterableToArray(o, minLen) { if (!o) return; if (typeof o === "string") return _arrayLikeToArray(o, minLen); var n = Object.prototype.toString.call(o).slice(8, -1); if (n === "Object" && o.constructor) n = o.constructor.name; if (n === "Map" || n === "Set") return Array.from(o); if (n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n)) return _arrayLikeToArray(o, minLen); }

function _arrayLikeToArray(arr, len) { if (len == null || len > arr.length) len = arr.length; for (var i = 0, arr2 = new Array(len); i < len; i++) { arr2[i] = arr[i]; } return arr2; }

function _iterableToArrayLimit(arr, i) { var _i = arr == null ? null : typeof Symbol !== "undefined" && arr[Symbol.iterator] || arr["@@iterator"]; if (_i == null) return; var _arr = []; var _n = true; var _d = false; var _s, _e; try { for (_i = _i.call(arr); !(_n = (_s = _i.next()).done); _n = true) { _arr.push(_s.value); if (i && _arr.length === i) break; } } catch (err) { _d = true; _e = err; } finally { try { if (!_n && _i["return"] != null) _i["return"](); } finally { if (_d) throw _e; } } return _arr; }

function _arrayWithHoles(arr) { if (Array.isArray(arr)) return arr; }

var Promise = require("bluebird");

var React = require("react");

module.exports = function Settings(_ref) {
  var oauth = _ref.oauth;

  var _React$useState = React.useState({}),
      _React$useState2 = _slicedToArray(_React$useState, 2),
      info = _React$useState2[0],
      setInfo = _React$useState2[1];

  var _React$useState3 = React.useState(""),
      _React$useState4 = _slicedToArray(_React$useState3, 2),
      errorMsg = _React$useState4[0],
      setError = _React$useState4[1];

  var _React$useState5 = React.useState("Fetching instance info"),
      _React$useState6 = _slicedToArray(_React$useState5, 2),
      statusMsg = _React$useState6[0],
      setStatus = _React$useState6[1];

  React.useEffect(function () {
    Promise["try"](function () {
      return oauth.apiRequest("/api/v1/instance", "GET");
    }).then(function (json) {
      setInfo(json);
    })["catch"](function (e) {
      setError(e.message);
      setStatus("");
    });
  }, []);

  function submit() {
    setStatus("PATCHing");
    setError("");
    return Promise["try"](function () {
      var formDataInfo = new FormData();
      Object.entries(info).forEach(function (_ref2) {
        var _ref3 = _slicedToArray(_ref2, 2),
            key = _ref3[0],
            val = _ref3[1];

        if (key == "contact_account") {
          key = "contact_username";
          val = val.username;
        }

        if (key == "email") {
          key = "contact_email";
        }

        if (_typeof(val) != "object") {
          formDataInfo.append(key, val);
        }
      });
      return oauth.apiRequest("/api/v1/instance", "PATCH", formDataInfo, "form");
    }).then(function (json) {
      setStatus("Config saved");
      console.log(json);
    })["catch"](function (e) {
      setError(e.message);
      setStatus("");
    });
  }

  return /*#__PURE__*/React.createElement("section", {
    className: "info login"
  }, /*#__PURE__*/React.createElement("h1", null, "Instance Information ", /*#__PURE__*/React.createElement("button", {
    onClick: submit
  }, "Save")), /*#__PURE__*/React.createElement("div", {
    className: "error accent"
  }, errorMsg), /*#__PURE__*/React.createElement("div", null, statusMsg), /*#__PURE__*/React.createElement("form", {
    onSubmit: function onSubmit(e) {
      return e.preventDefault();
    }
  }, editableObject(info)));
};

function editableObject(obj) {
  var path = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : [];
  var readOnlyKeys = ["uri", "version", "urls_streaming_api", "stats"];
  var hiddenKeys = ["contact_account_", "urls"];
  var explicitShownKeys = ["contact_account_username"];
  var implementedKeys = "title, contact_account_username, email, short_description, description, terms, avatar, header".split(", ");
  var listing = Object.entries(obj).map(function (_ref4) {
    var _ref5 = _slicedToArray(_ref4, 2),
        key = _ref5[0],
        val = _ref5[1];

    var fullkey = [].concat(_toConsumableArray(path), [key]).join("_");

    if (hiddenKeys.includes(fullkey) || hiddenKeys.includes(path.join("_") + "_") // also match just parent path
    ) {
      if (!explicitShownKeys.includes(fullkey)) {
        return null;
      }
    }

    if (Array.isArray(val)) {// FIXME: handle this
    } else if (_typeof(val) == "object") {
      return /*#__PURE__*/React.createElement(React.Fragment, {
        key: fullkey
      }, editableObject(val, [].concat(_toConsumableArray(path), [key])));
    }

    var isImplemented = "";

    if (!implementedKeys.includes(fullkey)) {
      isImplemented = " notImplemented";
    }

    var isReadOnly = readOnlyKeys.includes(fullkey) || readOnlyKeys.includes(path.join("_")) || isImplemented != "";
    var label = key.replace(/_/g, " ");

    if (path.length > 0) {
      label = "\xA0".repeat(4 * path.length) + label;
    }

    var inputProps;
    var changeFunc;

    if (val === true || val === false) {
      inputProps = {
        type: "checkbox",
        defaultChecked: val,
        disabled: isReadOnly
      };

      changeFunc = function changeFunc(e) {
        return e.target.checked;
      };
    } else if (val.length != 0 && !isNaN(val)) {
      inputProps = {
        type: "number",
        defaultValue: val,
        readOnly: isReadOnly
      };

      changeFunc = function changeFunc(e) {
        return e.target.value;
      };
    } else {
      inputProps = {
        type: "text",
        defaultValue: val,
        readOnly: isReadOnly
      };

      changeFunc = function changeFunc(e) {
        return e.target.value;
      };
    }

    function setRef(element) {
      if (element != null) {
        element.addEventListener("change", function (e) {
          obj[key] = changeFunc(e);
        });
      }
    }

    return /*#__PURE__*/React.createElement(React.Fragment, {
      key: fullkey
    }, /*#__PURE__*/React.createElement("label", {
      htmlFor: key,
      className: "capitalize"
    }, label), /*#__PURE__*/React.createElement("div", {
      className: isImplemented
    }, /*#__PURE__*/React.createElement("input", _extends({
      className: isImplemented,
      ref: setRef
    }, inputProps))));
  });
  return /*#__PURE__*/React.createElement(React.Fragment, null, path != "" && /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("b", null, path, ":"), " ", /*#__PURE__*/React.createElement("span", {
    id: "filler"
  })), listing);
}

},{"bluebird":15,"react":23}],10:[function(require,module,exports){
"use strict";

function _typeof(obj) { "@babel/helpers - typeof"; if (typeof Symbol === "function" && typeof Symbol.iterator === "symbol") { _typeof = function _typeof(obj) { return typeof obj; }; } else { _typeof = function _typeof(obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; }; } return _typeof(obj); }

function _toConsumableArray(arr) { return _arrayWithoutHoles(arr) || _iterableToArray(arr) || _unsupportedIterableToArray(arr) || _nonIterableSpread(); }

function _nonIterableSpread() { throw new TypeError("Invalid attempt to spread non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method."); }

function _iterableToArray(iter) { if (typeof Symbol !== "undefined" && iter[Symbol.iterator] != null || iter["@@iterator"] != null) return Array.from(iter); }

function _arrayWithoutHoles(arr) { if (Array.isArray(arr)) return _arrayLikeToArray(arr); }

function _slicedToArray(arr, i) { return _arrayWithHoles(arr) || _iterableToArrayLimit(arr, i) || _unsupportedIterableToArray(arr, i) || _nonIterableRest(); }

function _nonIterableRest() { throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method."); }

function _unsupportedIterableToArray(o, minLen) { if (!o) return; if (typeof o === "string") return _arrayLikeToArray(o, minLen); var n = Object.prototype.toString.call(o).slice(8, -1); if (n === "Object" && o.constructor) n = o.constructor.name; if (n === "Map" || n === "Set") return Array.from(o); if (n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n)) return _arrayLikeToArray(o, minLen); }

function _arrayLikeToArray(arr, len) { if (len == null || len > arr.length) len = arr.length; for (var i = 0, arr2 = new Array(len); i < len; i++) { arr2[i] = arr[i]; } return arr2; }

function _iterableToArrayLimit(arr, i) { var _i = arr == null ? null : typeof Symbol !== "undefined" && arr[Symbol.iterator] || arr["@@iterator"]; if (_i == null) return; var _arr = []; var _n = true; var _d = false; var _s, _e; try { for (_i = _i.call(arr); !(_n = (_s = _i.next()).done); _n = true) { _arr.push(_s.value); if (i && _arr.length === i) break; } } catch (err) { _d = true; _e = err; } finally { try { if (!_n && _i["return"] != null) _i["return"](); } finally { if (_d) throw _e; } } return _arr; }

function _arrayWithHoles(arr) { if (Array.isArray(arr)) return arr; }

var Promise = require("bluebird");

var React = require("react");

var fileDownload = require("js-file-download");

function sortBlocks(blocks) {
  return blocks.sort(function (a, b) {
    // alphabetical sort
    return a.domain.localeCompare(b.domain);
  });
}

function deduplicateBlocks(blocks) {
  var a = new Map();
  blocks.forEach(function (block) {
    a.set(block.id, block);
  });
  return Array.from(a.values());
}

module.exports = function Blocks(_ref) {
  var oauth = _ref.oauth;

  var _React$useState = React.useState([]),
      _React$useState2 = _slicedToArray(_React$useState, 2),
      blocks = _React$useState2[0],
      setBlocks = _React$useState2[1];

  var _React$useState3 = React.useState("Fetching blocks"),
      _React$useState4 = _slicedToArray(_React$useState3, 2),
      info = _React$useState4[0],
      setInfo = _React$useState4[1];

  var _React$useState5 = React.useState(""),
      _React$useState6 = _slicedToArray(_React$useState5, 2),
      errorMsg = _React$useState6[0],
      setError = _React$useState6[1];

  var _React$useState7 = React.useState(new Set()),
      _React$useState8 = _slicedToArray(_React$useState7, 2),
      checked = _React$useState8[0],
      setChecked = _React$useState8[1];

  React.useEffect(function () {
    Promise["try"](function () {
      return oauth.apiRequest("/api/v1/admin/domain_blocks", undefined, undefined, "GET");
    }).then(function (json) {
      setInfo("");
      setError("");
      setBlocks(sortBlocks(json));
    })["catch"](function (e) {
      setError(e.message);
      setInfo("");
    });
  }, []);
  var blockList = blocks.map(function (block) {
    function update(e) {
      var newChecked = new Set(checked.values());

      if (e.target.checked) {
        newChecked.add(block.id);
      } else {
        newChecked["delete"](block.id);
      }

      setChecked(newChecked);
    }

    return /*#__PURE__*/React.createElement(React.Fragment, {
      key: block.id
    }, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("input", {
      type: "checkbox",
      onChange: update,
      checked: checked.has(block.id)
    })), /*#__PURE__*/React.createElement("div", null, block.domain), /*#__PURE__*/React.createElement("div", null, new Date(block.created_at).toLocaleString()));
  });

  function clearChecked() {
    setChecked(new Set());
  }

  function undoChecked() {
    var amount = checked.size;

    if (confirm("Are you sure you want to remove ".concat(amount, " block(s)?"))) {
      setInfo("");
      Promise.map(Array.from(checked.values()), function (block) {
        console.log("deleting", block);
        return oauth.apiRequest("/api/v1/admin/domain_blocks/".concat(block), "DELETE");
      }).then(function (res) {
        console.log(res);
        setInfo("Deleted ".concat(amount, " blocks: ").concat(res.map(function (a) {
          return a.domain;
        }).join(", ")));
      })["catch"](function (e) {
        setError(e);
      });
      var newBlocks = blocks.filter(function (block) {
        if (checked.size > 0 && checked.has(block.id)) {
          checked["delete"](block.id);
          return false;
        } else {
          return true;
        }
      });
      setBlocks(newBlocks);
      clearChecked();
    }
  }

  return /*#__PURE__*/React.createElement("section", {
    className: "blocks"
  }, /*#__PURE__*/React.createElement("h1", null, "Blocks"), /*#__PURE__*/React.createElement("div", {
    className: "error accent"
  }, errorMsg), /*#__PURE__*/React.createElement("div", null, info), /*#__PURE__*/React.createElement(AddBlock, {
    oauth: oauth,
    blocks: blocks,
    setBlocks: setBlocks
  }), /*#__PURE__*/React.createElement("h3", null, "Blocks:"), /*#__PURE__*/React.createElement("div", {
    style: {
      display: "grid",
      gridTemplateColumns: "1fr auto"
    }
  }, /*#__PURE__*/React.createElement("span", {
    onClick: clearChecked,
    className: "accent",
    style: {
      alignSelf: "end"
    }
  }, "uncheck all"), /*#__PURE__*/React.createElement("button", {
    onClick: undoChecked
  }, "Unblock selected")), /*#__PURE__*/React.createElement("div", {
    className: "blocklist overflow"
  }, blockList), /*#__PURE__*/React.createElement(BulkBlocking, {
    oauth: oauth,
    blocks: blocks,
    setBlocks: setBlocks
  }));
};

function BulkBlocking(_ref2) {
  var oauth = _ref2.oauth,
      blocks = _ref2.blocks,
      setBlocks = _ref2.setBlocks;

  var _React$useState9 = React.useState(""),
      _React$useState10 = _slicedToArray(_React$useState9, 2),
      bulk = _React$useState10[0],
      setBulk = _React$useState10[1];

  var _React$useState11 = React.useState(new Map()),
      _React$useState12 = _slicedToArray(_React$useState11, 2),
      blockMap = _React$useState12[0],
      setBlockMap = _React$useState12[1];

  var _React$useState13 = React.useState(),
      _React$useState14 = _slicedToArray(_React$useState13, 2),
      output = _React$useState14[0],
      setOutput = _React$useState14[1];

  React.useEffect(function () {
    var newBlockMap = new Map();
    blocks.forEach(function (block) {
      newBlockMap.set(block.domain, block);
    });
    setBlockMap(newBlockMap);
  }, [blocks]);
  var fileRef = React.useRef();

  function error(e) {
    setOutput( /*#__PURE__*/React.createElement("div", {
      className: "error accent"
    }, e));
    throw e;
  }

  function fileUpload() {
    var reader = new FileReader();
    reader.addEventListener("load", function (e) {
      try {
        // TODO: use validatem?
        var json = JSON.parse(e.target.result);
        json.forEach(function (block) {
          console.log("block:", block);
        });
      } catch (e) {
        error(e.message);
      }
    });
    reader.readAsText(fileRef.current.files[0]);
  }

  React.useEffect(function () {
    if (fileRef && fileRef.current) {
      fileRef.current.addEventListener("change", fileUpload);
    }

    return function cleanup() {
      fileRef.current.removeEventListener("change", fileUpload);
    };
  });

  function textImport() {
    Promise["try"](function () {
      if (bulk[0] == "[") {
        // assume it's json
        return JSON.parse(bulk);
      } else {
        return bulk.split("\n").map(function (val) {
          return {
            domain: val.trim()
          };
        });
      }
    }).then(function (domains) {
      console.log(domains);
      var before = domains.length;
      setOutput("Importing ".concat(before, " domain(s)"));
      domains = domains.filter(function (_ref3) {
        var domain = _ref3.domain;
        return domain != "" && !blockMap.has(domain);
      });
      setOutput( /*#__PURE__*/React.createElement("span", null, output, /*#__PURE__*/React.createElement("br", null), "Deduplicated ".concat(before - domains.length, "/").concat(before, " with existing blocks, adding ").concat(domains.length, " block(s)")));

      if (domains.length > 0) {
        var data = new FormData();
        data.append("domains", new Blob([JSON.stringify(domains)], {
          type: "application/json"
        }), "import.json");
        return oauth.apiRequest("/api/v1/admin/domain_blocks?import=true", "POST", data, "form");
      }
    }).then(function (json) {
      console.log("bulk import result:", json);
      setBlocks(sortBlocks(deduplicateBlocks([].concat(_toConsumableArray(json), _toConsumableArray(blocks)))));
    })["catch"](function (e) {
      error(e.message);
    });
  }

  function textExport() {
    setBulk(blocks.reduce(function (str, val) {
      if (_typeof(str) == "object") {
        return str.domain;
      } else {
        return str + "\n" + val.domain;
      }
    }));
  }

  function jsonExport() {
    Promise["try"](function () {
      return oauth.apiRequest("/api/v1/admin/domain_blocks?export=true", "GET");
    }).then(function (json) {
      fileDownload(JSON.stringify(json), "block-export.json");
    })["catch"](function (e) {
      error(e);
    });
  }

  function textAreaUpdate(e) {
    setBulk(e.target.value);
  }

  return /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("h3", null, "Bulk import/export"), /*#__PURE__*/React.createElement("label", {
    htmlFor: "bulk"
  }, "Domains, one per line:"), /*#__PURE__*/React.createElement("textarea", {
    value: bulk,
    rows: 20,
    onChange: textAreaUpdate
  }), /*#__PURE__*/React.createElement("div", {
    className: "controls"
  }, /*#__PURE__*/React.createElement("button", {
    onClick: textImport
  }, "Import All From Field"), /*#__PURE__*/React.createElement("button", {
    onClick: textExport
  }, "Export To Field"), /*#__PURE__*/React.createElement("label", {
    className: "button",
    htmlFor: "upload"
  }, "Upload .json"), /*#__PURE__*/React.createElement("button", {
    onClick: jsonExport
  }, "Download .json")), output, /*#__PURE__*/React.createElement("input", {
    type: "file",
    id: "upload",
    className: "hidden",
    ref: fileRef
  }));
}

function AddBlock(_ref4) {
  var oauth = _ref4.oauth,
      blocks = _ref4.blocks,
      setBlocks = _ref4.setBlocks;

  var _React$useState15 = React.useState(""),
      _React$useState16 = _slicedToArray(_React$useState15, 2),
      domain = _React$useState16[0],
      setDomain = _React$useState16[1];

  var _React$useState17 = React.useState("suspend"),
      _React$useState18 = _slicedToArray(_React$useState17, 2),
      type = _React$useState18[0],
      setType = _React$useState18[1];

  var _React$useState19 = React.useState(false),
      _React$useState20 = _slicedToArray(_React$useState19, 2),
      obfuscated = _React$useState20[0],
      setObfuscated = _React$useState20[1];

  var _React$useState21 = React.useState(""),
      _React$useState22 = _slicedToArray(_React$useState21, 2),
      privateDescription = _React$useState22[0],
      setPrivateDescription = _React$useState22[1];

  var _React$useState23 = React.useState(""),
      _React$useState24 = _slicedToArray(_React$useState23, 2),
      publicDescription = _React$useState24[0],
      setPublicDescription = _React$useState24[1];

  function addBlock() {
    console.log("".concat(type, "ing"), domain);
    Promise["try"](function () {
      return oauth.apiRequest("/api/v1/admin/domain_blocks", "POST", {
        domain: domain,
        obfuscate: obfuscated,
        private_comment: privateDescription,
        public_comment: publicDescription
      }, "json");
    }).then(function (json) {
      setDomain("");
      setPrivateDescription("");
      setPublicDescription("");
      setBlocks([json].concat(_toConsumableArray(blocks)));
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

  return /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("h3", null, "Add Block:"), /*#__PURE__*/React.createElement("div", {
    className: "addblock"
  }, /*#__PURE__*/React.createElement("input", {
    id: "domain",
    placeholder: "instance",
    onChange: onDomainChange,
    value: domain,
    onKeyDown: onKeyDown
  }), /*#__PURE__*/React.createElement("select", {
    value: type,
    onChange: onTypeChange
  }, /*#__PURE__*/React.createElement("option", {
    id: "suspend"
  }, "Suspend"), /*#__PURE__*/React.createElement("option", {
    id: "silence"
  }, "Silence")), /*#__PURE__*/React.createElement("button", {
    onClick: addBlock
  }, "Add"), /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("label", {
    htmlFor: "private"
  }, "Private description:"), /*#__PURE__*/React.createElement("br", null), /*#__PURE__*/React.createElement("textarea", {
    id: "private",
    value: privateDescription,
    onChange: function onChange(e) {
      return setPrivateDescription(e.target.value);
    }
  })), /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("label", {
    htmlFor: "public"
  }, "Public description:"), /*#__PURE__*/React.createElement("br", null), /*#__PURE__*/React.createElement("textarea", {
    id: "public",
    value: publicDescription,
    onChange: function onChange(e) {
      return setPublicDescription(e.target.value);
    }
  })), /*#__PURE__*/React.createElement("div", {
    className: "single"
  }, /*#__PURE__*/React.createElement("label", {
    htmlFor: "obfuscate"
  }, "Obfuscate:"), /*#__PURE__*/React.createElement("input", {
    id: "obfuscate",
    type: "checkbox",
    value: obfuscated,
    onChange: function onChange(e) {
      return setObfuscated(e.target.checked);
    }
  }))));
} // function Blocklist() {
// 	return (
// 		<section className="blocklists">
// 			<h1>Blocklists</h1>
// 		</section>
// 	);
// }

},{"bluebird":15,"js-file-download":16,"react":23}],16:[function(require,module,exports){
module.exports = function(data, filename, mime, bom) {
    var blobData = (typeof bom !== 'undefined') ? [bom, data] : [data]
    var blob = new Blob(blobData, {type: mime || 'application/octet-stream'});
    if (typeof window.navigator.msSaveBlob !== 'undefined') {
        // IE workaround for "HTML7007: One or more blob URLs were
        // revoked by closing the blob for which they were created.
        // These URLs will no longer resolve as the data backing
        // the URL has been freed."
        window.navigator.msSaveBlob(blob, filename);
    }
    else {
        var blobURL = (window.URL && window.URL.createObjectURL) ? window.URL.createObjectURL(blob) : window.webkitURL.createObjectURL(blob);
        var tempLink = document.createElement('a');
        tempLink.style.display = 'none';
        tempLink.href = blobURL;
        tempLink.setAttribute('download', filename);

        // Safari thinks _blank anchor are pop ups. We only want to set _blank
        // target if the browser does not support the HTML5 download attribute.
        // This allows you to download files in desktop safari if pop up blocking
        // is enabled.
        if (typeof tempLink.download === 'undefined') {
            tempLink.setAttribute('target', '_blank');
        }

        document.body.appendChild(tempLink);
        tempLink.click();

        // Fixes "webkit blob resource error 1"
        setTimeout(function() {
            document.body.removeChild(tempLink);
            window.URL.revokeObjectURL(blobURL);
        }, 200)
    }
}

},{}],9:[function(require,module,exports){
"use strict";

function _slicedToArray(arr, i) { return _arrayWithHoles(arr) || _iterableToArrayLimit(arr, i) || _unsupportedIterableToArray(arr, i) || _nonIterableRest(); }

function _nonIterableRest() { throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method."); }

function _unsupportedIterableToArray(o, minLen) { if (!o) return; if (typeof o === "string") return _arrayLikeToArray(o, minLen); var n = Object.prototype.toString.call(o).slice(8, -1); if (n === "Object" && o.constructor) n = o.constructor.name; if (n === "Map" || n === "Set") return Array.from(o); if (n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n)) return _arrayLikeToArray(o, minLen); }

function _arrayLikeToArray(arr, len) { if (len == null || len > arr.length) len = arr.length; for (var i = 0, arr2 = new Array(len); i < len; i++) { arr2[i] = arr[i]; } return arr2; }

function _iterableToArrayLimit(arr, i) { var _i = arr == null ? null : typeof Symbol !== "undefined" && arr[Symbol.iterator] || arr["@@iterator"]; if (_i == null) return; var _arr = []; var _n = true; var _d = false; var _s, _e; try { for (_i = _i.call(arr); !(_n = (_s = _i.next()).done); _n = true) { _arr.push(_s.value); if (i && _arr.length === i) break; } } catch (err) { _d = true; _e = err; } finally { try { if (!_n && _i["return"] != null) _i["return"](); } finally { if (_d) throw _e; } } return _arr; }

function _arrayWithHoles(arr) { if (Array.isArray(arr)) return arr; }

var Promise = require("bluebird");

var React = require("react");

var oauthLib = require("./oauth");

module.exports = function Auth(_ref) {
  var setOauth = _ref.setOauth;

  var _React$useState = React.useState(""),
      _React$useState2 = _slicedToArray(_React$useState, 2),
      instance = _React$useState2[0],
      setInstance = _React$useState2[1];

  React.useEffect(function () {
    var isStillMounted = true; // check if current domain runs an instance

    var thisUrl = new URL(window.location.origin);
    thisUrl.pathname = "/api/v1/instance";
    fetch(thisUrl.href).then(function (res) {
      return res.json();
    }).then(function (json) {
      if (json && json.uri) {
        if (isStillMounted) {
          setInstance(json.uri);
        }
      }
    })["catch"](function (e) {
      console.error("caught", e); // no instance here
    });
    return function () {
      // cleanup function
      isStillMounted = false;
    };
  }, []);

  function doAuth() {
    var oauth = oauthLib({
      instance: instance,
      client_name: "GoToSocial Admin Panel",
      scope: ["admin"],
      website: window.location.href
    });
    setOauth(oauth);
    return Promise["try"](function () {
      return oauth.register();
    }).then(function () {
      return oauth.authorize();
    });
  }

  function updateInstance(e) {
    if (e.key == "Enter") {
      doAuth();
    } else {
      setInstance(e.target.value);
    }
  }

  return /*#__PURE__*/React.createElement("section", {
    className: "login"
  }, /*#__PURE__*/React.createElement("h1", null, "OAUTH Login:"), /*#__PURE__*/React.createElement("form", {
    onSubmit: function onSubmit(e) {
      return e.preventDefault();
    }
  }, /*#__PURE__*/React.createElement("label", {
    htmlFor: "instance"
  }, "Instance: "), /*#__PURE__*/React.createElement("input", {
    value: instance,
    onChange: updateInstance,
    id: "instance"
  }), /*#__PURE__*/React.createElement("button", {
    onClick: doAuth
  }, "Authenticate")));
};

},{"./oauth":12,"bluebird":15,"react":23}],12:[function(require,module,exports){
"use strict";

function _slicedToArray(arr, i) { return _arrayWithHoles(arr) || _iterableToArrayLimit(arr, i) || _unsupportedIterableToArray(arr, i) || _nonIterableRest(); }

function _nonIterableRest() { throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method."); }

function _unsupportedIterableToArray(o, minLen) { if (!o) return; if (typeof o === "string") return _arrayLikeToArray(o, minLen); var n = Object.prototype.toString.call(o).slice(8, -1); if (n === "Object" && o.constructor) n = o.constructor.name; if (n === "Map" || n === "Set") return Array.from(o); if (n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n)) return _arrayLikeToArray(o, minLen); }

function _arrayLikeToArray(arr, len) { if (len == null || len > arr.length) len = arr.length; for (var i = 0, arr2 = new Array(len); i < len; i++) { arr2[i] = arr[i]; } return arr2; }

function _iterableToArrayLimit(arr, i) { var _i = arr == null ? null : typeof Symbol !== "undefined" && arr[Symbol.iterator] || arr["@@iterator"]; if (_i == null) return; var _arr = []; var _n = true; var _d = false; var _s, _e; try { for (_i = _i.call(arr); !(_n = (_s = _i.next()).done); _n = true) { _arr.push(_s.value); if (i && _arr.length === i) break; } } catch (err) { _d = true; _e = err; } finally { try { if (!_n && _i["return"] != null) _i["return"](); } finally { if (_d) throw _e; } } return _arr; }

function _arrayWithHoles(arr) { if (Array.isArray(arr)) return arr; }

var Promise = require("bluebird");

function getCurrentUrl() {
  return window.location.origin + window.location.pathname; // strips ?query=string and #hash
}

module.exports = function oauthClient(config, initState) {
  /* config: 
  	instance: instance domain (https://testingtesting123.xyz)
  	client_name: "GoToSocial Admin Panel"
  	scope: []
  	website: 
  */
  var state = initState;

  if (initState == undefined) {
    state = localStorage.getItem("oauth");

    if (state == undefined) {
      state = {
        config: config
      };
      storeState();
    } else {
      state = JSON.parse(state);
    }
  }

  function storeState() {
    localStorage.setItem("oauth", JSON.stringify(state));
  }
  /* register app
  	/api/v1/apps
  */


  function register() {
    if (state.client_id != undefined) {
      return true; // we already have a registration
    }

    var url = new URL(config.instance);
    url.pathname = "/api/v1/apps";
    return fetch(url.href, {
      method: "POST",
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        client_name: config.client_name,
        redirect_uris: getCurrentUrl(),
        scopes: config.scope.join(" "),
        website: getCurrentUrl()
      })
    }).then(function (res) {
      if (res.status != 200) {
        throw res;
      }

      return res.json();
    }).then(function (json) {
      state.client_id = json.client_id;
      state.client_secret = json.client_secret;
      storeState();
    });
  }
  /* authorize:
  	/oauth/authorize
  		?client_id=CLIENT_ID
  		&redirect_uri=window.location.href
  		&response_type=code
  		&scope=admin
  */


  function authorize() {
    var url = new URL(config.instance);
    url.pathname = "/oauth/authorize";
    url.searchParams.set("client_id", state.client_id);
    url.searchParams.set("redirect_uri", getCurrentUrl());
    url.searchParams.set("response_type", "code");
    url.searchParams.set("scope", config.scope.join(" "));
    window.location.assign(url.href);
  }

  function callback() {
    if (state.access_token != undefined) {
      return; // we're already done :)
    }

    var params = new URL(window.location).searchParams;
    var token = params.get("code");

    if (token != null) {
      console.log("got token callback:", token);
    }

    return authorizeToken(token)["catch"](function (e) {
      console.log("Error processing oauth callback:", e);
      logout(); // just to be sure
    });
  }

  function authorizeToken(token) {
    var url = new URL(config.instance);
    url.pathname = "/oauth/token";
    return fetch(url.href, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        client_id: state.client_id,
        client_secret: state.client_secret,
        redirect_uri: getCurrentUrl(),
        grant_type: "authorization_code",
        code: token
      })
    }).then(function (res) {
      if (res.status != 200) {
        throw res;
      }

      return res.json();
    }).then(function (json) {
      state.access_token = json.access_token;
      storeState();
      window.location = getCurrentUrl(); // clear ?token=
    });
  }

  function isAuthorized() {
    return state.access_token != undefined;
  }

  function apiRequest(path, method, data) {
    var type = arguments.length > 3 && arguments[3] !== undefined ? arguments[3] : "json";

    if (!isAuthorized()) {
      throw new Error("Not Authenticated");
    }

    var url = new URL(config.instance);

    var _path$split = path.split("?"),
        _path$split2 = _slicedToArray(_path$split, 2),
        p = _path$split2[0],
        s = _path$split2[1];

    url.pathname = p;
    url.search = s;
    var headers = {
      "Authorization": "Bearer ".concat(state.access_token)
    };
    var body = data;

    if (type == "json" && body != undefined) {
      headers["Content-Type"] = "application/json";
      body = JSON.stringify(data);
    }

    return fetch(url.href, {
      method: method,
      headers: headers,
      body: body
    }).then(function (res) {
      return Promise.all([res.json(), res]);
    }).then(function (_ref) {
      var _ref2 = _slicedToArray(_ref, 2),
          json = _ref2[0],
          res = _ref2[1];

      if (res.status != 200) {
        if (json.error) {
          throw new Error(json.error);
        } else {
          throw new Error("".concat(res.status, ": ").concat(res.statusText));
        }
      } else {
        return json;
      }
    });
  }

  function logout() {
    var url = new URL(config.instance);
    url.pathname = "/oauth/revoke";
    return fetch(url.href, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        client_id: state.client_id,
        client_secret: state.client_secret,
        token: state.access_token
      })
    }).then(function (res) {
      if (res.status != 200) {
        // GoToSocial doesn't actually implement this route yet,
        // so error is to be expected
        return;
      }

      return res.json();
    })["catch"](function () {// see above
    }).then(function () {
      localStorage.removeItem("oauth");
      window.location = getCurrentUrl();
    });
  }

  return {
    register: register,
    authorize: authorize,
    callback: callback,
    isAuthorized: isAuthorized,
    apiRequest: apiRequest,
    logout: logout
  };
};

},{"bluebird":15}]},{},[11])
//# sourceMappingURL=data:application/json;charset:utf-8;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIm5vZGVfbW9kdWxlcy9mYWN0b3ItYnVuZGxlL25vZGVfbW9kdWxlcy9icm93c2VyLXBhY2svX3ByZWx1ZGUuanMiLCJwYW5lbHMvYWRtaW4vaW5kZXguanMiLCJwYW5lbHMvYWRtaW4vc3R5bGUuY3NzIiwicGFuZWxzL2FkbWluL3NldHRpbmdzLmpzIiwicGFuZWxzL2FkbWluL2Jsb2Nrcy5qcyIsInBhbmVscy9ub2RlX21vZHVsZXMvanMtZmlsZS1kb3dubG9hZC9maWxlLWRvd25sb2FkLmpzIiwicGFuZWxzL2FkbWluL2F1dGguanMiLCJwYW5lbHMvYWRtaW4vb2F1dGguanMiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUE7QUNBQTs7Ozs7Ozs7Ozs7Ozs7QUFFQSxJQUFNLE9BQU8sR0FBRyxPQUFPLENBQUMsVUFBRCxDQUF2Qjs7QUFDQSxJQUFNLEtBQUssR0FBRyxPQUFPLENBQUMsT0FBRCxDQUFyQjs7QUFDQSxJQUFNLFFBQVEsR0FBRyxPQUFPLENBQUMsV0FBRCxDQUF4Qjs7QUFFQSxJQUFNLFFBQVEsR0FBRyxPQUFPLENBQUMsWUFBRCxDQUF4Qjs7QUFDQSxJQUFNLElBQUksR0FBRyxPQUFPLENBQUMsUUFBRCxDQUFwQjs7QUFDQSxJQUFNLFFBQVEsR0FBRyxPQUFPLENBQUMsWUFBRCxDQUF4Qjs7QUFDQSxJQUFNLE1BQU0sR0FBRyxPQUFPLENBQUMsVUFBRCxDQUF0Qjs7QUFFQSxPQUFPLENBQUMsYUFBRCxDQUFQOztBQUVBLFNBQVMsR0FBVCxHQUFlO0FBQ2Qsd0JBQTBCLEtBQUssQ0FBQyxRQUFOLEVBQTFCO0FBQUE7QUFBQSxNQUFPLEtBQVA7QUFBQSxNQUFjLFFBQWQ7O0FBQ0EseUJBQTJCLEtBQUssQ0FBQyxRQUFOLENBQWUsS0FBZixDQUEzQjtBQUFBO0FBQUEsTUFBTyxPQUFQO0FBQUEsTUFBZ0IsT0FBaEI7O0FBQ0EseUJBQW9DLEtBQUssQ0FBQyxRQUFOLENBQWUsWUFBWSxDQUFDLE9BQWIsQ0FBcUIsT0FBckIsQ0FBZixDQUFw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QUUsS0FBSyxDQUFDLGFBRkQ7QUFHcEIsUUFBQSxLQUFLLEVBQUUsS0FBSyxDQUFDO0FBSE8sT0FBZjtBQUxnQixLQUFYLENBQUwsQ0FVSixJQVZJLENBVUMsVUFBQyxHQUFELEVBQVM7QUFDaEIsVUFBSSxHQUFHLENBQUMsTUFBSixJQUFjLEdBQWxCLEVBQXVCO0FBQ3RCO0FBQ0E7QUFDQTtBQUNBOztBQUNELGFBQU8sR0FBRyxDQUFDLElBQUosRUFBUDtBQUNBLEtBakJNLFdBaUJFLFlBQU0sQ0FDZDtBQUNBLEtBbkJNLEVBbUJKLElBbkJJLENBbUJDLFlBQU07QUFDYixNQUFBLFlBQVksQ0FBQyxVQUFiLENBQXdCLE9BQXhCO0FBQ0EsTUFBQSxNQUFNLENBQUMsUUFBUCxHQUFrQixhQUFhLEVBQS9CO0FBQ0EsS0F0Qk0sQ0FBUDtBQXVCQTs7QUFFRCxTQUFPO0FBQ04sSUFBQSxRQUFRLEVBQVIsUUFETTtBQUNJLElBQUEsU0FBUyxFQUFULFNBREo7QUFDZSxJQUFBLFFBQVEsRUFBUixRQURmO0FBQ3lCLElBQUEsWUFBWSxFQUFaLFlBRHpCO0FBQ3VDLElBQUEsVUFBVSxFQUFWLFVBRHZDO0FBQ21ELElBQUEsTUFBTSxFQUFOO0FBRG5ELEdBQVA7QUFHQSxDQS9MRCIsImZpbGUiOiJnZW5lcmF0ZWQuanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlc0NvbnRlbnQiOlsiKGZ1bmN0aW9uIGUodCxuLHIpe2Z1bmN0aW9uIHMobyx1KXtpZighbltvXSl7aWYoIXRbb10pe3ZhciBhPXR5cGVvZiByZXF1aXJlPT1cImZ1bmN0aW9uXCImJnJlcXVpcmU7aWYoIXUmJmEpcmV0dXJuIGEobywhMCk7aWYoaSlyZXR1cm4gaShvLCEwKTt2YXIgZj1uZXcgRXJyb3IoXCJDYW5ub3QgZmluZCBtb2R1bGUgJ1wiK28rXCInXCIpO3Rocm93IGYuY29kZT1cIk1PRFVMRV9OT1RfRk9VTkRcIixmfXZhciBsPW5bb109e2V4cG9ydHM6e319O3Rbb11bMF0uY2FsbChsLmV4cG9ydHMsZnVuY3Rpb24oZSl7dmFyIG49dFtvXVsxXVtlXTtyZXR1cm4gcyhuP246ZSl9LGwsbC5leHBvcnRzLGUsdCxuLHIpfXJldHVybiBuW29dLmV4cG9ydHN9dmFyIGk9dHlwZW9mIHJlcXVpcmU9PVwiZnVuY3Rpb25cIiYmcmVxdWlyZTtmb3IodmFyIG89MDtvPHIubGVuZ3RoO28rKylzKHJbb10pO3JldHVybiBzfSkiLCJcInVzZSBzdHJpY3RcIjtcblxuY29uc3QgUHJvbWlzZSA9IHJlcXVpcmUoXCJibHVlYmlyZFwiKTtcbmNvbnN0IFJlYWN0ID0gcmVxdWlyZShcInJlYWN0XCIpO1xuY29uc3QgUmVhY3REb20gPSByZXF1aXJlKFwicmVhY3QtZG9tXCIpO1xuXG5jb25zdCBvYXV0aExpYiA9IHJlcXVpcmUoXCIuL29hdXRoLmpzXCIpO1xuY29uc3QgQXV0aCA9IHJlcXVpcmUoXCIuL2F1dGhcIik7XG5jb25zdCBTZXR0aW5ncyA9IHJlcXVpcmUoXCIuL3NldHRpbmdzXCIpO1xuY29uc3QgQmxvY2tzID0gcmVxdWlyZShcIi4vYmxvY2tzXCIpO1xuXG5yZXF1aXJlKFwiLi9zdHlsZS5jc3NcIik7XG5cbmZ1bmN0aW9uIEFwcCgpIHtcblx0Y29uc3QgW29hdXRoLCBzZXRPYXV0aF0gPSBSZWFjdC51c2VTdGF0ZSgpO1xuXHRjb25zdCBbaGFzQXV0aCwgc2V0QXV0aF0gPSBSZWFjdC51c2VTdGF0ZShmYWxzZSk7XG5cdGNvbnN0IFtvYXV0aFN0YXRlLCBzZXRPYXV0aFN0YXRlXSA9IFJlYWN0LnVzZVN0YXRlKGxvY2FsU3RvcmFnZS5nZXRJdGVtKFwib2F1dGhcIikpO1xuXG5cdFJlYWN0LnVzZUVmZmVjdCgoKSA9PiB7XG5cdFx0bGV0IHN0YXRlID0gbG9jYWxTdG9yYWdlLmdldEl0ZW0oXCJvYXV0aFwiKTtcblx0XHRpZiAoc3RhdGUgIT0gdW5kZWZpbmVkKSB7XG5cdFx0XHRzdGF0ZSA9IEpTT04ucGFyc2Uoc3RhdGUpO1xuXHRcdFx0bGV0IHJlc3RvcmVkT2F1dGggPSBvYXV0aExpYihzdGF0ZS5jb25maWcsIHN0YXRlKTtcblx0XHRcdFByb21pc2UudHJ5KCgpID0+IHtcblx0XHRcdFx0cmV0dXJuIHJlc3RvcmVkT2F1dGguY2FsbGJhY2soKTtcblx0XHRcdH0pLnRoZW4oKCkgPT4ge1xuXHRcdFx0XHRzZXRBdXRoKHRydWUpO1xuXHRcdFx0fSk7XG5cdFx0XHRzZXRPYXV0aChyZXN0b3JlZE9hdXRoKTtcblx0XHR9XG5cdH0sIFtdKTtcblxuXHRpZiAoIWhhc0F1dGggJiYgb2F1dGggJiYgb2F1dGguaXNBdXRob3JpemVkKCkpIHtcblx0XHRzZXRBdXRoKHRydWUpO1xuXHR9XG5cblx0aWYgKG9hdXRoICYmIG9hdXRoLmlzQXV0aG9yaXplZCgpKSB7XG5cdFx0cmV0dXJuIDxBZG1pblBhbmVsIG9hdXRoPXtvYXV0aH0gLz47XG5cdH0gZWxzZSBpZiAob2F1dGhTdGF0ZSAhPSB1bmRlZmluZWQpIHtcblx0XHRyZXR1cm4gXCJwcm9jZXNzaW5nIG9hdXRoLi4uXCI7XG5cdH0gZWxzZSB7XG5cdFx0cmV0dXJuIDxBdXRoIHNldE9hdXRoPXtzZXRPYXV0aH0gLz47XG5cdH1cbn1cblxuZnVuY3Rpb24gQWRtaW5QYW5lbCh7b2F1dGh9KSB7XG5cdC8qIFxuXHRcdEZlYXR1cmVzOiAoaXNzdWUgIzc4KVxuXHRcdC0gWyBdIEluc3RhbmNlIGluZm9ybWF0aW9uIHVwZGF0aW5nXG5cdFx0XHQgIEdFVCAvYXBpL3YxL2luc3RhbmNlIFBBVENIIC9hcGkvdjEvaW5zdGFuY2Vcblx0XHQtIFsgXSBEb21haW4gYmxvY2sgY3JlYXRpb24sIHZpZXdpbmcsIGFuZCBkZWxldGlvblxuXHRcdFx0ICBHRVQgL2FwaS92MS9hZG1pbi9kb21haW5fYmxvY2tzXG5cdFx0XHQgIFBPU1QgL2FwaS92MS9hZG1pbi9kb21haW5fYmxvY2tzXG5cdFx0XHQgIEdFVCAvYXBpL3YxL2FkbWluL2RvbWFpbl9ibG9ja3MvRE9NQUlOX0JMT0NLX0lELCBERUxFVEUgL2FwaS92MS9hZG1pbi9kb21haW5fYmxvY2tzL0RPTUFJTl9CTE9DS19JRFxuXHRcdC0gWyBdIEJsb2NrbGlzdCBpbXBvcnQvZXhwb3J0XG5cdFx0XHQgIEdFVCAvYXBpL3YxL2FkbWluL2RvbWFpbl9ibG9ja3M/ZXhwb3J0PXRydWVcblx0XHRcdCAgUE9TVCBqc29uIGZpbGUgYXMgZm9ybSBmaWVsZCBkb21haW5zIHRvIC9hcGkvdjEvYWRtaW4vZG9tYWluX2Jsb2Nrc1xuXHQqL1xuXG5cdHJldHVybiAoXG5cdFx0PFJlYWN0LkZyYWdtZW50PlxuXHRcdFx0PExvZ291dCBvYXV0aD17b2F1dGh9Lz5cblx0XHRcdDxTZXR0aW5ncyBvYXV0aD17b2F1dGh9IC8+XG5cdFx0XHQ8QmxvY2tzIG9hdXRoPXtvYXV0aH0vPlxuXHRcdDwvUmVhY3QuRnJhZ21lbnQ+XG5cdCk7XG59XG5cbmZ1bmN0aW9uIExvZ291dCh7b2F1dGh9KSB7XG5cdHJldHVybiAoXG5cdFx0PGRpdj5cblx0XHRcdDxidXR0b24gb25DbGljaz17b2F1dGgubG9nb3V0fT5Mb2dvdXQ8L2J1dHRvbj5cblx0XHQ8L2Rpdj5cblx0KTtcbn1cblxuUmVhY3REb20ucmVuZGVyKDxBcHAvPiwgZG9jdW1lbnQuZ2V0RWxlbWVudEJ5SWQoXCJyb290XCIpKTsiLCJyZXF1aXJlKFwiLi4vLi4vbm9kZV9tb2R1bGVzL2ljc3NpZnkvZ2xvYmFsLWNzcy1sb2FkZXIuanNcIik7IG1vZHVsZS5leHBvcnRzID0ge307IiwiXCJ1c2Ugc3RyaWN0XCI7XG5cbmNvbnN0IFByb21pc2UgPSByZXF1aXJlKFwiYmx1ZWJpcmRcIik7XG5jb25zdCBSZWFjdCA9IHJlcXVpcmUoXCJyZWFjdFwiKTtcblxubW9kdWxlLmV4cG9ydHMgPSBmdW5jdGlvbiBTZXR0aW5ncyh7b2F1dGh9KSB7XG5cdGNvbnN0IFtpbmZvLCBzZXRJbmZvXSA9IFJlYWN0LnVzZVN0YXRlKHt9KTtcblx0Y29uc3QgW2Vycm9yTXNnLCBzZXRFcnJvcl0gPSBSZWFjdC51c2VTdGF0ZShcIlwiKTtcblx0Y29uc3QgW3N0YXR1c01zZywgc2V0U3RhdHVzXSA9IFJlYWN0LnVzZVN0YXRlKFwiRmV0Y2hpbmcgaW5zdGFuY2UgaW5mb1wiKTtcblxuXHRSZWFjdC51c2VFZmZlY3QoKCkgPT4ge1xuXHRcdFByb21pc2UudHJ5KCgpID0+IHtcblx0XHRcdHJldHVybiBvYXV0aC5hcGlSZXF1ZXN0KFwiL2FwaS92MS9pbnN0YW5jZVwiLCBcIkdFVFwiKTtcblx0XHR9KS50aGVuKChqc29uKSA9PiB7XG5cdFx0XHRzZXRJbmZvKGpzb24pO1xuXHRcdH0pLmNhdGNoKChlKSA9PiB7XG5cdFx0XHRzZXRFcnJvcihlLm1lc3NhZ2UpO1xuXHRcdFx0c2V0U3RhdHVzKFwiXCIpO1xuXHRcdH0pO1xuXHR9LCBbXSk7XG5cblx0ZnVuY3Rpb24gc3VibWl0KCkge1xuXHRcdHNldFN0YXR1cyhcIlBBVENIaW5nXCIpO1xuXHRcdHNldEVycm9yKFwiXCIpO1xuXHRcdHJldHVybiBQcm9taXNlLnRyeSgoKSA9PiB7XG5cdFx0XHRsZXQgZm9ybURhdGFJbmZvID0gbmV3IEZvcm1EYXRhKCk7XG5cdFx0XHRPYmplY3QuZW50cmllcyhpbmZvKS5mb3JFYWNoKChba2V5LCB2YWxdKSA9PiB7XG5cdFx0XHRcdGlmIChrZXkgPT0gXCJjb250YWN0X2FjY291bnRcIikge1xuXHRcdFx0XHRcdGtleSA9IFwiY29udGFjdF91c2VybmFtZVwiO1xuXHRcdFx0XHRcdHZhbCA9IHZhbC51c2VybmFtZTtcblx0XHRcdFx0fVxuXHRcdFx0XHRpZiAoa2V5ID09IFwiZW1haWxcIikge1xuXHRcdFx0XHRcdGtleSA9IFwiY29udGFjdF9lbWFpbFwiO1xuXHRcdFx0XHR9XG5cdFx0XHRcdGlmICh0eXBlb2YgdmFsICE9IFwib2JqZWN0XCIpIHtcblx0XHRcdFx0XHRmb3JtRGF0YUluZm8uYXBwZW5kKGtleSwgdmFsKTtcblx0XHRcdFx0fVxuXHRcdFx0fSk7XG5cdFx0XHRyZXR1cm4gb2F1dGguYXBpUmVxdWVzdChcIi9hcGkvdjEvaW5zdGFuY2VcIiwgXCJQQVRDSFwiLCBmb3JtRGF0YUluZm8sIFwiZm9ybVwiKTtcblx0XHR9KS50aGVuKChqc29uKSA9PiB7XG5cdFx0XHRzZXRTdGF0dXMoXCJDb25maWcgc2F2ZWRcIik7XG5cdFx0XHRjb25zb2xlLmxvZyhqc29uKTtcblx0XHR9KS5jYXRjaCgoZSkgPT4ge1xuXHRcdFx0c2V0RXJyb3IoZS5tZXNzYWdlKTtcblx0XHRcdHNldFN0YXR1cyhcIlwiKTtcblx0XHR9KTtcblx0fVxuXG5cdHJldHVybiAoXG5cdFx0PHNlY3Rpb24gY2xhc3NOYW1lPVwiaW5mbyBsb2dpblwiPlxuXHRcdFx0PGgxPkluc3RhbmNlIEluZm9ybWF0aW9uIDxidXR0b24gb25DbGljaz17c3VibWl0fT5TYXZlPC9idXR0b24+PC9oMT5cblx0XHRcdDxkaXYgY2xhc3NOYW1lPVwiZXJyb3IgYWNjZW50XCI+XG5cdFx0XHRcdHtlcnJvck1zZ31cblx0XHRcdDwvZGl2PlxuXHRcdFx0PGRpdj5cblx0XHRcdFx0e3N0YXR1c01zZ31cblx0XHRcdDwvZGl2PlxuXHRcdFx0PGZvcm0gb25TdWJtaXQ9eyhlKSA9PiBlLnByZXZlbnREZWZhdWx0KCl9PlxuXHRcdFx0XHR7ZWRpdGFibGVPYmplY3QoaW5mbyl9XG5cdFx0XHQ8L2Zvcm0+XG5cdFx0PC9zZWN0aW9uPlxuXHQpO1xufTtcblxuZnVuY3Rpb24gZWRpdGFibGVPYmplY3Qob2JqLCBwYXRoPVtdKSB7XG5cdGNvbnN0IHJlYWRPbmx5S2V5cyA9IFtcInVyaVwiLCBcInZlcnNpb25cIiwgXCJ1cmxzX3N0cmVhbWluZ19hcGlcIiwgXCJzdGF0c1wiXTtcblx0Y29uc3QgaGlkZGVuS2V5cyA9IFtcImNvbnRhY3RfYWNjb3VudF9cIiwgXCJ1cmxzXCJdO1xuXHRjb25zdCBleHBsaWNpdFNob3duS2V5cyA9IFtcImNvbnRhY3RfYWNjb3VudF91c2VybmFtZVwiXTtcblx0Y29uc3QgaW1wbGVtZW50ZWRLZXlzID0gXCJ0aXRsZSwgY29udGFjdF9hY2NvdW50X3VzZXJuYW1lLCBlbWFpbCwgc2hvcnRfZGVzY3JpcHRpb24sIGRlc2NyaXB0aW9uLCB0ZXJtcywgYXZhdGFyLCBoZWFkZXJcIi5zcGxpdChcIiwgXCIpO1xuXG5cdGxldCBsaXN0aW5nID0gT2JqZWN0LmVudHJpZXMob2JqKS5tYXAoKFtrZXksIHZhbF0pID0+IHtcblx0XHRsZXQgZnVsbGtleSA9IFsuLi5wYXRoLCBrZXldLmpvaW4oXCJfXCIpO1xuXG5cdFx0aWYgKFxuXHRcdFx0aGlkZGVuS2V5cy5pbmNsdWRlcyhmdWxsa2V5KSB8fFxuXHRcdFx0aGlkZGVuS2V5cy5pbmNsdWRlcyhwYXRoLmpvaW4oXCJfXCIpK1wiX1wiKSAvLyBhbHNvIG1hdGNoIGp1c3QgcGFyZW50IHBhdGhcblx0XHQpIHtcblx0XHRcdGlmICghZXhwbGljaXRTaG93bktleXMuaW5jbHVkZXMoZnVsbGtleSkpIHtcblx0XHRcdFx0cmV0dXJuIG51bGw7XG5cdFx0XHR9XG5cdFx0fVxuXG5cdFx0aWYgKEFycmF5LmlzQXJyYXkodmFsKSkge1xuXHRcdFx0Ly8gRklYTUU6IGhhbmRsZSB0aGlzXG5cdFx0fSBlbHNlIGlmICh0eXBlb2YgdmFsID09IFwib2JqZWN0XCIpIHtcblx0XHRcdHJldHVybiAoPFJlYWN0LkZyYWdtZW50IGtleT17ZnVsbGtleX0+XG5cdFx0XHRcdHtlZGl0YWJsZU9iamVjdCh2YWwsIFsuLi5wYXRoLCBrZXldKX1cblx0XHRcdDwvUmVhY3QuRnJhZ21lbnQ+KTtcblx0XHR9IFxuXG5cdFx0bGV0IGlzSW1wbGVtZW50ZWQgPSBcIlwiO1xuXHRcdGlmICghaW1wbGVtZW50ZWRLZXlzLmluY2x1ZGVzKGZ1bGxrZXkpKSB7XG5cdFx0XHRpc0ltcGxlbWVudGVkID0gXCIgbm90SW1wbGVtZW50ZWRcIjtcblx0XHR9XG5cblx0XHRsZXQgaXNSZWFkT25seSA9IChcblx0XHRcdHJlYWRPbmx5S2V5cy5pbmNsdWRlcyhmdWxsa2V5KSB8fFxuXHRcdFx0cmVhZE9ubHlLZXlzLmluY2x1ZGVzKHBhdGguam9pbihcIl9cIikpIHx8XG5cdFx0XHRpc0ltcGxlbWVudGVkICE9IFwiXCJcblx0XHQpO1xuXG5cdFx0bGV0IGxhYmVsID0ga2V5LnJlcGxhY2UoL18vZywgXCIgXCIpO1xuXHRcdGlmIChwYXRoLmxlbmd0aCA+IDApIHtcblx0XHRcdGxhYmVsID0gYFxcdTAwQTBgLnJlcGVhdCg0ICogcGF0aC5sZW5ndGgpICsgbGFiZWw7XG5cdFx0fVxuXG5cdFx0bGV0IGlucHV0UHJvcHM7XG5cdFx0bGV0IGNoYW5nZUZ1bmM7XG5cdFx0aWYgKHZhbCA9PT0gdHJ1ZSB8fCB2YWwgPT09IGZhbHNlKSB7XG5cdFx0XHRpbnB1dFByb3BzID0ge1xuXHRcdFx0XHR0eXBlOiBcImNoZWNrYm94XCIsXG5cdFx0XHRcdGRlZmF1bHRDaGVja2VkOiB2YWwsXG5cdFx0XHRcdGRpc2FibGVkOiBpc1JlYWRPbmx5XG5cdFx0XHR9O1xuXHRcdFx0Y2hhbmdlRnVuYyA9IChlKSA9PiBlLnRhcmdldC5jaGVja2VkO1xuXHRcdH0gZWxzZSBpZiAodmFsLmxlbmd0aCAhPSAwICYmICFpc05hTih2YWwpKSB7XG5cdFx0XHRpbnB1dFByb3BzID0ge1xuXHRcdFx0XHR0eXBlOiBcIm51bWJlclwiLFxuXHRcdFx0XHRkZWZhdWx0VmFsdWU6IHZhbCxcblx0XHRcdFx0cmVhZE9ubHk6IGlzUmVhZE9ubHlcblx0XHRcdH07XG5cdFx0XHRjaGFuZ2VGdW5jID0gKGUpID0+IGUudGFyZ2V0LnZhbHVlO1xuXHRcdH0gZWxzZSB7XG5cdFx0XHRpbnB1dFByb3BzID0ge1xuXHRcdFx0XHR0eXBlOiBcInRleHRcIixcblx0XHRcdFx0ZGVmYXVsdFZhbHVlOiB2YWwsXG5cdFx0XHRcdHJlYWRPbmx5OiBpc1JlYWRPbmx5XG5cdFx0XHR9O1xuXHRcdFx0Y2hhbmdlRnVuYyA9IChlKSA9PiBlLnRhcmdldC52YWx1ZTtcblx0XHR9XG5cblx0XHRmdW5jdGlvbiBzZXRSZWYoZWxlbWVudCkge1xuXHRcdFx0aWYgKGVsZW1lbnQgIT0gbnVsbCkge1xuXHRcdFx0XHRlbGVtZW50LmFkZEV2ZW50TGlzdGVuZXIoXCJjaGFuZ2VcIiwgKGUpID0+IHtcblx0XHRcdFx0XHRvYmpba2V5XSA9IGNoYW5nZUZ1bmMoZSk7XG5cdFx0XHRcdH0pO1xuXHRcdFx0fVxuXHRcdH1cblxuXHRcdHJldHVybiAoXG5cdFx0XHQ8UmVhY3QuRnJhZ21lbnQga2V5PXtmdWxsa2V5fT5cblx0XHRcdFx0PGxhYmVsIGh0bWxGb3I9e2tleX0gY2xhc3NOYW1lPVwiY2FwaXRhbGl6ZVwiPntsYWJlbH08L2xhYmVsPlxuXHRcdFx0XHQ8ZGl2IGNsYXNzTmFtZT17aXNJbXBsZW1lbnRlZH0+XG5cdFx0XHRcdFx0PGlucHV0IGNsYXNzTmFtZT17aXNJbXBsZW1lbnRlZH0gcmVmPXtzZXRSZWZ9IHsuLi5pbnB1dFByb3BzfSAvPlxuXHRcdFx0XHQ8L2Rpdj5cblx0XHRcdDwvUmVhY3QuRnJhZ21lbnQ+XG5cdFx0KTtcblx0fSk7XG5cdHJldHVybiAoXG5cdFx0PFJlYWN0LkZyYWdtZW50PlxuXHRcdFx0e3BhdGggIT0gXCJcIiAmJlxuXHRcdFx0XHQ8PjxiPntwYXRofTo8L2I+IDxzcGFuIGlkPVwiZmlsbGVyXCI+PC9zcGFuPjwvPlxuXHRcdFx0fVxuXHRcdFx0e2xpc3Rpbmd9XG5cdFx0PC9SZWFjdC5GcmFnbWVudD5cblx0KTtcbn0iLCJcInVzZSBzdHJpY3RcIjtcblxuY29uc3QgUHJvbWlzZSA9IHJlcXVpcmUoXCJibHVlYmlyZFwiKTtcbmNvbnN0IFJlYWN0ID0gcmVxdWlyZShcInJlYWN0XCIpO1xuY29uc3QgZmlsZURvd25sb2FkID0gcmVxdWlyZShcImpzLWZpbGUtZG93bmxvYWRcIik7XG5cbmZ1bmN0aW9uIHNvcnRCbG9ja3MoYmxvY2tzKSB7XG5cdHJldHVybiBibG9ja3Muc29ydCgoYSwgYikgPT4geyAvLyBhbHBoYWJldGljYWwgc29ydFxuXHRcdHJldHVybiBhLmRvbWFpbi5sb2NhbGVDb21wYXJlKGIuZG9tYWluKTtcblx0fSk7XG59XG5cbmZ1bmN0aW9uIGRlZHVwbGljYXRlQmxvY2tzKGJsb2Nrcykge1xuXHRsZXQgYSA9IG5ldyBNYXAoKTtcblx0YmxvY2tzLmZvckVhY2goKGJsb2NrKSA9PiB7XG5cdFx0YS5zZXQoYmxvY2suaWQsIGJsb2NrKTtcblx0fSk7XG5cdHJldHVybiBBcnJheS5mcm9tKGEudmFsdWVzKCkpO1xufVxuXG5tb2R1bGUuZXhwb3J0cyA9IGZ1bmN0aW9uIEJsb2Nrcyh7b2F1dGh9KSB7XG5cdGNvbnN0IFtibG9ja3MsIHNldEJsb2Nrc10gPSBSZWFjdC51c2VTdGF0ZShbXSk7XG5cdGNvbnN0IFtpbmZvLCBzZXRJbmZvXSA9IFJlYWN0LnVzZVN0YXRlKFwiRmV0Y2hpbmcgYmxvY2tzXCIpO1xuXHRjb25zdCBbZXJyb3JNc2csIHNldEVycm9yXSA9IFJlYWN0LnVzZVN0YXRlKFwiXCIpO1xuXHRjb25zdCBbY2hlY2tlZCwgc2V0Q2hlY2tlZF0gPSBSZWFjdC51c2VTdGF0ZShuZXcgU2V0KCkpO1xuXG5cdFJlYWN0LnVzZUVmZmVjdCgoKSA9PiB7XG5cdFx0UHJvbWlzZS50cnkoKCkgPT4ge1xuXHRcdFx0cmV0dXJuIG9hdXRoLmFwaVJlcXVlc3QoXCIvYXBpL3YxL2FkbWluL2RvbWFpbl9ibG9ja3NcIiwgdW5kZWZpbmVkLCB1bmRlZmluZWQsIFwiR0VUXCIpO1xuXHRcdH0pLnRoZW4oKGpzb24pID0+IHtcblx0XHRcdHNldEluZm8oXCJcIik7XG5cdFx0XHRzZXRFcnJvcihcIlwiKTtcblx0XHRcdHNldEJsb2Nrcyhzb3J0QmxvY2tzKGpzb24pKTtcblx0XHR9KS5jYXRjaCgoZSkgPT4ge1xuXHRcdFx0c2V0RXJyb3IoZS5tZXNzYWdlKTtcblx0XHRcdHNldEluZm8oXCJcIik7XG5cdFx0fSk7XG5cdH0sIFtdKTtcblxuXHRsZXQgYmxvY2tMaXN0ID0gYmxvY2tzLm1hcCgoYmxvY2spID0+IHtcblx0XHRmdW5jdGlvbiB1cGRhdGUoZSkge1xuXHRcdFx0bGV0IG5ld0NoZWNrZWQgPSBuZXcgU2V0KGNoZWNrZWQudmFsdWVzKCkpO1xuXHRcdFx0aWYgKGUudGFyZ2V0LmNoZWNrZWQpIHtcblx0XHRcdFx0bmV3Q2hlY2tlZC5hZGQoYmxvY2suaWQpO1xuXHRcdFx0fSBlbHNlIHtcblx0XHRcdFx0bmV3Q2hlY2tlZC5kZWxldGUoYmxvY2suaWQpO1xuXHRcdFx0fVxuXHRcdFx0c2V0Q2hlY2tlZChuZXdDaGVja2VkKTtcblx0XHR9XG5cblx0XHRyZXR1cm4gKFxuXHRcdFx0PFJlYWN0LkZyYWdtZW50IGtleT17YmxvY2suaWR9PlxuXHRcdFx0XHQ8ZGl2PjxpbnB1dCB0eXBlPVwiY2hlY2tib3hcIiBvbkNoYW5nZT17dXBkYXRlfSBjaGVja2VkPXtjaGVja2VkLmhhcyhibG9jay5pZCl9PjwvaW5wdXQ+PC9kaXY+XG5cdFx0XHRcdDxkaXY+e2Jsb2NrLmRvbWFpbn08L2Rpdj5cblx0XHRcdFx0PGRpdj57KG5ldyBEYXRlKGJsb2NrLmNyZWF0ZWRfYXQpKS50b0xvY2FsZVN0cmluZygpfTwvZGl2PlxuXHRcdFx0PC9SZWFjdC5GcmFnbWVudD5cblx0XHQpO1xuXHR9KTtcblxuXHRmdW5jdGlvbiBjbGVhckNoZWNrZWQoKSB7XG5cdFx0c2V0Q2hlY2tlZChuZXcgU2V0KCkpO1xuXHR9XG5cblx0ZnVuY3Rpb24gdW5kb0NoZWNrZWQoKSB7XG5cdFx0bGV0IGFtb3VudCA9IGNoZWNrZWQuc2l6ZTtcblx0XHRpZihjb25maXJtKGBBcmUgeW91IHN1cmUgeW91IHdhbnQgdG8gcmVtb3ZlICR7YW1vdW50fSBibG9jayhzKT9gKSkge1xuXHRcdFx0c2V0SW5mbyhcIlwiKTtcblx0XHRcdFByb21pc2UubWFwKEFycmF5LmZyb20oY2hlY2tlZC52YWx1ZXMoKSksIChibG9jaykgPT4ge1xuXHRcdFx0XHRjb25zb2xlLmxvZyhcImRlbGV0aW5nXCIsIGJsb2NrKTtcblx0XHRcdFx0cmV0dXJuIG9hdXRoLmFwaVJlcXVlc3QoYC9hcGkvdjEvYWRtaW4vZG9tYWluX2Jsb2Nrcy8ke2Jsb2NrfWAsIFwiREVMRVRFXCIpO1xuXHRcdFx0fSkudGhlbigocmVzKSA9PiB7XG5cdFx0XHRcdGNvbnNvbGUubG9nKHJlcyk7XG5cdFx0XHRcdHNldEluZm8oYERlbGV0ZWQgJHthbW91bnR9IGJsb2NrczogJHtyZXMubWFwKChhKSA9PiBhLmRvbWFpbikuam9pbihcIiwgXCIpfWApO1xuXHRcdFx0fSkuY2F0Y2goKGUpID0+IHtcblx0XHRcdFx0c2V0RXJyb3IoZSk7XG5cdFx0XHR9KTtcblxuXHRcdFx0bGV0IG5ld0Jsb2NrcyA9IGJsb2Nrcy5maWx0ZXIoKGJsb2NrKSA9PiB7XG5cdFx0XHRcdGlmIChjaGVja2VkLnNpemUgPiAwICYmIGNoZWNrZWQuaGFzKGJsb2NrLmlkKSkge1xuXHRcdFx0XHRcdGNoZWNrZWQuZGVsZXRlKGJsb2NrLmlkKTtcblx0XHRcdFx0XHRyZXR1cm4gZmFsc2U7XG5cdFx0XHRcdH0gZWxzZSB7XG5cdFx0XHRcdFx0cmV0dXJuIHRydWU7XG5cdFx0XHRcdH1cblx0XHRcdH0pO1xuXHRcdFx0c2V0QmxvY2tzKG5ld0Jsb2Nrcyk7XG5cdFx0XHRjbGVhckNoZWNrZWQoKTtcblx0XHR9XG5cdH1cblxuXHRyZXR1cm4gKFxuXHRcdDxzZWN0aW9uIGNsYXNzTmFtZT1cImJsb2Nrc1wiPlxuXHRcdFx0PGgxPkJsb2NrczwvaDE+XG5cdFx0XHQ8ZGl2IGNsYXNzTmFtZT1cImVycm9yIGFjY2VudFwiPntlcnJvck1zZ308L2Rpdj5cblx0XHRcdDxkaXY+e2luZm99PC9kaXY+XG5cdFx0XHQ8QWRkQmxvY2sgb2F1dGg9e29hdXRofSBibG9ja3M9e2Jsb2Nrc30gc2V0QmxvY2tzPXtzZXRCbG9ja3N9IC8+XG5cdFx0XHQ8aDM+QmxvY2tzOjwvaDM+XG5cdFx0XHQ8ZGl2IHN0eWxlPXt7ZGlzcGxheTogXCJncmlkXCIsIGdyaWRUZW1wbGF0ZUNvbHVtbnM6IFwiMWZyIGF1dG9cIn19PlxuXHRcdFx0XHQ8c3BhbiBvbkNsaWNrPXtjbGVhckNoZWNrZWR9IGNsYXNzTmFtZT1cImFjY2VudFwiIHN0eWxlPXt7YWxpZ25TZWxmOiBcImVuZFwifX0+dW5jaGVjayBhbGw8L3NwYW4+XG5cdFx0XHRcdDxidXR0b24gb25DbGljaz17dW5kb0NoZWNrZWR9PlVuYmxvY2sgc2VsZWN0ZWQ8L2J1dHRvbj5cblx0XHRcdDwvZGl2PlxuXHRcdFx0PGRpdiBjbGFzc05hbWU9XCJibG9ja2xpc3Qgb3ZlcmZsb3dcIj5cblx0XHRcdFx0e2Jsb2NrTGlzdH1cblx0XHRcdDwvZGl2PlxuXHRcdFx0PEJ1bGtCbG9ja2luZyBvYXV0aD17b2F1dGh9IGJsb2Nrcz17YmxvY2tzfSBzZXRCbG9ja3M9e3NldEJsb2Nrc30vPlxuXHRcdDwvc2VjdGlvbj5cblx0KTtcbn07XG5cbmZ1bmN0aW9uIEJ1bGtCbG9ja2luZyh7b2F1dGgsIGJsb2Nrcywgc2V0QmxvY2tzfSkge1xuXHRjb25zdCBbYnVsaywgc2V0QnVsa10gPSBSZWFjdC51c2VTdGF0ZShcIlwiKTtcblx0Y29uc3QgW2Jsb2NrTWFwLCBzZXRCbG9ja01hcF0gPSBSZWFjdC51c2VTdGF0ZShuZXcgTWFwKCkpO1xuXHRjb25zdCBbb3V0cHV0LCBzZXRPdXRwdXRdID0gUmVhY3QudXNlU3RhdGUoKTtcblxuXHRSZWFjdC51c2VFZmZlY3QoKCkgPT4ge1xuXHRcdGxldCBuZXdCbG9ja01hcCA9IG5ldyBNYXAoKTtcblx0XHRibG9ja3MuZm9yRWFjaCgoYmxvY2spID0+IHtcblx0XHRcdG5ld0Jsb2NrTWFwLnNldChibG9jay5kb21haW4sIGJsb2NrKTtcblx0XHR9KTtcblx0XHRzZXRCbG9ja01hcChuZXdCbG9ja01hcCk7XG5cdH0sIFtibG9ja3NdKTtcblxuXHRjb25zdCBmaWxlUmVmID0gUmVhY3QudXNlUmVmKCk7XG5cblx0ZnVuY3Rpb24gZXJyb3IoZSkge1xuXHRcdHNldE91dHB1dCg8ZGl2IGNsYXNzTmFtZT1cImVycm9yIGFjY2VudFwiPntlfTwvZGl2Pik7XG5cdFx0dGhyb3cgZTtcblx0fVxuXG5cdGZ1bmN0aW9uIGZpbGVVcGxvYWQoKSB7XG5cdFx0bGV0IHJlYWRlciA9IG5ldyBGaWxlUmVhZGVyKCk7XG5cdFx0cmVhZGVyLmFkZEV2ZW50TGlzdGVuZXIoXCJsb2FkXCIsIChlKSA9PiB7XG5cdFx0XHR0cnkge1xuXHRcdFx0XHQvLyBUT0RPOiB1c2UgdmFsaWRhdGVtP1xuXHRcdFx0XHRsZXQganNvbiA9IEpTT04ucGFyc2UoZS50YXJnZXQucmVzdWx0KTtcblx0XHRcdFx0anNvbi5mb3JFYWNoKChibG9jaykgPT4ge1xuXHRcdFx0XHRcdGNvbnNvbGUubG9nKFwiYmxvY2s6XCIsIGJsb2NrKTtcblx0XHRcdFx0fSk7XG5cdFx0XHR9IGNhdGNoKGUpIHtcblx0XHRcdFx0ZXJyb3IoZS5tZXNzYWdlKTtcblx0XHRcdH1cblx0XHR9KTtcblx0XHRyZWFkZXIucmVhZEFzVGV4dChmaWxlUmVmLmN1cnJlbnQuZmlsZXNbMF0pO1xuXHR9XG5cblx0UmVhY3QudXNlRWZmZWN0KCgpID0+IHtcblx0XHRpZiAoZmlsZVJlZiAmJiBmaWxlUmVmLmN1cnJlbnQpIHtcblx0XHRcdGZpbGVSZWYuY3VycmVudC5hZGRFdmVudExpc3RlbmVyKFwiY2hhbmdlXCIsIGZpbGVVcGxvYWQpO1xuXHRcdH1cblx0XHRyZXR1cm4gZnVuY3Rpb24gY2xlYW51cCgpIHtcblx0XHRcdGZpbGVSZWYuY3VycmVudC5yZW1vdmVFdmVudExpc3RlbmVyKFwiY2hhbmdlXCIsIGZpbGVVcGxvYWQpO1xuXHRcdH07XG5cdH0pO1xuXG5cdGZ1bmN0aW9uIHRleHRJbXBvcnQoKSB7XG5cdFx0UHJvbWlzZS50cnkoKCkgPT4ge1xuXHRcdFx0aWYgKGJ1bGtbMF0gPT0gXCJbXCIpIHtcblx0XHRcdFx0Ly8gYXNzdW1lIGl0J3MganNvblxuXHRcdFx0XHRyZXR1cm4gSlNPTi5wYXJzZShidWxrKTtcblx0XHRcdH0gZWxzZSB7XG5cdFx0XHRcdHJldHVybiBidWxrLnNwbGl0KFwiXFxuXCIpLm1hcCgodmFsKSA9PiB7XG5cdFx0XHRcdFx0cmV0dXJuIHtcblx0XHRcdFx0XHRcdGRvbWFpbjogdmFsLnRyaW0oKVxuXHRcdFx0XHRcdH07XG5cdFx0XHRcdH0pO1xuXHRcdFx0fVxuXHRcdH0pLnRoZW4oKGRvbWFpbnMpID0+IHtcblx0XHRcdGNvbnNvbGUubG9nKGRvbWFpbnMpO1xuXHRcdFx0bGV0IGJlZm9yZSA9IGRvbWFpbnMubGVuZ3RoO1xuXHRcdFx0c2V0T3V0cHV0KGBJbXBvcnRpbmcgJHtiZWZvcmV9IGRvbWFpbihzKWApO1xuXHRcdFx0ZG9tYWlucyA9IGRvbWFpbnMuZmlsdGVyKCh7ZG9tYWlufSkgPT4ge1xuXHRcdFx0XHRyZXR1cm4gKGRvbWFpbiAhPSBcIlwiICYmICFibG9ja01hcC5oYXMoZG9tYWluKSk7XG5cdFx0XHR9KTtcblx0XHRcdHNldE91dHB1dCg8c3Bhbj57b3V0cHV0fTxici8+e2BEZWR1cGxpY2F0ZWQgJHtiZWZvcmUgLSBkb21haW5zLmxlbmd0aH0vJHtiZWZvcmV9IHdpdGggZXhpc3RpbmcgYmxvY2tzLCBhZGRpbmcgJHtkb21haW5zLmxlbmd0aH0gYmxvY2socylgfTwvc3Bhbj4pO1xuXHRcdFx0aWYgKGRvbWFpbnMubGVuZ3RoID4gMCkge1xuXHRcdFx0XHRsZXQgZGF0YSA9IG5ldyBGb3JtRGF0YSgpO1xuXHRcdFx0XHRkYXRhLmFwcGVuZChcImRvbWFpbnNcIiwgbmV3IEJsb2IoW0pTT04uc3RyaW5naWZ5KGRvbWFpbnMpXSwge3R5cGU6IFwiYXBwbGljYXRpb24vanNvblwifSksIFwiaW1wb3J0Lmpzb25cIik7XG5cdFx0XHRcdHJldHVybiBvYXV0aC5hcGlSZXF1ZXN0KFwiL2FwaS92MS9hZG1pbi9kb21haW5fYmxvY2tzP2ltcG9ydD10cnVlXCIsIFwiUE9TVFwiLCBkYXRhLCBcImZvcm1cIik7XG5cdFx0XHR9XG5cdFx0fSkudGhlbigoanNvbikgPT4ge1xuXHRcdFx0Y29uc29sZS5sb2coXCJidWxrIGltcG9ydCByZXN1bHQ6XCIsIGpzb24pO1xuXHRcdFx0c2V0QmxvY2tzKHNvcnRCbG9ja3MoZGVkdXBsaWNhdGVCbG9ja3MoWy4uLmpzb24sIC4uLmJsb2Nrc10pKSk7XG5cdFx0fSkuY2F0Y2goKGUpID0+IHtcblx0XHRcdGVycm9yKGUubWVzc2FnZSk7XG5cdFx0fSk7XG5cdH1cblxuXHRmdW5jdGlvbiB0ZXh0RXhwb3J0KCkge1xuXHRcdHNldEJ1bGsoYmxvY2tzLnJlZHVjZSgoc3RyLCB2YWwpID0+IHtcblx0XHRcdGlmICh0eXBlb2Ygc3RyID09IFwib2JqZWN0XCIpIHtcblx0XHRcdFx0cmV0dXJuIHN0ci5kb21haW47XG5cdFx0XHR9IGVsc2Uge1xuXHRcdFx0XHRyZXR1cm4gc3RyICsgXCJcXG5cIiArIHZhbC5kb21haW47XG5cdFx0XHR9XG5cdFx0fSkpO1xuXHR9XG5cblx0ZnVuY3Rpb24ganNvbkV4cG9ydCgpIHtcblx0XHRQcm9taXNlLnRyeSgoKSA9PiB7XG5cdFx0XHRyZXR1cm4gb2F1dGguYXBpUmVxdWVzdChcIi9hcGkvdjEvYWRtaW4vZG9tYWluX2Jsb2Nrcz9leHBvcnQ9dHJ1ZVwiLCBcIkdFVFwiKTtcblx0XHR9KS50aGVuKChqc29uKSA9PiB7XG5cdFx0XHRmaWxlRG93bmxvYWQoSlNPTi5zdHJpbmdpZnkoanNvbiksIFwiYmxvY2stZXhwb3J0Lmpzb25cIik7XG5cdFx0fSkuY2F0Y2goKGUpID0+IHtcblx0XHRcdGVycm9yKGUpO1xuXHRcdH0pO1xuXHR9XG5cblx0ZnVuY3Rpb24gdGV4dEFyZWFVcGRhdGUoZSkge1xuXHRcdHNldEJ1bGsoZS50YXJnZXQudmFsdWUpO1xuXHR9XG5cblx0cmV0dXJuIChcblx0XHQ8UmVhY3QuRnJhZ21lbnQ+XG5cdFx0XHQ8aDM+QnVsayBpbXBvcnQvZXhwb3J0PC9oMz5cblx0XHRcdDxsYWJlbCBodG1sRm9yPVwiYnVsa1wiPkRvbWFpbnMsIG9uZSBwZXIgbGluZTo8L2xhYmVsPlxuXHRcdFx0PHRleHRhcmVhIHZhbHVlPXtidWxrfSByb3dzPXsyMH0gb25DaGFuZ2U9e3RleHRBcmVhVXBkYXRlfT48L3RleHRhcmVhPlxuXHRcdFx0PGRpdiBjbGFzc05hbWU9XCJjb250cm9sc1wiPlxuXHRcdFx0XHQ8YnV0dG9uIG9uQ2xpY2s9e3RleHRJbXBvcnR9PkltcG9ydCBBbGwgRnJvbSBGaWVsZDwvYnV0dG9uPlxuXHRcdFx0XHQ8YnV0dG9uIG9uQ2xpY2s9e3RleHRFeHBvcnR9PkV4cG9ydCBUbyBGaWVsZDwvYnV0dG9uPlxuXHRcdFx0XHQ8bGFiZWwgY2xhc3NOYW1lPVwiYnV0dG9uXCIgaHRtbEZvcj1cInVwbG9hZFwiPlVwbG9hZCAuanNvbjwvbGFiZWw+XG5cdFx0XHRcdDxidXR0b24gb25DbGljaz17anNvbkV4cG9ydH0+RG93bmxvYWQgLmpzb248L2J1dHRvbj5cblx0XHRcdDwvZGl2PlxuXHRcdFx0e291dHB1dH1cblx0XHRcdDxpbnB1dCB0eXBlPVwiZmlsZVwiIGlkPVwidXBsb2FkXCIgY2xhc3NOYW1lPVwiaGlkZGVuXCIgcmVmPXtmaWxlUmVmfT48L2lucHV0PlxuXHRcdDwvUmVhY3QuRnJhZ21lbnQ+XG5cdCk7XG59XG5cbmZ1bmN0aW9uIEFkZEJsb2NrKHtvYXV0aCwgYmxvY2tzLCBzZXRCbG9ja3N9KSB7XG5cdGNvbnN0IFtkb21haW4sIHNldERvbWFpbl0gPSBSZWFjdC51c2VTdGF0ZShcIlwiKTtcblx0Y29uc3QgW3R5cGUsIHNldFR5cGVdID0gUmVhY3QudXNlU3RhdGUoXCJzdXNwZW5kXCIpO1xuXHRjb25zdCBbb2JmdXNjYXRlZCwgc2V0T2JmdXNjYXRlZF0gPSBSZWFjdC51c2VTdGF0ZShmYWxzZSk7XG5cdGNvbnN0IFtwcml2YXRlRGVzY3JpcHRpb24sIHNldFByaXZhdGVEZXNjcmlwdGlvbl0gPSBSZWFjdC51c2VTdGF0ZShcIlwiKTtcblx0Y29uc3QgW3B1YmxpY0Rlc2NyaXB0aW9uLCBzZXRQdWJsaWNEZXNjcmlwdGlvbl0gPSBSZWFjdC51c2VTdGF0ZShcIlwiKTtcblxuXHRmdW5jdGlvbiBhZGRCbG9jaygpIHtcblx0XHRjb25zb2xlLmxvZyhgJHt0eXBlfWluZ2AsIGRvbWFpbik7XG5cdFx0UHJvbWlzZS50cnkoKCkgPT4ge1xuXHRcdFx0cmV0dXJuIG9hdXRoLmFwaVJlcXVlc3QoXCIvYXBpL3YxL2FkbWluL2RvbWFpbl9ibG9ja3NcIiwgXCJQT1NUXCIsIHtcblx0XHRcdFx0ZG9tYWluOiBkb21haW4sXG5cdFx0XHRcdG9iZnVzY2F0ZTogb2JmdXNjYXRlZCxcblx0XHRcdFx0cHJpdmF0ZV9jb21tZW50OiBwcml2YXRlRGVzY3JpcHRpb24sXG5cdFx0XHRcdHB1YmxpY19jb21tZW50OiBwdWJsaWNEZXNjcmlwdGlvblxuXHRcdFx0fSwgXCJqc29uXCIpO1xuXHRcdH0pLnRoZW4oKGpzb24pID0+IHtcblx0XHRcdHNldERvbWFpbihcIlwiKTtcblx0XHRcdHNldFByaXZhdGVEZXNjcmlwdGlvbihcIlwiKTtcblx0XHRcdHNldFB1YmxpY0Rlc2NyaXB0aW9uKFwiXCIpO1xuXHRcdFx0c2V0QmxvY2tzKFtqc29uLCAuLi5ibG9ja3NdKTtcblx0XHR9KTtcblx0fVxuXG5cdGZ1bmN0aW9uIG9uRG9tYWluQ2hhbmdlKGUpIHtcblx0XHRzZXREb21haW4oZS50YXJnZXQudmFsdWUpO1xuXHR9XG5cblx0ZnVuY3Rpb24gb25UeXBlQ2hhbmdlKGUpIHtcblx0XHRzZXRUeXBlKGUudGFyZ2V0LnZhbHVlKTtcblx0fVxuXG5cdGZ1bmN0aW9uIG9uS2V5RG93bihlKSB7XG5cdFx0aWYgKGUua2V5ID09IFwiRW50ZXJcIikge1xuXHRcdFx0YWRkQmxvY2soKTtcblx0XHR9XG5cdH1cblxuXHRyZXR1cm4gKFxuXHRcdDxSZWFjdC5GcmFnbWVudD5cblx0XHRcdDxoMz5BZGQgQmxvY2s6PC9oMz5cblx0XHRcdDxkaXYgY2xhc3NOYW1lPVwiYWRkYmxvY2tcIj5cblx0XHRcdFx0PGlucHV0IGlkPVwiZG9tYWluXCIgcGxhY2Vob2xkZXI9XCJpbnN0YW5jZVwiIG9uQ2hhbmdlPXtvbkRvbWFpbkNoYW5nZX0gdmFsdWU9e2RvbWFpbn0gb25LZXlEb3duPXtvbktleURvd259IC8+XG5cdFx0XHRcdDxzZWxlY3QgdmFsdWU9e3R5cGV9IG9uQ2hhbmdlPXtvblR5cGVDaGFuZ2V9PlxuXHRcdFx0XHRcdDxvcHRpb24gaWQ9XCJzdXNwZW5kXCI+U3VzcGVuZDwvb3B0aW9uPlxuXHRcdFx0XHRcdDxvcHRpb24gaWQ9XCJzaWxlbmNlXCI+U2lsZW5jZTwvb3B0aW9uPlxuXHRcdFx0XHQ8L3NlbGVjdD5cblx0XHRcdFx0PGJ1dHRvbiBvbkNsaWNrPXthZGRCbG9ja30+QWRkPC9idXR0b24+XG5cdFx0XHRcdDxkaXY+XG5cdFx0XHRcdFx0PGxhYmVsIGh0bWxGb3I9XCJwcml2YXRlXCI+UHJpdmF0ZSBkZXNjcmlwdGlvbjo8L2xhYmVsPjxici8+XG5cdFx0XHRcdFx0PHRleHRhcmVhIGlkPVwicHJpdmF0ZVwiIHZhbHVlPXtwcml2YXRlRGVzY3JpcHRpb259IG9uQ2hhbmdlPXsoZSkgPT4gc2V0UHJpdmF0ZURlc2NyaXB0aW9uKGUudGFyZ2V0LnZhbHVlKX0+PC90ZXh0YXJlYT5cblx0XHRcdFx0PC9kaXY+XG5cdFx0XHRcdDxkaXY+XG5cdFx0XHRcdFx0PGxhYmVsIGh0bWxGb3I9XCJwdWJsaWNcIj5QdWJsaWMgZGVzY3JpcHRpb246PC9sYWJlbD48YnIvPlxuXHRcdFx0XHRcdDx0ZXh0YXJlYSBpZD1cInB1YmxpY1wiIHZhbHVlPXtwdWJsaWNEZXNjcmlwdGlvbn0gb25DaGFuZ2U9eyhlKSA9PiBzZXRQdWJsaWNEZXNjcmlwdGlvbihlLnRhcmdldC52YWx1ZSl9PjwvdGV4dGFyZWE+XG5cdFx0XHRcdDwvZGl2PlxuXHRcdFx0XHQ8ZGl2IGNsYXNzTmFtZT1cInNpbmdsZVwiPlxuXHRcdFx0XHRcdDxsYWJlbCBodG1sRm9yPVwib2JmdXNjYXRlXCI+T2JmdXNjYXRlOjwvbGFiZWw+XG5cdFx0XHRcdFx0PGlucHV0IGlkPVwib2JmdXNjYXRlXCIgdHlwZT1cImNoZWNrYm94XCIgdmFsdWU9e29iZnVzY2F0ZWR9IG9uQ2hhbmdlPXsoZSkgPT4gc2V0T2JmdXNjYXRlZChlLnRhcmdldC5jaGVja2VkKX0vPlxuXHRcdFx0XHQ8L2Rpdj5cblx0XHRcdDwvZGl2PlxuXHRcdDwvUmVhY3QuRnJhZ21lbnQ+XG5cdCk7XG59XG5cbi8vIGZ1bmN0aW9uIEJsb2NrbGlzdCgpIHtcbi8vIFx0cmV0dXJuIChcbi8vIFx0XHQ8c2VjdGlvbiBjbGFzc05hbWU9XCJibG9ja2xpc3RzXCI+XG4vLyBcdFx0XHQ8aDE+QmxvY2tsaXN0czwvaDE+XG4vLyBcdFx0PC9zZWN0aW9uPlxuLy8gXHQpO1xuLy8gfSIsIm1vZHVsZS5leHBvcnRzID0gZnVuY3Rpb24oZGF0YSwgZmlsZW5hbWUsIG1pbWUsIGJvbSkge1xuICAgIHZhciBibG9iRGF0YSA9ICh0eXBlb2YgYm9tICE9PSAndW5kZWZpbmVkJykgPyBbYm9tLCBkYXRhXSA6IFtkYXRhXVxuICAgIHZhciBibG9iID0gbmV3IEJsb2IoYmxvYkRhdGEsIHt0eXBlOiBtaW1lIHx8ICdhcHBsaWNhdGlvbi9vY3RldC1zdHJlYW0nfSk7XG4gICAgaWYgKHR5cGVvZiB3aW5kb3cubmF2aWdhdG9yLm1zU2F2ZUJsb2IgIT09ICd1bmRlZmluZWQnKSB7XG4gICAgICAgIC8vIElFIHdvcmthcm91bmQgZm9yIFwiSFRNTDcwMDc6IE9uZSBvciBtb3JlIGJsb2IgVVJMcyB3ZXJlXG4gICAgICAgIC8vIHJldm9rZWQgYnkgY2xvc2luZyB0aGUgYmxvYiBmb3Igd2hpY2ggdGhleSB3ZXJlIGNyZWF0ZWQuXG4gICAgICAgIC8vIFRoZXNlIFVSTHMgd2lsbCBubyBsb25nZXIgcmVzb2x2ZSBhcyB0aGUgZGF0YSBiYWNraW5nXG4gICAgICAgIC8vIHRoZSBVUkwgaGFzIGJlZW4gZnJlZWQuXCJcbiAgICAgICAgd2luZG93Lm5hdmlnYXRvci5tc1NhdmVCbG9iKGJsb2IsIGZpbGVuYW1lKTtcbiAgICB9XG4gICAgZWxzZSB7XG4gICAgICAgIHZhciBibG9iVVJMID0gKHdpbmRvdy5VUkwgJiYgd2luZG93LlVSTC5jcmVhdGVPYmplY3RVUkwpID8gd2luZG93LlVSTC5jcmVhdGVPYmplY3RVUkwoYmxvYikgOiB3aW5kb3cud2Via2l0VVJMLmNyZWF0ZU9iamVjdFVSTChibG9iKTtcbiAgICAgICAgdmFyIHRlbXBMaW5rID0gZG9jdW1lbnQuY3JlYXRlRWxlbWVudCgnYScpO1xuICAgICAgICB0ZW1wTGluay5zdHlsZS5kaXNwbGF5ID0gJ25vbmUnO1xuICAgICAgICB0ZW1wTGluay5ocmVmID0gYmxvYlVSTDtcbiAgICAgICAgdGVtcExpbmsuc2V0QXR0cmlidXRlKCdkb3dubG9hZCcsIGZpbGVuYW1lKTtcblxuICAgICAgICAvLyBTYWZhcmkgdGhpbmtzIF9ibGFuayBhbmNob3IgYXJlIHBvcCB1cHMuIFdlIG9ubHkgd2FudCB0byBzZXQgX2JsYW5rXG4gICAgICAgIC8vIHRhcmdldCBpZiB0aGUgYnJvd3NlciBkb2VzIG5vdCBzdXBwb3J0IHRoZSBIVE1MNSBkb3dubG9hZCBhdHRyaWJ1dGUuXG4gICAgICAgIC8vIFRoaXMgYWxsb3dzIHlvdSB0byBkb3dubG9hZCBmaWxlcyBpbiBkZXNrdG9wIHNhZmFyaSBpZiBwb3AgdXAgYmxvY2tpbmdcbiAgICAgICAgLy8gaXMgZW5hYmxlZC5cbiAgICAgICAgaWYgKHR5cGVvZiB0ZW1wTGluay5kb3dubG9hZCA9PT0gJ3VuZGVmaW5lZCcpIHtcbiAgICAgICAgICAgIHRlbXBMaW5rLnNldEF0dHJpYnV0ZSgndGFyZ2V0JywgJ19ibGFuaycpO1xuICAgICAgICB9XG5cbiAgICAgICAgZG9jdW1lbnQuYm9keS5hcHBlbmRDaGlsZCh0ZW1wTGluayk7XG4gICAgICAgIHRlbXBMaW5rLmNsaWNrKCk7XG5cbiAgICAgICAgLy8gRml4ZXMgXCJ3ZWJraXQgYmxvYiByZXNvdXJjZSBlcnJvciAxXCJcbiAgICAgICAgc2V0VGltZW91dChmdW5jdGlvbigpIHtcbiAgICAgICAgICAgIGRvY3VtZW50LmJvZHkucmVtb3ZlQ2hpbGQodGVtcExpbmspO1xuICAgICAgICAgICAgd2luZG93LlVSTC5yZXZva2VPYmplY3RVUkwoYmxvYlVSTCk7XG4gICAgICAgIH0sIDIwMClcbiAgICB9XG59XG4iLCJcInVzZSBzdHJpY3RcIjtcblxuY29uc3QgUHJvbWlzZSA9IHJlcXVpcmUoXCJibHVlYmlyZFwiKTtcbmNvbnN0IFJlYWN0ID0gcmVxdWlyZShcInJlYWN0XCIpO1xuY29uc3Qgb2F1dGhMaWIgPSByZXF1aXJlKFwiLi9vYXV0aFwiKTtcblxubW9kdWxlLmV4cG9ydHMgPSBmdW5jdGlvbiBBdXRoKHtzZXRPYXV0aH0pIHtcblx0Y29uc3QgWyBpbnN0YW5jZSwgc2V0SW5zdGFuY2UgXSA9IFJlYWN0LnVzZVN0YXRlKFwiXCIpO1xuXG5cdFJlYWN0LnVzZUVmZmVjdCgoKSA9PiB7XG5cdFx0bGV0IGlzU3RpbGxNb3VudGVkID0gdHJ1ZTtcblx0XHQvLyBjaGVjayBpZiBjdXJyZW50IGRvbWFpbiBydW5zIGFuIGluc3RhbmNlXG5cdFx0bGV0IHRoaXNVcmwgPSBuZXcgVVJMKHdpbmRvdy5sb2NhdGlvbi5vcmlnaW4pO1xuXHRcdHRoaXNVcmwucGF0aG5hbWUgPSBcIi9hcGkvdjEvaW5zdGFuY2VcIjtcblx0XHRmZXRjaCh0aGlzVXJsLmhyZWYpXG5cdFx0XHQudGhlbigocmVzKSA9PiByZXMuanNvbigpKVxuXHRcdFx0LnRoZW4oKGpzb24pID0+IHtcblx0XHRcdFx0aWYgKGpzb24gJiYganNvbi51cmkpIHtcblx0XHRcdFx0XHRpZiAoaXNTdGlsbE1vdW50ZWQpIHtcblx0XHRcdFx0XHRcdHNldEluc3RhbmNlKGpzb24udXJpKTtcblx0XHRcdFx0XHR9XG5cdFx0XHRcdH1cblx0XHRcdH0pXG5cdFx0XHQuY2F0Y2goKGUpID0+IHtcblx0XHRcdFx0Y29uc29sZS5lcnJvcihcImNhdWdodFwiLCBlKTtcblx0XHRcdFx0Ly8gbm8gaW5zdGFuY2UgaGVyZVxuXHRcdFx0fSk7XG5cdFx0cmV0dXJuICgpID0+IHtcblx0XHRcdC8vIGNsZWFudXAgZnVuY3Rpb25cblx0XHRcdGlzU3RpbGxNb3VudGVkID0gZmFsc2U7XG5cdFx0fTtcblx0fSwgW10pO1xuXG5cdGZ1bmN0aW9uIGRvQXV0aCgpIHtcblx0XHRsZXQgb2F1dGggPSBvYXV0aExpYih7XG5cdFx0XHRpbnN0YW5jZTogaW5zdGFuY2UsXG5cdFx0XHRjbGllbnRfbmFtZTogXCJHb1RvU29jaWFsIEFkbWluIFBhbmVsXCIsXG5cdFx0XHRzY29wZTogW1wiYWRtaW5cIl0sXG5cdFx0XHR3ZWJzaXRlOiB3aW5kb3cubG9jYXRpb24uaHJlZlxuXHRcdH0pO1xuXHRcdHNldE9hdXRoKG9hdXRoKTtcblxuXHRcdHJldHVybiBQcm9taXNlLnRyeSgoKSA9PiB7XG5cdFx0XHRyZXR1cm4gb2F1dGgucmVnaXN0ZXIoKTtcblx0XHR9KS50aGVuKCgpID0+IHtcblx0XHRcdHJldHVybiBvYXV0aC5hdXRob3JpemUoKTtcblx0XHR9KTtcblx0fVxuXG5cdGZ1bmN0aW9uIHVwZGF0ZUluc3RhbmNlKGUpIHtcblx0XHRpZiAoZS5rZXkgPT0gXCJFbnRlclwiKSB7XG5cdFx0XHRkb0F1dGgoKTtcblx0XHR9IGVsc2Uge1xuXHRcdFx0c2V0SW5zdGFuY2UoZS50YXJnZXQudmFsdWUpO1xuXHRcdH1cblx0fVxuXG5cdHJldHVybiAoXG5cdFx0PHNlY3Rpb24gY2xhc3NOYW1lPVwibG9naW5cIj5cblx0XHRcdDxoMT5PQVVUSCBMb2dpbjo8L2gxPlxuXHRcdFx0PGZvcm0gb25TdWJtaXQ9eyhlKSA9PiBlLnByZXZlbnREZWZhdWx0KCl9PlxuXHRcdFx0XHQ8bGFiZWwgaHRtbEZvcj1cImluc3RhbmNlXCI+SW5zdGFuY2U6IDwvbGFiZWw+XG5cdFx0XHRcdDxpbnB1dCB2YWx1ZT17aW5zdGFuY2V9IG9uQ2hhbmdlPXt1cGRhdGVJbnN0YW5jZX0gaWQ9XCJpbnN0YW5jZVwiLz5cblx0XHRcdFx0PGJ1dHRvbiBvbkNsaWNrPXtkb0F1dGh9PkF1dGhlbnRpY2F0ZTwvYnV0dG9uPlxuXHRcdFx0PC9mb3JtPlxuXHRcdDwvc2VjdGlvbj5cblx0KTtcbn07IiwiXCJ1c2Ugc3RyaWN0XCI7XG5cbmNvbnN0IFByb21pc2UgPSByZXF1aXJlKFwiYmx1ZWJpcmRcIik7XG5cbmZ1bmN0aW9uIGdldEN1cnJlbnRVcmwoKSB7XG5cdHJldHVybiB3aW5kb3cubG9jYXRpb24ub3JpZ2luICsgd2luZG93LmxvY2F0aW9uLnBhdGhuYW1lOyAvLyBzdHJpcHMgP3F1ZXJ5PXN0cmluZyBhbmQgI2hhc2hcbn1cblxubW9kdWxlLmV4cG9ydHMgPSBmdW5jdGlvbiBvYXV0aENsaWVudChjb25maWcsIGluaXRTdGF0ZSkge1xuXHQvKiBjb25maWc6IFxuXHRcdGluc3RhbmNlOiBpbnN0YW5jZSBkb21haW4gKGh0dHBzOi8vdGVzdGluZ3Rlc3RpbmcxMjMueHl6KVxuXHRcdGNsaWVudF9uYW1lOiBcIkdvVG9Tb2NpYWwgQWRtaW4gUGFuZWxcIlxuXHRcdHNjb3BlOiBbXVxuXHRcdHdlYnNpdGU6IFxuXHQqL1xuXG5cdGxldCBzdGF0ZSA9IGluaXRTdGF0ZTtcblx0aWYgKGluaXRTdGF0ZSA9PSB1bmRlZmluZWQpIHtcblx0XHRzdGF0ZSA9IGxvY2FsU3RvcmFnZS5nZXRJdGVtKFwib2F1dGhcIik7XG5cdFx0aWYgKHN0YXRlID09IHVuZGVmaW5lZCkge1xuXHRcdFx0c3RhdGUgPSB7XG5cdFx0XHRcdGNvbmZpZ1xuXHRcdFx0fTtcblx0XHRcdHN0b3JlU3RhdGUoKTtcblx0XHR9IGVsc2Uge1xuXHRcdFx0c3RhdGUgPSBKU09OLnBhcnNlKHN0YXRlKTtcblx0XHR9XG5cdH1cblxuXHRmdW5jdGlvbiBzdG9yZVN0YXRlKCkge1xuXHRcdGxvY2FsU3RvcmFnZS5zZXRJdGVtKFwib2F1dGhcIiwgSlNPTi5zdHJpbmdpZnkoc3RhdGUpKTtcblx0fVxuXG5cdC8qIHJlZ2lzdGVyIGFwcFxuXHRcdC9hcGkvdjEvYXBwc1xuXHQqL1xuXHRmdW5jdGlvbiByZWdpc3RlcigpIHtcblx0XHRpZiAoc3RhdGUuY2xpZW50X2lkICE9IHVuZGVmaW5lZCkge1xuXHRcdFx0cmV0dXJuIHRydWU7IC8vIHdlIGFscmVhZHkgaGF2ZSBhIHJlZ2lzdHJhdGlvblxuXHRcdH1cblx0XHRsZXQgdXJsID0gbmV3IFVSTChjb25maWcuaW5zdGFuY2UpO1xuXHRcdHVybC5wYXRobmFtZSA9IFwiL2FwaS92MS9hcHBzXCI7XG5cblx0XHRyZXR1cm4gZmV0Y2godXJsLmhyZWYsIHtcblx0XHRcdG1ldGhvZDogXCJQT1NUXCIsXG5cdFx0XHRoZWFkZXJzOiB7XG5cdFx0XHRcdCdDb250ZW50LVR5cGUnOiAnYXBwbGljYXRpb24vanNvbidcblx0XHRcdH0sXG5cdFx0XHRib2R5OiBKU09OLnN0cmluZ2lmeSh7XG5cdFx0XHRcdGNsaWVudF9uYW1lOiBjb25maWcuY2xpZW50X25hbWUsXG5cdFx0XHRcdHJlZGlyZWN0X3VyaXM6IGdldEN1cnJlbnRVcmwoKSxcblx0XHRcdFx0c2NvcGVzOiBjb25maWcuc2NvcGUuam9pbihcIiBcIiksXG5cdFx0XHRcdHdlYnNpdGU6IGdldEN1cnJlbnRVcmwoKVxuXHRcdFx0fSlcblx0XHR9KS50aGVuKChyZXMpID0+IHtcblx0XHRcdGlmIChyZXMuc3RhdHVzICE9IDIwMCkge1xuXHRcdFx0XHR0aHJvdyByZXM7XG5cdFx0XHR9XG5cdFx0XHRyZXR1cm4gcmVzLmpzb24oKTtcblx0XHR9KS50aGVuKChqc29uKSA9PiB7XG5cdFx0XHRzdGF0ZS5jbGllbnRfaWQgPSBqc29uLmNsaWVudF9pZDtcblx0XHRcdHN0YXRlLmNsaWVudF9zZWNyZXQgPSBqc29uLmNsaWVudF9zZWNyZXQ7XG5cdFx0XHRzdG9yZVN0YXRlKCk7XG5cdFx0fSk7XG5cdH1cblx0XG5cdC8qIGF1dGhvcml6ZTpcblx0XHQvb2F1dGgvYXV0aG9yaXplXG5cdFx0XHQ/Y2xpZW50X2lkPUNMSUVOVF9JRFxuXHRcdFx0JnJlZGlyZWN0X3VyaT13aW5kb3cubG9jYXRpb24uaHJlZlxuXHRcdFx0JnJlc3BvbnNlX3R5cGU9Y29kZVxuXHRcdFx0JnNjb3BlPWFkbWluXG5cdCovXG5cdGZ1bmN0aW9uIGF1dGhvcml6ZSgpIHtcblx0XHRsZXQgdXJsID0gbmV3IFVSTChjb25maWcuaW5zdGFuY2UpO1xuXHRcdHVybC5wYXRobmFtZSA9IFwiL29hdXRoL2F1dGhvcml6ZVwiO1xuXHRcdHVybC5zZWFyY2hQYXJhbXMuc2V0KFwiY2xpZW50X2lkXCIsIHN0YXRlLmNsaWVudF9pZCk7XG5cdFx0dXJsLnNlYXJjaFBhcmFtcy5zZXQoXCJyZWRpcmVjdF91cmlcIiwgZ2V0Q3VycmVudFVybCgpKTtcblx0XHR1cmwuc2VhcmNoUGFyYW1zLnNldChcInJlc3BvbnNlX3R5cGVcIiwgXCJjb2RlXCIpO1xuXHRcdHVybC5zZWFyY2hQYXJhbXMuc2V0KFwic2NvcGVcIiwgY29uZmlnLnNjb3BlLmpvaW4oXCIgXCIpKTtcblxuXHRcdHdpbmRvdy5sb2NhdGlvbi5hc3NpZ24odXJsLmhyZWYpO1xuXHR9XG5cdFxuXHRmdW5jdGlvbiBjYWxsYmFjaygpIHtcblx0XHRpZiAoc3RhdGUuYWNjZXNzX3Rva2VuICE9IHVuZGVmaW5lZCkge1xuXHRcdFx0cmV0dXJuOyAvLyB3ZSdyZSBhbHJlYWR5IGRvbmUgOilcblx0XHR9XG5cdFx0bGV0IHBhcmFtcyA9IChuZXcgVVJMKHdpbmRvdy5sb2NhdGlvbikpLnNlYXJjaFBhcmFtcztcblx0XG5cdFx0bGV0IHRva2VuID0gcGFyYW1zLmdldChcImNvZGVcIik7XG5cdFx0aWYgKHRva2VuICE9IG51bGwpIHtcblx0XHRcdGNvbnNvbGUubG9nKFwiZ290IHRva2VuIGNhbGxiYWNrOlwiLCB0b2tlbik7XG5cdFx0fVxuXG5cdFx0cmV0dXJuIGF1dGhvcml6ZVRva2VuKHRva2VuKVxuXHRcdFx0LmNhdGNoKChlKSA9PiB7XG5cdFx0XHRcdGNvbnNvbGUubG9nKFwiRXJyb3IgcHJvY2Vzc2luZyBvYXV0aCBjYWxsYmFjazpcIiwgZSk7XG5cdFx0XHRcdGxvZ291dCgpOyAvLyBqdXN0IHRvIGJlIHN1cmVcblx0XHRcdH0pO1xuXHR9XG5cblx0ZnVuY3Rpb24gYXV0aG9yaXplVG9rZW4odG9rZW4pIHtcblx0XHRsZXQgdXJsID0gbmV3IFVSTChjb25maWcuaW5zdGFuY2UpO1xuXHRcdHVybC5wYXRobmFtZSA9IFwiL29hdXRoL3Rva2VuXCI7XG5cdFx0cmV0dXJuIGZldGNoKHVybC5ocmVmLCB7XG5cdFx0XHRtZXRob2Q6IFwiUE9TVFwiLFxuXHRcdFx0aGVhZGVyczoge1xuXHRcdFx0XHRcIkNvbnRlbnQtVHlwZVwiOiBcImFwcGxpY2F0aW9uL2pzb25cIlxuXHRcdFx0fSxcblx0XHRcdGJvZHk6IEpTT04uc3RyaW5naWZ5KHtcblx0XHRcdFx0Y2xpZW50X2lkOiBzdGF0ZS5jbGllbnRfaWQsXG5cdFx0XHRcdGNsaWVudF9zZWNyZXQ6IHN0YXRlLmNsaWVudF9zZWNyZXQsXG5cdFx0XHRcdHJlZGlyZWN0X3VyaTogZ2V0Q3VycmVudFVybCgpLFxuXHRcdFx0XHRncmFudF90eXBlOiBcImF1dGhvcml6YXRpb25fY29kZVwiLFxuXHRcdFx0XHRjb2RlOiB0b2tlblxuXHRcdFx0fSlcblx0XHR9KS50aGVuKChyZXMpID0+IHtcblx0XHRcdGlmIChyZXMuc3RhdHVzICE9IDIwMCkge1xuXHRcdFx0XHR0aHJvdyByZXM7XG5cdFx0XHR9XG5cdFx0XHRyZXR1cm4gcmVzLmpzb24oKTtcblx0XHR9KS50aGVuKChqc29uKSA9PiB7XG5cdFx0XHRzdGF0ZS5hY2Nlc3NfdG9rZW4gPSBqc29uLmFjY2Vzc190b2tlbjtcblx0XHRcdHN0b3JlU3RhdGUoKTtcblx0XHRcdHdpbmRvdy5sb2NhdGlvbiA9IGdldEN1cnJlbnRVcmwoKTsgLy8gY2xlYXIgP3Rva2VuPVxuXHRcdH0pO1xuXHR9XG5cblx0ZnVuY3Rpb24gaXNBdXRob3JpemVkKCkge1xuXHRcdHJldHVybiAoc3RhdGUuYWNjZXNzX3Rva2VuICE9IHVuZGVmaW5lZCk7XG5cdH1cblxuXHRmdW5jdGlvbiBhcGlSZXF1ZXN0KHBhdGgsIG1ldGhvZCwgZGF0YSwgdHlwZT1cImpzb25cIikge1xuXHRcdGlmICghaXNBdXRob3JpemVkKCkpIHtcblx0XHRcdHRocm93IG5ldyBFcnJvcihcIk5vdCBBdXRoZW50aWNhdGVkXCIpO1xuXHRcdH1cblx0XHRsZXQgdXJsID0gbmV3IFVSTChjb25maWcuaW5zdGFuY2UpO1xuXHRcdGxldCBbcCwgc10gPSBwYXRoLnNwbGl0KFwiP1wiKTtcblx0XHR1cmwucGF0aG5hbWUgPSBwO1xuXHRcdHVybC5zZWFyY2ggPSBzO1xuXHRcdGxldCBoZWFkZXJzID0ge1xuXHRcdFx0XCJBdXRob3JpemF0aW9uXCI6IGBCZWFyZXIgJHtzdGF0ZS5hY2Nlc3NfdG9rZW59YFxuXHRcdH07XG5cdFx0bGV0IGJvZHkgPSBkYXRhO1xuXHRcdGlmICh0eXBlID09IFwianNvblwiICYmIGJvZHkgIT0gdW5kZWZpbmVkKSB7XG5cdFx0XHRoZWFkZXJzW1wiQ29udGVudC1UeXBlXCJdID0gXCJhcHBsaWNhdGlvbi9qc29uXCI7XG5cdFx0XHRib2R5ID0gSlNPTi5zdHJpbmdpZnkoZGF0YSk7XG5cdFx0fVxuXHRcdHJldHVybiBmZXRjaCh1cmwuaHJlZiwge1xuXHRcdFx0bWV0aG9kLFxuXHRcdFx0aGVhZGVycyxcblx0XHRcdGJvZHlcblx0XHR9KS50aGVuKChyZXMpID0+IHtcblx0XHRcdHJldHVybiBQcm9taXNlLmFsbChbcmVzLmpzb24oKSwgcmVzXSk7XG5cdFx0fSkudGhlbigoW2pzb24sIHJlc10pID0+IHtcblx0XHRcdGlmIChyZXMuc3RhdHVzICE9IDIwMCkge1xuXHRcdFx0XHRpZiAoanNvbi5lcnJvcikge1xuXHRcdFx0XHRcdHRocm93IG5ldyBFcnJvcihqc29uLmVycm9yKTtcblx0XHRcdFx0fSBlbHNlIHtcblx0XHRcdFx0XHR0aHJvdyBuZXcgRXJyb3IoYCR7cmVzLnN0YXR1c306ICR7cmVzLnN0YXR1c1RleHR9YCk7XG5cdFx0XHRcdH1cblx0XHRcdH0gZWxzZSB7XG5cdFx0XHRcdHJldHVybiBqc29uO1xuXHRcdFx0fVxuXHRcdH0pO1xuXHR9XG5cblx0ZnVuY3Rpb24gbG9nb3V0KCkge1xuXHRcdGxldCB1cmwgPSBuZXcgVVJMKGNvbmZpZy5pbnN0YW5jZSk7XG5cdFx0dXJsLnBhdGhuYW1lID0gXCIvb2F1dGgvcmV2b2tlXCI7XG5cdFx0cmV0dXJuIGZldGNoKHVybC5ocmVmLCB7XG5cdFx0XHRtZXRob2Q6IFwiUE9TVFwiLFxuXHRcdFx0aGVhZGVyczoge1xuXHRcdFx0XHRcIkNvbnRlbnQtVHlwZVwiOiBcImFwcGxpY2F0aW9uL2pzb25cIlxuXHRcdFx0fSxcblx0XHRcdGJvZHk6IEpTT04uc3RyaW5naWZ5KHtcblx0XHRcdFx0Y2xpZW50X2lkOiBzdGF0ZS5jbGllbnRfaWQsXG5cdFx0XHRcdGNsaWVudF9zZWNyZXQ6IHN0YXRlLmNsaWVudF9zZWNyZXQsXG5cdFx0XHRcdHRva2VuOiBzdGF0ZS5hY2Nlc3NfdG9rZW4sXG5cdFx0XHR9KVxuXHRcdH0pLnRoZW4oKHJlcykgPT4ge1xuXHRcdFx0aWYgKHJlcy5zdGF0dXMgIT0gMjAwKSB7XG5cdFx0XHRcdC8vIEdvVG9Tb2NpYWwgZG9lc24ndCBhY3R1YWxseSBpbXBsZW1lbnQgdGhpcyByb3V0ZSB5ZXQsXG5cdFx0XHRcdC8vIHNvIGVycm9yIGlzIHRvIGJlIGV4cGVjdGVkXG5cdFx0XHRcdHJldHVybjtcblx0XHRcdH1cblx0XHRcdHJldHVybiByZXMuanNvbigpO1xuXHRcdH0pLmNhdGNoKCgpID0+IHtcblx0XHRcdC8vIHNlZSBhYm92ZVxuXHRcdH0pLnRoZW4oKCkgPT4ge1xuXHRcdFx0bG9jYWxTdG9yYWdlLnJlbW92ZUl0ZW0oXCJvYXV0aFwiKTtcblx0XHRcdHdpbmRvdy5sb2NhdGlvbiA9IGdldEN1cnJlbnRVcmwoKTtcblx0XHR9KTtcblx0fVxuXG5cdHJldHVybiB7XG5cdFx0cmVnaXN0ZXIsIGF1dGhvcml6ZSwgY2FsbGJhY2ssIGlzQXV0aG9yaXplZCwgYXBpUmVxdWVzdCwgbG9nb3V0XG5cdH07XG59O1xuIl19
