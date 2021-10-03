(function(f){if(typeof exports==="object"&&typeof module!=="undefined"){module.exports=f()}else if(typeof define==="function"&&define.amd){define([],f)}else{var g;if(typeof window!=="undefined"){g=window}else if(typeof global!=="undefined"){g=global}else if(typeof self!=="undefined"){g=self}else{g=this}g.reactGo = f()}})(function(){var define,module,exports;return (function(){function r(e,n,t){function o(i,f){if(!n[i]){if(!e[i]){var c="function"==typeof require&&require;if(!f&&c)return c(i,!0);if(u)return u(i,!0);var a=new Error("Cannot find module '"+i+"'");throw a.code="MODULE_NOT_FOUND",a}var p=n[i]={exports:{}};e[i][0].call(p.exports,function(r){var n=e[i][1][r];return o(n||r)},p,p.exports,r,e,n,t)}return n[i].exports}for(var u="function"==typeof require&&require,i=0;i<t.length;i++)o(t[i]);return o}return r})()({1:[function(require,module,exports){
"use strict";

function TemplateIndex(props) {
  return /*#__PURE__*/React.createElement(Page, props, /*#__PURE__*/React.createElement("main", {
    className: "lightgray"
  }, /*#__PURE__*/React.createElement("section", null, /*#__PURE__*/React.createElement("h1", null, "Home to ", /*#__PURE__*/React.createElement("span", {
    className: "count"
  }, instance.Stats.user_count), " users who posted ", /*#__PURE__*/React.createElement("span", {
    className: "count"
  }, instance.Stats.status_count), " statuses, federating with  ", /*#__PURE__*/React.createElement("span", {
    className: "count"
  }, instance.Stats.domain_count), " other instances."), /*#__PURE__*/React.createElement("div", {
    className: "short-description",
    __dangerouslySetInnerHTML: {
      __html: instance.ShortDescription
    }
  })), /*#__PURE__*/React.createElement("section", {
    className: "apps"
  }, /*#__PURE__*/React.createElement("p", null, "GoToSocial does not provide its own frontend, but implements the Mastodon client API. You can use this server through a variety of clients:"), /*#__PURE__*/React.createElement("div", {
    className: "applist"
  }, /*#__PURE__*/React.createElement("div", {
    className: "entry"
  }, /*#__PURE__*/React.createElement("svg", {
    className: "logo redraw",
    xmlns: "http://www.w3.org/2000/svg",
    viewBox: "0 0 10000 10000"
  }, /*#__PURE__*/React.createElement("path", {
    d: "M9212 5993H5987V823c1053 667 2747 2177 3225 5170zM3100 2690A12240 12240 0 01939 6035h2161zm676 7210h2448a3067 3067 0 003067-3067H5052V627a527 527 0 00-1052 0v6206H709a3067 3067 0 003067 3067z"
  })), /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("h2", null, "Pinafore"), /*#__PURE__*/React.createElement("p", null, "Pinafore is a web client designed for speed and simplicity."), /*#__PURE__*/React.createElement("a", {
    className: "button",
    href: "https://pinafore.social/settings/instances/add"
  }, "Use Pinafore"))), /*#__PURE__*/React.createElement("div", {
    className: "entry"
  }, /*#__PURE__*/React.createElement("img", {
    className: "logo",
    src: "/assets/tusky.svg",
    alt: "The Tusky mascot, a cartoon elephant tooting happily"
  }), /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("h2", null, "Tusky"), /*#__PURE__*/React.createElement("p", null, "Tusky is a lightweight mobile client for Android"), /*#__PURE__*/React.createElement("a", {
    className: "button",
    href: "https://tusky.app"
  }, "Get Tusky")))))));
}

function Page(_ref) {
  var stylesheets = _ref.stylesheets,
      instance = _ref.instance,
      children = _ref.children;
  return /*#__PURE__*/React.createElement("html", {
    lang: "en"
  }, /*#__PURE__*/React.createElement("head", null, /*#__PURE__*/React.createElement("meta", {
    charset: "UTF-8"
  }), /*#__PURE__*/React.createElement("meta", {
    "http-equiv": "X-UA-Compatible",
    content: "IE=edge"
  }), /*#__PURE__*/React.createElement("meta", {
    name: "viewport",
    content: "width=device-width, initial-scale=1.0"
  }), /*#__PURE__*/React.createElement("meta", {
    name: "og:title",
    content: "GoToSocial Testing Instance"
  }), /*#__PURE__*/React.createElement("meta", {
    name: "og:description",
    content: ""
  }), /*#__PURE__*/React.createElement("meta", {
    name: "viewport",
    content: "width=device-width, initial-scale=1.0"
  }), /*#__PURE__*/React.createElement("link", {
    rel: "stylesheet",
    href: "/assets/base.css"
  }), /*#__PURE__*/React.createElement("link", {
    rel: "shortcut icon",
    href: "/assets/logo.png",
    type: "image/png"
  }), /*#__PURE__*/React.createElement("title", null, instance.Title, " - GoToSocial")), /*#__PURE__*/React.createElement("body", null, /*#__PURE__*/React.createElement("header", null, /*#__PURE__*/React.createElement("img", {
    src: "/assets/logo.png",
    alt: "Instance Logo"
  }), /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("h1", null, instance.Title)), /*#__PURE__*/React.createElement("div", null)), children, /*#__PURE__*/React.createElement("footer", null, /*#__PURE__*/React.createElement("div", {
    id: "version"
  }, "GoToSocial: ", /*#__PURE__*/React.createElement("span", {
    className: "accent"
  }, instance.Version), /*#__PURE__*/React.createElement("br", null), /*#__PURE__*/React.createElement("a", {
    href: "https://github.com/superseriousbusiness/gotosocial"
  }, "Source Code")), /*#__PURE__*/React.createElement("div", {
    id: "contact"
  }, "Contact: ", /*#__PURE__*/React.createElement("a", {
    href: instance.ContactAccount.URL,
    className: "nounderline"
  }, instance.ContactAccount.Username), /*#__PURE__*/React.createElement("br", null)), /*#__PURE__*/React.createElement("div", {
    id: "email"
  }, "Email: ", /*#__PURE__*/React.createElement("a", {
    href: "mailto:".concat(instance.Email),
    className: "nounderline"
  }, instance.Email), /*#__PURE__*/React.createElement("br", null)))));
}

;
module.exports = {
  TemplateIndex: TemplateIndex,
  Page: Page
};

},{}]},{},[1])(1)
});
